package server

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/zenja/gsr/searcher"
)

func pageID2StartIndex(pageID int) int {
	if pageID == 1 {
		return 1
	}
	return (pageID - 1) * 10
}

func startIndex2PageID(startIndex int) int {
	if startIndex <= 9 {
		return 1
	}
	return (startIndex / 10) + 1
}

func isSamePage(startIndex int, pageID int) bool {
	return startIndex2PageID(startIndex) == pageID
}

var funcMap = template.FuncMap{
	"pageID2StartIndex": pageID2StartIndex,
	"isSamePage":        isSamePage,
}

var ser *searcher.Searcher
var temps = template.Must(template.New("").Funcs(funcMap).ParseFiles(
	"template/index.html", "template/search.html", "template/error.html"))

type QueryAndSearchResults struct {
	Query         string
	PageIDs       []int
	CurrentPageID int
	Result        *searcher.SearchResult
}

func NewQueryAndSearchResults(query string, results *searcher.SearchResult) *QueryAndSearchResults {
	totalResults := results.TotalResults
	var numPage int
	if totalResults < 100 {
		numPage = totalResults / 10
		if totalResults%10 != 0 {
			numPage++
		}
	}
	numPage = 10
	var pageIDs []int
	for i := 1; i <= numPage; i++ {
		pageIDs = append(pageIDs, i)
	}
	return &QueryAndSearchResults{
		Query:         query,
		PageIDs:       pageIDs,
		CurrentPageID: startIndex2PageID(results.StartIndex),
		Result:        results,
	}
}

func StartServer(port int, apiKey string, engineID string, timeout time.Duration) {
	ser = searcher.New(apiKey, engineID, timeout)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var startIndex int
	if len(r.URL.Query().Get("start")) == 0 {
		startIndex = 1
	} else {
		startIndex, _ = strconv.Atoi(r.URL.Query().Get("start"))
		if startIndex <= 0 {
			startIndex = 1
		}
	}

	// if there is no query string, display index.html
	if len(query) == 0 {
		if err := temps.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	results, err := ser.SearchFrom(query, startIndex)
	// display error page if got error in search
	if err != nil {
		log.Print(err)
		if err := temps.ExecuteTemplate(w, "error.html", err); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// display search results page
	if err := temps.ExecuteTemplate(w, "search.html", NewQueryAndSearchResults(query, results)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
