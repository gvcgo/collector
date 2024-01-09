package sites

import (
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

func (e *EDomains) testDomains() {
	// TODO: cloudflare SSL/TSL.
}

func (e *EDomains) Run() {
	e.testDomains()
	if e.handler != nil {
		e.handler(e.result)
	}
}
