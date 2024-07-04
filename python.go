package python

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

type Python struct {
	Executable string
	Env        []string
}

// NewPython returns a new instance of Python
func NewPython() *Python {
	return &Python{
		Executable: "python",
	}
}

func (p *Python) SetExecutable(executable string) {
	p.Executable = executable
}

// Execute runs python cmd and executes the path supplied
func (p *Python) Execute(path string, args ...string) error {
	commandArgs := []string{path}
	commandArgs = append(commandArgs, args...)

	cmd := exec.Command(p.Executable, commandArgs...)
	cmd.Env = os.Environ()

	if len(p.Env) > 0 {
		cmd.Env = append(cmd.Env, p.Env...)
	}

	stdOut, errPipe := cmd.StdoutPipe()

	if errPipe != nil {
		errorMessage := fmt.Sprintf("Unable to get the stdoutput for the pipe: %s", errPipe)
		return errors.New(errorMessage)
	}

	stdErr, errPipe2 := cmd.StderrPipe()

	if errPipe2 != nil {
		errorMessage := fmt.Sprintf("Unable to get the stdoutput for the pipe: %s", errPipe2)
		return errors.New(errorMessage)
	}

	err := cmd.Start()

	if err != nil {
		errorMessage := fmt.Sprintf("Unable to start python process: %s", err)
		return errors.New(errorMessage)
	}

	go copyFromTo(stdOut, os.Stdout)
	go copyFromTo(stdErr, os.Stderr)

	errRun := cmd.Wait()

	if errRun != nil {
		errorMessage := fmt.Sprintf("Python terminated prematurely: %s", errRun)
		return errors.New(errorMessage)
	}

	state := cmd.ProcessState
	if !state.Success() {
		errorMessage := fmt.Sprintf("Python did not completed sucessfully: %s", state.String())
		return errors.New(errorMessage)
	}

	return nil
}

func copyFromTo(r io.Reader, w io.Writer) {
	var copy []byte
	buf := make([]byte, 1024, 1024)
	for {
		_, errOut := writeOrDie(w, r, buf, copy)

		if errOut == io.EOF {
			return
		} else if errOut != nil {
			log.Fatalf("Unable to read the output with error: %v", errOut)
			return
		}
	}
}

func writeOrDie(w io.Writer, r io.Reader, buf []byte, out []byte) ([]byte, error) {

	n, err := r.Read(buf[:])
	if n > 0 {
		d := buf[:n]
		out = append(out, d...)
		_, err := w.Write(d)
		if err != nil {
			return out, err
		}
	}
	return out, err
}
