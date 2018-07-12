package probes

import (
	"bytes"
	"os/exec"
	"time"
)

type TracerouteExperiment struct {
	Target     string
	ResultChan chan *TracerouteResult
}

type TracerouteResult struct {
	Command string
	Target  string
	Args    []string
	Stdout  string
	Stderr  string
}

func (tre *TracerouteExperiment) Run() (ExperimentResult, error) {
	var cmd string
	var args []string
	var err error
	var result *TracerouteResult

	// Find the proper command for this system. Command varies across systems.
	// May not be on some limited systems (e.g., Arch base install).
	if cmd, err = exec.LookPath("traceroute"); err == nil {
		args = []string{tre.Target}
	} else if cmd, err = exec.LookPath("tracert.exe"); err == nil {
		args = []string{tre.Target}
	} else if cmd, err = exec.LookPath("tracert"); err == nil {
		args = []string{tre.Target}
	} else {
		cmd = ""
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
			result = &TracerouteResult{cmd, tre.Target, args, stdout.String(), stderr.String()}
		}
	} else {
		// Otherwise we can't do anything
		result = nil
	}
	tre.sendResponse(result)
	return result, err
}

func (tre *TracerouteExperiment) GetCost() int {
	return 3
}

func (tre *TracerouteExperiment) GetTable() string {
	return "traceroute"
}

func (tre *TracerouteExperiment) sendResponse(result *TracerouteResult) {
	if tre.ResultChan != nil {
		select {
		case tre.ResultChan <- result:
			break
		case <-time.After(time.Minute):
			break
		}
	}
}

func (tre *TracerouteExperiment) GetVersion() string {
	return "0.0.1"
}
