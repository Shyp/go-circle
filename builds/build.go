package build

import (
  "fmt"
  "time"

  "github.com/Shyp/go-circle"
  "github.com/Shyp/go-git"
  "github.com/Shyp/cli/ansi"
)

func inSlice(a string, list []string) bool {
  for _, b := range list {
    if b == a {
        return true
    }
  }
  return false
}

// GetBuilds gets the status of the 5 most recent Circle builds for a branch
func GetBuilds(branch string) error {
  // Different statuses Circle builds can have
  green := []string{"fixed", "success"}
  grey  := []string{"retried", "canceled", "not_run"}
  red   := []string{"infrastructure_fail", "timedout", "failed", "no_tests"}
  blue  := []string{"running"}
  // purple := []string{"queued", "not_running", "scheduled"}

  tip, err := git.Tip(branch)
  _ = tip
  // This throws if the branch doesn't exist
  if err != nil {
    return err
  }

  fmt.Println("\nFetching recent builds for", branch, "starting with most recent commit\n")
  // Give CircleCI a little bit of time to start
  time.Sleep(1 * time.Second)

  remote, err := git.GetRemoteURL("origin")
  if err != nil {
    return err
  }

  cr, err := circle.GetTree(remote.Path, remote.RepoName, branch)
  if err != nil {
    return err
  }

  sum := 0

  // Limited to 5 most recent builds. Feature would be to pass in number
  // of builds to fetch via command line args
  for i := 0; i < 5; i++ {
    build := (*cr)[i]
    ghUrl, url, status := build.CompareURL, build.BuildURL, build.Status

    // Based on the status of the build, change the color of status print out
    if inSlice(status, green) {
      status = ansi.Green(status)
    } else if inSlice(status, grey) {
      status = ansi.Grey(status)
    } else if inSlice(status, red) {
      status = ansi.Red(status)
    } else if inSlice(status, blue) {
      status = ansi.Blue(status)
    } else {
      status = ansi.Purple(status)
    }

    fmt.Println(url, status, ghUrl)

    sum += i
  }

  fmt.Println("\nMost recent build statuses fetched!")

  return nil
}

// CancelBuild cancels a build (as specified by the build number)
func CancelBuild(org string, project string, buildNum int) string {
  fmt.Println("\nCanceling build", buildNum, "for", project, "\n")
  build, err := circle.CancelBuild(org, project, buildNum)
  _ = build

  if err != nil {
    return ""
  }

  fmt.Println("Verify status by running `shyp builds`")
  return ""
}
