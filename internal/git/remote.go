package git

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

type RepoInfo struct {
	Workspace string
	RepoSlug  string
}

var (
	sshPattern   = regexp.MustCompile(`git@bitbucket\.org:([^/]+)/([^/]+?)(?:\.git)?$`)
	httpsPattern = regexp.MustCompile(`https://[^@]*@?bitbucket\.org/([^/]+)/([^/]+?)(?:\.git)?$`)
)

func DetectRepo() (*RepoInfo, error) {
	url, err := getRemoteURL("origin")
	if err != nil {
		return nil, fmt.Errorf("infra: failed to get remote URL: %w", err)
	}
	return ParseRemoteURL(url)
}

func ParseRemoteURL(url string) (*RepoInfo, error) {
	url = strings.TrimSpace(url)

	if matches := sshPattern.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{
			Workspace: matches[1],
			RepoSlug:  matches[2],
		}, nil
	}

	if matches := httpsPattern.FindStringSubmatch(url); matches != nil {
		return &RepoInfo{
			Workspace: matches[1],
			RepoSlug:  matches[2],
		}, nil
	}

	return nil, fmt.Errorf("domain: could not parse Bitbucket remote URL: %s", url)
}

func getRemoteURL(name string) (string, error) {
	cmd := exec.Command("git", "remote", "get-url", name)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}
