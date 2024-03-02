package versions

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"

	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/collector/pkgs/confs"
	"github.com/gvcgo/collector/pkgs/upload"
	"github.com/gvcgo/collector/pkgs/utils"
)

const (
	FlutterVersionFileName string = "flutter.version.json"
)

type FRelease struct {
	Version string `json:"version"`
	Stable  string `json:"channel"`
	Arch    string `json:"dart_sdk_arch"`
	Uri     string `json:"archive"`
	Sha256  string `json:"sha256"`
}

type FVersions struct {
	BaseUrl  string      `json:"base_url"`
	Releases []*FRelease `json:"releases"`
}

/*
https://storage.googleapis.com/flutter_infra_release/releases/releases_{linux/macos/windows}.json
https://storage.flutter-io.cn/flutter_infra_release/releases/releases_{linux/macos/windows}.json
*/
type Flutter struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
}

func NewFlutter(cnf *confs.CollectorConf) (f *Flutter) {
	f = &Flutter{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://storage.googleapis.com/flutter_infra_release/releases/releases_%s.json",
	}
	if UseCNSource() {
		f.homepage = "https://storage.flutter-io.cn/flutter_infra_release/releases/releases_%s.json"
	}
	if confs.EnableProxyOrNot() {
		pxy := f.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		f.fetcher.Proxy = pxy
	}
	return
}

func (f *Flutter) GetVersions() {
	platforms := []string{"linux", "macos", "windows"}
	for _, platform := range platforms {
		f.fetcher.SetUrl(fmt.Sprintf(f.homepage, platform))
		if resp := f.fetcher.Get(); resp != nil {
			versionList := FVersions{}
			content, _ := io.ReadAll(resp.RawBody())
			if err := json.Unmarshal(content, &versionList); err != nil {
				gprint.PrintError(fmt.Sprintf("Parse content from %s failed.", f.fetcher.Url))
				continue
			}
			if len(versionList.Releases) > 0 {
			INNER:
				for _, rr := range versionList.Releases {
					ver := &VFile{}
					ver.Os = utils.ParsePlatform(platform)
					ver.Arch = utils.ParseArch(rr.Arch)
					if ver.Arch == "" {
						continue INNER
					}
					ver.Sum = rr.Sha256
					if ver.Sum != "" {
						ver.SumType = "sha256"
					}
					ver.Url, _ = url.JoinPath(versionList.BaseUrl, rr.Uri)
					ver.Extra = rr.Stable
					if vlist, ok := f.versions[rr.Version]; !ok || vlist == nil {
						f.versions[rr.Version] = []*VFile{}
					}
					f.versions[rr.Version] = append(f.versions[rr.Version], ver)
				}
			}
		}
	}
}

func (f *Flutter) FetchAll() {
	f.GetVersions()
}

func (f *Flutter) Upload() {
	if len(f.versions) > 0 {
		fPath := filepath.Join(f.cnf.DirPath(), FlutterVersionFileName)
		if content, err := json.MarshalIndent(f.versions, "", "  "); err == nil && content != nil {
			os.WriteFile(fPath, content, os.ModePerm)
			f.uploader.Upload(fPath)
		}
	}
}
