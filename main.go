package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
	"github.com/manifoldco/promptui"
)

type SelectorID struct {
	Method           string
	QueryParameter   string
	RequestParameter string
	URLParameter     string
}

var ts = strings.TrimSpace

const (
	EN = "English"
	JA = "Japanese"
)

func main() {
	_, lang, err := askLanguage()
	if err != nil {
		return
	}

	parent := colly.NewCollector(
		colly.AllowedDomains("developer.nulab.com"),
		colly.CacheDir("./.cache"),
	)
	child := parent.Clone()

	parent.OnHTML("#apiNavigation a", func(e *colly.HTMLElement) {
		u := e.Attr("href") // Get the URL of the API detail page
		child.Visit(e.Request.AbsoluteURL(u))
	})

	c := getPostmanCollectionTemplate()
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

	if err := parent.Visit(getDeveloperSiteURL(lang)); err != nil {
		log.Fatal(err)
	}

	savePostmanCollection(c)
}

func askLanguage() (int, string, error) {
	prompt := promptui.Select{
		Label: "Select language",
		Items: []string{EN, JA},
	}

	return prompt.Run()
}

func getDeveloperSiteURL(lang string) string {
	url := "https://developer.nulab.com/docs/backlog"
	if lang == JA {
		url = "https://developer.nulab.com/ja/docs/backlog"
	}
	return url
}

func getSelectorID(lang string) SelectorID {
	d := SelectorID{
		Method:           "#method",
		QueryParameter:   "#query-parameters",
		RequestParameter: "#form-parameters",
		URLParameter:     "#url-parameters",
	}
	if lang == JA {
		d = SelectorID{
			Method:           "#メソッド",
			QueryParameter:   "#クエリパラメーター",
			RequestParameter: "#リクエストパラメーター",
			URLParameter:     "#url-パラメーター",
		}
	}
	return d
}

func getPostmanCollectionTemplate() *PostmanCollection {
	return &PostmanCollection{
		Info: PostmanCollectionInfo{
			Name:   "Backlog API",
			Schema: "https://schema.getpostman.com/json/collection/v2.1.0/collection.json",
		},
		Auth: PostmanCollectionAuth{
			Type: "apikey",
			APIKey: []PostmanCollectionAPIKey{
				{
					Key:   "in",
					Value: "query",
					Type:  "string",
				},
				{
					Key:   "key",
					Value: "apiKey",
					Type:  "string",
				},
				{
					Key:   "value",
					Value: "{{api_key}}",
					Type:  "string",
				},
			},
		},
	}
}

func findName(e *colly.HTMLElement) string {
	name := ts(e.DOM.Find("h1").Text())
	if name == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		name = ts(e.DOM.Find("h2").First().Text())
	}
	return name
}

func findDescription(e *colly.HTMLElement) string {
	description := ts(e.DOM.Find("h1 + p").Text())
	if description == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		description = ts(e.DOM.Find("h2 + p").First().Text())
	}
	return description
}

func findMethod(e *colly.HTMLElement, selectorID string) string {
	selector := selectorID + " + pre"
	return ts(e.DOM.Find(selector).Text())
}

func findURL(e *colly.HTMLElement) string {
	return ts(e.DOM.Find("#url + pre").Text())
}

func findQuery(e *colly.HTMLElement, selectorID string) *[]PostmanCollectionKeyValue {
	selector := selectorID + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findVariable(e *colly.HTMLElement, selectorID string) *[]PostmanCollectionKeyValue {
	selector := selectorID + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findBody(e *colly.HTMLElement, selectorID string) *PostmanCollectionBody {
	selector := selectorID + " + pre + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	body := PostmanCollectionBody{
		Mode:       "urlencoded",
		URLEncoded: *keyValue,
	}
	return &body
}

func buildKeyValue(tr *goquery.Selection) *[]PostmanCollectionKeyValue {
	var annotationMatches []string
	annotationPattern := regexp.MustCompile(`\s*(\(.+\))`)

	list := []PostmanCollectionKeyValue{}
	tr.Each(func(index int, s *goquery.Selection) {
		annotationMatches = nil
		kv := PostmanCollectionKeyValue{}
		td := s.Find("td")
		td.Each(func(index int, s *goquery.Selection) {
			text := ts(s.Text())
			switch index {
			case 0: // td[0] is parameter name
				kv.Key = annotationPattern.ReplaceAllString(text, "")
				annotationMatches = annotationPattern.FindStringSubmatch(text)
			case 1: // td[1] is parameter type
				kv.Value = "<" + text + ">"
			case 2: // td[2] is parameter description
				kv.Description = text
				if len(annotationMatches) > 0 {
					kv.Description += " " + annotationMatches[1]
				}
			default:
				break
			}
		})
		list = append(list, kv)
	})
	return &list
}

func buildItem(name string, description string, method string, url string, query *[]PostmanCollectionKeyValue, variable *[]PostmanCollectionKeyValue, body *PostmanCollectionBody) *PostmanCollectionItem {
	fmt.Printf("%s -> %s %s\n", name, method, url)
	path := strings.Split(url, "/")[1:] // path[0] is "", so remove it
	return &PostmanCollectionItem{
		Name: name,
		Request: PostmanCollectionRequest{
			Method: method,
			Header: []string{},
			Body:   *body,
			URL: PostmanCollectionURL{
				Raw:      url,
				Host:     []string{"{{base_url}}"},
				Path:     path,
				Query:    *query,
				Variable: *variable,
			},
			Description: description,
		},
	}
}

func savePostmanCollection(collection *PostmanCollection) {
	f, err := os.Create("backlog_api_postman_collection.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := json.MarshalIndent(*collection, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = f.Write(b); err != nil {
		log.Fatal(err)
	}
}
