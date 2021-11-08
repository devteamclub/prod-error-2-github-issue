package prod_error_2_github_issue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
	"os"
	"time"
)

type PubSubMessage struct {
	Data []byte `json:"data"`
}
type ProductionError struct {
	InsertId    string `json:"insertId"`
	JsonPayload struct {
		Error  string `json:"Error"`
		Locals struct {
			CurrentBattleId int         `json:"current-battle-id"`
			CurrentCrew     interface{} `json:"current-crew"`
			CurrentEvent    struct {
				Admins          interface{} `json:"admins"`
				Battles         interface{} `json:"battles"`
				ContractStatus  string      `json:"contractStatus"`
				DateEnd         int         `json:"dateEnd"`
				DateStart       int         `json:"dateStart"`
				Description     string      `json:"description"`
				Djs             interface{} `json:"djs"`
				Feature         bool        `json:"feature"`
				Hosts           interface{} `json:"hosts"`
				Id              int         `json:"id"`
				Image           string      `json:"image"`
				Judges          interface{} `json:"judges"`
				LocationAddress string      `json:"locationAddress"`
				LocationLat     int         `json:"locationLat"`
				LocationLng     int         `json:"locationLng"`
				OrgId           int         `json:"orgId"`
				Organizer       struct {
					Avatar     string `json:"avatar"`
					DancerName string `json:"dancerName"`
					Id         int    `json:"id"`
				} `json:"organizer"`
				PaymentMethod struct {
				} `json:"paymentMethod"`
				Period string `json:"period"`
				Status string `json:"status"`
				Tiers  []struct {
					Cost        int    `json:"cost"`
					Description string `json:"description"`
					Id          int    `json:"id"`
					Title       string `json:"title"`
				} `json:"tiers"`
				Title   string `json:"title"`
				Website string `json:"website"`
			} `json:"current-event"`
			CurrentEventId     int         `json:"current-event-id"`
			CurrentOrg         interface{} `json:"current-org"`
			CurrentPermissions struct {
				AgreementSigned           int  `json:"agreementSigned"`
				IsPayCheckboxChecked      bool `json:"isPayCheckboxChecked"`
				IsShowCheckInButton       bool `json:"isShowCheckInButton"`
				IsShowCheckInOnDayButton  bool `json:"isShowCheckInOnDayButton"`
				IsShowCheckInPeopleButton bool `json:"isShowCheckInPeopleButton"`
				IsShowEditButton          bool `json:"isShowEditButton"`
				IsShowPayCheckbox         bool `json:"isShowPayCheckbox"`
				IsShowPurchaseButton      bool `json:"isShowPurchaseButton"`
				IsShowViewTicketButton    bool `json:"isShowViewTicketButton"`
				IsUserAdmin               bool `json:"isUserAdmin"`
				IsUserCheckedIn           bool `json:"isUserCheckedIn"`
				IsUserOrgAdmin            bool `json:"isUserOrgAdmin"`
				WaiverSigned              int  `json:"waiverSigned"`
			} `json:"current-permissions"`
			CurrentUser struct {
				Active     bool   `json:"active"`
				Avatar     string `json:"avatar"`
				Birthday   string `json:"birthday"`
				Card       string `json:"card"`
				City       string `json:"city"`
				Country    string `json:"country"`
				DancerName string `json:"dancerName"`
				Email      string `json:"email"`
				FullName   string `json:"fullName"`
				Id         int    `json:"id"`
				Provider   string `json:"provider"`
				Socials    struct {
				} `json:"socials"`
				State string `json:"state"`
				Uid   string `json:"uid"`
			} `json:"current-user"`
		} `json:"Locals"`
		Stack string `json:"Stack"`
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

func CreateGithubIssue(ctx context.Context, m PubSubMessage) error {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("github_token")},
	)
	tc := oauth2.NewClient(ctx, ts)
	githubClient := github.NewClient(tc)

	var issueTitle string
	var issueBody string
	var newError ProductionError
	err := json.Unmarshal(m.Data, &newError)
	if err != nil {
		issueTitle = "Production error"
		issueBody = string(m.Data)
	} else {
		issueTitle = fmt.Sprintf("Prod: %s", newError.JsonPayload.Error)
		issueBody = fmt.Sprintf("Stack:\n%s\nLocals:\n%s", newError.JsonPayload.Stack, newError.JsonPayload.Locals)
	}

	newIssue := github.IssueRequest{Title: &issueTitle, Body: &issueBody}
	_, _, err = githubClient.Issues.Create(ctx, os.Getenv("github_user"), os.Getenv("github_repo"), &newIssue)

	if err != nil {
		panic(err)
	}
	return err
}
