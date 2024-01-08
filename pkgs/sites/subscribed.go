package sites

import (
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

func NewSubVPN(cfg *confs.CollectorConf) (sv *SubscribedVPNs) {
	sv = &SubscribedVPNs{
		result:  []string{},
		cnf:     cfg,
		fetcher: request.NewFetcher(),
	}
	return
}

func (s *SubscribedVPNs) Type() SiteType {
	return Subscribed
}

// TODO: fetch subscribed urls.
func (s *SubscribedVPNs) fetch() {}

func (s *SubscribedVPNs) SetHandler(h func([]string)) {
	s.handler = h
}

func (s *SubscribedVPNs) Run() {
	s.fetch()
	if s.handler != nil {
		s.handler(s.result)
	}
}
