package cli

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Runner struct {
	Home string
}

func NewRunner() Runner {
	home := os.Getenv("CLASHCTL_HOME")
	if home == "" {
		if wd, err := os.Getwd(); err == nil {
			home = wd
		}
	}
	return Runner{Home: home}
}

func (r Runner) Clashctl(args ...string) Result {
	quoted := make([]string, 0, len(args))
	for _, arg := range args {
		quoted = append(quoted, shellQuote(arg))
	}
	cmdline := fmt.Sprintf(". %s >/dev/null 2>&1; clashctl %s", shellQuote(r.Home+"/scripts/cmd/clashctl.sh"), strings.Join(quoted, " "))
	cmd := exec.Command("bash", "-lc", cmdline)
	cmd.Env = append(os.Environ(), "CLASHCTL_HOME="+r.Home)
	cmd.Stdin = strings.NewReader("")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	err := cmd.Run()
	return Result{Command: "clashctl " + strings.Join(args, " "), Output: strings.TrimSpace(out.String()), Err: err}
}

type Result struct {
	Command string
	Output  string
	Err     error
}

func (r Result) Text() string {
	var b strings.Builder
	b.WriteString("$ ")
	b.WriteString(r.Command)
	b.WriteString("\n\n")
	if r.Output != "" {
		b.WriteString(r.Output)
	} else {
		b.WriteString("(无输出)")
	}
	if r.Err != nil {
		b.WriteString("\n\n退出状态：")
		b.WriteString(r.Err.Error())
	}
	return b.String()
}

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}
