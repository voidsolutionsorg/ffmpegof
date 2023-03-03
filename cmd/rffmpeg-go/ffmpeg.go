package main

import (
	"fmt"
	//"os"
	//"os/signal"
	"os/exec"
	"io"
	//"strings"
	//"syscall"

	"github.com/aleksasiriski/rffmpeg-go/processor"
)

// signum="", frame=""
func cleanup(pid int, proc *processor.Processor) (error, error) {
	errStates := proc.RemoveStatesByPid(pid)
	errProcesses := proc.RemoveProcessesByPid(pid)
	return errStates, errProcesses
}

func generateSshCommand(config Config, targetHostname string) []string {
	sshCommand := make([]string, 0)

	// Add SSH component
	sshCommand = append(sshCommand, config.Commands.Ssh)
	sshCommand = append(sshCommand, "-q")
	sshCommand = append(sshCommand, "-t")

	// Set our connection details
	sshCommand = append(sshCommand, []string{"-o", "ConnectTimeout=1"}...)
	sshCommand = append(sshCommand, []string{"-o", "ConnectionAttempts=1"}...)
	sshCommand = append(sshCommand, []string{"-o", "StrictHostKeyChecking=no"}...)
	sshCommand = append(sshCommand, []string{"-o", "UserKnownHostsFile=/dev/null"}...)

	// Use SSH control persistence to keep sessions alive for subsequent commands
	if config.Remote.Persist > 0 {
		sshCommand = append(sshCommand, []string{"-o", "ControlMaster=auto"}...)
		sshCommand = append(sshCommand, []string{"-o", fmt.Sprintf("ControlPath=%s/ssh-%r@%h:%p", config.Directories.Persist)}...)
		sshCommand = append(sshCommand, []string{"-o", fmt.Sprintf("ControlPersist=%s", config.Remote.Persist)}...)
	}

	// Add the remote config args
	sshCommand = append(sshCommand, config.Remote.Args...)

	// Add user+host string
	sshCommand = append(sshCommand, fmt.Sprintf("%s@%s", config.Remote.User, targetHostname))

	return sshCommand
}

func runCommand(commandArray []string, stdin io.Reader, stdout io.Writer, stderr io.Writer) *exec.Cmd {
	commandName := commandArray[0]
	commandArgs := commandArray[1:]
	cmd := exec.Command(commandName, commandArgs...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd
}

/*func runRemoteFfmpeg(config Config, proc *processor.Processor, cmd string, args []string, target processor.Host) int {
	rffmpegSshCommand := generateSshCommand(config, target.Hostname)
	rffmpegFfmpegCommand := make([]string, 0)

	for _, cmd := range config.PreCommands {
		if cmd != nil {
			rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, cmd)
		}
	}

	stdin := os.Stdin
	stderr := os.Stderr

	if strings.Contains(cmd, "ffprobe") {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.FfprobeCommand)
		stdout := os.Stdout
	} else {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.FfmpegCommand)
		stdout := os.Stderr
	}
}


func runFfmpeg(config Config, proc *processor.Processor, cmd string, args []string) error {
	var worker conc.WaitGroup
	worker.Go(func() {
		log.Info().
			Msg(fmt.Sprintf("Starting rffmpeg as %s with args: %s", cmd, strings.Join(args[:], " ")))

		target, err := getTargetHost(config, proc)
		if err != nil {
			return err
		}

		if target.Hostname == nil || target.Hostname == "localhost" {
			ret := runLocalFfmpeg(config, proc, cmd, args)
		} else {
			ret := runRemoteFfmpeg(config, proc, cmd, args, target)
		}

		cleanup()
		if ret.returncode == 0 {
			log.Info().
				Str("returncode", ret.returncode).
				Msg("Finished rffmpeg")
		} else {
			log.Error().
				Str("returncode", ret.returncode).
				Msg("Finished rffmpeg with return code ")
		}
		exit(ret.returncode)
	})

	// handle interrupt signal
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	<-quitChannel
	cleanup()
	return nil
}*/
