package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strings"
)

const goDownloadAPI = "https://go.dev/dl/?mode=json&include=all"

// Version represents a Go version
type Version struct {
	Version string `json:"version"`
	Stable  bool   `json:"stable"`
	Files   []File `json:"files"`
}

// File represents a downloadable file for a Go version
type File struct {
	Filename string `json:"filename"`
	OS       string `json:"os"`
	Arch     string `json:"arch"`
	Version  string `json:"version"`
	SHA256   string `json:"sha256"`
	Size     int64  `json:"size"`
	Kind     string `json:"kind"`
}

// FetchVersions fetches all available Go versions from go.dev
func FetchVersions() ([]Version, error) {
	resp, err := http.Get(goDownloadAPI)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch versions: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var versions []Version
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return versions, nil
}

// GetStableVersions returns only stable versions
func GetStableVersions(versions []Version) []Version {
	var stable []Version
	for _, v := range versions {
		if v.Stable {
			stable = append(stable, v)
		}
	}
	return stable
}

// GetMajorVersions groups versions by major version and returns the latest of each
func GetMajorVersions(versions []Version) []Version {
	majorMap := make(map[string]Version)

	for _, v := range versions {
		major := extractMajorVersion(v.Version)
		if existing, ok := majorMap[major]; !ok || compareVersions(v.Version, existing.Version) > 0 {
			majorMap[major] = v
		}
	}

	var result []Version
	for _, v := range majorMap {
		result = append(result, v)
	}

	sort.Slice(result, func(i, j int) bool {
		return compareVersions(result[i].Version, result[j].Version) > 0
	})

	return result
}

// GetFile returns the file info for current OS/Arch
func GetFile(version Version) (*File, error) {
	targetOS := runtime.GOOS
	targetArch := runtime.GOARCH

	for _, f := range version.Files {
		if f.OS == targetOS && f.Arch == targetArch && f.Kind == "archive" {
			return &f, nil
		}
	}

	return nil, fmt.Errorf("no download available for %s/%s", targetOS, targetArch)
}

// FormatVersion returns a clean version string
func FormatVersion(version string) string {
	return strings.TrimPrefix(version, "go")
}

func extractMajorVersion(version string) string {
	v := strings.TrimPrefix(version, "go")
	parts := strings.Split(v, ".")
	if len(parts) >= 2 {
		return parts[0] + "." + parts[1]
	}
	return v
}

func compareVersions(a, b string) int {
	aParts := parseVersion(a)
	bParts := parseVersion(b)

	for i := 0; i < len(aParts) && i < len(bParts); i++ {
		if aParts[i] > bParts[i] {
			return 1
		}
		if aParts[i] < bParts[i] {
			return -1
		}
	}

	return len(aParts) - len(bParts)
}

func parseVersion(version string) []int {
	v := strings.TrimPrefix(version, "go")
	v = strings.Split(v, "rc")[0]
	v = strings.Split(v, "beta")[0]

	parts := strings.Split(v, ".")
	var result []int
	for _, p := range parts {
		var num int
		_, _ = fmt.Sscanf(p, "%d", &num)
		result = append(result, num)
	}
	return result
}
