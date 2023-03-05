package main

import (
	"fmt"
	"os"
	"time"

	"github.com/alecthomas/kong"
	"github.com/aleksasiriski/rffmpeg-go/processor"
	"github.com/rs/zerolog/log"
)

type Add struct {
	Name   string `help:"Name of the server." short:"n" optional:""`
	Weight int    `help:"Weight of the server." short:"w" default:"1" optional:""`
	Host   string `arg:"" name:"host" help:"Hostname or IP." required:""`
}

type Rm struct {
	Name string `arg:"" name:"host" help:"Name of the server." required:""`
}

type Cli struct {
	Add    Add      `cmd:"" help:"Add host."`
	Rm     Rm       `cmd:"" help:"Remove host."`
	Status struct{} `cmd:"" help:"Status of all hosts."`
}

func addHost(config Config, proc *processor.Processor, info Add) error {
	if info.Name == "" {
		info.Name = info.Host
	}

	err := proc.AddHost(processor.Host{
		Servername: info.Name,
		Hostname:   info.Host,
		Weight:     info.Weight,
		Created:    time.Now(),
	})

	return err
}

func removeHost(config Config, proc *processor.Processor, info Rm) error {
	err := proc.RemoveHost(processor.Host{
		Servername: info.Name,
	})

	return err
}

type StatusMapping struct {
	Id           int
	Servername   string
	Hostname     string
	Weight       int
	CurrentState string
	Commands     []processor.Process
}

func status(config Config, proc *processor.Processor) error {
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
			Id:           0,
			Servername:   "localhost (fallback)",
			Hostname:     "localhost (fallback)",
			Weight:       0,
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
			Id:           host.Id,
			Servername:   host.Servername,
			Hostname:     host.Hostname,
			Weight:       host.Weight,
			CurrentState: currentState,
			Commands:     processes,
		})
	}

	fmt.Println(statusMappings)

	return err
}

func runControl(config Config, proc *processor.Processor) {
	// parse cli
	cli := Cli{}

	ctx := kong.Parse(&cli,
		kong.Name("rffmpeg"),
		kong.Description("Remote ffmpeg"),
		kong.UsageOnError(),
		kong.ConfigureHelp(kong.HelpOptions{
			Summary: true,
			Compact: true,
		}),
	)

	if err := ctx.Validate(); err != nil {
		log.Error().
			Err(err).
			Msg("Failed parsing cli:")
		os.Exit(1)
	}

	// functions based on arguments
	switch ctx.Command() {
	case "add <host>":
		{
			err := addHost(config, proc, cli.Add)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed adding host:")
			} else {
				log.Info().
					Msg("Succesfully added host")
			}
		}
	case "rm <host>":
		{
			err := removeHost(config, proc, cli.Rm)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed removing host:")
			} else {
				log.Info().
					Msg("Succesfully removed host")
			}
		}
	case "status":
		{
			err := status(config, proc)
			if err != nil {
				log.Error().
					Err(err).
					Msg("Failed reading status:")
			}
		}
	default:
		{
			log.Fatal().
				Err(fmt.Errorf("%s", ctx.Command())).
				Msg("Invalid command:")
		}
	}
}
