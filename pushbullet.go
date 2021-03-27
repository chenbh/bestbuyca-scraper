package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

const api = "https://api.pushbullet.com/v2/pushes"

type payload struct {
	Type  string `json:"type"`
	Title string `json:"title"`
	Body  string `json:"body"`
	URL   string `json:"url"`
}

func notify(p product, token string) error {
	return push(payload{
		Type:  "link",
		Title: "Product in stock",
		Body:  fmt.Sprintf("%v: $%v", p.Name, p.SalePrice),
		URL:   p.ProductUrl,
	}, token)
}

func push(p payload, token string) error {
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(p)
	if err != nil {
		return fmt.Errorf("encoding: %v", err)
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", api, buf)
	if err != nil {
		return fmt.Errorf("creating request: %v", err)
	}

	req.Header.Set("Access-Token", token)
	req.Header.Set("Content-Type", "application/json")

	_, err = client.Do(req)
	return err
}
