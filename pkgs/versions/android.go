package versions

import (
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

const (
	SDKManagerVersionFileName string = "sdkmanager.version.json"
	SDKManagerBaseUrl         string = "https://dl.google.com/android/repository"
)

/*
sdkmanager: https://developer.android.com/tools/sdkmanager?hl=zh-cn
download:
https://developer.android.google.cn/studio?hl=zh-cn
https://developer.android.com/studio?hl=en
*/
type SDKManager struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	versions Versions
	fetcher  *request.Fetcher
	homepage string
}

func NewSDKManager(cnf *confs.CollectorConf) (s *SDKManager) {
	s = &SDKManager{
		cnf:      cnf,
		uploader: upload.NewUploader(cnf),
		versions: Versions{},
		fetcher:  request.NewFetcher(),
		homepage: "https://developer.android.com/studio?hl=en",
	}
	if confs.EnableProxyOrNot() {
		pxy := s.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		s.fetcher.Proxy = pxy
	}
	return
}

func (s *SDKManager) FetchAll() {

}

func (s *SDKManager) Upload() {

}
