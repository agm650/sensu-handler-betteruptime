package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/apex/log"
	jsoniter "github.com/json-iterator/go"

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

	options = []sensu.ConfigOption{
		&sensu.PluginConfigOption[string]{
			Path:      "apitoken",
			Env:       "BETTER_UPTIME_TOKEN",
			Argument:  "token",
			Shorthand: "t",
			Default:   "",
			Usage:     "API Token for BetterUptime",
			Value:     &plugin.Token,
		},
		&sensu.PluginConfigOption[string]{
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

	var incident Incident
	incident.Email = true
	incident.Push = true
	incident.Call = false
	incident.TeamWait = 1

	ctx.Warnf("executing handler with --token %s", plugin.Token)

	if len(event.Entity.Annotations["betteruptime/config/name"]) > 0 {
		incident.Name = event.Entity.Annotations["betteruptime/config/name"]
	} else if len(event.Check.Annotations["betteruptime/config/name"]) > 0 {
		incident.Name = event.Check.Annotations["betteruptime/config/name"]
	} else {
		incident.Name = event.Entity.Name + " - " + event.Check.Name
	}

	if len(event.Check.Annotations["betteruptime/config/summary"]) > 0 {
		incident.Summary = event.Check.Annotations["betteruptime/config/summary"]
	} else {
		incident.Summary = event.Check.Name
	}

	if len(event.Check.Annotations["betteruptime/config/description"]) > 0 {
		incident.Description = event.Check.Annotations["betteruptime/config/description"]
	} else {
		incident.Description = event.Check.Output
	}

	if len(event.Check.Annotations["betteruptime/config/email"]) > 0 {
		incident.Requester_email = event.Check.Annotations["betteruptime/config/email"]
	} else {
		incident.Requester_email = "my-great-email@example.com"
	}

	if len(event.Check.Annotations["betteruptime/config/sendpush"]) > 0 {
		incident.Push = true
	} else {
		incident.Push = false
	}

	if len(event.Check.Annotations["betteruptime/config/sendmail"]) > 0 {
		incident.Email = true
	} else {
		incident.Email = false
	}

	if len(event.Check.Annotations["betteruptime/config/call"]) > 0 {
		incident.Call = true
	} else {
		incident.Call = false
	}

	if len(event.Check.Annotations["betteruptime/config/teamwait"]) > 0 {
		val, err := strconv.Atoi(event.Check.Annotations["betteruptime/config/teamwait"])
		if err != nil {
			incident.TeamWait = 0
		} else {
			incident.TeamWait = val
		}
	} else {
		incident.TeamWait = 0
	}

	if event.Check.Status == 2 {
		err := createIncident(plugin.URL, plugin.Token, incident)
		if err != nil {
			return err
		}
	} else {
		err := resolveIncident(plugin.URL, plugin.Token, incident)
		if err != nil {
			return err
		}
	}

	return nil
}

func createIncident(urlIncident string, token string, incident Incident) error {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "createIncident",
	})

	var method string = "POST"
	var incidentReply ServerIncident

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

	body, err := io.ReadAll(resp.Body)
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

func resolveIncident(urlIncident string, token string, incident Incident) error {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "resolveIncident",
	})

	var method string = "POST"
	var incidentReply ServerIncident

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

	body, err := io.ReadAll(resp.Body)
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

//lint:ignore U1000 Ignore unused function temporarily for debugging
func getIncidents(urlIncident string, token string) ([]Incident, error) {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "getIncidents",
	})

	var method string = "POST"

	reqURL, err := url.Parse(urlIncident)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}

	var resp *http.Response
	req, err := http.NewRequest(method, reqURL.String(), nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	resp, err = client.Do(req)
	if err != nil {
		// fmt.Println("Error with the API Request")
		return nil, err
	}

	rawData, err := io.ReadAll(resp.Body)
	if err != nil {
		// fmt.Println("Error with the API Request")
		return nil, err
	}
	incidents, err := ExtractIncidents(rawData)
	if err != nil {
		return nil, err
	}
	ctx.Infof("retrived %d incidents", len(incidents))

	return incidents, nil
}

func ExtractIncidents(data []byte) ([]Incident, error) {
	ctx := log.WithFields(log.Fields{
		"file":     "main.go",
		"function": "ExtractIncidents",
	})

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	var incidents []Incident

	err := json.Unmarshal(data, &incidents)
	if err != nil {
		return nil, err
	}

	ctx.Warnf("Extrating %d events", len(incidents))

	return incidents, nil
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
