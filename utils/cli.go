/**
 * This software and associated documentation files (the “Software”),
 * including GFI AppManager, is the property of GFI USA, LLC and its affiliates.
 * No part of the Software may be copied, modified, distributed, sold, or otherwise
 * used except as expressly permitted by the terms of the software license agreement.
 */

package utils

import (
	"bytes"
	"context"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/trilogy-group/gfi-agent-sdk/constants"
	"github.com/trilogy-group/gfi-agent-sdk/logger"
)

const DefaultTimeout = 120 // 2 minutes

type CLI struct {
	Name string
	Args []string
	Envs map[string]string
	Cmd  *exec.Cmd
}

func NewCLIWithEnvs(name string, args []string, envs map[string]string) *CLI {
	return &CLI{
		Name: name,
		Args: args,
		Envs: envs,
	}
}

func NewCLI(name string, args []string) *CLI {
	return &CLI{Name: name, Args: args, Envs: map[string]string{}}
}

func (C *CLI) runWithTimeout(timeout int) (string, string, error) {
	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, C.Name, C.Args...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	for k, v := range C.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	C.Cmd = cmd
	err := C.Cmd.Run()
	C.Cmd = nil

	return strings.TrimSuffix(outb.String(), "\n"), strings.TrimSuffix(errb.String(), "\n"), err
}

func (C *CLI) Run(timeout ...int) (string, string, error) {
	if len(timeout) > 0 {
		return C.runWithTimeout(timeout[0])
	}
	return C.runWithTimeout(DefaultTimeout)
}

func GetFullPidPath(filename string) string {
	return filepath.Join(constants.GFIAgentDataDir, filename)
}

func (C *CLI) Start() error {
	return C.StartWithPidFile("")
}

func (C *CLI) KillPreviouslyRunProcess(pidPath string) error {
	pid, err := GetPidFileContents(pidPath)
	if err != nil {
		logger.Logger.Infof("No running process found with pid file %s: %s", pidPath, err.Error())
		return nil
	}

	if IsPidRunning(pid) {
		logger.Logger.Infof("Killing previously run process with pid %d", pid)
		err = KillProcess(pid)
		if err != nil {
			return err
		}
	}

	return nil
}

func (C *CLI) StartWithPidFile(pidFile string) error {
	pidPath := ""
	if len(pidFile) > 0 {
		pidPath = GetFullPidPath(pidFile)
		err := C.KillPreviouslyRunProcess(pidPath)
		if err != nil {
			logger.Logger.Errorf("Failed to kill process: %s", err.Error())
		}
	}

	cmd := exec.Command(C.Name, C.Args...)
	for k, v := range C.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	C.Cmd = cmd

	if len(pidPath) > 0 {
		err := WritePidFile(pidPath, cmd.Process.Pid)
		if err != nil {
			logger.Logger.Errorf("Failed to write pid file %s, error: %s", pidPath, err.Error())
		}
	}

	return nil
}

func (C *CLI) StartDetached() error {
	cmd := exec.Command(C.Name, C.Args...)
	cmd.SysProcAttr = GetDetachedStartAttributes()
	for k, v := range C.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	if err := cmd.Start(); err != nil {
		return err
	}
	C.Cmd = cmd

	return nil
}

func (C *CLI) Stop() error {
	if err := C.Cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}

func (C *CLI) StopWithPidFile(pidFile string) error {
	err := C.Stop()
	if err == nil {
		pidPath := GetFullPidPath(pidFile)
		err = RemovePidFile(pidPath)
		if err != nil {
			logger.Logger.Errorf("Failed to remove pid file %s, error: %s", pidFile, err.Error())
		}
	}
	return err
}

func (C *CLI) Wait() error {
	return C.Cmd.Wait()
}

func (C *CLI) PID() int {
	return C.Cmd.Process.Pid
}

func (C *CLI) Output() ([]byte, error) {
	cmd := exec.Command(C.Name, C.Args...)
	for k, v := range C.Envs {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	output, err := cmd.Output()
	if err != nil {
		return output, err
	}
	C.Cmd = cmd
	return output, nil
}
