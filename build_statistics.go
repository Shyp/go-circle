package circle

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh/terminal"
)

func roundDuration(d CircleDuration, unit time.Duration) time.Duration {
	return ((time.Duration(d) + unit/2) / unit) * unit
}

const stepWidth = 45

var stepPadding = fmt.Sprintf("%%-%ds", stepWidth)

func isatty() bool {
	return terminal.IsTerminal(int(os.Stdout.Fd()))
}

func BuildStatistics(cb *CircleBuild) string {
	var b bytes.Buffer
	b.WriteString(fmt.Sprintf(stepPadding, "Step"))
	l := stepWidth
	for i := uint8(0); i < cb.Parallel; i++ {
		b.WriteString(fmt.Sprintf("%-8d", i))
		l += 8
	}
	b.WriteString(fmt.Sprintf("\n%s\n", strings.Repeat("=", l)))
	for _, step := range cb.Steps {
		var stepName string
		if len(step.Name) > stepWidth-2 {
			stepName = fmt.Sprintf("%s… ", step.Name[:(stepWidth-2)])
		} else {
			stepName = fmt.Sprintf(stepPadding, step.Name)
		}
		b.WriteString(stepName)
		for _, action := range step.Actions {
			var dur time.Duration
			if time.Duration(action.Runtime) > time.Minute {
				dur = roundDuration(action.Runtime, time.Second)
			} else {
				dur = roundDuration(action.Runtime, time.Millisecond*10)
			}
			var durString string
			if (action.Status == "failed" || action.Status == "timedout") && isatty() {
				durString = fmt.Sprintf("\033[38;05;160m%-8s\033[0m", dur.String())
			} else {
				durString = fmt.Sprintf("%-8s", dur.String())
			}
			b.WriteString(durString)
		}
		b.WriteString("\n")
	}
	return b.String()
}
