// Package source acquires only an explicitly named local file or URL, hashes
// its bytes, and stores an immutable content-addressed artifact. It performs no
// crawling, connector discovery, or model work.
package source

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/repository"
)

const DefaultMaxBytes int64 = 64 << 20

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Artifact struct {
	Original  string `json:"original"`
	Digest    string `json:"digest"`
	Bytes     []byte `json:"-"`
	MediaType string `json:"media_type,omitempty"`
}

type Stored struct {
	SourceID     string `json:"source_id"`
	Digest       string `json:"digest"`
	AbsolutePath string `json:"absolute_path"`
	Resource     string `json:"resource"`
	Created      bool   `json:"created"`
}

func Acquire(ctx context.Context, explicit string, client Doer, maxBytes int64) (Artifact, error) {
	if strings.TrimSpace(explicit) == "" {
		return Artifact{}, errors.New("source must not be empty")
	}
	if maxBytes <= 0 {
		maxBytes = DefaultMaxBytes
	}
	parsed, err := url.Parse(explicit)
	isWebURL := strings.HasPrefix(strings.ToLower(explicit), "http://") || strings.HasPrefix(strings.ToLower(explicit), "https://")
	if isWebURL {
		if err != nil {
			return Artifact{}, fmt.Errorf("parse explicit source URL: %w", err)
		}
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return Artifact{}, fmt.Errorf("unsupported source URL scheme %q", parsed.Scheme)
		}
		if client == nil {
			return Artifact{}, errors.New("explicit URL source requires an HTTP client")
		}
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, parsed.String(), nil)
		if err != nil {
			return Artifact{}, fmt.Errorf("create explicit source request: %w", err)
		}
		response, err := client.Do(request)
		if err != nil {
			return Artifact{}, fmt.Errorf("fetch explicit source URL: %w", err)
		}
		defer response.Body.Close()
		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return Artifact{}, fmt.Errorf("fetch explicit source URL: HTTP %d", response.StatusCode)
		}
		data, err := readBounded(response.Body, maxBytes)
		if err != nil {
			return Artifact{}, err
		}
		return newArtifact(parsed.String(), response.Header.Get("Content-Type"), data), nil
	}
	if strings.Contains(explicit, "://") {
		scheme := strings.SplitN(explicit, "://", 2)[0]
		return Artifact{}, fmt.Errorf("unsupported source URL scheme %q", scheme)
	}
	absolute, err := filepath.Abs(explicit)
	if err != nil {
		return Artifact{}, fmt.Errorf("resolve local source %q: %w", explicit, err)
	}
	file, err := os.Open(absolute)
	if err != nil {
		return Artifact{}, fmt.Errorf("open local source %q: %w", absolute, err)
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return Artifact{}, fmt.Errorf("inspect local source %q: %w", absolute, err)
	}
	if !info.Mode().IsRegular() {
		return Artifact{}, fmt.Errorf("local source %q must be a regular file", absolute)
	}
	data, err := readBounded(file, maxBytes)
	if err != nil {
		return Artifact{}, err
	}
	return newArtifact(absolute, http.DetectContentType(data), data), nil
}

func Store(selected repository.Repository, artifact Artifact) (Stored, error) {
	digestHex := strings.TrimPrefix(artifact.Digest, "sha256:")
	if len(digestHex) != sha256.Size*2 {
		return Stored{}, errors.New("artifact digest must be sha256:<64 hex characters>")
	}
	if calculated := sha256.Sum256(artifact.Bytes); hex.EncodeToString(calculated[:]) != digestHex {
		return Stored{}, errors.New("artifact bytes do not match digest")
	}
	relative := filepath.Join("references", "raw", digestHex+".source")
	absolute := filepath.Join(selected.Bundle, relative)
	created := false
	if existing, err := os.ReadFile(absolute); err == nil {
		existingDigest := sha256.Sum256(existing)
		if hex.EncodeToString(existingDigest[:]) != digestHex {
			return Stored{}, fmt.Errorf("content-addressed artifact %q has unexpected bytes", absolute)
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := repository.AtomicWriteFile(absolute, artifact.Bytes, 0o444); err != nil {
			return Stored{}, err
		}
		created = true
	} else {
		return Stored{}, fmt.Errorf("inspect content-addressed artifact: %w", err)
	}
	return Stored{
		SourceID:     "source-" + digestHex[:16],
		Digest:       artifact.Digest,
		AbsolutePath: absolute,
		Resource:     "/" + filepath.ToSlash(relative),
		Created:      created,
	}, nil
}

func StableUID(namespace, value string) string {
	digest := sha256.Sum256([]byte(namespace + "\x00" + value))
	bytes := append([]byte(nil), digest[:16]...)
	bytes[6] = (bytes[6] & 0x0f) | 0x50
	bytes[8] = (bytes[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", bytes[0:4], bytes[4:6], bytes[6:8], bytes[8:10], bytes[10:16])
}

func readBounded(reader io.Reader, maxBytes int64) ([]byte, error) {
	limited := io.LimitReader(reader, maxBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("read explicit source: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("explicit source exceeds %d byte limit", maxBytes)
	}
	return data, nil
}

func newArtifact(original, mediaType string, data []byte) Artifact {
	digest := sha256.Sum256(data)
	return Artifact{
		Original:  original,
		Digest:    "sha256:" + hex.EncodeToString(digest[:]),
		Bytes:     append([]byte(nil), data...),
		MediaType: mediaType,
	}
}
