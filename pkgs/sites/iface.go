package sites

type ISite interface {
	SetHandler(handler func([]string))
	Run()
	Type() SiteType
}

type SiteType string
