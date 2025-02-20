package music

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/jeffscottbrown/goapple/model"
	"github.com/patrickmn/go-cache"
)

var searchCache = cache.New(5*time.Minute, 10*time.Minute)

const AppleMusicAPI = "https://itunes.apple.com/search"

func SearchApple(searchTerm string, limit string) (model.SearchResult, string) {
	cachKey := fmt.Sprintf("%s-%s", searchTerm, limit)
	if cachedData, found := searchCache.Get(cachKey); found {
		slog.Debug("Cache hit", "searchTerm", searchTerm)
		return cachedData.(model.SearchResult), ""
	}

	fullURL := createSearchUrl(searchTerm, limit)

	slog.Debug("Querying Apple API", "url", fullURL)

	resp, err := http.Get(fullURL)

	var errorMessage string
	var result model.SearchResult
	if err != nil {
		errorMessage = "Failed to fetch data"
	} else {
		defer resp.Body.Close()

		if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
			errorMessage = "Failed to parse JSON"
		}
		searchCache.Set(cachKey, result, cache.DefaultExpiration)
	}
	return result, errorMessage
}

func createSearchUrl(searchTerm string, limit string) string {
	params := createRequestParameters(searchTerm, limit)

	fullURL := AppleMusicAPI + "?" + params.Encode()
	return fullURL
}

func createRequestParameters(searchTerm string, limit string) url.Values {
	params := url.Values{}
	params.Add("term", searchTerm)
	params.Add("media", "music")
	params.Add("entity", "album")
	params.Add("limit", limit)
	return params
}
