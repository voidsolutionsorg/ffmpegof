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

func (p *Processor) GetHostsByField(field string) ([]Host, error) {
	return p.store.SelectHostsWhere(field)
}

func (p *Processor) GetHostsIdByField(field string) ([]Host, error) {
	return p.store.SelectHostsIdWhere(field)
}

// processes
func (p *Processor) AddProcess(process Process) error {
	return p.store.InsertProcess(process)
}

func (p *Processor) RemoveProcesses() error {
	return p.store.DeleteProcesses()
}

func (p *Processor) RemoveProcess(process Process) error {
	return p.store.DeleteProcess(process)
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

func (p *Processor) GetProcessesByField(field string) ([]Process, error) {
	return p.store.SelectProcessesWhere(field)
}

func (p *Processor) GetProcessesIdByField(field string) ([]Process, error) {
	return p.store.SelectProcessesIdWhere(field)
}

// states
func (p *Processor) AddState(state State) error {
	return p.store.InsertState(state)
}

func (p *Processor) RemoveStates() error {
	return p.store.DeleteStates()
}

func (p *Processor) RemoveState(state State) error {
	return p.store.DeleteState(state)
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

func (p *Processor) GetStatesByField(field string) ([]State, error) {
	return p.store.SelectStatesWhere(field)
}

func (p *Processor) GetStatesIdByField(field string) ([]State, error) {
	return p.store.SelectStatesIdWhere(field)
}
