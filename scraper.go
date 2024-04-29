package main

import (
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
		name = ts(e.DOM.Find("h2").First().Text())
	}
	return name
}

func findDesc(e *colly.HTMLElement) string {
	return ts(e.DOM.Find("p").First().Text())
}

func findReq(e *colly.HTMLElement) []string {
	return strings.Split(ts(e.DOM.Find("pre").First().Text()), " ") // Split to HTTP method and URL
}

func findQuery(e *colly.HTMLElement, config *Config) *[]PostmanCollectionKeyValue {
	selector := config.SelectorID.QueryParameter + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findVariable(e *colly.HTMLElement, config *Config) *[]PostmanCollectionKeyValue {
	selector := config.SelectorID.URLParameter + " + table tbody tr"
	tr := e.DOM.Find(selector)
	keyValue := buildKeyValue(tr)
	return keyValue
}

func findBody(e *colly.HTMLElement, config *Config) *PostmanCollectionBody {
	selector := config.SelectorID.RequestParameter + " + pre + table tbody tr"
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
