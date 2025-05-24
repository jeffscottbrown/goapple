package music

import (
	"net/http"
	"testing"

	"github.com/h2non/gock"
	"github.com/stretchr/testify/assert"
)

func TestSearchApple_ApiSuccess(t *testing.T) {
	defer gock.OffAll()

	searchTerm := "examples"
	limit := "5"

	responseJson := `{
		"results": [
			{
				"artistName": "First Test Artist",
				"collectionName": "First Test Album",
				"collectionViewUrl": "http://first-example.com/"
			}, {
				"artistName": "Second Test Artist",
				"collectionName": "Second Test Album",
				"collectionViewUrl": "http://second-example.com/"
			}
		]
	}`
	gock.New(AppleMusicAPI).
		MatchParam("term", searchTerm).
		MatchParam("media", "music").
		MatchParam("entity", "album").
		MatchParam("limit", limit).
		Reply(http.StatusOK).
		JSON(responseJson)

	result, errorMessage := SearchApple(searchTerm, limit)

	assert.Equal(t, 2, len(result.Albums))
	assert.Equal(t, "First Test Artist", result.Albums[0].ArtistName)
	assert.Equal(t, "First Test Album", result.Albums[0].AlbumTitle)
	assert.Equal(t, "http://first-example.com/", result.Albums[0].Url)
	assert.Equal(t, "Second Test Artist", result.Albums[1].ArtistName)
	assert.Equal(t, "Second Test Album", result.Albums[1].AlbumTitle)
	assert.Equal(t, "http://second-example.com/", result.Albums[1].Url)
	assert.Empty(t, errorMessage)
	assert.True(t, gock.IsDone())
}

func TestSearchApple_ApiFailure(t *testing.T) {
	defer gock.OffAll()

	searchTerm := "failures"
	limit := "5"

	gock.New(AppleMusicAPI).
		MatchParam("term", searchTerm).
		MatchParam("media", "music").
		MatchParam("entity", "album").
		MatchParam("limit", limit).
		Reply(http.StatusInternalServerError)

	result, errorMessage := SearchApple(searchTerm, limit)

	assert.Empty(t, result)
	assert.Equal(t, "Failed to fetch data", errorMessage)
	assert.True(t, gock.IsDone())
}

func TestSearchApple_InvalidJson(t *testing.T) {
	defer gock.OffAll()

	searchTerm := "invalidjson"
	limit := "5"

	gock.New(AppleMusicAPI).
		MatchParam("term", searchTerm).
		MatchParam("media", "music").
		MatchParam("entity", "album").
		MatchParam("limit", limit).
		Reply(http.StatusOK).
		BodyString("invalid json")

	result, errorMessage := SearchApple(searchTerm, limit)

	assert.Empty(t, result)
	assert.Equal(t, "Failed to parse JSON", errorMessage)
	assert.True(t, gock.IsDone())
}
