package sites

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gvcgo/goutils/pkgs/gtea/gprint"
	"github.com/gvcgo/goutils/pkgs/request"
	"github.com/gvcgo/collector/pkgs/confs"
)

const (
	EdgeDomains    SiteType = "edge_domains"
	RawEdgeDomains SiteType = "raw_edge_domains"
)

type EDomains struct {
	result    []string
	handler   func([]string)
	cnf       *confs.CollectorConf
	sender    chan string
	lock      *sync.Mutex
	ipNetList []*net.IPNet
}

func NewEDomains(cnf *confs.CollectorConf) (ed *EDomains) {
	ed = &EDomains{
		cnf:       cnf,
		result:    []string{},
		lock:      &sync.Mutex{},
		ipNetList: []*net.IPNet{},
	}
	return
}

func (e *EDomains) Type() SiteType {
	return EdgeDomains
}

func (e *EDomains) SetHandler(h func([]string)) {
	e.handler = h
}

func (e *EDomains) sendDomains() {
	e.sender = make(chan string, 100)
	for _, d := range e.cnf.GetRawDomains() {
		e.sender <- d
	}
	close(e.sender)
}

func (e *EDomains) isCloudflareCDN(sUrl string) (ok bool) {
	if len(e.ipNetList) == 0 {
		for _, ipr := range e.cnf.GetCloudflareIPV4RangeList() {
			_, ipNet, _ := net.ParseCIDR(ipr)
			if ipNet != nil {
				e.ipNetList = append(e.ipNetList, ipNet)
			}
		}
	}
	if ip, err := net.ResolveIPAddr("ip", sUrl); err == nil {
		for _, ipNet := range e.ipNetList {
			if ok = ipNet.Contains(ip.IP); ok {
				return
			}
		}
	} else {
		gprint.PrintWarning("Failed to parse IP: %s", sUrl)
	}
	return
}

func (e *EDomains) verifyDomain(sUrl string) {
	if sUrl == "" {
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second,
	}

	u := sUrl
	if !strings.HasPrefix(sUrl, "https://") {
		u = "https://" + sUrl
	}
	if resp, err := client.Get(u); err == nil && resp != nil {
		// if len(resp.TLS.PeerCertificates) > 0 {
		// 	certInfo := resp.TLS.PeerCertificates[0]
		// 	if strings.Contains(strings.ToLower(certInfo.Subject.String()), "cloudflare") {
		// 		gprint.PrintSuccess(sUrl)
		// 		e.lock.Lock()
		// 		e.result = append(e.result, sUrl)
		// 		e.lock.Unlock()
		// 	} else {
		// 		gprint.PrintInfo("No cloudflare: %s", sUrl)
		// 	}
		// }
		if e.isCloudflareCDN(sUrl) {
			gprint.PrintSuccess(sUrl)
			e.lock.Lock()
			e.result = append(e.result, sUrl)
			e.lock.Unlock()
		}
		if resp.Body != nil {
			resp.Body.Close()
		}
	} else {
		gprint.PrintWarning("%+v", err)
	}
}

func (e *EDomains) domains() {
	for {
		select {
		case sUrl, ok := <-e.sender:
			if sUrl == "" || !ok {
				return
			}
			e.verifyDomain(sUrl)
		default:
			time.Sleep(time.Millisecond * 10)
		}
	}
}

func (e *EDomains) Run() {
	go e.sendDomains()
	time.Sleep(time.Millisecond * 100)
	for i := 0; i < runtime.NumCPU()*2; i++ {
		e.domains()
	}
	if e.handler != nil {
		gprint.PrintInfo("Total: %d", len(e.result))
		e.handler(e.result)
	}
}

func TestEDomains() {
	cnf := &confs.CollectorConf{}
	d := NewEDomains(cnf)
	d.SetHandler(func(result []string) {
		fmt.Println(result)
		fmt.Println("Total: ", len(result))
	})
	d.Run()
}

/*
Collecting websites using cloudflare SSL/TSL.

https://trends.builtwith.com/websitelist/Cloudflare-SSL
*/
type EDCollector struct {
	startUrls []string
	result    map[string]struct{}
	handler   func([]string)
	fetcher   *request.Fetcher
	cnf       *confs.CollectorConf
	urls      map[string]struct{}
}

func NewEDCollector(cnf *confs.CollectorConf) (ec *EDCollector) {
	ec = &EDCollector{
		cnf:     cnf,
		result:  map[string]struct{}{},
		fetcher: request.NewFetcher(),
		startUrls: []string{
			"https://trends.builtwith.com/websitelist/Cloudflare-CDN",
			"https://trends.builtwith.com/websitelist/Cloudflare-SSL",
			"https://trends.builtwith.com/websitelist/Cloudflare",
		},
		urls: map[string]struct{}{},
	}
	if gconv.Bool(os.Getenv(confs.ToEnableProxyEnvName)) {
		if ec.cnf.ProxyURI == "" {
			ec.cnf.ProxyURI = DefaultProxy
		}
		ec.fetcher.Proxy = ec.cnf.ProxyURI
	}
	return
}

func (e *EDCollector) Type() SiteType {
	return RawEdgeDomains
}

func (e *EDCollector) SetHandler(h func([]string)) {
	e.handler = h
}

func (e *EDCollector) GetResult() []string {
	r := []string{}
	for s := range e.result {
		r = append(r, s)
	}
	return r
}

func (e *EDCollector) GetWebsites() {
	for sUrl := range e.urls {
		gprint.PrintInfo("Fetch: %s", sUrl)
		e.fetcher.SetUrl(sUrl)
		if respStr, rCode := e.fetcher.GetString(); rCode == 200 {
			if doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(respStr)); err == nil && doc != nil {
				tr := doc.Find("table").Find("tr")
				tr.Each(func(_ int, s *goquery.Selection) {
					u := s.Find("td").First().Next().Text()
					if u != "" && !strings.Contains(u, "...") {
						e.result[u] = struct{}{}
					}
				})
			}
		}
	}
}

func (e *EDCollector) Run() {
	e.result = map[string]struct{}{}
	for _, sUrl := range e.startUrls {
		e.fetcher.SetUrl(sUrl)
		if respStr, rCode := e.fetcher.GetString(); rCode == 200 {
			// os.WriteFile("test.html", []byte(respStr), 0666)
			if doc, err := goquery.NewDocumentFromReader(bytes.NewBufferString(respStr)); err == nil && doc != nil {
				div := doc.Find("div.card-body").First()
				div.Find("li").Find("a").Each(func(_ int, s *goquery.Selection) {
					if u := s.AttrOr("href", ""); u != "" {
						if strings.HasPrefix(u, "//") {
							u = "https:" + u
						}
						e.urls[u] = struct{}{}
					}
				})
			}

		}
	}
	e.GetWebsites()
	if e.handler != nil {
		gprint.PrintInfo("Total rawDomains: %d", len(e.GetResult()))
		e.handler(e.GetResult())
	}
}

func TestEDCollector() {
	cnf := &confs.CollectorConf{}
	ec := NewEDCollector(cnf)
	ec.Run()
	fmt.Println(ec.GetResult())
	fmt.Println("Total: ", len(ec.GetResult()))
}

func TestTDomains() {
	cnf := &confs.CollectorConf{}
	ipList := cnf.GetCloudflareIPV4RangeList()
	ipNetList := []*net.IPNet{}
	for _, ipr := range ipList {
		_, ipNet, _ := net.ParseCIDR(ipr)
		if ipNet != nil {
			ipNetList = append(ipNetList, ipNet)
		}
	}

	domainList := strings.Split(confs.RawEdDomains, "\n")
	for _, d := range domainList {
		if ip, err := net.ResolveIPAddr("ip", d); err == nil {
			for _, ipNet := range ipNetList {
				if ipNet.Contains(ip.IP) {
					fmt.Println(d, " IP: ", ip.IP.String())
				}
			}
		}
	}
}
