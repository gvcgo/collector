package confs

import (
	"os"
	"path/filepath"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

type StorageType int

const (
	StorageGithub      StorageType = 0
	StorageGitee       StorageType = 1
	ConfigFileName     string      = "config.json"
	VPNFileName        string      = "conf.txt"
	SubscriberFileName string      = "subscribers.txt"
	DomainFileName     string      = "domains.txt"
	WorkDirName        string      = ".pxycollector"
)

type CollectorConf struct {
	Type      StorageType `json,koanf:"type"`
	Token     string      `json,koanf:"token"`
	Repo      string      `json,koanf:"repo"`
	CryptoKey string      `json,koanf:"crypto_key"`
	ProxyURI  string      `json,koanf:"proxy_uri"`
	dirpath   string
	k         *koanfer.JsonKoanfer
}

func NewCollectorConf() (cc *CollectorConf) {
	homeDir, _ := os.UserHomeDir()
	cc = &CollectorConf{
		dirpath: filepath.Join(homeDir, WorkDirName),
	}
	cc.initiate()
	return
}

func (c *CollectorConf) initiate() {
	if ok, _ := gutils.PathIsExist(c.dirpath); !ok {
		os.MkdirAll(c.dirpath, os.ModePerm)
	}
	confPath := filepath.Join(c.dirpath, ConfigFileName)
	c.k, _ = koanfer.NewKoanfer(confPath)
	if ok, _ := gutils.PathIsExist(confPath); !ok {
		if err := c.Save(); err != nil {
			gprint.PrintError("%+v", err)
			return
		}
	}

	subPath := filepath.Join(c.dirpath, SubscriberFileName)
	if ok, _ := gutils.PathIsExist(subPath); !ok {
		os.WriteFile(subPath, []byte(Subscribers), os.ModePerm)
	}

	domainPath := filepath.Join(c.dirpath, DomainFileName)
	if ok, _ := gutils.PathIsExist(domainPath); !ok {
		os.WriteFile(domainPath, []byte(Domains), os.ModePerm)
	}

	c.Load()
	// TODO: setup configs for collector.
}

func (c *CollectorConf) Load() error {
	return c.k.Load(c)
}

func (c *CollectorConf) Save() error {
	return c.k.Save(c)
}

// TODO: setup Crypto Key for collector

// TODO: add domains.

// TODO: add subscriber list.
