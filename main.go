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

type Dictionary map[string]string

const (
	EN = "English"
	JA = "Japanese"
)

func main() {
	_, lang, err := askLanguage()
	if err != nil {
		return
	}
	d := getDictionary(lang)

	collection := getPostmanCollectionTemplate()

	parentController := colly.NewCollector(colly.AllowedDomains("developer.nulab.com"))
	childController := parentController.Clone()

	parentController.OnHTML(d["linkSelector"], func(e *colly.HTMLElement) {
		link := e.Attr("href")
		childController.Visit(e.Request.AbsoluteURL(link))
	})

	childController.OnHTML("div.content", func(e *colly.HTMLElement) {
		name := findName(e)
		description := findDescription(e)
		method := findMethod(e, d)
		url := findURL(e)
		query := findQuery(e, d)
		variable := findVariable(e, d)
		body := findBody(e, d)
		item := buildItem(name, description, method, url, query, variable, body)
		collection.Item = append(collection.Item, *item)
	})

	parentController.Visit(d["url"])

	savePostmanCollection(collection)
}

func getDictionary(language string) Dictionary {
	if language == JA {
		return Dictionary{
			"linkSelector":       "a[href^='/ja/docs/backlog/api/2/']",
			"methodId":           "#メソッド",
			"queryParameterId":   "#クエリパラメーター",
			"requestParameterId": "#リクエストパラメーター",
			"url":                "https://developer.nulab.com/ja/docs/backlog",
			"urlParameterId":     "#url-パラメーター",
		}
	}

	return Dictionary{
		"linkSelector":       "a[href^='/docs/backlog/api/2/']",
		"methodId":           "#method",
		"queryParameterId":   "#query-parameters",
		"requestParameterId": "#form-parameters",
		"url":                "https://developer.nulab.com/docs/backlog",
		"urlParameterId":     "#url-parameters",
	}
}

func askLanguage() (int, string, error) {
	prompt := promptui.Select{
		Label: "Select language",
		Items: []string{EN, JA},
	}

	return prompt.Run()
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

func t(s string) string {
	return strings.TrimSpace(s)
}

func findName(e *colly.HTMLElement) string {
	name := t(e.DOM.Find("h1").Text())
	if name == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		name = t(e.DOM.Find("h2").First().Text())
	}
	return name
}

func findDescription(e *colly.HTMLElement) string {
	description := t(e.DOM.Find("h1 + p").Text())
	if description == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		description = t(e.DOM.Find("h2 + p").First().Text())
	}
	return description
}

func findMethod(e *colly.HTMLElement, d Dictionary) string {
	selector := d["methodId"] + " + pre"
	return t(e.DOM.Find(selector).Text())
}

func findURL(e *colly.HTMLElement) string {
	return t(e.DOM.Find("#url + pre").Text())
}

func findQuery(e *colly.HTMLElement, d Dictionary) *[]PostmanCollectionKeyValue {
	selector := d["queryParameterId"] + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findVariable(e *colly.HTMLElement, d Dictionary) *[]PostmanCollectionKeyValue {
	selector := d["urlParameterId"] + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findBody(e *colly.HTMLElement, d Dictionary) *PostmanCollectionBody {
	selector := d["requestParameterId"] + " + pre + table tbody tr"
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
			text := t(s.Text())
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
