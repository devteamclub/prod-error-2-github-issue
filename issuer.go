package prod_error_2_github_issue

import (
	"context"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
	"os"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func CreateGithubIssue(ctx context.Context, m PubSubMessage) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)
	data := string(m.Data)

	issueTitle := "Staging error"
	if len(data) > 40 {
		issueTitle = "Staging error: " + data[:40]
	}
	issueBody := data
	newIssue := github.IssueRequest{Title: &issueTitle, Body: &issueBody}
	_, _, err := githubClient.Issues.Create(ctx, os.Getenv("github_user"), os.Getenv("github_repo"), &newIssue)
	log.Println(err)
}
