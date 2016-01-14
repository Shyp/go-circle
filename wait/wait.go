package wait

import (
	"fmt"
	"net"
	"net/url"
	"time"

	"github.com/Shyp/go-circle"
	"github.com/Shyp/go-git"
	"github.com/kevinburke/bigtext"
)

func roundDuration(d time.Duration, unit time.Duration) time.Duration {
	return ((d + unit/2) / unit) * unit
}

func getStats(org string, project string, buildNum int) string {
	build, err := circle.GetBuild(org, project, buildNum)
	if err != nil {
		return ""
	}
	return circle.BuildStatistics(build)
}

// isHttpError checks if the given error is a request timeout or a network
// failure - in those cases we want to just retry the request.
func isHttpError(err error) bool {
	if err == nil {
		return false
	}
	// some net.OpError's are wrapped in a url.Error
	if uerr, ok := err.(*url.Error); ok {
		err = uerr.Err
	}
	switch err := err.(type) {
	default:
		return false
	case *net.OpError:
		return err.Op == "dial" && err.Net == "tcp"
	case *net.DNSError:
		return true
	// Catchall, this needs to go last.
	case net.Error:
		return err.Timeout() || err.Temporary()
	}
	return false
}

func Wait(branch string) error {
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	tip, err := git.Tip(branch)
	if err != nil {
		return err
	}
	fmt.Println("Waiting for latest build on", branch, "to complete")
	// Give CircleCI a little bit of time to start
	time.Sleep(1 * time.Second)
	for {
		cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
		if err != nil {
			if isHttpError(err) {
				fmt.Printf("Caught network error: %s. Continuing\n", err.Error())
				time.Sleep(2 * time.Second)
				continue
			}
			return err
		}
		if len(*cr) == 0 {
			return fmt.Errorf("No results, are you sure there are tests for %s/%s?\n",
				remote.Path, remote.RepoName)
			break
		}
		latestBuild := (*cr)[0]
		var vcsLen int
		var tipLen int
		if len(latestBuild.VCSRevision) > 8 {
			vcsLen = 8
		} else {
			vcsLen = len(latestBuild.VCSRevision)
		}
		if len(tip) > 8 {
			tipLen = 8
		} else {
			tipLen = len(tip)
		}
		if latestBuild.VCSRevision[:tipLen] != tip {
			fmt.Printf("Latest build in Circle is %s, waiting for %s...\n",
				latestBuild.VCSRevision[:vcsLen], tip[:tipLen])
			time.Sleep(5 * time.Second)
			continue
		}
		var duration time.Duration
		if latestBuild.StartTime.Valid {
			if latestBuild.StopTime.Valid {
				duration = latestBuild.StopTime.Time.Sub(latestBuild.StartTime.Time)
			} else {
				duration = time.Now().Sub(latestBuild.StartTime.Time)
			}
			duration = roundDuration(duration, time.Second)
		}
		if latestBuild.Passed() {
			fmt.Printf("Build on %s succeeded!\n\n", branch)
			fmt.Printf(getStats(remote.Path, remote.RepoName, latestBuild.BuildNum))
			fmt.Printf("\nTests on %s took %s. Quitting.\n", branch, duration.String())
			bigtext.Display(fmt.Sprintf("%s done", branch))
			break
		} else if latestBuild.Failed() {
			fmt.Printf(getStats(remote.Path, remote.RepoName, latestBuild.BuildNum))
			fmt.Printf("\nURL: %s\n", latestBuild.BuildURL)
			err = fmt.Errorf("Build on %s failed!\n\n", branch)
			bigtext.Display("build failed")
			return err
		} else {
			if latestBuild.Status == "running" {
				fmt.Printf("Running (%s elapsed)\n", duration.String())
			} else {
				fmt.Printf("Status is %s, trying again\n", latestBuild.Status)
			}
			if float32(duration) < (2.5 * float32(time.Minute)) {
				time.Sleep(10 * time.Second)
			} else {
				time.Sleep(5 * time.Second)
			}
		}
	}
	return nil
}
