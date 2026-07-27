package selfmanage

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

const (
	DefaultRepository = "CGuiho/buda"
	DefaultAPIBase    = "https://api.github.com"
)

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Asset struct {
	Name        string `json:"name"`
	DownloadURL string `json:"browser_download_url"`
	Size        int64  `json:"size"`
}

type Release struct {
	TagName     string    `json:"tag_name"`
	Name        string    `json:"name"`
	PublishedAt time.Time `json:"published_at"`
	HTMLURL     string    `json:"html_url"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
	Assets      []Asset   `json:"assets"`
	Version     string    `json:"version"`
}

type Catalog struct {
	Client     Doer
	BaseURL    string
	Repository string
}

var semverPattern = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(?:-([0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*))?(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`)

func VersionFromTag(tag string) (string, bool) {
	const prefix = "buda/v"
	if !strings.HasPrefix(tag, prefix) {
		return "", false
	}
	version := strings.TrimPrefix(tag, prefix)
	return version, strictSemver(version)
}

func strictSemver(version string) bool {
	match := semverPattern.FindStringSubmatch(version)
	if match == nil || match[4] == "" {
		return match != nil
	}
	for _, identifier := range strings.Split(match[4], ".") {
		if len(identifier) > 1 && identifier[0] == '0' {
			if _, err := strconv.ParseUint(identifier, 10, 64); err == nil {
				return false
			}
		}
	}
	return true
}

func CompareVersions(left, right string) int {
	l := semverPattern.FindStringSubmatch(strings.TrimPrefix(left, "v"))
	r := semverPattern.FindStringSubmatch(strings.TrimPrefix(right, "v"))
	if l == nil || r == nil {
		return strings.Compare(left, right)
	}
	for index := 1; index <= 3; index++ {
		lv, _ := strconv.ParseUint(l[index], 10, 64)
		rv, _ := strconv.ParseUint(r[index], 10, 64)
		if lv < rv {
			return -1
		}
		if lv > rv {
			return 1
		}
	}
	return comparePrerelease(l[4], r[4])
}

func comparePrerelease(left, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return 1
	}
	if right == "" {
		return -1
	}
	lparts, rparts := strings.Split(left, "."), strings.Split(right, ".")
	for index := 0; index < len(lparts) && index < len(rparts); index++ {
		if lparts[index] == rparts[index] {
			continue
		}
		li, lerr := strconv.ParseUint(lparts[index], 10, 64)
		ri, rerr := strconv.ParseUint(rparts[index], 10, 64)
		switch {
		case lerr == nil && rerr == nil && li < ri:
			return -1
		case lerr == nil && rerr == nil:
			return 1
		case lerr == nil:
			return -1
		case rerr == nil:
			return 1
		default:
			return strings.Compare(lparts[index], rparts[index])
		}
	}
	if len(lparts) < len(rparts) {
		return -1
	}
	return 1
}

func (catalog Catalog) Releases(ctx context.Context) ([]Release, error) {
	client := catalog.Client
	if client == nil {
		client = &http.Client{Timeout: 30 * time.Second}
	}
	base := strings.TrimRight(catalog.BaseURL, "/")
	if base == "" {
		base = DefaultAPIBase
	}
	repository := catalog.Repository
	if repository == "" {
		repository = DefaultRepository
	}
	seen := map[string]bool{}
	result := []Release{}
	for page := 1; page <= 100; page++ {
		url := fmt.Sprintf("%s/repos/%s/releases?per_page=100&page=%d", base, repository, page)
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return nil, fmt.Errorf("create GitHub release request: %w", err)
		}
		request.Header.Set("Accept", "application/vnd.github+json")
		request.Header.Set("User-Agent", "guiho-buda-go")
		response, err := client.Do(request)
		if err != nil {
			return nil, fmt.Errorf("fetch GitHub releases: %w", err)
		}
		if response.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(io.LimitReader(response.Body, 4<<10))
			_ = response.Body.Close()
			return nil, fmt.Errorf("GitHub releases returned %s: %s", response.Status, strings.TrimSpace(string(body)))
		}
		var batch []Release
		decoder := json.NewDecoder(io.LimitReader(response.Body, 8<<20))
		if err := decoder.Decode(&batch); err != nil {
			_ = response.Body.Close()
			return nil, fmt.Errorf("decode GitHub releases: %w", err)
		}
		var trailing any
		if err := decoder.Decode(&trailing); err != io.EOF {
			_ = response.Body.Close()
			return nil, fmt.Errorf("decode GitHub releases: expected exactly one JSON document")
		}
		if err := response.Body.Close(); err != nil {
			return nil, fmt.Errorf("close GitHub release response: %w", err)
		}
		for _, release := range batch {
			version, ok := VersionFromTag(release.TagName)
			if !ok || release.Draft || seen[release.TagName] {
				continue
			}
			release.Version = version
			release.Prerelease = release.Prerelease || strings.Contains(version, "-")
			seen[release.TagName] = true
			result = append(result, release)
		}
		if len(batch) < 100 {
			sort.Slice(result, func(i, j int) bool { return CompareVersions(result[i].Version, result[j].Version) > 0 })
			return result, nil
		}
	}
	return nil, fmt.Errorf("GitHub release pagination exceeded 100 pages")
}
