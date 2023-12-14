package runwait

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

type OutputStr string

func RunWaitOutput(name string, arg ...string) (output OutputStr, err error) {
	return output.RunWait(name, arg...)
}

func (in OutputStr) RunWait(name string, arg ...string) (output OutputStr, err error) {
	cmd := exec.Command(name, arg...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	input := string(in)
	if input != "" {
		cmd.Stdin = strings.NewReader(input)
	}

	err = cmd.Run()
	if err != nil {
		if stdErrStr := stderr.String(); stdErrStr != "" {
			err = errors.New(stdErrStr)
		}
	} else {
		output = OutputStr(out.String())
	}

	return
}

func (in OutputStr) FindAllStringSubmatch(expr string, n int) [][]string {
	re := regexp.MustCompile(expr)
	return re.FindAllStringSubmatch(string(in), -1)
}

func (in OutputStr) FindStringSubmatch(expr string) []string {
	re := regexp.MustCompile(expr)
	return re.FindStringSubmatch(string(in))
}

func (in OutputStr) Split(sep string) []string {
	return strings.Split(string(in), sep)
}

func (in OutputStr) Lines() []string {
	sep := "\n"
	if runtime.GOOS == "windows" {
		sep = "\r\n"
	}
	return in.Split(sep)
}

func (in OutputStr) ForEachLine(breaker func(s string) (breakLoop bool)) (breakStr string) {
	lines := in.Lines()
	var tmp string
	for i := range lines {
		tmp = lines[i]
		if breaker(tmp) {
			breakStr = tmp
			break
		}
	}

	return
}

// type KeyFuncs map[string]func(line string, idxBegin, idxRemain int)
type KeyFuncs map[string]func(remain string)

func (in OutputStr) ForEachLineIncludeAny(subStrs KeyFuncs, trimspace bool) {
	includeAny := func(line string) {
		for k, f := range subStrs {
			if idx := strings.Index(line, k); idx != -1 {
				// f(line, idx, idx+len(k))
				remain := line[idx+len(k):]
				if trimspace {
					remain = strings.TrimSpace(remain)
				}
				f(remain)
				break
			}
		}
	}

	lines := in.Lines()
	for i := range lines {
		includeAny(lines[i])
		// breakLoop := includeAny(lines[i])
		// if breakLoop {
		// 	break
		// }
	}
}

func (in OutputStr) ForEachLineReverse(breaker func(s string) (breakLoop bool)) (breakStr string) {
	lines := in.Lines()
	var tmp string
	for i := len(lines) - 1; i >= 0; i-- {
		tmp = lines[i]
		if breaker(tmp) {
			breakStr = tmp
			break
		}
	}

	return
}

func (in OutputStr) WriteFile(fn string) error {
	return os.WriteFile(fn, []byte(in), 0644)
}

type Command struct {
	Cmd          string
	Args         []string
	OutputFilter func(OutputStr) OutputStr
}

type PipelineRun struct {
	cmds []*Command
}

func Add(cmd string, arg ...string) *PipelineRun {
	p := &PipelineRun{}
	return p.Add(cmd, arg...)
}

func AddWithFilter(cmd string, outputFilter func(OutputStr) OutputStr, arg ...string) *PipelineRun {
	p := &PipelineRun{}
	return p.AddWithFilter(cmd, outputFilter, arg...)
}

func (pl *PipelineRun) Add(cmd string, arg ...string) *PipelineRun {
	return pl.AddWithFilter(cmd, nil, arg...)
}

func (pl *PipelineRun) AddWithFilter(cmd string, outputFilter func(OutputStr) OutputStr, arg ...string) *PipelineRun {
	pl.cmds = append(pl.cmds, &Command{
		Cmd:          cmd,
		Args:         arg,
		OutputFilter: outputFilter,
	})
	return pl
}

func (pl *PipelineRun) RunWait() (output OutputStr, err error) {
	for _, cmd := range pl.cmds {
		if cmd == nil || cmd.Cmd == "" {
			continue
		}

		output, err = output.RunWait(cmd.Cmd, cmd.Args...)
		if err != nil {
			break
		}

		if cmd.OutputFilter != nil {
			output = cmd.OutputFilter(output)
		}
	}

	return
}
