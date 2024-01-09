package sites

import (
	"os"
	"strings"

	"github.com/gogf/gf/v2/util/gconv"
	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/proxy-collector/pkgs/confs"
)

const (
	Subscribed SiteType = "subscribed"
)

type SubscribedVPNs struct {
	result  []string
	fetcher *request.Fetcher
	handler func([]string)
	cnf     *confs.CollectorConf
}

func NewSubVPN(cnf *confs.CollectorConf) (sv *SubscribedVPNs) {
	sv = &SubscribedVPNs{
		result:  []string{},
		cnf:     cnf,
		fetcher: request.NewFetcher(),
	}
	if gconv.Bool(os.Getenv(confs.ToEnableProxyEnvName)) {
		sv.fetcher.Proxy = sv.cnf.ProxyURI
	}
	return
}

func (s *SubscribedVPNs) Type() SiteType {
	return Subscribed
}

func (s *SubscribedVPNs) SetHandler(h func([]string)) {
	s.handler = h
}

// Fetches subscribed urls.
func (s *SubscribedVPNs) fetch() {
	if s.cnf != nil {
		for _, sUrl := range s.cnf.GetSubs() {
			subUrl := confs.HandleSubscribedUrl(sUrl, s.cnf)
			if subUrl == "" {
				continue
			}
			gprint.PrintInfo("Getting: %s", subUrl)
			s.fetcher.SetUrl(subUrl)
			if content, statusCode := s.fetcher.GetString(); len(content) > 0 {
				decryptedContent := crypt.DecodeBase64(content)
				if len(decryptedContent) == 0 && len(content) > 500 && !strings.Contains(content, "</html>") {
					// fmt.Println(content)
					for _, encryptedContent := range strings.Split(content, "\n") {
						decryptedContent = crypt.DecodeBase64(strings.TrimSpace(encryptedContent))
						for _, rawUri := range strings.Split(decryptedContent, "\n") {
							if strings.Contains(rawUri, "://") {
								s.result = append(s.result, strings.TrimSpace(rawUri))
							}
						}
					}
				} else if len(content) > 800 && !strings.Contains(content, "</html>") {
					for _, rawUri := range strings.Split(decryptedContent, "\n") {
						if strings.Contains(rawUri, "://") {
							s.result = append(s.result, strings.TrimSpace(rawUri))
						}
					}
				}
			} else {
				gprint.PrintError("status code: %d", statusCode)
			}
		}
	}
}

func (s *SubscribedVPNs) Run() {
	s.fetch()
	if s.handler != nil {
		s.handler(s.result)
	}
}
