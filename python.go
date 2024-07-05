package python

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"sync"
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

	err := receiveLogsAndErrorsFromPipe(cmd)
	if err != nil {
		errorMessage := fmt.Sprintf("Python terminated prematurely: %v", err)
		return errors.New(errorMessage)
	}

	state := cmd.ProcessState
	if !state.Success() {
		errorMessage := fmt.Sprintf("Python did not complete sucessfully: %s", state.String())
		return errors.New(errorMessage)
	}

	return nil
}

func receiveLogsAndErrorsFromPipe(cmd *exec.Cmd) error {
	rOut, _ := cmd.StdoutPipe()
	rErr, _ := cmd.StderrPipe()
	scannerOut := bufio.NewScanner(rOut)
	scannerErr := bufio.NewScanner(rErr)
	var wg sync.WaitGroup
	wg.Add(2)

	errChan := make(chan string, 1)
	errLine := ""

	go func() {
		for scannerOut.Scan() {
			line := scannerOut.Text()
			fmt.Println(line)
		}
		wg.Done()
	}()

	go func() {
		for scannerErr.Scan() {
			lineErr := scannerErr.Text()
			errLine = fmt.Sprintf("%s \n %s", errLine, lineErr)
		}
		errChan <- errLine
		wg.Done()
	}()

	err := cmd.Start()
	if err != nil {
		return err
	}

	wg.Wait()

	errPython := <-errChan
	if errPython != "" {
		return errors.New(errPython)
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
