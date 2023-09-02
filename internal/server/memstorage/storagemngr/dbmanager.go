package storagemngr

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/server/config"
	_ "github.com/jackc/pgx/v4/stdlib"
)

type DataBaseManager struct {
	database *sql.DB
	log      *logger.BaseLogger
}

const (
	schemaName    = "metrics"
	gaugesTable   = "gauges"
	countersTable = "counters"
)

func CreateDataBaseManager(cfg *config.Config, log *logger.BaseLogger) (StorageManager, error) {
	(*log).Info("[storagemngr:CreateDataBaseManager] Open database with settings: '%s'", cfg.DataBaseDSN)
	database, err := sql.Open("pgx", cfg.DataBaseDSN)
	if err != nil {
		return nil, err
	}

	manager := &DataBaseManager{database: database, log: log}

	if !manager.CheckConnection() {
		return manager, fmt.Errorf("[storagemngr:CreateDataBaseManager] Database doesn't respond")
	}

	if err = manager.createSchema(); err != nil {
		return manager, err
	}

	if err = manager.createTables(); err != nil {
		return manager, err

	}

	return manager, nil
}

func (m *DataBaseManager) createSchema() error {
	if _, err := m.database.Exec(`CREATE SCHEMA IF NOT EXISTS ` + schemaName + `;`); err != nil {
		return err
	}

	if _, err := m.database.Exec(`SET search_path TO ` + schemaName); err != nil {
		return err
	}

	return nil
}

func (m *DataBaseManager) createTables() error {
	if _, err := m.database.Exec(`CREATE TABLE IF NOT EXISTS ` + gaugesTable + ` (id TEXT PRIMARY KEY, value float8);`); err != nil {
		return err
	}

	if _, err := m.database.Exec(`CREATE TABLE IF NOT EXISTS ` + countersTable + ` (id TEXT PRIMARY KEY, value int8);`); err != nil {
		return err
	}

	return nil
}

func (m *DataBaseManager) Close() error {
	return m.database.Close()
}

func (m *DataBaseManager) CheckConnection() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := m.database.PingContext(ctx); err != nil {
		return false
	}
	return true
}

func (m *DataBaseManager) SaveMetricsInStorage(gaugesValues map[string]float64, countersValues map[string]int64) {
	for key, value := range gaugesValues {
		if err := m.saveMetric(key, value); err != nil {
			(*m.log).Info("[DataBaseManager:SaveMetricsInStorage] Failed to save gauge metric in database name: %s, value: %v, error: %s",
				key, value, err)
		}
	}

	for key, value := range countersValues {
		if err := m.saveMetric(key, value); err != nil {
			(*m.log).Info("[DataBaseManager:SaveMetricsInStorage] Failed to save counter metric in database name: %s, value: %v, error: %s",
				key, value, err)
		}
	}
}

func (m *DataBaseManager) RestoreDataFromStorage() (map[string]float64, map[string]int64) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	if err := m.restoreDataInMap(gaugesTable, gauges); err != nil {
		(*m.log).Info("[DataBaseManager:RestoreDataFromStorage] Failed to get gauge metrics data from database (error: %s)", err)
	}

	if err := m.restoreDataInMap(countersTable, counters); err != nil {
		(*m.log).Info("[DataBaseManager:RestoreDataFromStorage] Failed to get gauge metrics data from database (with error: %s)", err)
	}

	return gauges, counters
}

func (m *DataBaseManager) restoreDataInMap(tableName string, mapDest interface{}) error {
	rows, err := m.database.Query(`SELECT * FROM ` + tableName)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		switch dest := mapDest.(type) {
		case map[string]float64:
			var name string
			var value float64
			err = rows.Scan(&name, &value)
			if err != nil {
				return err
			}

			dest[name] = value
		case map[string]int64:
			var name string
			var value int64
			err = rows.Scan(&name, &value)
			if err != nil {
				return err
			}

			dest[name] = value
		default:
			panic("Unknown value type")
		}
	}

	err = rows.Err()
	return err
}

func (m *DataBaseManager) saveMetric(name string, value interface{}) error {
	exists, err := m.checkMetricExists(name, value)
	if err != nil {
		return err
	}

	if exists {
		err = m.updateMetric(name, value)
	} else {
		err = m.insertMetric(name, value)
	}

	return err
}

func (m *DataBaseManager) checkMetricExists(name string, value interface{}) (bool, error) {
	var exists bool
	requestBody := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1);`, getMetricsTableName(value))
	err := m.database.QueryRow(requestBody, name).Scan(&exists)
	return exists, err
}

func (m *DataBaseManager) insertMetric(name string, value interface{}) error {
	requestBody := fmt.Sprintf(`INSERT INTO %s (id, value) VALUES ($1, $2);`, getMetricsTableName(value))
	var err error
	if getMetricsTableName(value) == gaugesTable {
		_, err = m.database.Exec(requestBody, name, value.(float64))
	} else {
		_, err = m.database.Exec(requestBody, name, value.(int64))
	}
	return err
}

func (m *DataBaseManager) updateMetric(name string, value interface{}) error {
	requestBody := fmt.Sprintf(`UPDATE %s SET value = $1 WHERE id = $2;`, getMetricsTableName(value))
	var err error
	if getMetricsTableName(value) == gaugesTable {
		_, err = m.database.Exec(requestBody, value.(float64), name)
	} else {
		_, err = m.database.Exec(requestBody, value.(int64), name)
	}
	return err
}

func getMetricsTableName(value interface{}) string {
	switch value.(type) {
	case int64:
		return countersTable
	case float64:
		return gaugesTable
	default:
		//TODO is it correct or better to handle wrong type?
		panic("Unknown value type")
	}
}
