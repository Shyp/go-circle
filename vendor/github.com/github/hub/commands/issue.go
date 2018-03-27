package commands

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/github/hub/git"
	"github.com/github/hub/github"
	"github.com/github/hub/ui"
	"github.com/github/hub/utils"
)

var (
	cmdIssue = &Command{
		Run: listIssues,
		Usage: `
issue [-a <ASSIGNEE>] [-c <CREATOR>] [-@ <USER>] [-s <STATE>] [-f <FORMAT>] [-M <MILESTONE>] [-l <LABELS>] [-d <DATE>] [-o <SORT_KEY> [-^]] [-L <LIMIT>]
issue create [-oc] [-m <MESSAGE>|-F <FILE>] [-a <USERS>] [-M <MILESTONE>] [-l <LABELS>]
issue labels [--color]
`,
		Long: `Manage GitHub issues for the current project.

## Commands:

With no arguments, show a list of open issues.

	* _create_:
		Open an issue in the current project.

	* _labels_:
		List the labels available in this repository.

## Options:
	-a, --assignee <ASSIGNEE>
		Display only issues assigned to <ASSIGNEE>.

		When opening an issue, this can be a comma-separated list of people to
		assign to the new issue.

	-c, --creator <CREATOR>
		Display only issues created by <CREATOR>.

	-@, --mentioned <USER>
		Display only issues mentioning <USER>.

	-s, --state <STATE>
		Display issues with state <STATE> (default: "open").

	-f, --format <FORMAT>
		Pretty print the contents of the issues using format <FORMAT> (default:
		"%sC%>(8)%i%Creset  %t%  l%n"). See the "PRETTY FORMATS" section of the
		git-log manual for some additional details on how placeholders are used in
		format. The available placeholders for issues are:

		%I: issue number

		%i: issue number prefixed with "#"

		%U: the URL of this issue

		%S: state (i.e. "open", "closed")

		%sC: set color to red or green, depending on issue state.

		%t: title

		%l: colored labels

		%L: raw, comma-separated labels

		%b: body

		%au: login name of author

		%as: comma-separated list of assignees

		%Mn: milestone number

		%Mt: milestone title

		%NC: number of comments

		%Nc: number of comments wrapped in parentheses, or blank string if zero.

		%cD: created date-only (no time of day)

		%cr: created date, relative

		%ct: created date, UNIX timestamp

		%cI: created date, ISO 8601 format

		%uD: updated date-only (no time of day)

		%ur: updated date, relative

		%ut: updated date, UNIX timestamp

		%uI: updated date, ISO 8601 format

	-m, --message <MESSAGE>
		Use the first line of <MESSAGE> as issue title, and the rest as issue description.

	-F, --file <FILE>
		Read the issue title and description from <FILE>.

	-e, --edit
		Further edit the contents of <FILE> in a text editor before submitting.

	-o, --browse
		Open the new issue in a web browser.

	-c, --copy
		Put the URL of the new issue to clipboard instead of printing it.

	-M, --milestone <ID>
		Display only issues for a GitHub milestone with id <ID>.

		When opening an issue, add this issue to a GitHub milestone with id <ID>.

	-l, --labels <LABELS>
		Display only issues with certain labels.

		When opening an issue, add a comma-separated list of labels to this issue.

	-d, --since <DATE>
		Display only issues updated on or after <DATE> in ISO 8601 format.

	-o, --sort <SORT_KEY>
		Sort displayed issues by "created" (default), "updated" or "comments".

	-^ --sort-ascending
		Sort by ascending dates instead of descending.

	-L, --limit <LIMIT>
		Display only the first <LIMIT> issues.

	--include-pulls
		Include pull requests as well as issues.

	--color
		Enable colored output for labels list.
`,
	}

	cmdCreateIssue = &Command{
		Key:   "create",
		Run:   createIssue,
		Usage: "issue create [-o] [-m <MESSAGE>|-F <FILE>] [-a <USERS>] [-M <MILESTONE>] [-l <LABELS>]",
		Long:  "Open an issue in the current project.",
	}

	cmdLabel = &Command{
		Key:   "labels",
		Run:   listLabels,
		Usage: "issue labels [--color]",
		Long:  "List the labels available in this repository.",
	}

	flagIssueAssignee,
	flagIssueState,
	flagIssueFormat,
	flagIssueMessage,
	flagIssueMilestoneFilter,
	flagIssueCreator,
	flagIssueMentioned,
	flagIssueLabelsFilter,
	flagIssueSince,
	flagIssueSort,
	flagIssueFile string

	flagIssueEdit,
	flagIssueCopy,
	flagIssueBrowse,
	flagIssueSortAscending bool
	flagIssueIncludePulls bool

	flagIssueMilestone uint64

	flagIssueAssignees,
	flagIssueLabels listFlag

	flagIssueLimit int

	flagLabelsColorize bool
)

func init() {
	cmdCreateIssue.Flag.StringVarP(&flagIssueMessage, "message", "m", "", "MESSAGE")
	cmdCreateIssue.Flag.StringVarP(&flagIssueFile, "file", "F", "", "FILE")
	cmdCreateIssue.Flag.Uint64VarP(&flagIssueMilestone, "milestone", "M", 0, "MILESTONE")
	cmdCreateIssue.Flag.VarP(&flagIssueLabels, "label", "l", "LABEL")
	cmdCreateIssue.Flag.VarP(&flagIssueAssignees, "assign", "a", "ASSIGNEE")
	cmdCreateIssue.Flag.BoolVarP(&flagIssueBrowse, "browse", "o", false, "BROWSE")
	cmdCreateIssue.Flag.BoolVarP(&flagIssueCopy, "copy", "c", false, "COPY")
	cmdCreateIssue.Flag.BoolVarP(&flagIssueEdit, "edit", "e", false, "EDIT")

	cmdIssue.Flag.StringVarP(&flagIssueAssignee, "assignee", "a", "", "ASSIGNEE")
	cmdIssue.Flag.StringVarP(&flagIssueState, "state", "s", "", "STATE")
	cmdIssue.Flag.StringVarP(&flagIssueFormat, "format", "f", "%sC%>(8)%i%Creset  %t%  l%n", "FORMAT")
	cmdIssue.Flag.StringVarP(&flagIssueMilestoneFilter, "milestone", "M", "", "MILESTONE")
	cmdIssue.Flag.StringVarP(&flagIssueCreator, "creator", "c", "", "CREATOR")
	cmdIssue.Flag.StringVarP(&flagIssueMentioned, "mentioned", "@", "", "USER")
	cmdIssue.Flag.StringVarP(&flagIssueLabelsFilter, "labels", "l", "", "LABELS")
	cmdIssue.Flag.StringVarP(&flagIssueSince, "since", "d", "", "DATE")
	cmdIssue.Flag.StringVarP(&flagIssueSort, "sort", "o", "created", "SORT_KEY")
	cmdIssue.Flag.BoolVarP(&flagIssueSortAscending, "sort-ascending", "^", false, "SORT_KEY")
	cmdIssue.Flag.BoolVarP(&flagIssueIncludePulls, "include-pulls", "", false, "INCLUDE_PULLS")
	cmdIssue.Flag.IntVarP(&flagIssueLimit, "limit", "L", -1, "LIMIT")

	cmdLabel.Flag.BoolVarP(&flagLabelsColorize, "color", "", false, "COLORIZE")

	cmdIssue.Use(cmdCreateIssue)
	cmdIssue.Use(cmdLabel)
	CmdRunner.Use(cmdIssue)
}

func listIssues(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	if args.Noop {
		ui.Printf("Would request list of issues for %s\n", project)
	} else {
		flagFilters := map[string]string{
			"state":     flagIssueState,
			"assignee":  flagIssueAssignee,
			"milestone": flagIssueMilestoneFilter,
			"creator":   flagIssueCreator,
			"mentioned": flagIssueMentioned,
			"labels":    flagIssueLabelsFilter,
			"sort":      flagIssueSort,
		}
		filters := map[string]interface{}{}
		for flag, filter := range flagFilters {
			if cmd.FlagPassed(flag) {
				filters[flag] = filter
			}
		}

		if flagIssueSortAscending {
			filters["direction"] = "asc"
		}

		if cmd.FlagPassed("since") {
			if sinceTime, err := time.ParseInLocation("2006-01-02", flagIssueSince, time.Local); err == nil {
				filters["since"] = sinceTime.Format(time.RFC3339)
			} else {
				filters["since"] = flagIssueSince
			}
		}

		issues, err := gh.FetchIssues(project, filters, flagIssueLimit, func(issue *github.Issue) bool {
			return issue.PullRequest == nil || flagIssueIncludePulls
		})
		utils.Check(err)

		maxNumWidth := 0
		for _, issue := range issues {
			if numWidth := len(strconv.Itoa(issue.Number)); numWidth > maxNumWidth {
				maxNumWidth = numWidth
			}
		}

		colorize := ui.IsTerminal(os.Stdout)
		for _, issue := range issues {
			ui.Printf(formatIssue(issue, flagIssueFormat, colorize))
		}
	}

	args.NoForward()
}

func formatIssuePlaceholders(issue github.Issue, colorize bool) map[string]string {
	var stateColorSwitch string
	if colorize {
		issueColor := 32
		if issue.State == "closed" {
			issueColor = 31
		}
		stateColorSwitch = fmt.Sprintf("\033[%dm", issueColor)
	}

	var labelStrings []string
	var rawLabels []string
	for _, label := range issue.Labels {
		if !colorize {
			labelStrings = append(labelStrings, fmt.Sprintf(" %s ", label.Name))
			continue
		}
		color, err := utils.NewColor(label.Color)
		if err != nil {
			utils.Check(err)
		}

		labelStrings = append(labelStrings, colorizeLabel(label, color))
		rawLabels = append(rawLabels, label.Name)
	}

	var assignees []string
	for _, assignee := range issue.Assignees {
		assignees = append(assignees, assignee.Login)
	}

	var milestoneNumber, milestoneTitle string
	if issue.Milestone != nil {
		milestoneNumber = fmt.Sprintf("%d", issue.Milestone.Number)
		milestoneTitle = issue.Milestone.Title
	}

	var numCommentsWrapped string
	numComments := fmt.Sprintf("%d", issue.Comments)
	if issue.Comments > 0 {
		numCommentsWrapped = fmt.Sprintf("(%d)", issue.Comments)
	}

	var createdDate, createdAtISO8601, createdAtUnix, createdAtRelative,
		updatedDate, updatedAtISO8601, updatedAtUnix, updatedAtRelative string
	if !issue.CreatedAt.IsZero() {
		createdDate = issue.CreatedAt.Format("02 Jan 2006")
		createdAtISO8601 = issue.CreatedAt.Format(time.RFC3339)
		createdAtUnix = fmt.Sprintf("%d", issue.CreatedAt.Unix())
		createdAtRelative = utils.TimeAgo(issue.CreatedAt)
	}
	if !issue.UpdatedAt.IsZero() {
		updatedDate = issue.UpdatedAt.Format("02 Jan 2006")
		updatedAtISO8601 = issue.UpdatedAt.Format(time.RFC3339)
		updatedAtUnix = fmt.Sprintf("%d", issue.UpdatedAt.Unix())
		updatedAtRelative = utils.TimeAgo(issue.UpdatedAt)
	}

	return map[string]string{
		"I":  fmt.Sprintf("%d", issue.Number),
		"i":  fmt.Sprintf("#%d", issue.Number),
		"U":  issue.HtmlUrl,
		"S":  issue.State,
		"sC": stateColorSwitch,
		"t":  issue.Title,
		"l":  strings.Join(labelStrings, " "),
		"L":  strings.Join(rawLabels, ", "),
		"b":  issue.Body,
		"au": issue.User.Login,
		"as": strings.Join(assignees, ", "),
		"Mn": milestoneNumber,
		"Mt": milestoneTitle,
		"NC": numComments,
		"Nc": numCommentsWrapped,
		"cD": createdDate,
		"cI": createdAtISO8601,
		"ct": createdAtUnix,
		"cr": createdAtRelative,
		"uD": updatedDate,
		"uI": updatedAtISO8601,
		"ut": updatedAtUnix,
		"ur": updatedAtRelative,
	}
}

func formatIssue(issue github.Issue, format string, colorize bool) string {
	placeholders := formatIssuePlaceholders(issue, colorize)
	return ui.Expand(format, placeholders, colorize)
}

func createIssue(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	messageBuilder := &github.MessageBuilder{
		Filename: "ISSUE_EDITMSG",
		Title:    "issue",
	}

	messageBuilder.AddCommentedSection(fmt.Sprintf(`Creating an issue for %s

Write a message for this issue. The first block of
text is the title and the rest is the description.`, project))

	if cmd.FlagPassed("message") {
		messageBuilder.Message = flagIssueMessage
		messageBuilder.Edit = flagIssueEdit
	} else if cmd.FlagPassed("file") {
		messageBuilder.Message, err = msgFromFile(flagIssueFile)
		utils.Check(err)
		messageBuilder.Edit = flagIssueEdit
	} else {
		messageBuilder.Edit = true

		workdir, _ := git.WorkdirName()
		if workdir != "" {
			template, err := github.ReadTemplate(github.IssueTemplate, workdir)
			utils.Check(err)
			if template != "" {
				messageBuilder.Message = template
			}
		}

	}

	title, body, err := messageBuilder.Extract()
	utils.Check(err)

	if title == "" {
		utils.Check(fmt.Errorf("Aborting creation due to empty issue title"))
	}

	params := map[string]interface{}{
		"title": title,
		"body":  body,
	}

	if len(flagIssueLabels) > 0 {
		params["labels"] = flagIssueLabels
	}

	if len(flagIssueAssignees) > 0 {
		params["assignees"] = flagIssueAssignees
	}

	if flagIssueMilestone > 0 {
		params["milestone"] = flagIssueMilestone
	}

	args.NoForward()
	if args.Noop {
		ui.Printf("Would create issue `%s' for %s\n", params["title"], project)
	} else {
		issue, err := gh.CreateIssue(project, params)
		utils.Check(err)

		printBrowseOrCopy(args, issue.HtmlUrl, flagIssueBrowse, flagIssueCopy)
	}

	messageBuilder.Cleanup()
}

func listLabels(cmd *Command, args *Args) {
	localRepo, err := github.LocalRepo()
	utils.Check(err)

	project, err := localRepo.MainProject()
	utils.Check(err)

	gh := github.NewClient(project.Host)

	args.NoForward()
	if args.Noop {
		ui.Printf("Would request list of labels for %s\n", project)
		return
	}

	labels, err := gh.FetchLabels(project)
	utils.Check(err)

	for _, label := range labels {
		ui.Printf(formatLabel(label, flagLabelsColorize))
	}
}

func formatLabel(label github.IssueLabel, colorize bool) string {
	if colorize {
		if color, err := utils.NewColor(label.Color); err == nil {
			return fmt.Sprintf("%s\n", colorizeLabel(label, color))
		}
	}
	return fmt.Sprintf("%s\n", label.Name)
}

func colorizeLabel(label github.IssueLabel, color *utils.Color) string {
	return fmt.Sprintf("\033[38;5;%d;48;2;%d;%d;%dm %s \033[m",
		getSuitableLabelTextColor(color), color.Red, color.Green, color.Blue, label.Name)
}

func getSuitableLabelTextColor(color *utils.Color) int {
	if color.Brightness() < 0.65 {
		return 15 // white text
	}
	return 16 // black text
}
