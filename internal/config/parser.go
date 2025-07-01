package config

import (
	"context"
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

func getNetworkTomlResource(ctx context.Context, path string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) (string, error) {
	u, err := url.Parse(path)
	if err != nil {
		return "", errors.New("failed to parse url")
	}

	if u.Scheme == "https" && isGitHubBlobURL(path) {
		log.Debug("Converting GitHub blob URL to raw", "original", path)
		rawURL := convertGitHubBlobToRaw(path)
		if rawURL != "" {
			log.Debug("Converted to raw URL", "raw_url", rawURL)
			return getHTTPSTomlResource(ctx, rawURL, log)
		}
	}

	if isGitRepoPath(path) {
		log.Debug("Processing as Git repository path", "path", path)
		return getGitTomlResource(ctx, path, log)
	}

	if u.Scheme == "https" {
		log.Debug("Processing as HTTPS URL", "url", path)
		return getHTTPSTomlResource(ctx, path, log)
	}

	if u.Scheme == "git" {
		log.Debug("Processing as Git URL", "url", path)
		return getGitTomlResource(ctx, path, log)
	}

	if u.Scheme != "" && u.Scheme != "https" && u.Scheme != "git" {
		return "", errors.New("cannot use non-secure protocol")
	}

	return "", errors.New("failed to retrieve the network resource contents")
}

func getHTTPSTomlResource(ctx context.Context, urlStr string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) (string, error) {
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

	log.Debug("Creating HTTP request", "url", urlStr)
	req, err := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	log.Debug("Sending HTTP request")
	resp, err := client.Do(req)
	if err != nil {
		return "", errors.New("failed to fetch network resource")
	}
	defer resp.Body.Close()

	log.Debug("Received HTTP response", "status", resp.StatusCode, "content_length", resp.ContentLength)
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

	if strings.HasPrefix(path, "https://github.com/") {
		return true
	}

	if strings.Contains(path, "/") {
		parts := strings.Split(path, "/")
		if len(parts) == 2 && parts[0] != "" && parts[1] != "" && !strings.Contains(parts[1], ".") {
			return true
		}
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

func getGitTomlResource(ctx context.Context, path string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) (string, error) {
	log.Debug("Parsing Git path", "path", path)
	gitPath := parseGitPath(ctx, path, log)
	if gitPath == "" {
		return "", errors.New("invalid git repository path")
	}

	log.Debug("Resolved Git path", "git_path", gitPath)

	if !strings.Contains(gitPath, "/") || len(strings.Split(gitPath, "/")) <= 2 {
		repoURL := "https://github.com/" + strings.Join(strings.Split(gitPath, "/")[:2], "/")
		log.Debug("Verifying repository exists", "repo_url", repoURL)
		cmd := exec.CommandContext(ctx, "git", "ls-remote", "--exit-code", repoURL)
		_, err := cmd.Output()
		if err != nil {
			return "", errors.New("git repository not found or not accessible")
		}
		log.Debug("Repository verified")
	}

	rawURL := "https://raw.githubusercontent.com/" + gitPath
	log.Debug("Fetching from raw GitHub URL", "raw_url", rawURL)
	return getHTTPSTomlResource(ctx, rawURL, log)
}

func getDefaultBranch(ctx context.Context, owner, repo string) string {
	branches := []string{"main", "master", "develop", "dev"}

	for _, branch := range branches {
		cmd := exec.CommandContext(ctx, "git", "ls-remote", fmt.Sprintf("https://github.com/%s/%s", owner, repo), branch)
		output, err := cmd.Output()
		if err == nil && len(output) > 0 {
			return branch
		}
	}

	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo), nil)
	if err != nil {
		return ""
	}

	resp, err := http.DefaultClient.Do(req)
	if err == nil {
		defer resp.Body.Close()
		if resp.StatusCode == 200 {
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

func parseGitPath(ctx context.Context, path string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) string {
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
		log.Debug("Searching for config file in repository", "owner", owner, "repo", repo)
		defaultBranch := getDefaultBranch(ctx, owner, repo)
		if defaultBranch == "" {
			defaultBranch = "main"
		}
		log.Debug("Using default branch", "branch", defaultBranch)

		commonFiles := []string{"fjrd.config.toml", "fjrd.toml", "config.toml"}
		searchPaths := []string{"", "fjrd/", ".fjrd/", "config/"}

		for _, searchPath := range searchPaths {
			for _, file := range commonFiles {
				testURL := fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/%s/%s%s", owner, repo, defaultBranch, searchPath, file)
				log.Debug("Testing config file location", "url", testURL)
				
				req, err := http.NewRequestWithContext(ctx, "HEAD", testURL, nil)
				if err != nil {
					continue
				}

				resp, err := http.DefaultClient.Do(req)
				if err == nil && resp.StatusCode == 200 {
					resp.Body.Close()
					foundPath := fmt.Sprintf("%s/%s/%s/%s%s", owner, repo, defaultBranch, searchPath, file)
					log.Info("Found config file", "path", foundPath)
					return foundPath
				}
				if resp != nil {
					resp.Body.Close()
				}
			}
		}
		log.Warn("No config file found in repository")
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

func LoadConfig(ctx context.Context, location string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) (*FjrdConfig, error) {
	log.Info("Loading configuration", "location", location)
	
	pathType := determinePathType(location)
	log.Debug("Determined path type", "type", pathType.String(), "location", location)
	
	content, err := resolveTomlResource(ctx, location, log)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve config location: %w", err)
	}

	log.Debug("Configuration content loaded", "size", len(content))

	var cfg FjrdConfig
	if err := parseConfig(content, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	log.Info("Configuration parsed successfully", "version", cfg.Version)
	return &cfg, nil
}

func resolveTomlResource(ctx context.Context, location string, log interface{ Info(string, ...any); Debug(string, ...any); Warn(string, ...any) }) (string, error) {
	pathType := determinePathType(location)
	switch pathType {
	case PathTypeLocal:
		log.Debug("Reading local file", "path", location)
		tomlBody, err := getLocalTomlFile(location)
		if err != nil {
			return "", errors.New("failed to read local toml file")
		}
		log.Debug("Local file read successfully", "size", len(tomlBody))
		return tomlBody, nil
	case PathTypeNetwork:
		log.Info("Fetching remote configuration", "url", location)
		tomlBody, err := getNetworkTomlResource(ctx, location, log)
		if err != nil {
			return "", errors.New("failed to read remote toml file")
		}
		log.Info("Remote configuration fetched successfully", "size", len(tomlBody))
		return tomlBody, nil
	case PathTypeNonExistent:
		return "", errors.New("failed to read location path")
	default:
		return "", errors.New("unknown location provided")
	}
}

func parseConfig(content string, cfg *FjrdConfig) error {
	err := goToml.Unmarshal([]byte(content), &cfg)
	if err != nil {
		return err
	}
	if err = cfg.Validate(); err != nil {
		return err
	}
	return nil
}