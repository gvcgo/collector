package confs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moqsien/goutils/pkgs/gtea/gprint"
	"github.com/moqsien/goutils/pkgs/gutils"
	"github.com/moqsien/goutils/pkgs/koanfer"
)

const (
	// https://cdn.jsdelivr.net
	ToEnableJsdelivrEnvName string = "ENABLE_JS_DELIVR"
	// proxy
	ToEnableProxyEnvName string = "ENABLE_PROXY"
)

type StorageType int

const (
	StorageGithub      StorageType = 1
	StorageGitee       StorageType = 2
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

	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); !ok {
		// To save default subscriber list.
		os.WriteFile(subPath, []byte(SubscribedUrls), os.ModePerm)
	}

	domainPath := c.domainPath()
	if ok, _ := gutils.PathIsExist(domainPath); !ok {
		// To save default domain list.
		os.WriteFile(domainPath, []byte(EdDomains), os.ModePerm)
	}

	c.Load()
	// Setup configs for collector.
	c.setup()
}

func (c *CollectorConf) subPath() string {
	return filepath.Join(c.dirpath, SubscriberFileName)
}

func (c *CollectorConf) domainPath() string {
	return filepath.Join(c.dirpath, DomainFileName)
}

func (c *CollectorConf) setup() {
	if c.Token == "" || c.Repo == "" {
		fmt.Println("Please choose storage type: ")
		fmt.Println("1. Github. (default)")
		fmt.Println("2. Gitee.")
		var sType string
		fmt.Scanln(&sType)
		switch sType {
		case "2":
			c.Type = StorageGitee
		default:
			c.Type = StorageGithub
		}
		fmt.Println("Please enter your github/gitee access-token: ")
		var token string
		fmt.Scanln(&token)
		if token != "" {
			c.Token = token
		}
		fmt.Println("Please enter your github/gitee repo name: ")
		var repo string
		fmt.Scanln(&repo)
		if repo != "" {
			c.Repo = repo
		}
		c.Save()

		// To reset the Crypto Key or not.
		fmt.Println("Do you want to reset the Crypto Key? (y/n)")
		var ok string
		fmt.Scanln(&ok)
		switch ok {
		case "y", "Y", "yes", "Yes":
			c.ResetCryptoKey()
		default:
			gprint.PrintWarning("Invalid input.")
		}
	}
}

func (c *CollectorConf) Load() error {
	return c.k.Load(c)
}

func (c *CollectorConf) Save() error {
	return c.k.Save(c)
}

func (c *CollectorConf) ResetCryptoKey() {
	c.Load()
	c.CryptoKey = gutils.RandomString(16)
	gprint.PrintInfo("CryptoKey: %s", c.CryptoKey)
	c.Save()
}

func (c *CollectorConf) ShowCryptoKey() {
	c.Load()
	gprint.PrintInfo("CryptoKey: %s", c.CryptoKey)
}

// Domains for cloudflare edgetunnels.
func (c *CollectorConf) ShowDomains() {
	dPath := c.domainPath()
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		gprint.PrintInfo("Domain list for cloudflare edgetunnels: ")
		fmt.Println(gprint.YellowStr(string(content)))
	} else {
		gprint.PrintError("No domain list for edgetunnels available.")
	}
}

func (c *CollectorConf) GetDomains() (r []string) {
	dPath := c.domainPath()
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		r = strings.Split(string(content), "\n")
	} else {
		os.WriteFile(dPath, []byte(EdDomains), os.ModePerm)
		r = strings.Split(EdDomains, "\n")
	}
	return r
}

func (c *CollectorConf) AddDomains(domains ...string) {
	dPath := c.domainPath()
	newStr := strings.Join(domains, "\n")
	if ok, _ := gutils.PathIsExist(dPath); ok {
		content, _ := os.ReadFile(dPath)
		s := string(content) + newStr
		os.WriteFile(dPath, []byte(s), os.ModePerm)
	} else {
		os.WriteFile(dPath, []byte(EdDomains+newStr), os.ModePerm)
	}
}

// Subscriber list.
func (c *CollectorConf) ShowSubs() {
	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		gprint.PrintInfo("Subscribed urls: ")
		fmt.Println(gprint.YellowStr(string(content)))
	} else {
		gprint.PrintError("No subscribed urls available.")
	}
}

func (c *CollectorConf) GetSubs() (r []string) {
	subPath := c.subPath()
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		r = strings.Split(string(content), "\n")
	} else {
		os.WriteFile(subPath, []byte(SubscribedUrls), os.ModePerm)
		r = strings.Split(SubscribedUrls, "\n")
	}
	return r
}

func (c *CollectorConf) AddSubs(subs ...string) {
	subPath := c.domainPath()
	newStr := strings.Join(subs, "\n")
	if ok, _ := gutils.PathIsExist(subPath); ok {
		content, _ := os.ReadFile(subPath)
		s := string(content) + newStr
		os.WriteFile(subPath, []byte(s), os.ModePerm)
	} else {
		os.WriteFile(subPath, []byte(EdDomains+newStr), os.ModePerm)
	}
}
