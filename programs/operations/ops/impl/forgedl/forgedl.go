package forgedl

import (
	"crypto/sha1"
	"fmt"
	"github.com/pufferpanel/apufferi/config"
	"github.com/pufferpanel/apufferi/logging"
	"github.com/pufferpanel/pufferd/commons"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
)

const JAR_DOWNLOAD = "https://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar"
const SHA1_DOWNLOAD = "https://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar.sha1"

type ForgeDl struct {
	Version string
	Filename string
}

type ForgeDlOperationFactory struct {
}

func (of ForgeDlOperationFactory) Key() string {
	return "forgedl"
}

func (op ForgeDl) Run(env environments.Environment) error {
	cacheDir := config.GetStringOrDefault("cache", "tmp")
	os.MkdirAll(cacheDir, 0755)

	jarDownload := strings.Replace(JAR_DOWNLOAD, "${version}", op.Version, -1)
	sha1Download := strings.Replace(SHA1_DOWNLOAD, "${version}", op.Version, -1)

	fileCache := path.Join(cacheDir, "forge-"+op.Version+".jar")

	useCache := true
	f, err := os.Open(fileCache)
	defer commons.Close(f)
	//cache was readable, so validate
	if err == nil {
		h := sha1.New()
		if _, err := io.Copy(h, f); err != nil {
			log.Fatal(err)
		}
		commons.Close(f)

		actualHash := fmt.Sprintf("%x", h.Sum(nil))

		client := &http.Client{}
		logging.Develf("Downloading hash from %s", sha1Download)
		response, err := client.Get(sha1Download)
		defer commons.CloseResponse(response)
		if err != nil {
			useCache = false
		} else {
			data := make([]byte, 40)
			_, err := response.Body.Read(data)
			expectedHash := string(data)

			if err != nil {
				useCache = false
			} else if expectedHash != actualHash {
				logging.Warnf("Forge cache expected %s but was actually %s", expectedHash, actualHash)
				useCache = false
			}
		}
	} else if !os.IsNotExist(err) {
		logging.Warnf("Cached file is not readable, will download (%s)", fileCache)
	} else {
		useCache = false
	}

	//if we can't use cache, redownload it to the cache
	if !useCache {
		logging.Infof("Downloading new version of forge and caching to %s", fileCache)
		env.DisplayToConsole("Downloading Forge: " + jarDownload)
		err = commons.DownloadFileToCache(jarDownload, fileCache)
		if err != nil {
			return err
		}
	}

	//copy from the cache
	return commons.CopyFile(fileCache, path.Join(env.GetRootDirectory(), op.Filename))
}

func (of ForgeDlOperationFactory) Create(op ops.CreateOperation) ops.Operation {
	version := op.OperationArgs["version"].(string)
	filename := op.OperationArgs["target"].(string)

	return ForgeDl{Version: version, Filename: filename}
}

var Factory ForgeDlOperationFactory
