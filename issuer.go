package prod_error_2_github_issue

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}

func Init() {
	_, isFound := os.LookupEnv("github_token")
	if isFound == false {
		panic(errors.New("couldn't find a token"))
	}
	_, isFound = os.LookupEnv("github_user")
	if isFound == false {
		panic(errors.New("couldn't find a username"))
	}
	_, isFound = os.LookupEnv("github_repo")
	if isFound == false {
		panic(errors.New("couldn't find a repository name"))
	}
}

func CreateGithubIssue(ctx context.Context, m PubSubMessage) {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	errorMessage := string(m.Data)

	var issueTitle string
	if len(errorMessage) < 20 {
		issueTitle = "Production error"
	} else {
		issueTitle = fmt.Sprintf("%s: %s...", issueTitle, errorMessage[:20])
	}
	issueBody := errorMessage

	newIssue := github.IssueRequest{Title: &issueTitle, Body: &issueBody}
	_, _, err := githubClient.Issues.Create(ctx, os.Getenv("github_user"), os.Getenv("github_repo"), &newIssue)

	if err != nil {
		panic(err)
	}
}
