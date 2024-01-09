package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/moqsien/goutils/pkgs/crypt"
	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/proxy-collector/pkgs/confs"
	"github.com/moqsien/proxy-collector/pkgs/sites"
	"github.com/moqsien/proxy-collector/pkgs/upload"
	"github.com/moqsien/vpnparser/pkgs/outbound"
)

func HandleQuery(rawUri string) (result string) {
	result = rawUri
	if !strings.Contains(rawUri, "?") {
		return
	}
	sList := strings.Split(rawUri, "?")
	query := sList[1]
	if strings.Contains(query, ";") && !strings.Contains(query, "&") {
		result = sList[0] + "?" + strings.ReplaceAll(sList[1], ";", "&")
	}
	return
}

func HandleRawUri(rawUri string) (result string) {
	if strings.HasPrefix(rawUri, "vmess://") {
		if r := crypt.DecodeBase64(strings.Split(rawUri, "://")[1]); r != "" {
			result = "vmess://" + r
		}
		return
	}

	if strings.Contains(rawUri, "\u0026") {
		rawUri = strings.ReplaceAll(rawUri, "\u0026", "&")
	}
	if strings.Contains(rawUri, "amp;") {
		rawUri = strings.ReplaceAll(rawUri, "amp;", "")
	}
	rawUri, _ = url.QueryUnescape(rawUri)
	r, err := url.Parse(rawUri)
	result = rawUri
	if err != nil {
		gprint.PrintError("%+v", err)
		return
	}

	host := r.Host
	uname := r.User.Username()
	passw, hasPassword := r.User.Password()

	if !strings.Contains(rawUri, "@") {
		if hostDecrypted := crypt.DecodeBase64(host); hostDecrypted != "" {
			result = strings.ReplaceAll(rawUri, host, hostDecrypted)
		}
	} else if uname != "" && !hasPassword && !strings.Contains(uname, "-") {
		if unameDecrypted := crypt.DecodeBase64(uname); unameDecrypted != "" {
			result = strings.ReplaceAll(rawUri, uname, unameDecrypted)
		}
	} else {
		if passwDecrypted := crypt.DecodeBase64(passw); passwDecrypted != "" {
			result = strings.ReplaceAll(rawUri, passw, passwDecrypted)
		}
	}

	if strings.Contains(result, "%") {
		result, _ = url.QueryUnescape(result)
	}
	result = HandleQuery(result)
	if strings.Contains(result, "127.0.0.1") || strings.Contains(result, "127.0.0.0") {
		return ""
	}
	return
}

type SiteRunner struct {
	Result        *outbound.Result `json:"vpn_list"`
	domainList    []string
	rawDomainList []string
	result        map[string]struct{}
	cnf           *confs.CollectorConf
	sites         []sites.ISite
	uploader      *upload.Uploader
}

func NewSiteRunner(cnf *confs.CollectorConf) (sr *SiteRunner) {
	sr = &SiteRunner{
		Result:        outbound.NewResult(),
		domainList:    []string{},
		rawDomainList: []string{},
		result:        map[string]struct{}{},
		cnf:           cnf,
		uploader:      upload.NewUploader(cnf),
	}
	return
}

func (s *SiteRunner) AddSite(st sites.ISite) {
	s.sites = append(s.sites, st)
}

func (s *SiteRunner) wrapItem(rawUri string) *outbound.ProxyItem {
	item := outbound.NewItem(rawUri)
	if strings.HasPrefix(item.Address, "127.0.") {
		return nil
	}
	item.GetOutbound()
	return item
}

func (s *SiteRunner) Run() {
	s.result = map[string]struct{}{}
	s.Result = outbound.NewResult()

	s.domainList = []string{}
	s.rawDomainList = []string{}

	for _, st := range s.sites {
		switch st.Type() {
		case sites.Subscribed, sites.FreeFQ:
			st.SetHandler(func(result []string) {
				for _, rawUri := range result {
					rawUri = HandleRawUri(rawUri)
					proxyItem := s.wrapItem(rawUri)
					proxyStr := fmt.Sprintf("%s%s:%d", proxyItem.Scheme, proxyItem.Address, proxyItem.Port)
					if _, ok := s.result[proxyStr]; !ok {
						s.Result.AddItem(proxyItem)
						s.result[proxyStr] = struct{}{}
					}
				}
			})
			st.Run()
			s.doProxy()
		case sites.RawEdgeDomains:
			st.SetHandler(func(rr []string) {
				s.rawDomainList = append(s.rawDomainList, rr...)
			})
			st.Run()
			s.doRawDomains()
		case sites.EdgeDomains:
			st.SetHandler(func(rr []string) {
				s.domainList = append(s.domainList, rr...)
			})
			st.Run()
			s.doDomains()
		default:
		}
		st.Run()
	}
}

func (s *SiteRunner) doProxy() {
	gprint.PrintSuccess("Total Proxies: %d", s.Result.Len())
	gprint.PrintSuccess(
		"vmess[%d]; vless[%d]; ss[%d]; trojan[%d]; ssr[%d]",
		s.Result.VmessTotal,
		s.Result.VlessTotal,
		s.Result.SSTotal,
		s.Result.TrojanTotal,
		s.Result.SSRTotal,
	)
	fPath := s.cnf.VPNFilePath()
	var cstZone = time.FixedZone("CST", 8*3600)
	now := time.Now().In(cstZone)
	s.Result.UpdateAt = now.Format("2006-01-02 15:04:05")
	if s.Result.Len() <= 0 {
		return
	}
	if content, err := json.Marshal(s.Result); err == nil {
		gprint.PrintWarning("neobox key: %s", s.cnf.CryptoKey)
		cc := crypt.NewCrptWithKey([]byte(s.cnf.CryptoKey))
		if r, err := cc.AesEncrypt([]byte(content)); err == nil {
			if err = os.WriteFile(fPath, r, os.ModePerm); err == nil {
				s.uploader.Upload(fPath)
			}
		}
	} else {
		gprint.PrintError("marshal failed: %+v", err)
	}
}

func (s *SiteRunner) doRawDomains() {
	if len(s.rawDomainList) == 0 {
		return
	}
	s.cnf.AddRawDomains(s.domainList...)
	s.uploader.Upload(s.cnf.RawDomainPath())
}

func (s *SiteRunner) doDomains() {
	if len(s.domainList) == 0 {
		return
	}
	fPath := s.cnf.DomainPath()
	content := strings.Join(s.domainList, "\n")
	if err := os.WriteFile(fPath, []byte(content), os.ModePerm); err == nil {
		s.uploader.Upload(fPath)
	}
	s.uploader.Upload(fPath)
}
