package versions

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
)

const (
	JuliaVersionFileName string = "julia.version.json"
)

type JItem struct {
	Url  string `json:"url"`
	Kind string `json:"kind"`
	Arch string `json:"arch"`
	Sum  string `json:"sha256"`
	Os   string `json:"os"`
}

type JVersion struct {
	Files  []*JItem `json:"files"`
	Stable any      `json:"stable"`
}

type JVersionList map[string]*JVersion

/*
https://julialang-s3.julialang.org/bin/versions.json
https://mirrors.tuna.tsinghua.edu.cn/julia-releases/bin/versions.json
*/
type Julia struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
}

func NewJulia(cnf *confs.CollectorConf) (j *Julia) {
	j = &Julia{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://julialang-s3.julialang.org/bin/versions.json",
	}
	if confs.EnableProxyOrNot() {
		pxy := j.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		j.fetcher.Proxy = pxy
	}
	if UseCNSource() {
		j.homepage = "https://mirrors.tuna.tsinghua.edu.cn/julia-releases/bin/versions.json"
	}
	return
}

func (j *Julia) GetVersions() {
	j.fetcher.SetUrl(j.homepage)
	j.fetcher.Timeout = 180 * time.Second
	versionList := &JVersionList{}
	if resp := j.fetcher.Get(); resp != nil {
		content, _ := io.ReadAll(resp.RawBody())
		if err := json.Unmarshal(content, &versionList); err != nil {
			gprint.PrintError(fmt.Sprintf("Parse content from %s failed.", j.fetcher.Url))
			return
		}
	}
	if len(*versionList) > 0 {
		for vName, fList := range *versionList {
			var extraStr string
			if gconv.Bool(fList.Stable) {
				extraStr = "stable"
			}
			for _, jFile := range fList.Files {
				if jFile.Kind == "archive" && !strings.HasSuffix(jFile.Url, ".dmg") {
					archStr := utils.ParseArch(jFile.Arch)
					platform := utils.ParsePlatform(jFile.Os)
					if archStr == "" || platform == "" {
						continue
					}
					ver := &VFile{}
					ver.Url = jFile.Url
					ver.Arch = archStr
					ver.Os = platform
					ver.Sum = jFile.Sum
					if ver.Sum != "" {
						ver.SumType = "sha256"
					}
					ver.Extra = extraStr
					if vlist, ok := j.versions[vName]; !ok || vlist == nil {
						j.versions[vName] = []*VFile{}
					}
					j.versions[vName] = append(j.versions[vName], ver)
				}
			}
		}
	}
}

func (j *Julia) FetchAll() {
	j.GetVersions()
}

func (j *Julia) Upload() {
	if len(j.versions) > 0 {
		fPath := filepath.Join(j.cnf.DirPath(), JuliaVersionFileName)
		if content, err := json.MarshalIndent(j.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			j.uploader.Upload(fPath)
		}
	}
}
