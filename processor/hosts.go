package processor

import (
	"errors"
	"fmt"

	"database/sql"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func sqlUpsertHost(dbType string) string {
	switch dbType {
	case "sqlite":
		return `INSERT INTO hosts (servername, hostname, weight, created)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (servername) DO UPDATE SET
				    hostname = excluded.hostname,
				    weight = excluded.weight,
				    created = excluded.created
				`
	case "postgres":
		return `INSERT INTO hosts (servername, hostname, weight, created)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (servername) DO UPDATE SET
				    hostname = excluded.hostname,
				    weight = excluded.weight,
				    created = excluded.created
				`
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
}

func (store *datastore) UpsertHost(host Host) error {
	tx, err := store.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(sqlUpsertHost(store.dbType), host.Servername, host.Hostname, host.Weight, host.Created); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic(rollbackErr)
		}
		return err
	}

	return tx.Commit()
}

func sqlDeleteHost(dbType string) string {
	switch dbType {
	case "sqlite":
		return `DELETE FROM hosts WHERE servername=?`
	case "postgres":
		return `DELETE FROM hosts WHERE servername=$1`
	default:
		panic(fmt.Errorf("Incorrect database type!"))
	}
}

func (store *datastore) DeleteHost(host Host) error {
	_, err := store.Exec(sqlDeleteHost(store.dbType), host.Servername)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}

	return nil
}

const sqlGetHostsRemaining = `SELECT COUNT(id) FROM hosts`

func (store *datastore) GetHostsRemaining() (int, error) {
	row := store.QueryRow(sqlGetHostsRemaining)

	remaining := 0
	err := row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("get remaining hosts: %w", err)
	}

	return remaining, nil
}

const sqlGetHosts = `SELECT * FROM hosts ORDER BY created ASC`

func (store *datastore) GetHosts() (hosts []Host, err error) {
	rows, err := store.Query(sqlGetHosts)
	if err != nil {
		return hosts, err
	}

	defer rows.Close()
	for rows.Next() {
		host := Host{}
		err = rows.Scan(&host.Id, &host.Servername, &host.Hostname, &host.Weight, &host.Created)
		if err != nil {
			return hosts, err
		}

		hosts = append(hosts, host)
	}

	return hosts, rows.Err()
}
