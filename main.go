package main

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
	"github.com/manifoldco/promptui"
)

func main() {
	parent := colly.NewCollector(
		colly.AllowedDomains("developer.nulab.com"),
		colly.CacheDir("./.cache"),
	)
	child := parent.Clone()

	parent.OnHTML("#apiNavigation a", func(e *colly.HTMLElement) {
		url := e.Attr("href") // Get the URL of the API detail page
		child.Visit(e.Request.AbsoluteURL(url))
	})

	c := NewPostmanCollection()
	lang, err := askLang()
	if err != nil {
		log.Fatal(err)
	}
	conf := NewConfig(lang)
	child.OnHTML("#contents", func(e *colly.HTMLElement) {
		name := findName(e)
		desc := findDesc(e)
		req := findReq(e)
		method := req[0]
		url := req[1]
		query := findQuery(e, conf)
		variable := findVariable(e, conf)
		body := findBody(e, conf)

		i := NewPostmanCollectionItem(name, desc, method, url, query, variable, body)
		c.Item = append(c.Item, *i)

		fmt.Printf("%s -> %s %s\n", name, method, url)
	})

	if err := parent.Visit(conf.URL); err != nil {
		log.Fatal(err)
	}

	c.Save()
}

func askLang() (Language, error) {
	prompt := promptui.Select{
		Label: "Select language",
		Items: []Language{EN, JA},
	}

	_, lang, err := prompt.Run()
	return Language(lang), err
}
