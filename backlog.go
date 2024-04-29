package main

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly/v2"
)

var ts = strings.TrimSpace

func findName(e *colly.HTMLElement) string {
	name := ts(e.DOM.Find("h1").Text())
	if name == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		return ts(e.DOM.Find("h2").First().Text())
	}

	return name
}

func findDescription(e *colly.HTMLElement) string {
	description := ts(e.DOM.Find("h1 + p").Text())
	if description == "" {
		// for HTML structure misalignment
		// https://developer.nulab.com/ja/docs/backlog/api/2/get-issue-participant-list
		return ts(e.DOM.Find("h2 + p").First().Text())
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
