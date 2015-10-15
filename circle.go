package circle

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const baseUri = "https://circleci.com/api/v1/project"
const VERSION = "0.1"

type CircleNullTime struct {
	Time  time.Time
	Valid bool
}

// Necessary because Circle returns "null" for some time instances
func (nt *CircleNullTime) UnmarshalJSON(b []byte) error {
	if string(b) == "null" {
		nt.Valid = false
		return nil
	}
	var t time.Time
	err := json.Unmarshal(b, &t)
	if err != nil {
		return err
	}
	nt.Valid = true
	nt.Time = t
	return nil
}

type Build struct {
	BuildURL  string         `json:"build_url"`
	Status    string         `json:"status"`
	StartTime CircleNullTime `json:"start_time"`
	StopTime  CircleNullTime `json:"stop_time"`
}

func getTreeUri(org string, project string, branch string, token string) string {
	return fmt.Sprintf("%s/%s/%s/tree/%s?circle-token=%s", baseUri, org, project, branch, token)
}

type CircleResponse []Build

func GetTree(org string, project string, branch string) (*CircleResponse, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getTreeUri(org, project, branch, token)
	client := http.Client{
		Timeout: 10 * time.Second,
	}
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("User-Agent", fmt.Sprintf("circle-command-line-client/%s", VERSION))
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("request error")
		return nil, err
	}
	defer resp.Body.Close()
	var cr CircleResponse
	d := json.NewDecoder(resp.Body)
	err = d.Decode(&cr)
	return &cr, err
}
