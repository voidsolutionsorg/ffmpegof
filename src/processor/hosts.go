package processor

import (
	"errors"
	"fmt"

	"database/sql"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

func sqlUpsertHost(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `INSERT INTO hosts (servername, hostname, weight, created)
				VALUES (?, ?, ?, ?)
				ON CONFLICT (servername) DO UPDATE SET
				    hostname = excluded.hostname,
				    weight = excluded.weight,
				    created = excluded.created
				`, nil
	case "postgres":
		return `INSERT INTO hosts (servername, hostname, weight, created)
				VALUES ($1, $2, $3, $4)
				ON CONFLICT (servername) DO UPDATE SET
				    hostname = excluded.hostname,
				    weight = excluded.weight,
				    created = excluded.created
				`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) UpsertHost(host Host) error {
	sqlUpsertHost, err := sqlUpsertHost(store.dbType)
	if err != nil {
		return err
	}

	tx, err := store.Begin()
	if err != nil {
		return err
	}

	if _, err = tx.Exec(sqlUpsertHost, host.Servername, host.Hostname, host.Weight, host.Created); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			panic(rollbackErr)
		}
		return err
	}

	return tx.Commit()
}

// func sqlDeleteHosts(dbType string) (string, error) {
// 	switch dbType {
// 	case "sqlite":
// 		return `DELETE FROM hosts`, nil
// 	case "postgres":
// 		return `DELETE FROM hosts`, nil
// 	default:
// 		return "", fmt.Errorf("incorrect database type")
// 	}
// }

func (store *datastore) DeleteHosts() error {
	sqlDeleteHost, err := sqlDeleteHost(store.dbType)
	if err != nil {
		return err
	}

	_, err = store.Exec(sqlDeleteHost)
	if err != nil {
		return fmt.Errorf("delete hosts: %w", err)
	}

	return nil
}

func sqlDeleteHost(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `DELETE FROM hosts WHERE servername=?`, nil
	case "postgres":
		return `DELETE FROM hosts WHERE servername=$1`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) DeleteHost(host Host) error {
	sqlDeleteHost, err := sqlDeleteHost(store.dbType)
	if err != nil {
		return err
	}

	_, err = store.Exec(sqlDeleteHost, host.Servername)
	if err != nil {
		return fmt.Errorf("delete host: %w", err)
	}

	return nil
}

func sqlSelectCountHosts(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT COUNT(id) FROM hosts`, nil
	case "postgres":
		return `SELECT COUNT(id) FROM hosts`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectCountHosts() (int, error) {
	sqlSelectCountHosts, err := sqlSelectCountHosts(store.dbType)
	if err != nil {
		return 0, err
	}

	row := store.QueryRow(sqlSelectCountHosts)

	remaining := 0
	err = row.Scan(&remaining)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return remaining, nil
	case err != nil:
		return remaining, fmt.Errorf("select count hosts: %w", err)
	}

	return remaining, nil
}

func sqlSelectHosts(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT * FROM hosts ORDER BY created ASC`, nil
	case "postgres":
		return `SELECT * FROM hosts ORDER BY created ASC`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectHosts() (hosts []Host, err error) {
	sqlSelectHosts, err := sqlSelectHosts(store.dbType)
	if err != nil {
		return hosts, err
	}

	rows, err := store.Query(sqlSelectHosts)
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

func sqlSelectHostsWhere(dbType string) (string, error) {
	switch dbType {
	case "sqlite":
		return `SELECT %s FROM hosts WHERE %s=? ORDER BY created ASC`, nil
	case "postgres":
		return `SELECT %s FROM hosts WHERE %s=$1 ORDER BY created ASC`, nil
	default:
		return "", fmt.Errorf("incorrect database type")
	}
}

func (store *datastore) SelectHostsWhere(fieldType string, field string) (hosts []Host, err error) {
	sqlSelectHostsWhere, err := sqlSelectHostsWhere(store.dbType)
	if err != nil {
		return hosts, err
	}
	sqlSelectHostsWhere = fmt.Sprintf(sqlSelectHostsWhere, "*", fieldType)

	rows, err := store.Query(sqlSelectHostsWhere, field)
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

func (store *datastore) SelectHostsIdWhere(fieldType string, field string) (hosts []Host, err error) {
	sqlSelectHostsWhere, err := sqlSelectHostsWhere(store.dbType)
	if err != nil {
		return hosts, err
	}
	sqlSelectHostsWhere = fmt.Sprintf(sqlSelectHostsWhere, "id", fieldType)

	rows, err := store.Query(sqlSelectHostsWhere, field)
	if err != nil {
		return hosts, err
	}

	defer rows.Close()
	for rows.Next() {
		host := Host{}
		err = rows.Scan(&host.Id)
		if err != nil {
			return hosts, err
		}

		hosts = append(hosts, host)
	}

	return hosts, rows.Err()
}
