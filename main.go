package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/gocolly/colly/v2"
	"github.com/manifoldco/promptui"
)

type SelectorID struct {
	Method           string
	QueryParameter   string
	RequestParameter string
	URLParameter     string
}

const (
	EN = "English"
	JA = "Japanese"
)

func main() {
	_, lang, err := selectLanguage()
	if err != nil {
		log.Fatal(err)
	}

	c, err := createPostmanCollection("Backlog API", lang)
	if err != nil {
		log.Fatal(err)
	}

	err = savePostmanCollection(c)
	if err != nil {
		log.Fatal(err)
	}
}

func selectLanguage() (int, string, error) {
	prompt := promptui.Select{
		Label: "Select language",
		Items: []string{EN, JA},
	}
	return prompt.Run()
}

func getDeveloperSiteURL(lang string) string {
	if lang == JA {
		return "https://developer.nulab.com/ja/docs/backlog"
	}
	return "https://developer.nulab.com/docs/backlog"
}

func getSelectorID(lang string) SelectorID {
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

func createPostmanCollection(collectionName string, lang string) (*PostmanCollection, error) {
	parent := colly.NewCollector(
		colly.AllowedDomains("developer.nulab.com"),
		colly.CacheDir("./.cache"),
	)
	child := parent.Clone()

	parent.OnHTML("#apiNavigation a", func(e *colly.HTMLElement) {
		url := e.Attr("href") // URL of each API
		child.Visit(e.Request.AbsoluteURL(url))
	})

	c := newPostmanCollection(collectionName)
	s := getSelectorID(lang)

	child.OnHTML("#contents", func(e *colly.HTMLElement) {
		name := findName(e)
		description := findDescription(e)
		method := findMethod(e, s.Method)
		url := findURL(e)
		query := findQuery(e, s.QueryParameter)
		variable := findVariable(e, s.URLParameter)
		body := findBody(e, s.RequestParameter)
		item := buildItem(name, description, method, url, query, variable, body)
		c.Item = append(c.Item, *item)
	})

	url := getDeveloperSiteURL(lang)
	if err := parent.Visit(url); err != nil {
		return nil, err
	}

	return c, nil
}

func savePostmanCollection(collection *PostmanCollection) error {
	f, err := os.Create("backlog_api_postman_collection.json")
	if err != nil {
		return err
	}
	defer f.Close()

	b, err := json.MarshalIndent(*collection, "", "  ")
	if err != nil {
		return err
	}

	if _, err = f.Write(b); err != nil {
		return err
	}

	return nil
}
