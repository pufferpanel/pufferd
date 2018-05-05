package sftp

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/sftp"
	utils "github.com/pufferpanel/apufferi/common"
)

type requestPrefix struct {
	prefix string
}

func CreateRequestPrefix(prefix string) sftp.Handlers {
	h := requestPrefix{prefix: prefix}

	return sftp.Handlers{h, h, h, h}
}

func (rp requestPrefix) Fileread(request *sftp.Request) (io.ReaderAt, error) {
	file, err := rp.getFile(request.Filepath, os.O_RDONLY, 0644)
	return file, err
}

func (rp requestPrefix) Filewrite(request *sftp.Request) (io.WriterAt, error) {
	file, err := rp.getFile(request.Filepath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	return file, err
}

func (rp requestPrefix) Filecmd(request *sftp.Request) error {
	sourceName, err := rp.validate(request.Filepath)
	if err != nil {
		return rp.maskError(err)
	}
	var targetName string
	if request.Target != "" {
		targetName, err = rp.validate(request.Target)
		if err != nil {
			return rp.maskError(err)
		}
	}
	switch request.Method {
	case "SetStat", "Setstat":
		{
			return nil
		}
	case "Rename":
		{
			return os.Rename(sourceName, targetName)
		}
	case "Rmdir":
		{
			return os.RemoveAll(sourceName)
		}
	case "Mkdir":
		{
			return os.Mkdir(sourceName, 0755)
		}
	case "Symlink":
		{
			return nil
		}
	case "Remove":
		{
			return os.Remove(sourceName)
		}
	default:
		return errors.New(fmt.Sprintf("Unknown request method: %v", request.Method))
	}
}

func (rp requestPrefix) Filelist(request *sftp.Request) (sftp.ListerAt, error) {
	sourceName, err := rp.validate(request.Filepath)
	if err != nil {
		return nil, rp.maskError(err)
	}
	switch request.Method {
	case "List":
		{
			file, err := os.Open(sourceName)
			if err != nil {
				return nil, rp.maskError(err)
			}
			files, err := file.Readdir(0)
			if err != nil {
				return nil, err
			} else {
				return listerat(files), nil
			}
		}
	case "Stat":
		{
			file, err := os.Open(sourceName)
			if err != nil {
				return nil, rp.maskError(err)
			}
			fi, err := file.Stat()
			if err != nil {
				return nil, rp.maskError(err)
			}
			return listerat([]os.FileInfo{fi}), nil
		}
	case "Readlink":
		{
			target, err := os.Readlink(sourceName)
			if err != nil {
				return nil, rp.maskError(err)
			}
			file, err := os.Open(target)
			if err != nil {
				return nil, rp.maskError(err)
			}
			fi, err := file.Stat()
			if err != nil {
				return nil, rp.maskError(err)
			}
			return listerat([]os.FileInfo{fi}), nil
		}
	default:
		return nil, errors.New(fmt.Sprintf("Unknown request method: %s", request.Method))
	}
}

func (rp requestPrefix) getFile(path string, flags int, mode os.FileMode) (*os.File, error) {
	filePath, err := rp.validate(path)
	folderPath := filepath.Dir(filePath)
	if err != nil {
		return nil, rp.maskError(err)
	}

	if flags&os.O_CREATE == 1  {
		if _, err := os.Stat(folderPath); os.IsNotExist(err) {
			os.MkdirAll(folderPath, 0755)
		}
	}
	file, err := os.OpenFile(filePath, flags, mode)
	if err != nil {
		return nil, rp.maskError(err)
	}
	return file, err
}

func (rp requestPrefix) validate(path string) (string, error) {
	ok, path := rp.tryPrefix(path)
	if !ok {
		return "", errors.New("access denied")
	}
	return path, nil
}

func (rp requestPrefix) tryPrefix(path string) (bool, string) {
	newPath := filepath.Clean(filepath.Join(rp.prefix, path))
	if utils.EnsureAccess(newPath, rp.prefix) {
		return true, newPath
	} else {
		return false, ""
	}
}

func (rp requestPrefix) stripPrefix(path string) string {
	newStr := strings.Replace(path, rp.prefix, "", -1)
	if len(newStr) == 0 {
		newStr = "/"
	}
	return newStr
}

func (rp requestPrefix) maskError(err error) error {
	return errors.New(rp.stripPrefix(err.Error()))
}

type listerat []os.FileInfo

// Modeled after strings.Reader's ReadAt() implementation
func (f listerat) ListAt(ls []os.FileInfo, offset int64) (int, error) {
	var n int
	if offset >= int64(len(f)) {
		return 0, io.EOF
	}
	n = copy(ls, f[offset:])
	if n < len(ls) {
		return n, io.EOF
	}
	return n, nil
}