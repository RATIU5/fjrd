package toml

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type PathType int

const (
	PathTypeUnknown PathType = iota
	PathTypeNetwork
	PathTypeLocal
	PathTypeNonExistent
)

func (pt PathType) String() string {
	switch pt {
	case PathTypeNetwork:
		return "Network Path"
	case PathTypeLocal:
		return "Local Path"
	case PathTypeNonExistent:
		return "Non-existent Path"
	default:
		return "Unknown"
	}
}

func determinePathType(path string) PathType {
	if path == "" {
		return PathTypeUnknown
	}

	if isNetworkPath(path) {
		return PathTypeNetwork
	}

	_, err := os.Stat(path)
	if err != nil {
		return PathTypeNonExistent
	}

	return PathTypeLocal
}

func isNetworkPath(str string) bool {
	u, err := url.Parse(str)
	if err == nil && u.Scheme != "" {
		return true
	}

	if isGitRepoPath(str) {
		return true
	}

	return false
}

func getLocalTomlFile(path string) (string, error) {
	stats, err := os.Stat(path)
	if err != nil {
		return "", errors.New("failed to read stats from path")
	}

	if stats.IsDir() {
		return "", errors.New("expected file, got directory")
	}

	if filepath.Ext(path) != ".toml" {
		return "", errors.New("expected file to be of type 'toml'")
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("failed to read local toml file contents")
	}

	return string(body), nil
}

func getNetworkTomlResource(path string) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", errors.New("failed to parse url")
	}

	if u.Scheme == "https" {
		return getHTTPSTomlResource(path)
	}

	if u.Scheme == "git" {
		return getGitTomlResource(path)
	}

	if isGitRepoPath(path) {
		return getGitTomlResource(path)
	}

	if u.Scheme != "" && u.Scheme != "https" && u.Scheme != "git" {
		return "", errors.New("cannot use non-secure protocol")
	}

	return "", errors.New("failed to retrieve the network resource contents")
}

func getHTTPSTomlResource(urlStr string) (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			if req.URL.Scheme != "https" {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	resp, err := client.Get(urlStr)
	if err != nil {
		return "", errors.New("failed to fetch network resource")
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.New("received non-200 status from network resource")
	}

	if !strings.HasSuffix(urlStr, ".toml") {
		return "", errors.New("URL must point to a .toml file")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("failed to read network resource body")
	}

	return string(body), nil
}

func isGitRepoPath(path string) bool {
	if strings.HasPrefix(path, "./") || strings.HasPrefix(path, "/") || strings.HasPrefix(path, "../") {
		return false
	}

	if strings.HasPrefix(path, "github.com/") || strings.HasPrefix(path, "www.github.com/") {
		return true
	}

	if strings.Contains(path, "/") && !strings.Contains(path, ".") {
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" {
			return true
		}
		if len(parts) >= 4 && parts[0] != "" && parts[1] != "" {
			return true
		}
	}

	return false
}

func getGitTomlResource(path string) (string, error) {
	gitPath := parseGitPath(path)
	if gitPath == "" {
		return "", errors.New("invalid git repository path")
	}

	cmd := exec.Command("git", "ls-remote", "--exit-code", "https://github.com/"+gitPath)
	_, err := cmd.Output()
	if err != nil {
		return "", errors.New("git repository not found or not accessible")
	}

	rawURL := "https://raw.githubusercontent.com/" + gitPath
	return getHTTPSTomlResource(rawURL)
}

func parseGitPath(path string) string {
	cleanPath := strings.TrimPrefix(path, "git://")
	cleanPath = strings.TrimPrefix(cleanPath, "https://")
	cleanPath = strings.TrimPrefix(cleanPath, "github.com/")
	cleanPath = strings.TrimPrefix(cleanPath, "www.github.com/")

	parts := strings.Split(cleanPath, "/")
	if len(parts) < 2 {
		return ""
	}

	owner := parts[0]
	repo := parts[1]

	if len(parts) == 2 {
		commonFiles := []string{"fjrd.config.toml", "fjrd.toml"}
		for _, file := range commonFiles {
			testURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/main/%s", owner, repo, file)
			resp, err := http.Head(testURL)
			if err == nil && resp.StatusCode == 200 {
				resp.Body.Close()
				return fmt.Sprintf("%s/%s/main/%s", owner, repo, file)
			}
			if resp != nil {
				resp.Body.Close()
			}
		}
		return ""
	}

	if len(parts) >= 4 {
		if len(parts) >= 5 && parts[2] == "blob" {
			branch := parts[3]
			filePath := strings.Join(parts[4:], "/")
			if strings.HasSuffix(filePath, ".toml") {
				return fmt.Sprintf("%s/%s/%s/%s", owner, repo, branch, filePath)
			}
		}
		return strings.Join(parts, "/")
	}

	return ""
}

func ResolveTomlResource(location string) (string, error) {
	pathType := determinePathType(location)
	switch pathType {
	case PathTypeLocal:
		tomlBody, err := getLocalTomlFile(location)
		if err != nil {
			return "", errors.New("failed to read local toml file")
		}
		return tomlBody, nil
	case PathTypeNetwork:
		tomlBody, err := getNetworkTomlResource(location)
		if err != nil {
			return "", errors.New("failed to read remote toml file")
		}
		return tomlBody, nil
	case PathTypeNonExistent:
		return "", errors.New("failed to read location path")
	default:
		return "", errors.New("unknown location provided")
	}
}
