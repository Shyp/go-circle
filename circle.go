package circle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var client http.Client

func init() {
	client = http.Client{
		Timeout: 10 * time.Second,
	}
}

const VERSION = "0.7"
const baseUri = "https://circleci.com/api/v1/project"

type TreeBuild struct {
	BuildNum    int            `json:"build_num"`
	BuildURL    string         `json:"build_url"`
	Status      string         `json:"status"`
	StartTime   CircleNullTime `json:"start_time"`
	StopTime    CircleNullTime `json:"stop_time"`
	VCSRevision string         `json:"vcs_revision"`
}

func (tb *TreeBuild) Passed() bool {
	return tb.Status == "success" || tb.Status == "fixed"
}

func (tb *TreeBuild) Failed() bool {
	return tb.Status == "failed" || tb.Status == "timedout"
}

type CircleBuild struct {
	Parallel uint8  `json:"parallel"`
	Steps    []Step `json:"steps"`
}

type Step struct {
	Name    string   `json:"name"`
	Actions []Action `json:"actions"`
}

type Action struct {
	Name      string         `json:"name"`
	OutputURL URL            `json:"output_url"`
	Runtime   CircleDuration `json:"run_time_millis"`
	Status    string         `json:"status"`
}

func (a *Action) Failed() bool {
	return a.Status == "failed" || a.Status == "timedout"
}

func getTreeUri(org string, project string, branch string, token string) string {
	return fmt.Sprintf("%s/%s/%s/tree/%s?circle-token=%s", baseUri, org, project, branch, token)
}

func getBuildUri(org string, project string, build int, token string) string {
	return fmt.Sprintf("%s/%s/%s/%d?circle-token=%s", baseUri, org, project, build, token)
}

type CircleTreeResponse []TreeBuild

func makeRequest(method, uri string) (io.ReadCloser, error) {
	req, err := http.NewRequest(method, uri, nil)
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
	return resp.Body, nil
}

func GetTree(org string, project string, branch string) (*CircleTreeResponse, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getTreeUri(org, project, branch, token)
	body, err := makeRequest("GET", uri)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var cr CircleTreeResponse
	d := json.NewDecoder(body)
	err = d.Decode(&cr)
	return &cr, err
}

func GetBuild(org string, project string, buildNum int) (*CircleBuild, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getBuildUri(org, project, buildNum, token)
	body, err := makeRequest("GET", uri)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var cb CircleBuild
	d := json.NewDecoder(body)
	err = d.Decode(&cb)
	return &cb, err
}
