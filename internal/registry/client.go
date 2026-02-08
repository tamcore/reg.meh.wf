package registry

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Client talks to the OCI distribution registry HTTP API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// New creates a new registry client.
func New(registryURL string) *Client {
	return &Client{
		baseURL:    strings.TrimRight(registryURL, "/"),
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type catalogResponse struct {
	Repositories []string `json:"repositories"`
}

type tagsResponse struct {
	Tags []string `json:"tags"`
}

// ListRepositories returns all repository names from the registry catalog.
func (c *Client) ListRepositories(ctx context.Context) ([]string, error) {
	var all []string
	url := fmt.Sprintf("%s/v2/_catalog?n=1000", c.baseURL)

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating catalog request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("listing catalog: %w", err)
		}

		var catalog catalogResponse
		err = json.NewDecoder(resp.Body).Decode(&catalog)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decoding catalog response: %w", err)
		}

		all = append(all, catalog.Repositories...)
		url = nextLink(resp, c.baseURL)
	}

	return all, nil
}

// ListTags returns all tags for a given repository.
func (c *Client) ListTags(ctx context.Context, repo string) ([]string, error) {
	var all []string
	url := fmt.Sprintf("%s/v2/%s/tags/list?n=1000", c.baseURL, repo)

	for url != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating tags request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("listing tags for %s: %w", repo, err)
		}

		var tags tagsResponse
		err = json.NewDecoder(resp.Body).Decode(&tags)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("decoding tags response: %w", err)
		}

		all = append(all, tags.Tags...)
		url = nextLink(resp, c.baseURL)
	}

	return all, nil
}

// nextLink parses the Link header for pagination.
// The registry returns: Link: </v2/_catalog?n=1000&last=repo>; rel="next"
func nextLink(resp *http.Response, baseURL string) string {
	link := resp.Header.Get("Link")
	if link == "" {
		return ""
	}

	// Parse format: </path>; rel="next"
	start := strings.Index(link, "<")
	end := strings.Index(link, ">")
	if start < 0 || end < 0 || end <= start {
		return ""
	}

	path := link[start+1 : end]
	if strings.HasPrefix(path, "/") {
		return baseURL + path
	}
	return path
}
