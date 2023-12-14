package runwait

import (
	"bytes"
	"errors"
	"os/exec"
	"strings"
)

func RunWait(name string, arg ...string) (output string, errExt string, err error) {
	cmd := exec.Command(name, arg...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		errExt = stderr.String()
	} else {
		output = out.String()
	}

	return
}

func RunWait2(name string, arg ...string) (output string, err error) {
	output, errExt, err := RunWait(name, arg...)
	if err != nil {
		err = errors.New(errExt)
	}
	return
}

func RunWaitWithStdIn(name string, input string, arg ...string) (output string, err error) {
	cmd := exec.Command(name, arg...)
	var out, stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	cmd.Stdin = strings.NewReader(input)

	err = cmd.Run()
	if err != nil {
		err = errors.New(stderr.String())
	} else {
		output = out.String()
	}

	return
}
