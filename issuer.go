package prod_error_2_github_issue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"log"
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

type Service struct {
	ServiceName string `json:"serviceName"`
	Repo        string `json:"repo"`
}

type Issuer struct {
	GithubClient   *github.Client
	GithubIssue    *github.Issue
	GithubOwner    string
	ProdError      ProductionError
	ServiceList    []Service
	ActualRepo     string
	ProductionType string
}

func init() {
	_, isFound := os.LookupEnv("GITHUB_TOKEN")
	if isFound == false {
		log.Fatalln(errors.New("couldn't find a token"))
	}
	_, isFound = os.LookupEnv("GITHUB_OWNER")
	if isFound == false {
		log.Fatalln(errors.New("couldn't find owner"))
	}

	reposList := make([]Service, 0)
	err := json.Unmarshal([]byte(os.Getenv("GITHUB_SERVICES")), &reposList)
	if err != nil || len(reposList) == 0 {
		log.Fatalln(errors.New("could not parse GITHUB_SERVICES or count of services = 0. Check provided value"))
	}
}

func CreateGithubIssue(ctx context.Context, m PubSubMessage) error {
	issuer := Issuer{}
	issuer.initGithubClient(ctx)
	issuer.buildIssueFromErrorMessage(m)
	ok := issuer.findActualRepo()
	if !ok {
		log.Fatalln(`issuerError: Possible reasons:  
						- Repository was not found. Please check map, contains repos names
						- Couldn't parse PubSub message successfully'`)
	}

	existingIssue, err := issuer.getExistingIssue(ctx)
	if err != nil {
		return err
	}
	if existingIssue != nil {
		return issuer.updateExistingIssue(ctx, existingIssue)
	}

	return issuer.publishNewIssue(ctx)
}

func (i *Issuer) initGithubClient(ctx context.Context) {
	tc := oauth2.NewClient(ctx, oauth2.StaticTokenSource(&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")}))
	i.GithubClient = github.NewClient(tc)
	i.GithubOwner = os.Getenv("GITHUB_OWNER")
	i.ProductionType = os.Getenv("ENV_TYPE")
	i.ServiceList = make([]Service, 0)
	err := json.Unmarshal([]byte(os.Getenv("GITHUB_SERVICES")), &i.ServiceList)
	if err != nil {
		log.Fatalln(err)
	}
	//Check for correct GITHUB_OWNER and GITHUB_TOKEN
	_, _, err = i.GithubClient.Repositories.List(ctx, i.GithubOwner, nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func (i *Issuer) buildIssueFromErrorMessage(m PubSubMessage) {
	var issueTitle string
	var issueBody string
	var newError ProductionError
	err := json.Unmarshal(m.Data, &newError)
	if err == nil {
		issueTitle = fmt.Sprintf("%s err: %s", i.ProductionType, newError.JsonPayload.Error)
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

	i.ProdError = newError
	i.GithubIssue = &github.Issue{Title: &issueTitle, Body: &issueBody}
}

func (i *Issuer) findActualRepo() bool {
	for _, s := range i.ServiceList {
		if s.ServiceName == i.ProdError.Resource.Labels.ServiceName {
			i.ActualRepo = s.Repo
			return true
		}
	}
	return false
}

func (i *Issuer) getExistingIssue(ctx context.Context) (*github.Issue, error) {
	issues, _, err := i.GithubClient.Issues.ListByRepo(ctx, i.GithubOwner, i.ActualRepo, nil)
	if err != nil {
		return nil, err
	}

	if len(issues) == 0 {
		return nil, nil
	}

	re, _ := regexp.Compile(fmt.Sprintf(`%s \(\d*\)$`, *i.GithubIssue.Title))
	for _, issue := range issues {
		if re.MatchString(*issue.Title) {
			return issue, nil
		}
	}
	return nil, err
}

func (i *Issuer) updateExistingIssue(ctx context.Context, existingIssue *github.Issue) error {
	newTitle, err := incrementCounter(*existingIssue.Title)
	if err != nil {
		return err
	}

	_, _, err = i.GithubClient.Issues.Edit(ctx, i.GithubOwner, i.ActualRepo, *existingIssue.Number,
		&github.IssueRequest{
			Title: &newTitle,
			Body:  existingIssue.Body,
		})
	return err
}

func (i *Issuer) publishNewIssue(ctx context.Context) error {
	// add error counter to the title
	*i.GithubIssue.Title = *i.GithubIssue.Title + ` (1)`
	newIssue := github.IssueRequest{
		Title: i.GithubIssue.Title,
		Body:  i.GithubIssue.Body,
	}
	_, _, err := i.GithubClient.Issues.Create(ctx, i.GithubOwner, i.ActualRepo, &newIssue)
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
