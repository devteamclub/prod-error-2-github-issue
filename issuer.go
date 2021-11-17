package prod_error_2_github_issue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"regexp"
	"strconv"
	"time"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}
type ProductionError struct {
	InsertId    string `json:"insertId"`
	JsonPayload struct {
		Error  string                 `json:"Error"`
		Locals map[string]interface{} `json:"Locals"`
		Stack  string                 `json:"Stack"`
	} `json:"jsonPayload"`
	LogName          string    `json:"logName"`
	ReceiveTimestamp time.Time `json:"receiveTimestamp"`
	Resource         struct {
		Labels struct {
			ConfigurationName string `json:"configuration_name"`
			Location          string `json:"location"`
			ProjectId         string `json:"project_id"`
			RevisionName      string `json:"revision_name"`
			ServiceName       string `json:"service_name"`
		} `json:"labels"`
		Type string `json:"type"`
	} `json:"resource"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

func init() {
	_, isFound := os.LookupEnv("GITHUB_TOKEN")
	if isFound == false {
		panic(errors.New("couldn't find a token"))
	}
	_, isFound = os.LookupEnv("GITHUB_USER")
	if isFound == false {
		panic(errors.New("couldn't find a username"))
	}
	_, isFound = os.LookupEnv("GITHUB_REPO")
	if isFound == false {
		panic(errors.New("couldn't find a repository name"))
	}
}

func CreateGithubIssue(ctx context.Context, m PubSubMessage) error {
	client := initGithubClient(ctx)
	err, issue := parseErrorMessage(m)

	var existingIssue *github.Issue
	if err == nil {
		existingIssue = checkIssueExists(ctx, *client, *issue.Title)
	}

	if existingIssue != nil {
		err := updateIssue(ctx, *client, existingIssue)
		if err != nil {
			panic(err)
		}
		return err
	}

	err = publishIssue(ctx, *client, issue)
	if err != nil {
		panic(err)
	}
	return nil
}

func initGithubClient(ctx context.Context) *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)
	return githubClient
}

func parseErrorMessage(m PubSubMessage) (error, *github.Issue) {
	var issueTitle string
	var issueBody string
	var newError ProductionError
	err := json.Unmarshal(m.Data, &newError)
	if err == nil {
		issueTitle = fmt.Sprintf("Prod err: %s", newError.JsonPayload.Error)
		beautifiedLocals, err := json.MarshalIndent(newError.JsonPayload.Locals, "", " ")
		if err == nil {
			issueBody = fmt.Sprintf("## Stack:\n```%s```\n## Locals:\n```%s\n```",
				newError.JsonPayload.Stack, string(beautifiedLocals))
		} else {
			issueBody = fmt.Sprintf("## Stack:\n```%s```\n## Locals:\n(Couldn't jsonify properly...)\n```%s\n```",
				newError.JsonPayload.Stack, newError.JsonPayload.Locals)
		}
	} else {
		issueTitle = "Production error"
		issueBody = string(m.Data)
	}

	var issue github.Issue
	issue.Title = &issueTitle
	issue.Body = &issueBody
	return err, &issue
}

func checkIssueExists(ctx context.Context, client github.Client, issueTitle string) *github.Issue {
	issues, _, err := client.Issues.ListByRepo(ctx, os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_REPO"), nil)
	if err != nil {
		return nil
	}

	re, _ := regexp.Compile(fmt.Sprintf(`%s \(\d*\)`, issueTitle))
	for _, issue := range issues {
		if re.MatchString(*issue.Title) {
			return issue
		}
	}
	return nil
}

func updateIssue(ctx context.Context, client github.Client, issue *github.Issue) error {
	newTitle, err := incrementCounter(*issue.Title)
	if err != nil {
		return err
	}

	_, _, err = client.Issues.Edit(ctx, os.Getenv("GITHUB_USER"),
		os.Getenv("GITHUB_REPO"), *issue.Number, &github.IssueRequest{Title: &newTitle, Body: issue.Body})
	if err != nil {
		return err
	}
	return nil
}

func publishIssue(ctx context.Context, client github.Client, issue *github.Issue) error {
	// add error counter to the title
	*issue.Title = *issue.Title + ` (1)`
	newIssue := github.IssueRequest{Title: issue.Title, Body: issue.Body}
	_, _, err := client.Issues.Create(ctx, os.Getenv("GITHUB_USER"), os.Getenv("GITHUB_REPO"), &newIssue)
	return err
}

func incrementCounter(t string) (string, error) {
	re, _ := regexp.Compile(`\(\d*\)$`)
	res := re.FindStringSubmatchIndex(t)
	counter, err := strconv.Atoi(t[res[0]+1 : res[1]-1])
	if err != nil {
		return "", err
	}
	counter = counter + 1
	newTitle := t[:res[0]] + "(" + strconv.Itoa(counter) + ")"
	return newTitle, nil
}
