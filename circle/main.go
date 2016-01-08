package main

import (
	"flag"
	"fmt"
	"os"

	circle "github.com/Shyp/go-circle"
	"github.com/Shyp/go-circle/wait"
	git "github.com/Shyp/go-git"
	"github.com/skratchdot/open-golang/open"
)

var help = `The circle binary interacts with a server that runs your tests.

Usage: 

	circle command [arguments]

The commands are:

	open            Open the latest branch build in a browser.
	update          Update to the latest version
	version         Print the current version
	wait            Wait for tests to finish on a branch.

Use "circle help [command]" for more information about a command.
`

func usage() {
	fmt.Fprintf(os.Stderr, help)
	flag.PrintDefaults()
	os.Exit(2)
}

func init() {
	flag.Usage = usage
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(1)
	}
}

// Given a set of command line args, return the git branch or an error. Returns
// the current git branch if no argument is specified
func getBranchFromArgs(args []string) (string, error) {
	if len(args) == 0 {
		return git.CurrentBranch()
	} else {
		return args[0], nil
	}
}

func doOpen(flags *flag.FlagSet) {
	args := flags.Args()
	branch, err := getBranchFromArgs(args)
	checkError(err)
	remote, err := git.GetRemoteURL("origin")
	checkError(err)
	cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
	checkError(err)
	if len(*cr) == 0 {
		fmt.Printf("No results, are you sure there are tests for %s/%s?\n",
			remote.Path, remote.RepoName)
		return
	}
	latestBuild := (*cr)[0]
	open.Start(latestBuild.BuildURL)
}

func main() {
	waitflags := flag.NewFlagSet("wait", flag.ExitOnError)
	openflags := flag.NewFlagSet("open", flag.ExitOnError)

	if len(os.Args) < 2 {
		usage()
		return
	}
	switch os.Args[1] {
	case "open":
		openflags.Parse(os.Args[2:])
		doOpen(openflags)
	case "update":
		err := equinoxUpdate()
		checkError(err)
	case "version":
		fmt.Fprintf(os.Stderr, "circle version %s\n", circle.VERSION)
		os.Exit(0)
	case "wait":
		waitflags.Parse(os.Args[2:])
		args := waitflags.Args()
		branch, err := getBranchFromArgs(args)
		checkError(err)
		err = wait.Wait(branch)
		checkError(err)
	default:
		usage()
	}
}
