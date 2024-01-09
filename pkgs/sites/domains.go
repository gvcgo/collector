package sites

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
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
}

func NewEDomains(cnf *confs.CollectorConf) (ed *EDomains) {
	ed = &EDomains{
		cnf:    cnf,
		result: []string{},
	}
	return
}

func (e *EDomains) Type() SiteType {
	return EdgeDomains
}

func (e *EDomains) SetHandler(h func([]string)) {
	e.handler = h
}

// TODO: Concurrently.
func (e *EDomains) testDomains() {
	for _, sUrl := range e.cnf.GetDomains() {
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
					e.result = append(e.result, sUrl)
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
}

func (e *EDomains) Run() {
	e.testDomains()
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
