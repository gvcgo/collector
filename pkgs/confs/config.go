package confs

type StorageType int

const (
	StorageGithub      StorageType = 0
	StorageGitee       StorageType = 1
	VPNFileName        string      = "conf.txt"
	SubscriberFileName string      = "subscribers.txt"
	DomainFileName     string      = "domains.txt"
	WorkDirName        string      = ".pxycollector"
)

type CollectorConf struct {
	Type     StorageType `json,koanf:"type"`
	Token    string      `json,koanf:"token"`
	Repo     string      `json,koanf:"repo"`
	CrytoKey string      `json,koanf:"crypto_key"`
	ProxyURI string      `json,koanf:"proxy_uri"`
}
