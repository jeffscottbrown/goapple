package music

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func setupWiremock(t *testing.T, stubPath string, stubBody []byte) (container testcontainers.Container, endpoint string, teardown func()) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "wiremock/wiremock:3.13.0",
		ExposedPorts: []string{"8080/tcp"},
		WaitingFor:   wait.ForListeningPort("8080/tcp"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start wiremock container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	port, err := container.MappedPort(ctx, "8080")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}
	endpoint = fmt.Sprintf("http://%s:%s", host, port.Port())

	// Register stub
	url := fmt.Sprintf("%s/__admin/mappings", endpoint)
	reqBody := stubBody
	resp, err := http.Post(url, "application/json", ioutil.NopCloser(
		// Use a reader for the stub body
		bytes.NewReader(reqBody),
	))
	if err != nil {
		t.Fatalf("failed to register wiremock stub: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		body, _ := ioutil.ReadAll(resp.Body)
		t.Fatalf("failed to register stub, status: %d, body: %s", resp.StatusCode, string(body))
	}

	teardown = func() {
		container.Terminate(ctx)
	}
	return
}

func TestSearch_WithWiremock(t *testing.T) {
	stub := map[string]interface{}{
		"request": map[string]interface{}{
			"method":  "GET",
			"urlPath": "/search",
			"queryParameters": map[string]interface{}{
				"term": map[string]interface{}{
					"equalTo": "dio",
				},
			},
		},
		"response": map[string]interface{}{
			"status": 200,
			"jsonBody": map[string]any{
				"results": []map[string]string{
					{
						"artistName":        "Dio",
						"collectionName":    "Holy Diver",
						"collectionViewUrl": "https://en.wikipedia.org/wiki/Holy_Diver",
					},
					{
						"artistName":        "Rainbow",
						"collectionName":    "Rising",
						"collectionViewUrl": "https://en.wikipedia.org/wiki/Rising_(Rainbow_album)",
					},
				},
			},
		},
	}
	stubBody, err := json.Marshal(stub)
	if err != nil {
		t.Fatalf("failed to marshal stub: %v", err)
	}

	_, endpoint, teardown := setupWiremock(t, "/v1/catalog/us/search", stubBody)
	defer teardown()

	// Set the Apple Music API base URL to the wiremock endpoint
	originalEndpoint := AppleMusicAPI
	AppleMusicAPI = endpoint + "/search"

	defer func() {
		// Restore the original endpoint after the test
		AppleMusicAPI = originalEndpoint
	}()

	searchTerm := "dio"
	limit := "5"

	result, errorMessage := SearchApple(searchTerm, limit)

	assert.Equal(t, 2, len(result.Albums))
	assert.Equal(t, "Dio", result.Albums[0].ArtistName)
	assert.Equal(t, "Holy Diver", result.Albums[0].AlbumTitle)
	assert.Equal(t, "https://en.wikipedia.org/wiki/Holy_Diver", result.Albums[0].Url)
	assert.Equal(t, "Rainbow", result.Albums[1].ArtistName)
	assert.Equal(t, "Rising", result.Albums[1].AlbumTitle)
	assert.Equal(t, "https://en.wikipedia.org/wiki/Rising_(Rainbow_album)", result.Albums[1].Url)
	assert.Empty(t, errorMessage)
}
