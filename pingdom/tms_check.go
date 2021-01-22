package pingdom

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"time"
)

// TmsCheckService provides an interface to Pingdom TMS checks.
type TmsCheckService struct {
	client *Client
}

// TmsCheck is an struct representing a TMS Check.
type TmsCheck struct {
	ID                       int       `json:"id"`
	Name                     string    `json:"name"`
	Steps                    []TmsStep `json:"steps"`
	Active                   bool      `json:"active,omitempty"`
	ContactIds               []int     `json:"contact_ids,omitempty"`
	CustomMessage            string    `json:"custom_message,omitempty"`
	IntegrationIds           []int     `json:"integration_ids,omitempty"`
	Interval                 int       `json:"interval,omitempty"`
	Region                   string    `json:"region,omitempty"`
	SendNotificationWhenDown int       `json:"send_notification_when_down,omitempty"`
	SeverityLevel            string    `json:"severity_level,omitempty"`
	Tags                     string    `json:"tags,omitempty"`
	TeamIds                  []int     `json:"team_ids,omitempty"`
}

type TmsStep struct {
	Function string            `json:"fn,omitempty"`
	Args     map[string]string `json:"args,omitempty"`
}

const (
	DESC Order      = "desc"
	ASC  Order      = "asc"
	HOUR Resolution = "hour"
	DAY  Resolution = "day"
	WEEK Resolution = "week"
)

type Order string
type Resolution string

type TmsStatusReportListByIdRequest struct {
	From  *time.Time
	To    *time.Time
	Order Order
}
type TmsStatusReportListRequest struct {
	From      *time.Time
	To        *time.Time
	Order     Order
	Limit     *int
	Offset    *int
	OmitEmpty bool
}
type TmsPerformanceReportRequest struct {
	From          *time.Time
	To            *time.Time
	Order         Order
	IncludeUptime bool
	Resolution    Resolution
}

func NewTmsCheck(name string, steps []TmsStep) *TmsCheck {
	check := &TmsCheck{
		Name:  name,
		Steps: steps,
	}

	// Defaults of the Pingdom TMS Check API
	check.Active = true
	check.Interval = 10
	check.Region = "us-east"
	check.SeverityLevel = "high"
	check.SendNotificationWhenDown = 1

	return check
}

// Valid determines whether the TmsCheck contains valid fields. This can be
// used to guard against sending illegal values to the Pingdom API.
func (ts *TmsCheck) Valid() error {
	if ts.Name == "" {
		return fmt.Errorf("Invalid value for `Name`.  Must contain non-empty string")
	}

	if ts.Region != "us-east" && ts.Region != "us-west" && ts.Region != "eu" && ts.Region != "au" {
		return fmt.Errorf("invalid value %v for `Region`, allowed values are [\"us-east\", \"us-west\", \"eu\", \"au\"]", ts.Region)
	}

	if ts.SeverityLevel != "high" && ts.SeverityLevel != "low" {
		return fmt.Errorf("invalid value %v for `SeverityLevel`, allowed values are [\"low\", \"high\"]", ts.SeverityLevel)
	}

	if ts.Interval != 5 && ts.Interval != 10 && ts.Interval != 20 && ts.Interval != 60 && ts.Interval != 720 && ts.Interval != 1440 {
		return fmt.Errorf("invalid value %v for `Interval`, allowed values are [5,10,20,60,720,1440]", ts.Interval)
	}

	return nil
}

// RenderForJSONAPI returns the JSON formatted version of this object that may be submitted to Pingdom
func (ts *TmsCheck) RenderForJSONAPI() string {
	tags := make([]string, 0)
	for _, t := range strings.Split(ts.Tags, ",") {
		tags = append(tags, strings.TrimSpace(t))
	}

	u := map[string]interface{}{
		"name":                        ts.Name,
		"steps":                       ts.Steps,
		"active":                      ts.Active,
		"contact_ids":                 ts.ContactIds,
		"custom_message":              ts.CustomMessage,
		"integration_ids":             ts.IntegrationIds,
		"interval":                    ts.Interval,
		"region":                      ts.Region,
		"severity_level":              ts.SeverityLevel,
		"send_notification_when_down": ts.SendNotificationWhenDown,
		"tags":                        tags,
		"team_ids":                    ts.TeamIds,
	}

	jsonBody, _ := json.Marshal(u)
	return string(jsonBody)
}

func (tr *TmsStatusReportListByIdRequest) Valid() error {
	if tr.To != nil && tr.From != nil && tr.To.Before(*tr.From) {
		return fmt.Errorf("from date should be earlier then to date")
	}

	switch tr.Order {
	case DESC, ASC, "":
	default:
		return fmt.Errorf("invalid order allowed values are: %s, %s", ASC, DESC)
	}

	return nil
}
func (tr *TmsStatusReportListByIdRequest) GetParams() map[string]string {
	m := map[string]string{}

	if tr.From != nil {
		m["from"] = tr.From.Format(time.RFC3339)
	}

	if tr.To != nil {
		m["to"] = tr.To.Format(time.RFC3339)
	}

	if len(tr.Order) > 0 {
		m["order"] = string(tr.Order)
	}

	return m
}
func (tr *TmsStatusReportListRequest) Valid() error {
	if tr.To != nil && tr.From != nil && tr.To.Before(*tr.From) {
		return fmt.Errorf("from date should be earlier then to date")
	}

	if tr.Offset != nil && *tr.Offset < 0 {
		return fmt.Errorf("offset should be greater equal 0")
	}
	if tr.Limit != nil && *tr.Limit <= 0 {
		return fmt.Errorf("limit should be greater 0")
	}

	switch tr.Order {
	case DESC, ASC, "":
	default:
		return fmt.Errorf("invalid order allowed values are: %s, %s", ASC, DESC)
	}
	return nil
}
func (tr *TmsStatusReportListRequest) GetParams() map[string]string {
	m := map[string]string{}

	if tr.From != nil {
		m["from"] = tr.From.Format(time.RFC3339)
	}

	if tr.To != nil {
		m["to"] = tr.To.Format(time.RFC3339)
	}

	if tr.Offset != nil {
		m["offset"] = strconv.Itoa(*tr.Offset)
	}

	if tr.Limit != nil {
		m["limit"] = strconv.Itoa(*tr.Limit)
	}

	if len(tr.Order) > 0 {
		m["order"] = string(tr.Order)
	}

	//default is false
	if tr.OmitEmpty {
		m["omit_empty"] = "true"
	}

	return m
}
func (tr *TmsPerformanceReportRequest) Valid() error {
	if tr.To != nil && tr.From != nil && tr.To.Before(*tr.From) {
		return fmt.Errorf("from date should be earlier then to date")
	}

	switch tr.Order {
	case DESC, ASC, "":
	default:
		return fmt.Errorf("invalid order allowed values are: %s, %s", ASC, DESC)
	}

	switch tr.Resolution {
	case HOUR, DAY, WEEK, "":
	default:
		return fmt.Errorf("invalid order allowed values are: %s, %s. %s", HOUR, DAY, WEEK)
	}
	return nil
}
func (tr *TmsPerformanceReportRequest) GetParams() map[string]string {
	m := map[string]string{}

	if tr.From != nil {
		m["from"] = tr.From.Format(time.RFC3339)
	}

	if tr.To != nil {
		m["to"] = tr.To.Format(time.RFC3339)
	}

	if len(tr.Order) > 0 {
		m["order"] = string(tr.Order)
	}

	//default is false
	if tr.IncludeUptime {
		m["include_uptime"] = "true"
	}

	if len(tr.Resolution) > 0 {
		m["resolution"] = string(tr.Resolution)
	}

	return m
}

// List returns a list of TMS checks from Pingdom.
func (cs *TmsCheckService) List(params ...map[string]string) ([]TmsCheck, error) {
	param := map[string]string{}
	if len(params) == 1 {
		param = params[0]
	}
	req, err := cs.client.NewRequest("GET", "/tms/check", param)
	if err != nil {
		return nil, err
	}

	resp, err := cs.client.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if err := validateResponse(resp); err != nil {
		return nil, err
	}

	bodyBytes, _ := ioutil.ReadAll(resp.Body)
	bodyString := string(bodyBytes)
	m := &listTmsChecksJSONResponse{}
	err = json.Unmarshal([]byte(bodyString), &m)

	checks := make([]TmsCheck, 0)

	for _, cr := range m.TmsChecks {
		checks = append(checks, *fromTmsCheckResponse(&cr))
	}

	return checks, nil
}

// Create a new TMS check.
func (cs *TmsCheckService) Create(check *TmsCheck) (*TmsCheck, error) {
	if err := check.Valid(); err != nil {
		return nil, err
	}

	req, err := cs.client.NewJSONRequest("POST", "/tms/check", check.RenderForJSONAPI())
	if err != nil {
		return nil, err
	}

	m := &tmsCheckDetailsJSONResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return fromTmsCheckResponse(m.TmsCheck), nil
}

// ReadCheck returns detailed information about a pingdom TMS check given its ID.
func (cs *TmsCheckService) Read(id int) (*TmsCheck, error) {
	req, err := cs.client.NewRequest("GET", "/tms/check/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}

	m := &tmsCheckDetailsJSONResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return fromTmsCheckResponse(m.TmsCheck), nil
}

func fromTmsCheckResponse(cr *TmsCheckResponse) *TmsCheck {
	check := &TmsCheck{
		ID:                       cr.ID,
		Name:                     cr.Name,
		Steps:                    cr.Steps,
		Active:                   cr.Active,
		ContactIds:               cr.ContactIds,
		CustomMessage:            cr.CustomMessage,
		IntegrationIds:           cr.IntegrationIds,
		Interval:                 cr.Interval,
		Region:                   cr.Region,
		SendNotificationWhenDown: cr.SendNotificationWhenDown,
		SeverityLevel:            cr.SeverityLevel,
		Tags:                     strings.Join(cr.Tags, ","),
		TeamIds:                  cr.TeamIds,
	}

	return check
}

// Update will update the TMS check represented by the given ID with the values
// in the given check.  You should submit the complete list of values in
// the given check parameter, not just those that have changed.
func (cs *TmsCheckService) Update(id int, tmsCheck *TmsCheck) (*TmsCheck, error) {
	if err := tmsCheck.Valid(); err != nil {
		return nil, err
	}

	req, err := cs.client.NewJSONRequest("PUT", "/tms/check/"+strconv.Itoa(id), tmsCheck.RenderForJSONAPI())
	if err != nil {
		return nil, err
	}

	m := &tmsCheckDetailsJSONResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}
	return fromTmsCheckResponse(m.TmsCheck), err
}

// Delete will delete the TMS check for the given ID.
func (cs *TmsCheckService) Delete(id int) (*PingdomResponse, error) {
	req, err := cs.client.NewRequest("DELETE", "/tms/check/"+strconv.Itoa(id), nil)
	if err != nil {
		return nil, err
	}

	m := &PingdomResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}
	return m, err
}

//Returns a status change report for all transaction checks in the current organization
func (cs *TmsCheckService) StatusReportList(request TmsStatusReportListRequest) (*TmsStatusChangeResponse, error) {
	if err := request.Valid(); err != nil {
		return nil, err
	}
	req, err := cs.client.NewRequest("GET", "/tms/check/report/status", request.GetParams())
	if err != nil {
		return nil, err
	}
	m := &TmsStatusChangeResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

//Returns a status change report for a single transaction checks in the current organization
func (cs *TmsCheckService) StatusReportById(id int, request TmsStatusReportListByIdRequest) (*TmsStatusChangeResponse, error) {
	if err := request.Valid(); err != nil {
		return nil, err
	}
	req, err := cs.client.NewRequest("GET", fmt.Sprintf("/tms/check/%s/report/status", strconv.Itoa(id)), request.GetParams())
	if err != nil {
		return nil, err
	}
	m := &TmsStatusChangeResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

//Returns a performance report for a single transaction checks in the current organization
func (cs *TmsCheckService) PerformanceReport(id int, request TmsPerformanceReportRequest) (*TmsPerformanceReportResponse, error) {
	if err := request.Valid(); err != nil {
		return nil, err
	}
	req, err := cs.client.NewRequest("GET", fmt.Sprintf("/tms/check/%s/report/performance", strconv.Itoa(id)), request.GetParams())
	if err != nil {
		return nil, err
	}
	m := &TmsPerformanceReportResponse{}
	_, err = cs.client.Do(req, m)
	if err != nil {
		return nil, err
	}

	return m, nil
}
