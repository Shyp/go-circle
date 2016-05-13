package circle

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/Shyp/go-types"
)

var client http.Client

func init() {
	client = http.Client{
		Timeout: 10 * time.Second,
	}
}

const VERSION = "0.20"
const baseUri = "https://circleci.com/api/v1/project"

type TreeBuild struct {
	BuildNum	int	`json:"build_num"`
	BuildURL	string	`json:"build_url"`
	CompareURL	string	`json:"compare"`
	// Tree builds have a `previous_successful_build` field but as far as I can
	// tell it is always null. Instead this field is set
	Previous	PreviousBuild	`json:"previous"`
	QueuedAt	types.NullTime	`json:"queued_at"`
	Status		string		`json:"status"`
	StartTime	types.NullTime	`json:"start_time"`
	StopTime	types.NullTime	`json:"stop_time"`
	UsageQueuedAt	types.NullTime	`json:"usage_queued_at"`
	VCSRevision	string		`json:"vcs_revision"`
}

func (tb TreeBuild) Passed() bool {
	return tb.Status == "success" || tb.Status == "fixed"
}

func (tb TreeBuild) NotRunning() bool {
	return tb.Status == "not_running" || tb.Status == "scheduled" || tb.Status == "queued"
}

func (tb TreeBuild) Running() bool {
	return tb.Status == "running"
}

func (tb TreeBuild) Failed() bool {
	return tb.Status == "failed" || tb.Status == "timedout" || tb.Status == "no_tests" || tb.Status == "infrastructure_fail"
}

type CircleBuild struct {
	Parallel		uint8		`json:"parallel"`
	PreviousSuccessfulBuild	PreviousBuild	`json:"previous_successful_build"`
	QueuedAt		types.NullTime	`json:"queued_at"`
	Steps			[]Step		`json:"steps"`
	UsageQueuedAt		types.NullTime	`json:"usage_queued_at"`
}

type PreviousBuild struct {
	BuildNum	int	`json:"build_num"`
	// would be neat to make this a time.Duration, easier to use the passed in
	// value.
	Status	string	`json:"status"`

	BuildDurationMs	int	`json:"build_time_millis"`
}

type Step struct {
	Name	string		`json:"name"`
	Actions	[]Action	`json:"actions"`
}

type Action struct {
	Name		string		`json:"name"`
	OutputURL	URL		`json:"output_url"`
	Runtime		CircleDuration	`json:"run_time_millis"`
	Status		string		`json:"status"`
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

func getCancelUri(org string, project string, build int, token string) string {
	return fmt.Sprintf("%s/%s/%s/%d/cancel?circle-token=%s", baseUri, org, project, build, token)
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
	var r io.Reader
	if os.Getenv("CIRCLE_DEBUG") == "true" {
		r = io.TeeReader(body, os.Stdout)
	} else {
		r = body
	}
	d := json.NewDecoder(r)
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
	var r io.Reader
	if os.Getenv("CIRCLE_DEBUG") == "true" {
		r = io.TeeReader(body, os.Stdout)
	} else {
		r = body
	}
	d := json.NewDecoder(r)
	var cb CircleBuild
	err = d.Decode(&cb)
	return &cb, err
}

func CancelBuild(org string, project string, buildNum int) (*CircleBuild, error) {
	token, err := getToken(org)
	if err != nil {
		return nil, err
	}
	uri := getCancelUri(org, project, buildNum, token)
	body, err := makeRequest("POST", uri)
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var r io.Reader
	if os.Getenv("CIRCLE_DEBUG") == "true" {
		r = io.TeeReader(body, os.Stdout)
	} else {
		r = body
	}
	d := json.NewDecoder(r)
	var cb CircleBuild
	err = d.Decode(&cb)
	return &cb, err
}
