package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/apex/log"

	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	Token string
	URL   string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-handler-betteruptime",
			Short:    "Sensu Handler for Better Uptime",
			Keyspace: "sensu.io/plugins/sensu-handler-betteruptime/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		&sensu.PluginConfigOption{
			Path:      "apitoken",
			Env:       "BETTER_UPTIME_TOKEN",
			Argument:  "token",
			Shorthand: "t",
			Default:   "",
			Usage:     "API Token for BetterUptime",
			Value:     &plugin.Token,
		},
		&sensu.PluginConfigOption{
			Path:      "url",
			Env:       "BETTER_UPTIME_URL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "https://betteruptime.com/api/v2/incidents",
			Usage:     "Incident API URL",
			Value:     &plugin.URL,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(plugin.Token) == 0 {
		return fmt.Errorf("--token or BETTER_UPTIME_TOKEN environment variable is required")
	}
	return nil
}

func executeHandler(event *types.Event) error {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "executeHandler",
	})

	var event_name string
	var summary string
	var description string
	var email string

	ctx.Warnf("executing handler with --token %s", plugin.Token)

	if len(event.Entity.Annotations["betteruptime/config/name"]) > 0 {
		event_name = event.Entity.Annotations["betteruptime/config/name"]
	} else if len(event.Check.Annotations["betteruptime/config/name"]) > 0 {
		event_name = event.Check.Annotations["betteruptime/config/name"]
	} else {
		event_name = event.Entity.Name
	}

	if len(event.Check.Annotations["betteruptime/config/summary"]) > 0 {
		summary = event.Check.Annotations["betteruptime/config/summary"]
	} else {
		summary = event.Check.Name
	}

	if len(event.Check.Annotations["betteruptime/config/description"]) > 0 {
		description = event.Check.Annotations["betteruptime/config/description"]
	} else {
		description = fmt.Sprintf("Trouble with %s for entity %s", event.Check.Name, event.Entity.Name)
	}

	if len(event.Check.Annotations["betteruptime/config/email"]) > 0 {
		email = event.Check.Annotations["betteruptime/config/email"]
	} else {
		email = "my-great-email@example.com"
	}

	err := createIncident(plugin.URL, plugin.Token, event_name, email, summary, description, true, true)
	if err != nil {
		return err
	}
	return nil
}

func createIncident(urlIncident string, token string, name string, email string, summary string, description string, sendEmail bool, sendPush bool) error {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "createIncident",
	})

	var method string = "POST"
	var incident Incident
	var incidentReply ServerIncident

	incident.Name = name
	incident.Summary = summary
	incident.Description = description
	incident.Requester_email = email
	incident.Email = sendEmail
	incident.Push = sendPush
	incident.Call = false
	incident.TeamWait = 1

	incidentJSON, err := json.Marshal(incident)
	if err != nil {
		return err
	}

	reqURL, err := url.Parse(urlIncident)
	if err != nil {
		return err
	}

	client := &http.Client{}

	var resp *http.Response
	req, err := http.NewRequest(method, reqURL.String(), bytes.NewBuffer(incidentJSON))
	if err != nil {
		return err
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	resp, err = client.Do(req)

	if err != nil {
		return err
	}
	ctx.Warnf("Incident creation return code: %d", resp.StatusCode)
	if resp.StatusCode != 201 {
		return fmt.Errorf("incident not created. Error code %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err = json.Unmarshal(body, &incidentReply)
	if err != nil {
		return err
	}

	ctx.Infof("incident %d created", incidentReply.Data.Id)

	return nil
}

type Incident struct {
	Requester_email string `json:"requester_email" yaml:"requester_email"`
	Name            string `json:"name" yaml:"name"`
	Summary         string `json:"summary" yaml:"summary"`
	Description     string `json:"description" yaml:"description"`
	Call            bool   `json:"call" yaml:"call"`
	Email           bool   `json:"email" yaml:"email"`
	Push            bool   `json:"push" yaml:"push"`
	TeamWait        int    `json:"team_wait" yaml:"team_wait"`
}

type ServerIncidentDataAttributes struct {
	Name                 string     `json:"name" yaml:"name"`
	Url                  *string    `json:"url" yaml:"url"`
	Http_method          *string    `json:"http_method" yaml:"http_method"`
	Cause                string     `json:"cause" yaml:"cause"`
	Incident_group_id    *int       `json:"incident_group_id" yaml:"incident_group_id"`
	Started_at           time.Time  `json:"started_at" yaml:"started_at"`
	Acknowledged_at      *time.Time `json:"acknowledged_at" yaml:"acknowledged_at"`
	Acknowledged_by      *string    `json:"acknowledged_by" yaml:"acknowledged_by"`
	Resolved_at          *time.Time `json:"resolved_at" yaml:"resolved_at"`
	Resolved_by          *string    `json:"resolved_by" yaml:"resolved_by"`
	Response_content     *string    `json:"response_content" yaml:"response_content"`
	Response_options     *string    `json:"response_options" yaml:"response_options"`
	Regions              *string    `json:"regions" yaml:"regions"`
	Response_url         *string    `json:"response_url" yaml:"response_url"`
	Screenshot_url       *string    `json:"screenshot_url" yaml:"screenshot_url"`
	Escalation_policy_id *string    `json:"escalation_policy_id" yaml:"escalation_policy_id"`
	Call                 bool       `json:"call" yaml:"call"`
	Sms                  bool       `json:"sms" yaml:"sms"`
	Email                bool       `json:"email" yaml:"email"`
	Push                 bool       `json:"push" yaml:"push"`
}

type ServerIncidentData struct {
	Id         int                          `json:"id" yaml:"id"`
	Type       string                       `json:"type" yaml:"type"`
	Attributes ServerIncidentDataAttributes `json:"attributes" yaml:"attributes"`
}

type ServerIncident struct {
	Data ServerIncidentData `json:"data" yaml:"data"`
}
