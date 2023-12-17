package processor

import (
	"database/sql"
	"time"

	"github.com/tminaorg/ffmpegof/src/migrate"
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
	store *datastore
	// processed int64
}

// version
func (p *Processor) GetVersion() (string, error) {
	return p.store.SelectVersion()
}

// hosts
func (p *Processor) AddHost(host Host) error {
	return p.store.UpsertHost(host)
}

func (p *Processor) RemoveHosts() error {
	return p.store.DeleteHosts()
}

func (p *Processor) RemoveHost(host Host) error {
	return p.store.DeleteHost(host)
}

func (p *Processor) NumberOfHosts() (int, error) {
	return p.store.SelectCountHosts()
}

func (p *Processor) GetHosts() ([]Host, error) {
	return p.store.SelectHosts()
}

func (p *Processor) GetHostsByField(field string, value string) ([]Host, error) {
	return p.store.SelectHostsWhere(field, value)
}

func (p *Processor) GetHostsIdByField(field string, value string) ([]Host, error) {
	return p.store.SelectHostsIdWhere(field, value)
}

// processes
func (p *Processor) AddProcess(process Process) error {
	return p.store.InsertProcess(process)
}

func (p *Processor) RemoveProcesses() error {
	return p.store.DeleteProcesses()
}

func (p *Processor) RemoveProcessesByField(field string, process Process) error {
	return p.store.DeleteProcessesWhere(field, process)
}

func (p *Processor) NumberOfProcesses() (int, error) {
	return p.store.SelectCountProcesses()
}

func (p *Processor) NumberOfProcessesFromHost(host Host) (int, error) {
	return p.store.SelectCountProcessesWhere(host)
}

func (p *Processor) GetProcesses() ([]Process, error) {
	return p.store.SelectProcesses()
}

func (p *Processor) GetProcessesId() ([]Process, error) {
	return p.store.SelectProcessesId()
}

func (p *Processor) GetProcessesFromHost(host Host) ([]Process, error) {
	return p.store.SelectProcessesWhere(host)
}

func (p *Processor) GetProcessesIdFromHost(host Host) ([]Process, error) {
	return p.store.SelectProcessesIdWhere(host)
}

// states
func (p *Processor) AddState(state State) error {
	return p.store.InsertState(state)
}

func (p *Processor) RemoveStates() error {
	return p.store.DeleteStates()
}

func (p *Processor) RemoveStatesByField(field string, state State) error {
	return p.store.DeleteStatesWhere(field, state)
}

func (p *Processor) NumberOfStates() (int, error) {
	return p.store.SelectCountStates()
}

func (p *Processor) NumberOfStatesFromHost(host Host) (int, error) {
	return p.store.SelectCountStatesWhere(host)
}

func (p *Processor) GetStates() ([]State, error) {
	return p.store.SelectStates()
}

func (p *Processor) GetStatesId() ([]State, error) {
	return p.store.SelectStatesId()
}

func (p *Processor) GetStatesFromHost(host Host) ([]State, error) {
	return p.store.SelectStatesWhere(host)
}

func (p *Processor) GetStatesIdFromHost(host Host) ([]State, error) {
	return p.store.SelectStatesIdWhere(host)
}
