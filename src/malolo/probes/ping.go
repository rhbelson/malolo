package probes

import (
	"bytes"
	"os/exec"
	"time"
)

type PingExperiment struct {
	Target     string
	ResultChan chan *PingResult
}

type PingResult struct {
	Command string
	Target  string
	Args    []string
	Stdout  string
	Stderr  string
}

func (pe *PingExperiment) Run() (ExperimentResult, error) {
	var cmd string
	var args []string
	var err error
	var result *PingResult

	// Find the proper command for this system.
	if cmd, err = exec.LookPath("ping.exe"); err == nil {
		args = []string{"-n", "1", pe.Target}
	} else if cmd, err = exec.LookPath("ping"); err == nil {
		args = []string{"-c", "1", pe.Target}
	}

	if cmd != "" {
		// If a traceroute command was found, run it.
		var stdout bytes.Buffer
		var stderr bytes.Buffer
		c := exec.Command(cmd, args...)
		c.Stdout = &stdout
		c.Stderr = &stderr
		err = c.Run()
		if err != nil {
			result = nil
		} else {
			result = &PingResult{cmd, pe.Target, args, stdout.String(), stderr.String()}
		}
	} else {
		// Otherwise we can't do anything
		result = nil
	}
	pe.sendResponse(result)
	return result, err
}

func (pe *PingExperiment) GetCost() int {
	return 1
}

func (pe *PingExperiment) GetTable() string {
	return "ping"
}

func (pe *PingExperiment) sendResponse(result *PingResult) {
	if pe.ResultChan != nil {
		select {
		case pe.ResultChan <- result:
			break
		case <-time.After(time.Minute):
			break
		}
	}
}

func (pe *PingExperiment) GetVersion() string {
	return "0.0.1"
}
