package searcher

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
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

func (s *Searcher) Search(query string) (*SearchResult, error) {
	return s.SearchFrom(query, 1)
}

func (s *Searcher) SearchFrom(query string, startIndex int) (*SearchResult, error) {
	qURL := fmt.Sprintf("https://www.googleapis.com/customsearch/v1?key=%s&cx=%s&q=%s&start=%d",
		s.apiKey, s.engineID, url.QueryEscape(query), startIndex)
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

	var singleResults []SingleSearchResult
	for _, res := range r.Items {
		singleResults = append(singleResults, SingleSearchResult{
			Title:            res.Title,
			Link:             res.Link,
			DisplayLink:      res.DisplayLink,
			HTMLFormattedURL: template.HTML(res.HtmlFormattedUrl),
			Snippet:          res.Snippet,
			HTMLSnippet:      template.HTML(strings.Replace(res.HtmlSnippet, "<br>", "", -1)),
		})
	}
	return &SearchResult{
		TotalResults:          int(r.SearchInformation.TotalResults),
		FormattedTotalResults: r.SearchInformation.FormattedTotalResults,
		SearchTime:            r.SearchInformation.SearchTime,
		FormattedSearchTime:   r.SearchInformation.FormattedSearchTime,
		Results:               singleResults,
		StartIndex:            int(r.Queries["request"][0].StartIndex),
	}, nil
}

type SearchResult struct {
	TotalResults          int
	FormattedTotalResults string
	SearchTime            float64
	FormattedSearchTime   string
	Results               []SingleSearchResult
	StartIndex            int
}

type SingleSearchResult struct {
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
