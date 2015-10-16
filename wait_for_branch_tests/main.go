package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	circle "github.com/Shyp/go-circle"
	"github.com/Shyp/go-git"
	"github.com/kevinburke/bigtext"
)

var help = `Usage: wait_for_circle [branch]

	branch: Name of the branch to wait for (defaults to "master")
`

func init() {
	flag.Usage = func() {
		fmt.Printf(help)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func roundDuration(d time.Duration, unit time.Duration) time.Duration {
	return ((d + unit/2) / unit) * unit
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
	for {
		cr, err := circle.GetTree("Shyp", "shyp_api", branch)
		checkError(err)
		latestBuild := (*cr)[0]
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
			fmt.Printf("Complete! Tests took %s. Quitting.\n", duration.String())
			bigtext.Display("done")
			break
		} else if latestBuild.Status == "failed" {
			fmt.Printf("Build failed! URL: %s\n", latestBuild.BuildURL)
			bigtext.Display("build failed")
			break
		} else {
			if latestBuild.Status == "running" {
				fmt.Printf("Running (%s elapsed)\n", duration.String())
			} else {
				fmt.Printf("Status is %s, trying again\n", latestBuild.Status)
			}
			time.Sleep(5 * time.Second)
		}
	}
}
