package processor

import (
	"errors"
	"fmt"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

const sqlGetProcessesRemaining = `SELECT COUNT(id) FROM processes`

func (store *datastore) GetProcessesRemaining() (int, error) {
	row := store.QueryRow(sqlGetProcessesRemaining)

	remaining := 0
	err := row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("get remaining processes: %w", err)
	}

	return remaining, nil
}

func sqlGetProcessesRemainingWhere(dbType string) string {
	switch dbType {
	case "sqlite":
		return `SELECT COUNT(id) FROM processes WHERE host_id=?`
	case "postgres":
		return `SELECT COUNT(id) FROM processes WHERE host_id=$1`
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
}

func (store *datastore) GetProcessesRemainingWhere(host Host) (int, error) {
	row := store.QueryRow(sqlGetProcessesRemainingWhere(store.dbType), host.Id)

	remaining := 0
	err := row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("get remaining processes where: %w", err)
	}

	return remaining, nil
}

const sqlGetProcesses = `SELECT * FROM processes`

func (store *datastore) GetProcesses() (processes []Process, err error) {
	rows, err := store.Query(sqlGetProcesses)
	if err != nil {
		return processes, err
	}

	defer rows.Close()
	for rows.Next() {
		process := Process{}
		err = rows.Scan(&process.Id, &process.HostId, &process.ProcessId, &process.Cmd)
		if err != nil {
			return processes, err
		}

		processes = append(processes, process)
	}

	return processes, rows.Err()
}

func sqlGetProcessesWhere(dbType string) string {
	switch dbType {
	case "sqlite":
		return `SELECT * FROM processes WHERE host_id=?`
	case "postgres":
		return `SELECT * FROM processes WHERE host_id=$1`
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
}

func (store *datastore) GetProcessesWhere(host Host) (processes []Process, err error) {
	rows, err := store.Query(sqlGetProcessesWhere(store.dbType), host.Id)
	if err != nil {
		return processes, err
	}

	defer rows.Close()
	for rows.Next() {
		process := Process{}
		err = rows.Scan(&process.Id, &process.HostId, &process.ProcessId, &process.Cmd)
		if err != nil {
			return processes, err
		}

		processes = append(processes, process)
	}

	return processes, rows.Err()
}
