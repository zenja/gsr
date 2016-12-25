package server

import (
	"fmt"
	"html/template"
	"net/http"
	"time"

	"github.com/zenja/gsr/searcher"
)

var ser *searcher.Searcher
var temps = template.Must(template.ParseFiles("template/index.html", "template/search.html", "template/error.html"))

type QueryAndSearchResults struct {
	Query  string
	Result *searcher.SearchResult
}

func StartServer(port int, apiKey string, engineID string, timeout time.Duration) {
	ser = searcher.New(apiKey, engineID, timeout)
	http.Handle("/public/", http.StripPrefix("/public/", http.FileServer(http.Dir("./public"))))
	http.HandleFunc("/", handle)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func handle(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	// if there is no query string, display index.html
	if len(query) == 0 {
		if err := temps.ExecuteTemplate(w, "index.html", nil); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	results, err := ser.Search(query)
	// display error page if got error in search
	if err != nil {
		if err := temps.ExecuteTemplate(w, "error.html", err); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// display search results page
	if err := temps.ExecuteTemplate(w, "search.html", QueryAndSearchResults{query, results}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
