package main

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
