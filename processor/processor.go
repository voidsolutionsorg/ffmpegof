package processor

import (
	"database/sql"
	"time"

	"github.com/aleksasiriski/rffmpeg-go/migrate"
)

type Config struct {
	Db     *sql.DB
	DbType string
	Mg     *migrate.Migrator
}

type Host struct {
	Id         int
	Servername string
	Hostname   string
	Weight     int
	Created    time.Time
}

type Process struct {
	Id        int
	HostId    int
	ProcessId int
	Cmd       string
}

type State struct {
	Id        int
	HostId    int
	ProcessId int
	State     string
}

func New(config Config) (*Processor, error) {
	store, err := newDatastore(config.Db, config.DbType, config.Mg)
	if err != nil {
		return nil, err
	}

	proc := &Processor{
		store: store,
	}
	return proc, nil
}

type Processor struct {
	store     *datastore
	processed int64
}

// hosts

func (p *Processor) AddHost(host Host) error {
	return p.store.UpsertHost(host)
}

func (p *Processor) RemoveHost(host Host) error {
	return p.store.DeleteHost(host)
}

func (p *Processor) NumberOfHosts() (int, error) {
	return p.store.GetHostsRemaining()
}

func (p *Processor) GetAllHosts() ([]Host, error) {
	return p.store.GetHosts()
}

// processes

func (p *Processor) NumberOfProcesses() (int, error) {
	return p.store.GetProcessesRemaining()
}

func (p *Processor) NumberOfProcessesFromHost(host Host) (int, error) {
	return p.store.GetProcessesRemainingWhere(host)
}

func (p *Processor) GetAllProcesses() ([]Process, error) {
	return p.store.GetProcesses()
}

func (p *Processor) GetAllProcessesFromHost(host Host) ([]Process, error) {
	return p.store.GetProcessesWhere(host)
}

// states

func (p *Processor) GetAllStatesFromHost(host Host) ([]State, error) {
	return p.store.GetStatesWhere(host)
}
