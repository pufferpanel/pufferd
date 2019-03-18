package forgedl

import (
	"github.com/pufferpanel/pufferd/commons"
	"github.com/pufferpanel/pufferd/environments"
	"github.com/pufferpanel/pufferd/environments/envs"
	"path"
	"strings"
)

const JAR_DOWNLOAD = "https://files.minecraftforge.net/maven/net/minecraftforge/forge/${version}/forge-${version}-installer.jar"

type ForgeDl struct {
	Version string
	Filename string
}


func (op ForgeDl) Run(env envs.Environment) error {
	jarDownload := strings.Replace(JAR_DOWNLOAD, "${version}", op.Version, -1)

	localFile, err := environments.DownloadViaMaven(jarDownload, env)
	if err != nil {
		return err
	}

	//copy from the cache
	return commons.CopyFile(localFile, path.Join(env.GetRootDirectory(), op.Filename))
}

