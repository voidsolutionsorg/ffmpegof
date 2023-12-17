package control

import (
	"fmt"
	"time"

	"github.com/alecthomas/kong"
	"github.com/rs/zerolog/log"
	"github.com/tminaorg/ffmpegof/src/processor"
)

func addHost(proc *processor.Processor, info Add) error {
	if info.Name == "" {
		info.Name = info.Host
	}

	return proc.AddHost(processor.Host{
		Servername: info.Name,
		Hostname:   info.Host,
		Weight:     info.Weight,
		Created:    time.Now(),
	})
}

func removeHost(proc *processor.Processor, info Remove) error {
	return proc.RemoveHost(processor.Host{
		Servername: info.Name,
	})
}

func printStatus(statusMappings []StatusMapping) {
	servernameLen := 11
	hostnameLen := 9
	idLen := 3
	weightLen := 7
	stateLen := 6
	for _, statusMapping := range statusMappings {
		if len(statusMapping.Servername)+1 > servernameLen {
			servernameLen = len(statusMapping.Servername) + 1
		}
		if len(statusMapping.Hostname)+1 > hostnameLen {
			hostnameLen = len(statusMapping.Hostname) + 1
		}
		if len(statusMapping.Id)+1 > idLen {
			idLen = len(statusMapping.Id) + 1
		}
		if len(statusMapping.Weight)+1 > weightLen {
			weightLen = len(statusMapping.Weight) + 1
		}
		if len(statusMapping.CurrentState)+1 > stateLen {
			stateLen = len(statusMapping.CurrentState) + 1
		}
	}

	fmt.Printf("%-s%-*s %-*s %-*s %-*s %-*s %-s%-s\n",
		"\033[1m",
		servernameLen,
		"Servername",
		hostnameLen,
		"Hostname",
		idLen,
		"ID",
		weightLen,
		"Weight",
		stateLen,
		"State",
		"Active Commands",
		"\033[0m",
	)

	for _, statusMapping := range statusMappings {
		firstCommand := "N/A"
		if len(statusMapping.Commands) > 0 {
			firstCommand = fmt.Sprintf("PID %-d: %-s", statusMapping.Commands[0].ProcessId, statusMapping.Commands[0].Cmd)
		}

		fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-s\n",
			servernameLen,
			statusMapping.Servername,
			hostnameLen,
			statusMapping.Hostname,
			idLen,
			statusMapping.Id,
			weightLen,
			statusMapping.Weight,
			stateLen,
			statusMapping.CurrentState,
			firstCommand,
		)

		if firstCommand != "N/A" {
			for index, command := range statusMapping.Commands {
				if index != 0 {
					formattedCommand := fmt.Sprintf("PID %d: %s", command.ProcessId, command.Cmd)
					fmt.Printf("%-*s %-*s %-*s %-*s %-*s %-s\n",
						servernameLen,
						"",
						hostnameLen,
						"",
						idLen,
						"",
						weightLen,
						"",
						stateLen,
						"",
						formattedCommand,
					)
				}
			}
		}
	}
}

func status(proc *processor.Processor) error {
	hosts, err := proc.GetHosts()
	if err != nil {
		return err
	}

	// Determine if there are any fallback processes running
	fallbackProcesses, err := proc.GetProcessesFromHost(processor.Host{
		Id: 0,
	})
	if err != nil {
		return err
	}

	// Generate a mapping dictionary of hosts and processes
	statusMappings := make([]StatusMapping, 0)

	if len(fallbackProcesses) > 0 {
		statusMappings = append(statusMappings, StatusMapping{
			Id:           "0",
			Servername:   "localhost (fallback)",
			Hostname:     "localhost",
			Weight:       "0",
			CurrentState: "fallback",
			Commands:     fallbackProcesses,
		})
	}

	for _, host := range hosts {
		// Get the latest state
		states, err := proc.GetStatesFromHost(host)
		if err != nil {
			return err
		}

		currentState := ""
		if len(states) == 0 {
			currentState = "idle"
		} else {
			currentState = states[0].State
		}

		// Get processes from host
		processes, err := proc.GetProcessesFromHost(host)
		if err != nil {
			return err
		}

		// Create the mappings entry
		statusMappings = append(statusMappings, StatusMapping{
			Id:           fmt.Sprintf("%d", host.Id),
			Servername:   host.Servername,
			Hostname:     host.Hostname,
			Weight:       fmt.Sprintf("%d", host.Weight),
			CurrentState: currentState,
			Commands:     processes,
		})
	}

	log.Info().Msg("Outputting status of hosts")
	printStatus(statusMappings)

	return err
}

func clear(proc *processor.Processor, info Clear) (error, error) {
	if info.Name != "" {
		hosts, err := proc.GetHostsIdByField("servername", info.Name)
		if err != nil {
			return err, err
		} else {
			return proc.RemoveProcessesByField("host_id", processor.Process{
					HostId: hosts[0].Id,
				}), proc.RemoveStatesByField("host_id", processor.State{
					HostId: hosts[0].Id,
				})
		}
	} else {
		return proc.RemoveProcesses(), proc.RemoveStates()
	}
}

func Run(proc *processor.Processor) {
	// parse cli
	cli := Cli{}

	ctx := kong.Parse(&cli,
		kong.Name("ffmpegof"),
		kong.Description("FFmpeg over Fabrics"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)

	if err := ctx.Validate(); err != nil {
		log.Fatal().Err(err).Msg("failed parsing cli")
	}

	// functions based on arguments
	switch ctx.Command() {
	case "add <host>":
		{
			err := addHost(proc, cli.Add)
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed adding host")
			} else {
				log.Info().
					Msg("succesfully added host")
			}
		}
	case "remove <name>":
		{
			err := removeHost(proc, cli.Remove)
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed removing host")
			} else {
				log.Info().
					Msg("succesfully removed host")
			}
		}
	case "status":
		{
			err := status(proc)
			if err != nil {
				log.Error().
					Err(err).
					Msg("failed reading status")
			}
		}
	case "clear":
		{
			errProcess, errState := clear(proc, cli.Clear)
			if errProcess != nil {
				log.Error().
					Err(errProcess).
					Msg("failed clearing processes")
			} else if errState != nil {
				log.Error().
					Err(errState).
					Msg("failed clearing states")
			} else {
				log.Info().
					Msg("succesfully cleared processes and states")
			}
		}
	default:
		{
			log.Fatal().
				Err(fmt.Errorf("%s", ctx.Command())).
				Msg("invalid command")
		}
	}
}
