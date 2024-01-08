package sites

import (
	"github.com/moqsien/goutils/pkgs/request"
	"github.com/moqsien/proxy-collector/pkgs/confs"
)

type SubscribeVPN struct {
	result  []string
	fetcher *request.Fetcher
	handler func([]string)
	cnf     *confs.CollectorConf
}

func NewSubVPN(cfg *confs.CollectorConf) (sv *SubscribeVPN) {
	sv = &SubscribeVPN{
		result:  []string{},
		cnf:     cfg,
		fetcher: request.NewFetcher(),
	}
	return
}

func (s *SubscribeVPN) fetch() {

}

func (s *SubscribeVPN) SetHandler(h func([]string)) {
	s.fetch()
	s.handler = h
}
