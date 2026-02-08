package registry

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRepositories(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/_catalog" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(catalogResponse{
			Repositories: []string{"app1", "app2"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	repos, err := c.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos, got %d", len(repos))
	}
	if repos[0] != "app1" || repos[1] != "app2" {
		t.Fatalf("unexpected repos: %v", repos)
	}
}

func TestListTags(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/myapp/tags/list" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(tagsResponse{
			Tags: []string{"1h", "30m", "latest"},
		})
	}))
	defer srv.Close()

	c := New(srv.URL)
	tags, err := c.ListTags(context.Background(), "myapp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(tags) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(tags))
	}
}

func TestListRepositories_Pagination(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Link", `</v2/_catalog?n=1000&last=app1>; rel="next"`)
			_ = json.NewEncoder(w).Encode(catalogResponse{
				Repositories: []string{"app1"},
			})
		} else {
			_ = json.NewEncoder(w).Encode(catalogResponse{
				Repositories: []string{"app2"},
			})
		}
	}))
	defer srv.Close()

	c := New(srv.URL)
	repos, err := c.ListRepositories(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(repos) != 2 {
		t.Fatalf("expected 2 repos across pages, got %d", len(repos))
	}
	if callCount != 2 {
		t.Fatalf("expected 2 API calls, got %d", callCount)
	}
}
