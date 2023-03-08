package processor

import (
	"errors"
	"fmt"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func sqlInsertState(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `INSERT INTO states (host_id, process_id, state) VALUES (?, ?, ?) `, nil
	case "postgres":
		return `INSERT INTO states (host_id, process_id, state) VALUES ($1, $2, $3) `, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) InsertState(state State) error {
	sqlInsertState, err := sqlInsertState(store.dbType)
	if err != nil {
		return err
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(sqlInsertState, state.HostId, state.ProcessId, state.State); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic(rollbackErr)
		}
		return err
	}

	return tx.Commit()
}

func sqlDeleteStates(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `DELETE FROM states`, nil
	case "postgres":
		return `DELETE FROM states`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) DeleteStates() error {
	sqlDeleteState, err := sqlDeleteStates(store.dbType)
	if err != nil {
		return err
	}

	_, err = store.Exec(sqlDeleteState)
	if err != nil {
		return fmt.Errorf("delete states: %w", err)
	}

	return nil
}

func sqlDeleteStatesWhere(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `DELETE FROM states WHERE %s=?`, nil
	case "postgres":
		return `DELETE FROM states WHERE %s=$1`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) DeleteStatesWhere(field string, state State) error {
	sqlDeleteStatesWhere, err := sqlDeleteStatesWhere(store.dbType)
	if err != nil {
		return err
	}
	sqlDeleteStatesWhere = fmt.Sprintf(sqlDeleteStatesWhere, field)

	if state.Id != 0 {
		_, err = store.Exec(sqlDeleteStatesWhere, state.Id)
	} else if state.ProcessId != 0 {
		_, err = store.Exec(sqlDeleteStatesWhere, state.ProcessId)
	}

	if err != nil {
		return fmt.Errorf("delete state: %w", err)
	}

	return nil
}

func sqlSelectCountStates(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT COUNT(id) FROM states`, nil
	case "postgres":
		return `SELECT COUNT(id) FROM states`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectCountStates() (int, error) {
	sqlSelectCountStates, err := sqlSelectCountStates(store.dbType)
	if err != nil {
		return 0, err
	}

	row := store.QueryRow(sqlSelectCountStates)

	remaining := 0
	err = row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("select count states: %w", err)
	}

	return remaining, nil
}

func sqlSelectCountStatesWhere(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT COUNT(id) FROM states WHERE host_id=?`, nil
	case "postgres":
		return `SELECT COUNT(id) FROM states WHERE host_id=$1`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectCountStatesWhere(host Host) (int, error) {
	sqlSelectCountStatesWhere, err := sqlSelectCountStatesWhere(store.dbType)
	if err != nil {
		return 0, err
	}

	row := store.QueryRow(sqlSelectCountStatesWhere, host.Id)

	remaining := 0
	err = row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("select count states where: %w", err)
	}

	return remaining, nil
}

func sqlSelectStates(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT %s FROM states`, nil
	case "postgres":
		return `SELECT %s FROM states`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectStates() (states []State, err error) {
	sqlSelectStates, err := sqlSelectStates(store.dbType)
	if err != nil {
		return states, err
	}
	sqlSelectStates = fmt.Sprintf(sqlSelectStates, "*")

	rows, err := store.Query(sqlSelectStates)
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

func (store *datastore) SelectStatesId() (states []State, err error) {
	sqlSelectStates, err := sqlSelectStates(store.dbType)
	if err != nil {
		return states, err
	}
	sqlSelectStates = fmt.Sprintf(sqlSelectStates, "id")

	rows, err := store.Query(sqlSelectStates)
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

func sqlSelectStatesWhere(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT %s FROM states WHERE host_id=? ORDER BY id DESC`, nil
	case "postgres":
		return `SELECT %s FROM states WHERE host_id=$1 ORDER BY id DESC`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectStatesWhere(host Host) (states []State, err error) {
	sqlSelectStatesWhere, err := sqlSelectStatesWhere(store.dbType)
	if err != nil {
		return states, err
	}
	sqlSelectStatesWhere = fmt.Sprintf(sqlSelectStatesWhere, "*")

	rows, err := store.Query(sqlSelectStatesWhere, host.Id)
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

func (store *datastore) SelectStatesIdWhere(host Host) (states []State, err error) {
	sqlSelectStatesWhere, err := sqlSelectStatesWhere(store.dbType)
	if err != nil {
		return states, err
	}
	sqlSelectStatesWhere = fmt.Sprintf(sqlSelectStatesWhere, "id")

	rows, err := store.Query(sqlSelectStatesWhere, host.Id)
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
