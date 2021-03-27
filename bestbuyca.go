package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// TODO: parse these out of https://www.bestbuy.ca/config.ashx
const (
	availabilityApiUrl = "https://www.bestbuy.ca/ecomm-api/availability/products"
	productApiUrl      = "https://www.bestbuy.ca/api/v2/json/product/"
	skuCollectionUrl   = "https://www.bestbuy.ca/api/v2/json/sku-collections/"
	searchApiUrl       = "https://www.bestbuy.ca/api/v2/json/search"

	userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.105 Safari/537.36"
)

type availability struct {
	Status      string `json:"status"`
	Purchasable bool   `json:"purchasable"`
}

type availabilities struct {
	Availabilities []struct {
		Pickup   availability `json:"pickup"`
		Shipping availability `json:"shipping"`
		Sku      string       `json:"sku"`
	} `json:"availabilities"`
}

func get(url string, result interface{}) error {
	client := &http.Client{}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("creating request for %v: %v", url, err)
	}

	// pretty sure their only effort at defeating bots is checking the user-agent header
	req.Header.Set("User-Agent", userAgent)
	res, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("getting %v: %v", url, err)
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		return fmt.Errorf("reading response: %v", err)
	}

	// their api response is utf-8 encoded for some reason
	// https://stackoverflow.com/a/31399046
	b = bytes.TrimPrefix(b, []byte("\xef\xbb\xbf")) // Or []byte{239, 187, 191}

	err = json.NewDecoder(bytes.NewReader(b)).Decode(result)
	if err != nil {
		return fmt.Errorf("decoding %v: %v", url, err)
	}
	return nil
}

func getAvailableSkus(skus []string) ([]string, error) {
	// ?skus=123|456|789
	url := availabilityApiUrl + fmt.Sprintf("?skus=%v", strings.Join(skus, "%7C"))
	var a availabilities

	err := get(url, &a)
	if err != nil {
		return nil, fmt.Errorf("availability: %v", err)
	}

	result := make([]string, 0)
	for _, v := range a.Availabilities {
		// maybe this will generate false alarms about stock available in nunavut, but oh well.
		if v.Pickup.Purchasable || v.Shipping.Purchasable {
			result = append(result, v.Sku)
		}
	}

	return result, nil
}

type product struct {
	Name       string  `json:"name"`
	Sku        string  `json:"sku"`
	SalePrice  float64 `json:"salePrice"`
	ProductUrl string  `json:"productUrl"`
}

func getProductFromSku(sku string) (product, error) {
	var p product
	err := get(productApiUrl+sku, &p)
	if err != nil {
		return product{}, fmt.Errorf("product: %v", err)
	}

	return p, nil
}

type results struct {
	Products []product `json:"products"`
}

func getSkusFromCollection(collectionId string) ([]string, error) {
	var c results
	// maybe the page size should be configurable?
	err := get(skuCollectionUrl+collectionId+"?pageSize=100", &c)
	if err != nil {
		return nil, fmt.Errorf("collection: %v", err)
	}

	result := make([]string, 0)
	for _, v := range c.Products {
		result = append(result, v.Sku)
	}

	return result, nil
}

// the query can be copied directly from the webpage
// e.g. ?path=category%253AComputers%2B%2526%2BTablets%253Bcategory%253APC%2BComponents%253Bcategory%253AGraphics%2BCards%253Bcustom0graphicscardmodel%253AGeForce%2BRTX%2B3060%257CGeForce%2BRTX%2B3060%2BTi%257CGeForce%2BRTX%2B3070
func getSkusFromSearch(query string) ([]string, error) {
	var r results
	err := get(searchApiUrl+query, &r)
	if err != nil {
		return nil, fmt.Errorf("search: %v", err)
	}

	result := make([]string, 0)
	for _, v := range r.Products {
		result = append(result, v.Sku)
	}

	return result, nil
}
