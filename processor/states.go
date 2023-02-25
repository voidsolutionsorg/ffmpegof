package processor

import (
	//"errors"
	"fmt"

	//"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func sqlGetStatesWhere(dbType string) string {
	switch dbType {
	case "sqlite":
		return `SELECT * FROM states WHERE host_id=?`
	case "postgres":
		return `SELECT * FROM states WHERE host_id=$1`
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
}

func (store *datastore) GetStatesWhere(host Host) (states []State, err error) {
	rows, err := store.Query(sqlGetStatesWhere(store.dbType), host.Id)
	if err != nil {
		return states, err
	}

	defer rows.Close()
	for rows.Next() {
		state := State{}
		err = rows.Scan(&state.Id, &state.HostId, &state.ProcessId, &state.State)
		if err != nil {
			return states, err
		}

		states = append(states, state)
	}

	return states, rows.Err()
}
