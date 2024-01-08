package sites

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/proxy-collector/pkgs/confs"
)

const (
	FreeFQ SiteType = "freefq"
)

type FreeFQVPNs struct {
	result  []string
	fetcher *request.Fetcher
	handler func([]string)
	cnf     *confs.CollectorConf
	urls    []string
	host    string
}

func NewFreeFQVPN(cnf *confs.CollectorConf) (fv *FreeFQVPNs) {
	fv = &FreeFQVPNs{
		result:  []string{},
		cnf:     cnf,
		fetcher: request.NewFetcher(),
		urls: []string{
			"https://freefq.com/v2ray/",
			"https://freefq.com/free-xray/",
			"https://freefq.com/free-ss/",
			"https://freefq.com/free-trojan/",
			"https://freefq.com/free-ssr/",
		},
		host: "https://freefq.com",
	}
	if gconv.Bool(os.Getenv(confs.ToEnableProxyEnvName)) {
		fv.fetcher.Proxy = fv.cnf.ProxyURI
	}
	return
}

func (f *FreeFQVPNs) Type() SiteType {
	return FreeFQ
}

func (f *FreeFQVPNs) SetHandler(h func([]string)) {
	f.handler = h
}

func (f *FreeFQVPNs) getUrl(sUrl string) (r string) {
	c := &http.Client{
		Timeout: time.Duration(30) * time.Second,
	}

	if gconv.Bool(os.Getenv(confs.ToEnableProxyEnvName)) {
		if f.cnf.ProxyURI == "" {
			f.cnf.ProxyURI = "http://127.0.0.1:2023"
		}
		u, _ := url.Parse(f.cnf.ProxyURI)
		c.Transport = &http.Transport{
			MaxIdleConns:    10,
			MaxConnsPerHost: 10,
			IdleConnTimeout: time.Duration(10) * time.Second,
			Proxy:           http.ProxyURL(u),
		}
	}

	if resp, err := c.Get(sUrl); err == nil {
		content, _ := io.ReadAll(resp.Body)
		r = string(content)
		resp.Body.Close()
	} else {
		gprint.PrintError("%+v", err)
	}
	return
}

func (f *FreeFQVPNs) getUrls() (r []string) {
	for _, sUrl := range f.urls {

		content := f.getUrl(sUrl)
		// fmt.Println(content)
		if doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(content)); err == nil && doc != nil {
			href := doc.Find("td.news_list").Find("ul").First().Find("li").First().Find("a").AttrOr("href", "")
			if href != "" {
				detailUrl := f.host + href
				gprint.PrintInfo("Dowload: %s", detailUrl)
				content = f.getUrl(detailUrl)
				if doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(content)); err == nil && doc != nil {
					fUrl := doc.Find("fieldset").Find("a").AttrOr("href", "")
					if fUrl != "" {
						r = append(r, fUrl)
					}
				}
			}
		}
	}
	return
}

func (f *FreeFQVPNs) getRawUris() {
	urls := f.getUrls()
	for _, u := range urls {
		content := f.getUrl(u)
		for _, rawUri := range strings.Split(content, "\n") {
			rawUri = strings.TrimSpace(rawUri)
			rawUri = strings.ReplaceAll(rawUri, "<br>", "")
			rawUri = strings.ReplaceAll(rawUri, "</p>", "")
			rawUri = strings.ReplaceAll(rawUri, "<p>", "")
			if strings.Contains(rawUri, "<script") {
				continue
			}
			if strings.HasPrefix(rawUri, "vmess://") {
				f.result = append(f.result, rawUri)
			} else if strings.HasPrefix(rawUri, "vless://") {
				f.result = append(f.result, rawUri)
			} else if strings.HasPrefix(rawUri, "ss://") {
				f.result = append(f.result, rawUri)
			} else if strings.HasPrefix(rawUri, "ssr://") {
				f.result = append(f.result, rawUri)
			} else if strings.HasPrefix(rawUri, "trojan://") {
				f.result = append(f.result, rawUri)
			}
		}
	}
}

func (f *FreeFQVPNs) Run() {
	f.getRawUris()
	if f.handler != nil {
		f.handler(f.result)
	}
}
