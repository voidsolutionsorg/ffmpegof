package ffmpeg

type HostMapping struct {
	Id           int
	Servername   string
	Hostname     string
	Weight       int
	CurrentState string
	MarkingPid   string
	Commands     []int
}
