package upload

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/gogf/gf/v2/encoding/gjson"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/storage"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
)

type Uploader struct {
	cnf     *confs.CollectorConf
	storage storage.IStorage
}

func NewUploader(cnf *confs.CollectorConf) (up *Uploader) {
	up = &Uploader{
		cnf: cnf,
	}
	up.initiate()
	return
}

func (u *Uploader) initiate() {
	switch u.cnf.Type {
	case confs.StorageGithub:
		if u.cnf.UserName != "" && u.cnf.Token != "" && u.cnf.Repo != "" {
			st := storage.NewGhStorage(u.cnf.UserName, u.cnf.Token)
			if u.cnf.ProxyURI != "" && gconv.Bool(os.Getenv(confs.ToEnableProxyEnvName)) {
				st.Proxy = u.cnf.ProxyURI
			}
			u.storage = st
		}
	case confs.StorageGitee:
		if u.cnf.UserName != "" && u.cnf.Token != "" && u.cnf.Repo != "" {
			u.storage = storage.NewGtStorage(u.cnf.UserName, u.cnf.Token)
		}
	default:
		gprint.PrintError("Unknown storage type: %v", u.cnf.Type)
	}
	if u.storage != nil {
		// Try to create the repo.
		content := u.storage.GetRepoInfo(u.cnf.Repo)
		if !strings.Contains(string(content), `"id":`) {
			u.storage.CreateRepo(u.cnf.Repo)
		}
	}
}

func (u *Uploader) Upload(localFilePath string) (r []byte) {
	if u.storage == nil {
		gprint.PrintError("Storage is not initialized, please check your configurations.")
		return
	}
	fileName := filepath.Base(localFilePath)
	content := u.storage.GetContents(u.cnf.Repo, "", fileName)
	shaStr := gjson.New(content).Get("sha").String()
	return u.storage.UploadFile(u.cnf.Repo, "", localFilePath, shaStr)
}
