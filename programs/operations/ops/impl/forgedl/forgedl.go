package forgedl

import (
	"github.com/pufferpanel/pufferd/commons"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/programs/operations/ops"
	"path"
	"strings"
)

const JAR_DOWNLOAD = "https://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar"

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
	jarDownload := strings.Replace(JAR_DOWNLOAD, "${version}", op.Version, -1)

	localFile, err := commons.DownloadViaMaven(jarDownload, env)
	if err != nil {
		return err
	}

	//copy from the cache
	return commons.CopyFile(localFile, path.Join(env.GetRootDirectory(), op.Filename))
}

func (of ForgeDlOperationFactory) Create(op ops.CreateOperation) ops.Operation {
	version := op.OperationArgs["version"].(string)
	filename := op.OperationArgs["target"].(string)

	return ForgeDl{Version: version, Filename: filename}
}

var Factory ForgeDlOperationFactory
