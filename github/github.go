package github

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/go-github/v37/github"
	"github.com/xabi/dependabot-syncher/dependabot"
	"golang.org/x/oauth2"
)

// TODO: Manage Github rate limit

type Github struct {
	cli   *github.Client
	owner string
}

func New(ctx context.Context, owner, token string) Github {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)

	return Github{github.NewClient(tc), owner}
}

func (g Github) ListAllOwnRepos(ctx context.Context) ([]*github.Repository, error) {
	opt := &github.RepositoryListOptions{
		Type: "sources",
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := g.cli.Repositories.List(ctx, g.owner, opt)
		if err != nil {
			return nil, err
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

func (g Github) SearchForFile(ctx context.Context, dep dependabot.DepFile) ([]*github.CodeResult, error) {
	queryParams := []string{fmt.Sprintf("user:%s", g.owner)}

	for _, f := range dep.FilesPattern {
		queryParams = append(queryParams, fmt.Sprintf("filename:%s", f))
	}
	if dep.Path != "" {
		queryParams = append(queryParams, fmt.Sprintf(" path:%s", dep.Path))
	}

	opt := &github.SearchOptions{}

	var allMatches []*github.CodeResult
	for {
		result, resp, err := g.cli.Search.Code(ctx, strings.Join(queryParams, " "), opt)
		if waitForRateLimit(err) {
			continue
		}
		if err != nil {
			return nil, err
		}
		allMatches = append(allMatches, result.CodeResults...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allMatches, nil
}

func (g Github) UpdateFile(ctx context.Context, commit, repo, pathFile string, file []byte) (err error) {
	defer func() {
		if waitForRateLimit(err) {
			err = g.UpdateFile(ctx, commit, repo, pathFile, file)
		}
	}()

	var sha *string
	meta, _, resp, err := g.cli.Repositories.GetContents(ctx, g.owner, repo, pathFile, nil)
	if err != nil && resp.StatusCode != http.StatusNotFound {
		return err
	} else if resp.StatusCode != http.StatusNotFound {
		if base64.StdEncoding.EncodeToString(file) == strings.ReplaceAll(*meta.Content, "\n", "") {
			// The version of the repo is the same, do not update
			return nil
		}
		sha = meta.SHA
	}

	_, _, err = g.cli.Repositories.UpdateFile(ctx, g.owner, repo, pathFile, &github.RepositoryContentFileOptions{
		Message: &commit,
		SHA:     sha,
		Content: file,
	})
	if err, ok := err.(*github.ErrorResponse); ok && err.Response.StatusCode == http.StatusForbidden {
		return nil
	}

	return err
}

// waitForRateLimit is a shitty solution to wait for rate limit
func waitForRateLimit(err error) bool {
	if _, ok := err.(*github.RateLimitError); !ok {
		return false
	}

	fmt.Println("Waiting 1 minute for rate limit")
	time.Sleep(time.Minute)

	return true
}
