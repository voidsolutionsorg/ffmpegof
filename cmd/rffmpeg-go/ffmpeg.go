package main

import (
	"fmt"
	"os"
	//"os/signal"
	"io"
	"os/exec"
	//"strings"
	//"syscall"

	"github.com/aleksasiriski/rffmpeg-go/processor"
	"github.com/rs/zerolog/log"
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
	targetHost := processor.Host{}

	hosts, err := proc.GetHosts()
	if err != nil {
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
					Str("command", testFfmpegCommand). // testFullCommand
					Msg(fmt.Sprintf("Marking host %s as bad due to: %w", hostMapping.Servername, err))
				err = proc.AddState(processor.State{
					HostId:    hostMapping.Id,
					ProcessId: os.Getpid(),
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
