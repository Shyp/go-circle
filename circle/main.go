package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	circle "github.com/Shyp/go-circle"
	"github.com/Shyp/go-circle/wait"
	git "github.com/Shyp/go-git"
	"github.com/skratchdot/open-golang/open"
	"golang.org/x/sync/errgroup"
)

const help = `The circle binary interacts with a server that runs your tests.

Usage: 

	circle command [arguments]

The commands are:

	open                Open the latest branch build in a browser.
	update              Update to the latest version
	version             Print the current version
	wait                Wait for tests to finish on a branch.
	download-artifacts  Download all artifacts.

Use "circle help [command]" for more information about a command.
`

const downloadUsage = `usage: download-artifacts <build-num>`

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
		log.Fatal(err)
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

func doDownload(flags *flag.FlagSet) error {
	buildStr := flags.Arg(0)
	val, err := strconv.Atoi(buildStr)
	if err != nil {
		return err
	}
	remote, err := git.GetRemoteURL("origin")
	if err != nil {
		return err
	}
	arts, err := circle.GetArtifactsForBuild(remote.Path, remote.RepoName, val)
	if err != nil {
		return err
	}
	var g errgroup.Group
	tempDir, err := ioutil.TempDir("", "circle-artifacts")
	if err != nil {
		return err
	}
	for _, art := range arts {
		art := art
		g.Go(func() error {
			return circle.DownloadArtifact(art, tempDir, remote.Path)
		})
	}

	if err := g.Wait(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return err
	}
	fmt.Fprintf(os.Stderr, "Wrote all artifacts for build %d to %s\n", val, tempDir)
	return nil
}

func main() {
	waitflags := flag.NewFlagSet("wait", flag.ExitOnError)
	waitflags.Usage = func() {
		fmt.Fprintf(os.Stderr, `usage: wait

Wait for builds to complete, then print a descriptive output on 
success or failure.
`)
		waitflags.PrintDefaults()
		os.Exit(2)
	}
	openflags := flag.NewFlagSet("open", flag.ExitOnError)
	downloadflags := flag.NewFlagSet("download-artifacts", flag.ExitOnError)
	downloadflags.Usage = func() {
		fmt.Fprintf(os.Stderr, "%s\n\n", downloadUsage)
		downloadflags.PrintDefaults()
		os.Exit(1)
	}

	flag.Parse()
	args := flag.Args()
	if len(args) < 1 {
		usage()
		return
	}
	switch flag.Arg(0) {
	case "open":
		openflags.Parse(os.Args[2:])
		doOpen(openflags)
	case "update":
		err := equinoxUpdate()
		checkError(err)
	case "version":
		fmt.Fprintf(os.Stderr, "circle version %s\n", circle.VERSION)
		os.Exit(1)
	case "wait":
		waitflags.Parse(os.Args[2:])
		args := waitflags.Args()
		branch, err := getBranchFromArgs(args)
		checkError(err)
		err = wait.Wait(branch)
		checkError(err)
	case "download-artifacts":
		if len(args) == 1 {
			fmt.Fprintf(os.Stderr, "usage: download-artifacts <build-number>\n")
			os.Exit(1)
		}
		downloadflags.Parse(args[1:])
		if err := doDownload(downloadflags); err != nil {
			log.Fatal(err)
		}

	default:
		usage()
	}
}
