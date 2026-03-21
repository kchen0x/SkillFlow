package update

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	platformgit "github.com/shinerio/skillflow/core/platform/git"
	"github.com/shinerio/skillflow/core/shared/logicalkey"
	skilldomain "github.com/shinerio/skillflow/core/skillcatalog/domain"
)

type CheckResult struct {
	SkillID   string
	HasUpdate bool
	LatestSHA string
}

type Checker struct {
	baseURL string
	client  *http.Client
}

func NewChecker(baseURL string, client *http.Client) *Checker {
	if baseURL == "" {
		baseURL = "https://api.github.com"
	}
	if client == nil {
		client = http.DefaultClient
	}
	return &Checker{baseURL: baseURL, client: client}
}

func (c *Checker) Check(ctx context.Context, sk *skilldomain.InstalledSkill) (CheckResult, error) {
	if !sk.IsGitHub() {
		return CheckResult{}, nil
	}
	owner, repo, subPath := parseSourceURL(sk.SourceURL, sk.SourceSubPath)
	reqURL := fmt.Sprintf("%s/repos/%s/%s/commits", c.baseURL, owner, repo)
	query := url.Values{}
	query.Set("per_page", "1")
	if normalized := logicalkey.NormalizeRepoSubPath(subPath); normalized != "" && normalized != "." {
		query.Set("path", normalized)
	}
	req, _ := http.NewRequestWithContext(ctx, "GET", reqURL+"?"+query.Encode(), nil)
	resp, err := c.client.Do(req)
	if err != nil {
		return CheckResult{}, err
	}
	defer resp.Body.Close()

	var commits []struct {
		SHA string `json:"sha"`
	}
	if err := decodeGitHubJSONResponse(resp, &commits); err != nil {
		return CheckResult{}, err
	}
	if len(commits) == 0 {
		return CheckResult{}, fmt.Errorf("github returned no commits for %s/%s path=%s", owner, repo, subPath)
	}
	latestSHA := commits[0].SHA
	return CheckResult{
		SkillID:   sk.ID,
		LatestSHA: latestSHA,
		HasUpdate: latestSHA != sk.SourceSHA,
	}, nil
}

func parseSourceURL(sourceURL, subPath string) (owner, repo, path string) {
	name, err := platformgit.ParseRepoName(sourceURL)
	if err != nil {
		return "", "", subPath
	}
	parts := strings.SplitN(name, "/", 2)
	if len(parts) != 2 {
		return "", "", subPath
	}
	owner = parts[0]
	repo = parts[1]
	return owner, repo, subPath
}

func decodeGitHubJSONResponse(resp *http.Response, target any) error {
	if resp.StatusCode != http.StatusOK {
		return githubStatusError(resp)
	}
	return json.NewDecoder(resp.Body).Decode(target)
}

func githubStatusError(resp *http.Response) error {
	var payload struct {
		Message string `json:"message"`
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8192))
	msg := strings.TrimSpace(string(body))
	if len(body) > 0 {
		if err := json.Unmarshal(body, &payload); err == nil && strings.TrimSpace(payload.Message) != "" {
			msg = strings.TrimSpace(payload.Message)
		}
	}
	if msg == "" {
		return fmt.Errorf("github status %d", resp.StatusCode)
	}
	return fmt.Errorf("github status %d: %s", resp.StatusCode, msg)
}
