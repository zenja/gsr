package searcher

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	gcs "google.golang.org/api/customsearch/v1"
	"google.golang.org/api/googleapi"
)

type Searcher struct {
	apiKey   string
	engineID string
	timeout  time.Duration
}

func New(apiKey string, engineID string, timeout time.Duration) *Searcher {
	return &Searcher{apiKey: apiKey, engineID: engineID, timeout: timeout}
}

func (s *Searcher) Search(query string) ([]SearchResult, error) {
	qURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s",
		s.apiKey, s.engineID, url.QueryEscape(query))
	client := http.Client{
		Timeout: s.timeout,
	}
	resp, err := client.Get(qURL)
	if err != nil {
		return nil, fmt.Errorf("failed to request %s: %s", qURL, err)
	}
	defer resp.Body.Close()

	if err := googleapi.CheckResponse(resp); err != nil {
		gErr := err.(*googleapi.Error)
		return nil, NewSearchError(gErr)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %s", err)
	}

	var r gcs.Search
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search result: %s", err)
	}

	var results []SearchResult
	for _, res := range r.Items {
		results = append(results, SearchResult{
			Title:            res.Title,
			Link:             res.Link,
			DisplayLink:      res.DisplayLink,
			HTMLFormattedURL: template.HTML(res.HtmlFormattedUrl),
			Snippet:          res.Snippet,
			HTMLSnippet:      template.HTML(res.HtmlSnippet),
		})
	}
	return results, nil
}

type SearchResult struct {
	Title            string
	Link             string
	DisplayLink      string
	HTMLFormattedURL template.HTML
	Snippet          string
	HTMLSnippet      template.HTML
}

type SearchError struct {
	Message          string
	AssembledMessage string
	Errors           []SearchErrorItem
}

type SearchErrorItem struct {
	Message string
	Reason  string
}

func (se SearchError) Error() string {
	return se.AssembledMessage
}

func NewSearchError(gErr *googleapi.Error) SearchError {
	var errors []SearchErrorItem
	for _, eItem := range gErr.Errors {
		errors = append(errors, SearchErrorItem{Message: eItem.Message, Reason: eItem.Reason})
	}
	return SearchError{Message: gErr.Message, AssembledMessage: gErr.Error(), Errors: errors}
}
