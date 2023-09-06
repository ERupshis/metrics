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

const (
	schemaName    = "metrics"
	gaugesTable   = "gauges"
	countersTable = "counters"

	insertStmt = "insert"
	updateStmt = "update"
	existStmt  = "exist"

	createDatabaseError = "create db: %w"
	createSchemaError   = "create db schema: %w"
	createTableError    = "create db table: %w"

	saveMetricsError           = "save metrics in db: %w"
	saveMetricsCreateStmtError = "create stmt: %w"
	saveMetricsUseStmtError    = "use stmt: %w"
	saveMetricError            = "save metric: %w"

	restoreMetricsError = "restore metrics from db: %w"
	restoreDataError    = "restore data from db response: %w"

	insertMetricError = "insert new metric name '%s', value '%s', table '%v': %w"
	updateMetricError = "update metric name '%s', value '%s', table '%v': %w"
)

type DataBaseManager struct {
	database *sql.DB
	log      logger.BaseLogger
}

func CreateDataBaseManager(cfg *config.Config, log logger.BaseLogger) (StorageManager, error) {
	log.Info("[storagemngr:CreateDataBaseManager] Open database with settings: '%s'", cfg.DataBaseDSN)
	database, err := sql.Open("pgx", cfg.DataBaseDSN)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	manager := &DataBaseManager{database: database, log: log}
	if _, err = manager.CheckConnection(); err != nil {
		return manager, fmt.Errorf(createDatabaseError, err)
	}

	if err = manager.createSchema(); err != nil {
		return manager, fmt.Errorf(createDatabaseError, err)
	}

	if err = manager.createTables(); err != nil {
		return manager, fmt.Errorf(createDatabaseError, err)
	}

	return manager, nil
}

func (m *DataBaseManager) createSchema() error {
	if _, err := m.database.Exec(`CREATE SCHEMA IF NOT EXISTS ` + schemaName + `;`); err != nil {
		return fmt.Errorf(createSchemaError, err)
	}

	if _, err := m.database.Exec(`SET search_path TO ` + schemaName); err != nil {
		return fmt.Errorf(createSchemaError, err)
	}

	return nil
}

func (m *DataBaseManager) createTables() error {
	if _, err := m.database.Exec(`CREATE TABLE IF NOT EXISTS ` + gaugesTable + ` (id TEXT PRIMARY KEY, value float8);`); err != nil {
		return fmt.Errorf(createTableError, err)
	}

	if _, err := m.database.Exec(`CREATE TABLE IF NOT EXISTS ` + countersTable + ` (id TEXT PRIMARY KEY, value int8);`); err != nil {
		return fmt.Errorf(createTableError, err)
	}

	return nil
}

func (m *DataBaseManager) Close() error {
	return m.database.Close()
}

func (m *DataBaseManager) CheckConnection() (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err := m.database.PingContext(ctx); err != nil {
		return false, err
	}
	return true, nil
}

func (m *DataBaseManager) SaveMetricsInStorage(gaugesValues map[string]interface{}, countersValues map[string]interface{}) error {
	m.log.Info("[DataBaseManager:SaveMetricsInStorage] start transaction")
	tx, err := m.database.Begin()
	if err != nil {
		return fmt.Errorf(saveMetricsError, err)
	}

	if err = m.saveMetrics(tx, gaugesTable, gaugesValues); err != nil {
		tx.Rollback()
		return fmt.Errorf(saveMetricsError, err)
	}

	if err = m.saveMetrics(tx, countersTable, countersValues); err != nil {
		tx.Rollback()
		return fmt.Errorf(saveMetricsError, err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf(saveMetricsError, err)
	}

	m.log.Info("[DataBaseManager:SaveMetricsInStorage] transaction completed")
	return nil
}

func (m *DataBaseManager) RestoreDataFromStorage() (map[string]float64, map[string]int64, error) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	m.log.Info("[DataBaseManager:RestoreDataFromStorage] start transaction")
	tx, err := m.database.Begin()
	if err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	if err = m.restoreDataInMap(tx, gaugesTable, gauges); err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	if err = m.restoreDataInMap(tx, countersTable, counters); err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	m.log.Info("[DataBaseManager:RestoreDataFromStorage] transaction completed")
	return gauges, counters, nil
}

func (m *DataBaseManager) restoreDataInMap(tx *sql.Tx, tableName string, mapDest interface{}) error {
	rows, err := tx.Query(`SELECT * FROM ` + tableName)
	if err != nil {
		return fmt.Errorf(restoreDataError, err)
	}
	defer rows.Close()

	for rows.Next() {
		switch dest := mapDest.(type) {
		case map[string]float64:
			var name string
			var value float64
			err = rows.Scan(&name, &value)
			if err != nil {
				return fmt.Errorf(restoreDataError, err)
			}

			dest[name] = value
		case map[string]int64:
			var name string
			var value int64
			err = rows.Scan(&name, &value)
			if err != nil {
				return fmt.Errorf(restoreDataError, err)
			}

			dest[name] = value
		default:
			panic("Unknown value type")
		}
	}

	err = rows.Err()
	if err != nil {
		return fmt.Errorf(restoreDataError, err)
	}
	return nil
}

func (m *DataBaseManager) saveMetrics(tx *sql.Tx, metricTable string, metricsValues map[string]interface{}) error {
	requestStmts, err := createDatabaseStmts(tx, metricTable)
	if err != nil {
		return fmt.Errorf(saveMetricsCreateStmtError, err)
	}
	defer closeDatabaseStmts(requestStmts)

	for key, value := range metricsValues {
		if err = m.saveMetric(requestStmts, key, value); err != nil {
			return fmt.Errorf(saveMetricsUseStmtError, err)
		}
	}

	return nil
}

func (m *DataBaseManager) saveMetric(stmts map[string]*sql.Stmt, name string, value interface{}) error {
	exists, err := m.checkMetricExists(stmts[existStmt], name, value)
	if err != nil {
		return fmt.Errorf(saveMetricError, err)
	}

	if exists {
		err = m.updateMetric(stmts[updateStmt], name, value)
	} else {
		err = m.insertMetric(stmts[insertStmt], name, value)
	}

	if err != nil {
		return fmt.Errorf(saveMetricError, err)
	}
	return nil
}

func (m *DataBaseManager) checkMetricExists(stmt *sql.Stmt, name string, _ interface{}) (bool, error) {
	var exists bool
	err := stmt.QueryRow(name).Scan(&exists)
	return exists, err
}

func (m *DataBaseManager) insertMetric(stmt *sql.Stmt, name string, value interface{}) error {
	var err error
	tableName := getMetricsTableName(value)
	if tableName == gaugesTable {
		_, err = stmt.Exec(name, value.(float64))
	} else {
		_, err = stmt.Exec(name, value.(int64))
	}

	if err != nil {
		return fmt.Errorf(insertMetricError, name, value, tableName, err)
	}
	return nil
}

func (m *DataBaseManager) updateMetric(stmt *sql.Stmt, name string, value interface{}) error {
	var err error
	tableName := getMetricsTableName(value)
	if tableName == gaugesTable {
		_, err = stmt.Exec(value.(float64), name)
	} else {
		_, err = stmt.Exec(value.(int64), name)
	}

	if err != nil {
		return fmt.Errorf(updateMetricError, name, value, tableName, err)
	}
	return nil
}

func getMetricsTableName(value interface{}) string {
	switch value.(type) {
	case int64:
		return countersTable
	case float64:
		return gaugesTable
	default:
		panic("unknown value type")
	}
}

func createDatabaseStmts(tx *sql.Tx, metricTable string) (map[string]*sql.Stmt, error) {
	existBody := fmt.Sprintf(`SELECT EXISTS (SELECT 1 FROM %s WHERE id = $1);`, metricTable)
	insertBody := fmt.Sprintf(`INSERT INTO %s (id, value) VALUES ($1, $2);`, metricTable)
	updateBody := fmt.Sprintf(`UPDATE %s SET value = $1 WHERE id = $2;`, metricTable)

	existStmtBody, err := tx.Prepare(existBody)
	if err != nil {
		return nil, fmt.Errorf("prepare metric exists statemnt: %w", err)
	}

	insertStmtBody, err := tx.Prepare(insertBody)
	if err != nil {
		return nil, fmt.Errorf("prepare metric insert statemnt: %w", err)
	}

	updateStmtBody, err := tx.Prepare(updateBody)
	if err != nil {
		return nil, fmt.Errorf("prepare metric update statemnt: %w", err)
	}

	return map[string]*sql.Stmt{
		existStmt:  existStmtBody,
		insertStmt: insertStmtBody,
		updateStmt: updateStmtBody,
	}, nil
}

func closeDatabaseStmts(stmts map[string]*sql.Stmt) {
	for _, stmt := range stmts {
		stmt.Close()
	}
}
