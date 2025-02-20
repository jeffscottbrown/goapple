package model

type SearchResult struct {
	Albums []struct {
		ArtistName string `json:"artistName"`
		AlbumTitle string `json:"collectionName"`
		Url        string `json:"collectionViewUrl"`
	} `json:"results"`
}
