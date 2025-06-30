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

	goToml "github.com/pelletier/go-toml/v2"
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

	if u.Scheme == "https" && isGitHubBlobURL(path) {
		rawURL := convertGitHubBlobToRaw(path)
		if rawURL != "" {
			return getHTTPSTomlResource(rawURL)
		}
	}

	if isGitRepoPath(path) {
		return getGitTomlResource(path)
	}

	if u.Scheme == "https" {
		return getHTTPSTomlResource(path)
	}

	if u.Scheme == "git" {
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

	// Check for full GitHub URLs
	if strings.HasPrefix(path, "https://github.com/") {
		return true
	}

	// Check for owner/repo patterns (but not if it's a file path with extension at the end)
	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		// Simple owner/repo format without file extension at the end
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.Contains(parts[1], ".") {
			return true
		}
		// Extended repo path
		if len(parts) >= 4 && parts[0] != "" && parts[1] != "" {
			return true
		}
	}

	return false
}

func isGitHubBlobURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}

	if u.Host != "github.com" {
		return false
	}

	pathParts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	return len(pathParts) >= 5 && pathParts[2] == "blob" && strings.HasSuffix(urlStr, ".toml")
}

func convertGitHubBlobToRaw(urlStr string) string {
	u, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	pathParts := strings.Split(strings.TrimPrefix(u.Path, "/"), "/")
	if len(pathParts) < 5 || pathParts[2] != "blob" {
		return ""
	}

	owner := pathParts[0]
	repo := pathParts[1]
	commit := pathParts[3]
	filePath := strings.Join(pathParts[4:], "/")

	return fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s", owner, repo, commit, filePath)
}

func getGitTomlResource(path string) (string, error) {
	gitPath := parseGitPath(path)
	if gitPath == "" {
		return "", errors.New("invalid git repository path")
	}

	// For simple owner/repo paths, the git ls-remote check is done in parseGitPath
	// Skip the check here if gitPath already includes the file path
	if !strings.Contains(gitPath, "/") || len(strings.Split(gitPath, "/")) <= 2 {
		cmd := exec.Command("git", "ls-remote", "--exit-code", "https://github.com/"+strings.Join(strings.Split(gitPath, "/")[:2], "/"))
		_, err := cmd.Output()
		if err != nil {
			return "", errors.New("git repository not found or not accessible")
		}
	}

	rawURL := "https://raw.githubusercontent.com/" + gitPath
	return getHTTPSTomlResource(rawURL)
}

func getDefaultBranch(owner, repo string) string {
	// Try common branch names in order
	branches := []string{"main", "master", "develop", "dev"}

	for _, branch := range branches {
		cmd := exec.Command("git", "ls-remote", fmt.Sprintf("https://github.com/%s/%s", owner, repo), branch)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return branch
		}
	}

	// Fallback: try to get the default branch from GitHub API (without auth)
	resp, err := http.Get(fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo))
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
			// Simple parsing - look for "default_branch" in JSON response
			body, err := io.ReadAll(resp.Body)
			if err == nil {
				bodyStr := string(body)
				if idx := strings.Index(bodyStr, `"default_branch":"`); idx != -1 {
					start := idx + len(`"default_branch":"`)
					end := strings.Index(bodyStr[start:], `"`)
					if end != -1 {
						return bodyStr[start : start+end]
					}
				}
			}
		}
	}

	return ""
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
		defaultBranch := getDefaultBranch(owner, repo)
		if defaultBranch == "" {
			defaultBranch = "main"
		}

		commonFiles := []string{"fjrd.config.toml", "fjrd.toml", "config.toml"}
		searchPaths := []string{"", "fjrd/", ".fjrd/", "config/"}

		for _, searchPath := range searchPaths {
			for _, file := range commonFiles {
				testURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s%s", owner, repo, defaultBranch, searchPath, file)
				resp, err := http.Head(testURL)
				if err == nil && resp.StatusCode == 200 {
					resp.Body.Close()
					return fmt.Sprintf("%s/%s/%s/%s%s", owner, repo, defaultBranch, searchPath, file)
				}
				if resp != nil {
					resp.Body.Close()
				}
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

func ParseConfig(content string, cfg *FjrdConfig) error {
	err := goToml.Unmarshal([]byte(content), &cfg)
	if err != nil {
		return err
	}
	if err = cfg.Validate(); err != nil {
		return err
	}
	return nil
}
