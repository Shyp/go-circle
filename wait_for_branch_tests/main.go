package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	circle "github.com/Shyp/go-circle"
	git "github.com/Shyp/go-git"
	"github.com/kevinburke/bigtext"
)

var help = `Usage: wait_for_circle [branch]

	branch: Name of the branch to wait for (defaults to your current git branch)
`

func init() {
	flag.Usage = func() {
		fmt.Printf(help)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}
}

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

func main() {
	flag.Parse()
	args := flag.Args()
	var branch string
	if len(args) > 1 {
		fmt.Fprintf(os.Stderr, "too many arguments provided\n")
		os.Exit(1)
	} else if len(args) == 0 {
		branchName, err := git.CurrentBranch()
		checkError(err)
		branch = branchName
	} else {
		branch = args[0]
	}
	remote, err := git.GetRemoteURL("origin")
	checkError(err)
	tip, err := git.Tip(branch)
	checkError(err)
	fmt.Println("Waiting for latest build on", branch, "to complete")
	// Give CircleCI a little bit of time to start
	time.Sleep(3 * time.Second)
	for {
		cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
		checkError(err)
		if len(*cr) == 0 {
			fmt.Printf("No results, are you sure there are tests for %s/%s?\n",
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
		if latestBuild.Status == "success" || latestBuild.Status == "fixed" {
			fmt.Printf("Build on %s complete!\n\n", branch)
			fmt.Printf(getStats(remote.Path, remote.RepoName, latestBuild.BuildNum))
			fmt.Printf("\nTests took %s. Quitting.\n", duration.String())
			bigtext.Display(fmt.Sprintf("%s done", branch))
			break
		} else if latestBuild.Status == "failed" || latestBuild.Status == "timedout" {
			fmt.Printf("Build failed!\n\n")
			fmt.Printf(getStats(remote.Path, remote.RepoName, latestBuild.BuildNum))
			fmt.Printf("\nURL: %s\n", latestBuild.BuildURL)
			bigtext.Display("build failed")
			break
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
}
