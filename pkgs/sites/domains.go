package sites

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/proxy-collector/pkgs/confs"
)

const (
	EdgeDomains SiteType = "edge_domains"
)

type EDomains struct {
	result  []string
	handler func([]string)
	cnf     *confs.CollectorConf
	sender  chan string
	lock    *sync.Mutex
}

func NewEDomains(cnf *confs.CollectorConf) (ed *EDomains) {
	ed = &EDomains{
		cnf:    cnf,
		result: []string{},
		lock:   &sync.Mutex{},
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
	for _, d := range e.cnf.GetDomains() {
		e.sender <- d
	}
	close(e.sender)
}

func (e *EDomains) domainTSL(sUrl string) {
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
	if !strings.HasPrefix(sUrl, "https://") {
		sUrl = "https://" + sUrl
	}
	if resp, err := client.Get(sUrl); err == nil && resp != nil {
		if len(resp.TLS.PeerCertificates) > 0 {
			certInfo := resp.TLS.PeerCertificates[0]
			if strings.Contains(strings.ToLower(certInfo.Subject.String()), "cloudflare") {
				gprint.PrintSuccess(sUrl)
				e.lock.Lock()
				e.result = append(e.result, sUrl)
				e.lock.Unlock()
			} else {
				gprint.PrintInfo("No cloudflare: %s", sUrl)
			}
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
			e.domainTSL(sUrl)
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
