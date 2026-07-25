package source

import (
	"context"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CGuiho/buda/internal/repository"
)

func TestAcquireAndStoreLocalSourceIdempotently(t *testing.T) {
	wiki := filepath.Join(t.TempDir(), "wiki")
	if _, err := repository.Initialize(wiki, repository.InitOptions{WikiID: "wiki"}); err != nil {
		t.Fatal(err)
	}
	selected, err := repository.Open(wiki)
	if err != nil {
		t.Fatal(err)
	}
	input := filepath.Join(t.TempDir(), "evidence.md")
	if err := os.WriteFile(input, []byte("durable evidence"), 0o644); err != nil {
		t.Fatal(err)
	}
	artifact, err := Acquire(context.Background(), input, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	first, err := Store(selected, artifact)
	if err != nil || !first.Created {
		t.Fatalf("first Store = %+v, %v", first, err)
	}
	second, err := Store(selected, artifact)
	if err != nil || second.Created || second.SourceID != first.SourceID {
		t.Fatalf("second Store = %+v, %v", second, err)
	}
}

func TestAcquireExplicitURLOnly(t *testing.T) {
	client := roundTripperFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.String() != "https://example.test/source" {
			t.Fatalf("URL = %s", request.URL)
		}
		return &http.Response{
			StatusCode: 200,
			Header:     http.Header{"Content-Type": []string{"text/plain"}},
			Body:       ioNopCloser{strings.NewReader("source")},
		}, nil
	})
	artifact, err := Acquire(context.Background(), "https://example.test/source", client, 1024)
	if err != nil || artifact.Original == "" || artifact.Digest == "" {
		t.Fatalf("Acquire = %+v, %v", artifact, err)
	}
	if _, err := Acquire(context.Background(), "ftp://example.test/source", client, 1024); err == nil {
		t.Fatal("unsupported URL scheme accepted")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (function roundTripperFunc) Do(request *http.Request) (*http.Response, error) {
	return function(request)
}

type ioNopCloser struct{ *strings.Reader }

func (ioNopCloser) Close() error { return nil }
