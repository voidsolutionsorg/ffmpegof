package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strings"
	"syscall"

	"github.com/aleksasiriski/rffmpeg-go/processor"
	"github.com/rs/zerolog/log"
	"github.com/sourcegraph/conc"
)

type HostMapping struct {
	Id           int
	Servername   string
	Hostname     string
	Weight       int
	CurrentState string
	MarkingPid   string
	Commands     []int
}

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

func removeFromSlice(slice []string, elemToRemove string) []string {
	for index, elem := range slice {
		if elem == elemToRemove {
			return append(slice[:index], slice[index+1:]...)
		}
	}
	return slice
}

func getTargetHost(config Config, proc *processor.Processor) (processor.Host, error) {
	targetHost := processor.Host{
		Id:         0,
		Servername: "localhost (fallback)",
		Hostname:   "localhost",
		Weight:     0,
	}

	hosts, err := proc.GetHosts()
	if err != nil || len(hosts) == 0 {
		return targetHost, err
	}

	hostMappings := make([]HostMapping, 0)

	for _, host := range hosts {
		states, err := proc.GetStatesFromHost(host)
		if err != nil {
			return targetHost, err
		}

		currentState := ""
		markingPid := ""
		if len(states) == 0 {
			currentState = "idle"
			markingPid = "N/A"
		} else {
			currentState = states[0].State
			markingPid = fmt.Sprintf("%d", states[0].ProcessId)
		}

		processes, err := proc.GetProcessesFromHost(host)
		if err != nil {
			return targetHost, err
		}

		commands := make([]int, 0)
		for _, process := range processes {
			commands = append(commands, process.ProcessId)
		}

		hostMappings = append(hostMappings, HostMapping{
			Id:           host.Id,
			Hostname:     host.Hostname,
			Weight:       host.Weight,
			Servername:   host.Servername,
			CurrentState: currentState,
			MarkingPid:   markingPid,
			Commands:     commands,
		})
	}

	lowestCount := 9999
	for _, hostMapping := range hostMappings {
		log.Debug().
			Msg(fmt.Sprintf("Trying host %s", hostMapping.Servername))

		if hostMapping.CurrentState == "bad" {
			log.Debug().
				Msg(fmt.Sprintf("Host previously marked bad by PID %s", hostMapping.MarkingPid))
			continue
		}
		if hostMapping.Hostname == "localhost" || hostMapping.Hostname == "127.0.0.1" {
			log.Debug().
				Msg("Running SSH test")

			testSshCommand := generateSshCommand(config, hostMapping.Hostname)
			testSshCommand = removeFromSlice(testSshCommand, "-q")
			testFfmpegCommand := config.Commands.Ffmpeg + "-version"
			testFullCommand := append(testSshCommand, testFfmpegCommand)
			testCommand := runCommand(testFullCommand, os.Stdin, os.Stdout, os.Stderr)
			err = testCommand.Run()
			if err != nil {
				// Mark the host as bad
				log.Warn().
					Str("command", strings.Join(testFullCommand, " ")).
					Msg(fmt.Sprintf("Marking host %s as bad due to: %w", hostMapping.Servername, err))
				err = proc.AddState(processor.State{
					HostId:    hostMapping.Id,
					ProcessId: config.Program.Pid,
					State:     "bad",
				})
				continue
			}
			log.Debug().
				Msg("SSH test succeeded")
		}

		// If the host state is idle, we can use it immediately
		if hostMapping.CurrentState == "idle" {
			targetHost.Id = hostMapping.Id
			targetHost.Servername = hostMapping.Servername
			targetHost.Hostname = hostMapping.Hostname
			log.Debug().
				Msg("Selecting host as idle")
			break
		}

		// Get the modified count of the host
		rawProcCount := len(hostMapping.Commands)
		weightedProcCount := rawProcCount / hostMapping.Weight

		// If this host is currently the least used, provisionally set it as the target
		if weightedProcCount < lowestCount {
			lowestCount = weightedProcCount
			targetHost.Id = hostMapping.Id
			targetHost.Servername = hostMapping.Servername
			targetHost.Hostname = hostMapping.Hostname
			log.Debug().
				Str("raw", fmt.Sprintf("%d", rawProcCount)).
				Str("weighted", fmt.Sprintf("%d", weightedProcCount)).
				Msg("Selecting host as current lowest proc count")
		}
	}

	log.Debug().
		Str("id", fmt.Sprintf("%d", targetHost.Id)).
		Str("servername", targetHost.Servername).
		Str("hostname", targetHost.Hostname).
		Msg("Found optimal host")
	return targetHost, err
}

func sliceContains(slice []string, elem string) bool {
	for _, iter := range slice {
		if iter == elem {
			return true
		}
	}
	return false
}

func runLocalFfmpeg(config Config, proc *processor.Processor, cmd string, args []string) (error, error, error) {
	rffmpegFfmpegCommand := make([]string, 0)

	// Prepare our default stdin/stdout/stderr
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	if strings.Contains(cmd, "ffprobe") {
		// If we're in ffprobe mode use that command and os.Stdout as stdout
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.Commands.Ffprobe)
	} else {
		// Otherwise, we use stderr as stdout
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.Commands.Ffmpeg)
		stdout = stderr
	}

	// Append all the passed arguments directly
	// Check for special flags that override the default stdout
	foundSpecialFlag := false
	for _, arg := range args {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, arg)

		if !foundSpecialFlag && sliceContains(config.Commands.SpecialFlags, arg) {
			stdout = os.Stdout
			foundSpecialFlag = true
		}
	}

	log.Info().
		Msg("Running command on localhost")

	log.Debug().
		Str("command", strings.Join(rffmpegFfmpegCommand, " ")).
		Msg("Localhost")

	fullCommand := cmd + " " + strings.Join(args, " ")
	errProcess := proc.AddProcess(processor.Process{
		HostId:    0,
		ProcessId: config.Program.Pid,
		Cmd:       fullCommand,
	})

	errState := proc.AddState(processor.State{
		HostId:    0,
		ProcessId: config.Program.Pid,
		State:     "active",
	})

	runnableCommand := runCommand(rffmpegFfmpegCommand, stdin, stdout, stderr)
	return runnableCommand.Run(), errProcess, errState
}

func runRemoteFfmpeg(config Config, proc *processor.Processor, cmd string, args []string, target processor.Host) (error, error, error) {
	rffmpegSshCommand := generateSshCommand(config, target.Hostname)
	rffmpegFfmpegCommand := make([]string, 0)

	// Add any pre commands
	for _, cmd := range config.Commands.Pre {
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, cmd)
	}

	// Prepare our default stdin/stdout/stderr
	stdin := os.Stdin
	stdout := os.Stdout
	stderr := os.Stderr

	if strings.Contains(cmd, "ffprobe") {
		// If we're in ffprobe mode use that command and os.Stdout as stdout
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.Commands.Ffprobe)
	} else {
		// Otherwise, we use stderr as stdout
		rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, config.Commands.Ffmpeg)
		stdout = stderr
	}

	// Append all the passed arguments with requoting of any problematic characters
	// Check for special flags that override the default stdout
	foundSpecialFlag := false
	re := regexp.MustCompile(`[*'()|\[\]\s]`)
	for _, arg := range args {
		// Match bad shell characters: * ' ( ) | [ ] or whitespace
		if re.Match([]byte(arg)) {
			rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, fmt.Sprintf("\"%s\"", arg))
		} else {
			rffmpegFfmpegCommand = append(rffmpegFfmpegCommand, arg)
		}

		if !foundSpecialFlag && sliceContains(config.Commands.SpecialFlags, arg) {
			stdout = os.Stdout
			foundSpecialFlag = true
		}
	}

	rffmpegFullCommand := append(rffmpegSshCommand, rffmpegFfmpegCommand...)

	log.Info().
		Msg(fmt.Sprintf("Running command on host %s", target.Servername))

	log.Debug().
		Str("command", strings.Join(rffmpegFullCommand, " ")).
		Msg("Remote")

	fullCommand := cmd + " " + strings.Join(args, " ")
	errProcess := proc.AddProcess(processor.Process{
		HostId:    target.Id,
		ProcessId: config.Program.Pid,
		Cmd:       fullCommand,
	})

	errState := proc.AddState(processor.State{
		HostId:    target.Id,
		ProcessId: config.Program.Pid,
		State:     "active",
	})

	runnableCommand := runCommand(rffmpegFullCommand, stdin, stdout, stderr)
	return runnableCommand.Run(), errProcess, errState
}

func runFfmpeg(config Config, proc *processor.Processor, cmd string, args []string) error {
	returnChannel := make(chan error, 1)
	var worker conc.WaitGroup
	worker.Go(func() {
		log.Info().
			Msg(fmt.Sprintf("Starting rffmpeg as %s with args: %s", cmd, strings.Join(args[:], " ")))

		target, err := getTargetHost(config, proc)
		if err != nil {
			log.Error().
				Err(err).
				Msg("Failed getting target host:")
		} else {
			ret := fmt.Errorf("not yet run")
			errProcess := fmt.Errorf("not yet run")
			errState := fmt.Errorf("not yet run")
			if target.Hostname == "localhost" || target.Hostname == "127.0.0.1" {
				ret, errProcess, errState = runLocalFfmpeg(config, proc, cmd, args)
			} else {
				ret, errProcess, errState = runRemoteFfmpeg(config, proc, cmd, args, target)
			}

			if errProcess != nil {
				log.Error().
					Err(errProcess).
					Msg("Failed adding process:")
			}
			if errState != nil {
				log.Error().
					Err(errState).
					Msg("Failed adding state:")
			}

			if ret != nil {
				log.Error().
					Err(ret).
					Msg("Finished rffmpeg with error:")
			} else {
				log.Info().
					Msg("Finished rffmpeg successfully")
			}
			returnChannel <- ret
		}
	})

	// handle interrupt signal
	quitChannel := make(chan os.Signal, 1)
	signal.Notify(quitChannel, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP)
	select {
	case <-quitChannel:
		{
			log.Warn().
				Msg("Forced quit executed")
		}
	case returnCode := <-returnChannel:
		{
			log.Info().
				Str("code", fmt.Sprintf("%w", returnCode)).
				Msg("Finished running command")
		}
	}

	errStates, errProcesses := cleanup(config.Program.Pid, proc)
	if errStates != nil {
		log.Error().
			Err(errStates).
			Msg("Error occured during cleanup of states:")
	}
	if errProcesses != nil {
		log.Error().
			Err(errProcesses).
			Msg("Error occured during cleanup of processes:")
	}

	return nil
}
