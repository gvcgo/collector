package versions

import (
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/proxy-collector/pkgs/confs"
	"github.com/gvcgo/proxy-collector/pkgs/upload"
)

/*
https://golang.google.cn/dl/
https://go.dev/dl/
*/
type Golang struct {
	cnf      *confs.CollectorConf
	uploader *upload.Uploader
	vList    []*Version
	fetcher  *request.Fetcher
	homepage string
}

func NewGolang(cnf *confs.CollectorConf) (g *Golang) {
	g = &Golang{
		cnf:      cnf,
		vList:    []*Version{},
		fetcher:  request.NewFetcher(),
		homepage: "https://go.dev/dl/",
	}
	if UseCNSource() {
		g.homepage = "https://golang.google.cn/dl/"
	}
	g.uploader = upload.NewUploader(cnf)
	if confs.EnableProxyOrNot() {
		pxy := g.cnf.ProxyURI
		if pxy == "" {
			pxy = confs.DefaultProxy
		}
		g.fetcher.Proxy = pxy
	}
	return
}
