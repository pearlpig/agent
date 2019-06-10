package cutil

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/PuerkitoBio/goquery"
)

// GetDocument .
func GetDocument(url string) (*goquery.Document, error) {
	client := http.Client{}
	client.Timeout = time.Duration(10) * time.Second
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

// Curl .
func Curl(url string) ([]byte, error) {
	client := http.Client{}
	client.Timeout = time.Duration(10) * time.Second
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}
	return body, nil
}
