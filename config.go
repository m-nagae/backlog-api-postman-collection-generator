package main

type Language string

const (
	EN Language = "English"
	JA Language = "Japanese"
)

type SelectorID struct {
	Method           string
	QueryParameter   string
	RequestParameter string
	URLParameter     string
}

type Config struct {
	URL        string
	SelectorID SelectorID
}

func NewConfig(lang Language) *Config {
	return &Config{
		URL:        devSiteURL(lang),
		SelectorID: selectorID(lang),
	}
}

func devSiteURL(lang Language) string {
	if lang == JA {
		return "https://developer.nulab.com/ja/docs/backlog"
	}
	return "https://developer.nulab.com/docs/backlog"
}

func selectorID(lang Language) SelectorID {
	if lang == JA {
		return SelectorID{
			Method:           "#メソッド",
			QueryParameter:   "#クエリパラメーター",
			RequestParameter: "#リクエストパラメーター",
			URLParameter:     "#url-パラメーター",
		}
	}
	return SelectorID{
		Method:           "#method",
		QueryParameter:   "#query-parameters",
		RequestParameter: "#form-parameters",
		URLParameter:     "#url-parameters",
	}
}
