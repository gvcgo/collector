package versions

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
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
