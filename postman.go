package main

import (
	"encoding/json"
	"log"
	"os"
	"strings"
)

type PostmanCollection struct {
	Info PostmanCollectionInfo   `json:"info"`
	Item []PostmanCollectionItem `json:"item"`
	Auth PostmanCollectionAuth   `json:"auth"`
}

type PostmanCollectionInfo struct {
	Name   string `json:"name"`
	Schema string `json:"schema"`
}

type PostmanCollectionItem struct {
	Name    string                   `json:"name"`
	Request PostmanCollectionRequest `json:"request"`
}

func NewPostmanCollectionItem(name string, description string, method string, url string, query *[]PostmanCollectionKeyValue, variable *[]PostmanCollectionKeyValue, body *PostmanCollectionBody) *PostmanCollectionItem {
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

type PostmanCollectionRequest struct {
	Method      string                `json:"method"`
	Header      []string              `json:"header"`
	Body        PostmanCollectionBody `json:"body"`
	URL         PostmanCollectionURL  `json:"url"`
	Description string                `json:"description"`
}
type PostmanCollectionBody struct {
	Mode       string                      `json:"mode"`
	URLEncoded []PostmanCollectionKeyValue `json:"urlencoded"`
}

type PostmanCollectionURL struct {
	Raw      string                      `json:"raw"`
	Host     []string                    `json:"host"`
	Path     []string                    `json:"path"`
	Query    []PostmanCollectionKeyValue `json:"query"`
	Variable []PostmanCollectionKeyValue `json:"variable"`
}

type PostmanCollectionKeyValue struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description"`
	Type        string `json:"type"`
}

type PostmanCollectionAuth struct {
	Type   string                    `json:"type"`
	APIKey []PostmanCollectionAPIKey `json:"apikey"`
}

type PostmanCollectionAPIKey struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type"`
}

func NewPostmanCollection() *PostmanCollection {
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

func (c *PostmanCollection) Save() {
	f, err := os.Create("backlog_api_postman_collection.json")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	b, err := json.MarshalIndent(*c, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	if _, err = f.Write(b); err != nil {
		log.Fatal(err)
	}
}
