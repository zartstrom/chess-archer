// Copyright 2013 Joe Walnes and the websocketd team.
// All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bufio"
	"io"
	"log"
    "os/exec"
	"syscall"
)

type LaunchedProcess struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
}

func launchCmd(commandName string, commandArgs []string, env []string) (*LaunchedProcess, error) {
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}

	err = cmd.Start()
	if err != nil {
		return nil, err
	}

	return &LaunchedProcess{cmd, stdin, stdout, stderr}, err
}

func NewProcessEndpoint(command string) *ProcessEndpoint {

    process, _ := launchCmd(command, []string{}, []string{})

	return &ProcessEndpoint{
		process:    process,
		bufferedIn: bufio.NewWriter(process.stdin),
		output:     make(chan string),
		input:      make(chan string),
		err:        make(chan bool),
    }
}

type ProcessEndpoint struct {
	process    *LaunchedProcess
	bufferedIn *bufio.Writer
	output     chan string
	input      chan string
	err        chan bool
	//log        *LogScope
}

func (pe *ProcessEndpoint) Output() chan string { return pe.output }
func (pe *ProcessEndpoint) Input() chan string { return pe.input }
func (pe *ProcessEndpoint) Err() chan bool { return pe.err }

func (pe *ProcessEndpoint) Terminate() {
	pe.process.stdin.Close()

	err := pe.process.cmd.Process.Signal(syscall.SIGINT)
	if err != nil {
		//pe.log.Debug("process", "Failed to Interrupt process %v: %s, attempting to kill", pe.process.cmd.Process.Pid, err)
		err = pe.process.cmd.Process.Kill()
		if err != nil {
            panic(err)
			//pe.log.Debug("process", "Failed to Kill process %v: %s", pe.process.cmd.Process.Pid, err)
		}
	}

	pe.process.cmd.Wait()
	if err != nil {
        panic(err)
		//pe.log.Debug("process", "Failed to reap process %v: %s", pe.process.cmd.Process.Pid, err)
	}
}

func (pe *ProcessEndpoint) Start() {
	go pe.log_stderr()
	go pe.process_stdout()
	go pe.process_stdin()
}

func (pe *ProcessEndpoint) process_stdin() {
    for msg := range pe.input {
        pe.bufferedIn.WriteString(msg)
        pe.bufferedIn.WriteString("\n")
        pe.bufferedIn.Flush()
    }
	close(pe.input)
}

func (pe *ProcessEndpoint) process_stdout() {
	bufin := bufio.NewReader(pe.process.stdout)
	for {
		str, err := bufin.ReadString('\n')
		if err != nil {
			if err != io.EOF {
                log.Println("process: Unexpected error while reading STDOUT from process: %s", err)
                panic(err)
			} else {
                log.Println("process: Process STDOUT closed")
                panic(err)
			}
			break
		}
		pe.output <- trimEOL(str)
	}
	close(pe.output)
}

func (pe *ProcessEndpoint) log_stderr() {
	bufstderr := bufio.NewReader(pe.process.stderr)
	for {
		//str, err := bufstderr.ReadString('\n')
		_, err := bufstderr.ReadString('\n')
		if err != nil {
			if err != io.EOF {
                panic(err)
				//pe.log.Error("process", "Unexpected error while reading STDERR from process: %s", err)
			} else {
                panic(err)
				//pe.log.Debug("process", "Process STDERR closed")
			}
			break
		}
		//pe.log.Error("stderr", "%s", trimEOL(str))
	}
}

// trimEOL cuts unixy style \n and windowsy style \r\n suffix from the string
func trimEOL(s string) string {
	lns := len(s)
	if lns > 0 && s[lns-1] == '\n' {
		lns--
		if lns > 0 && s[lns-1] == '\r' {
			lns--
		}
	}
	return s[0:lns]
}
