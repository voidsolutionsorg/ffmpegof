package control

import "github.com/tminaorg/ffmpegof/src/processor"

type Add struct {
	Name   string `help:"Name of the server." short:"n" optional:""`
	Weight int    `help:"Weight of the server." short:"w" default:"1" optional:""`
	Host   string `arg:"" name:"host" help:"Hostname or IP." required:""`
}

type Remove struct {
	Name string `arg:"" name:"name" help:"Name of the server." required:""`
}

type Clear struct {
	Name string `help:"Name of the server." short:"n" optional:""`
}

type Cli struct {
	Add    Add      `cmd:"" help:"Add host."`
	Remove Remove   `cmd:"" help:"Remove host."`
	Status struct{} `cmd:"" help:"Status of all hosts."`
	Clear  Clear    `cmd:"" help:"Clear processes and states."`
}

type StatusMapping struct {
	Id           string
	Servername   string
	Hostname     string
	Weight       string
	CurrentState string
	Commands     []processor.Process
}
