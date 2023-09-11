package storagemngr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/retryer"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

const (
	schemaName    = "metrics"
	gaugesTable   = "gauges"
	countersTable = "counters"

	insertStmt = "insert"
	updateStmt = "update"
	existStmt  = "exist"
)

var DatabaseErrorsToRetry = []error{
	errors.New(pgerrcode.UniqueViolation),
	errors.New(pgerrcode.ConnectionException),
	errors.New(pgerrcode.ConnectionDoesNotExist),
	errors.New(pgerrcode.ConnectionFailure),
	errors.New(pgerrcode.SQLClientUnableToEstablishSQLConnection),
	errors.New(pgerrcode.SQLServerRejectedEstablishmentOfSQLConnection),
	errors.New(pgerrcode.TransactionResolutionUnknown),
	errors.New(pgerrcode.ProtocolViolation),
}

type DataBaseManager struct {
	database *sql.DB
	log      logger.BaseLogger
}

func CreateDataBaseManager(ctx context.Context, cfg *config.Config, log logger.BaseLogger) (StorageManager, error) {
	log.Info("[storagemngr:CreateDataBaseManager] Open database with settings: '%s'", cfg.DataBaseDSN)
	createDatabaseError := "create db: %w"
	database, err := sql.Open("pgx", cfg.DataBaseDSN)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	driver, err := postgres.WithInstance(database, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	m, err := migrate.NewWithDatabaseInstance("file://db/migrations", "postgres", driver)
	if err != nil {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf(createDatabaseError, err)
	}

	manager := &DataBaseManager{database: database, log: log}
	if _, err = manager.CheckConnection(ctx); err != nil {
		return manager, fmt.Errorf(createDatabaseError, err)
	}

	return manager, nil
}

func (m *DataBaseManager) Close() error {
	return m.database.Close()
}

func (m *DataBaseManager) CheckConnection(ctx context.Context) (bool, error) {
	exec := func(context context.Context) error {
		return m.database.PingContext(context)
	}
	err := retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (m *DataBaseManager) SaveMetricsInStorage(ctx context.Context, gaugesValues map[string]interface{}, countersValues map[string]interface{}) error {
	m.log.Info("[DataBaseManager:SaveMetricsInStorage] start transaction")
	saveMetricsError := "save metrics in db: %w"
	tx, err := m.database.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf(saveMetricsError, err)
	}

	if err = m.saveMetrics(ctx, tx, gaugesTable, gaugesValues); err != nil {
		tx.Rollback()
		return fmt.Errorf(saveMetricsError, err)
	}

	if err = m.saveMetrics(ctx, tx, countersTable, countersValues); err != nil {
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

func (m *DataBaseManager) RestoreDataFromStorage(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	m.log.Info("[DataBaseManager:RestoreDataFromStorage] start transaction")
	restoreMetricsError := "restore metrics from db: %w"
	tx, err := m.database.BeginTx(ctx, nil)
	if err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	if err = m.restoreDataInMap(ctx, tx, gaugesTable, gauges); err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	if err = m.restoreDataInMap(ctx, tx, countersTable, counters); err != nil {
		tx.Rollback()
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	err = tx.Commit()
	if err != nil {
		return nil, nil, fmt.Errorf(restoreMetricsError, err)
	}

	m.log.Info("[DataBaseManager:RestoreDataFromStorage] transaction completed")
	return gauges, counters, nil
}

func (m *DataBaseManager) restoreDataInMap(ctx context.Context, tx *sql.Tx, tableName string, mapDest interface{}) error {
	restoreDataError := "restore data from db response: %w"
	var err error
	var rows *sql.Rows
	if rows != nil {
		//get rid of static check problem
		err = rows.Err()
	}

	query := func(context context.Context) error {
		rows, err = tx.QueryContext(ctx, `SELECT * FROM `+schemaName+"."+tableName)
		if rows != nil {
			//get rid of static check problem
			rows.Err()
		}
		return err
	}

	err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, query)
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

func (m *DataBaseManager) saveMetrics(ctx context.Context, tx *sql.Tx, metricTable string, metricsValues map[string]interface{}) error {
	requestStmts, err := createDatabaseStmts(ctx, tx, metricTable)
	if err != nil {
		return fmt.Errorf("create stmt: %w", err)
	}
	defer closeDatabaseStmts(requestStmts)

	for key, value := range metricsValues {
		if err = m.saveMetric(ctx, requestStmts, key, value); err != nil {
			return fmt.Errorf("use stmt: %w", err)
		}
	}

	return nil
}

func (m *DataBaseManager) saveMetric(ctx context.Context, stmts map[string]*sql.Stmt, name string, value interface{}) error {
	saveMetricError := "save metric: %w"
	exists, err := m.checkMetricExists(ctx, stmts[existStmt], name, value)
	if err != nil {
		return fmt.Errorf(saveMetricError, err)
	}

	if exists {
		err = m.updateMetric(ctx, stmts[updateStmt], name, value)
	} else {
		err = m.insertMetric(ctx, stmts[insertStmt], name, value)
	}

	if err != nil {
		return fmt.Errorf(saveMetricError, err)
	}
	return nil
}

func (m *DataBaseManager) checkMetricExists(ctx context.Context, stmt *sql.Stmt, name string, _ interface{}) (bool, error) {
	var exists bool
	var err error

	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&exists)
	}
	err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, query)

	return exists, err
}

func (m *DataBaseManager) insertMetric(ctx context.Context, stmt *sql.Stmt, name string, value interface{}) error {
	var err error
	tableName := getMetricsTableName(value)
	if tableName == gaugesTable {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, name, value.(float64))
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	} else {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, name, value.(int64))
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	}

	if err != nil {
		return fmt.Errorf("insert new metric name '%s', value '%s', table '%v': %w", name, value, tableName, err)
	}
	return nil
}

func (m *DataBaseManager) updateMetric(ctx context.Context, stmt *sql.Stmt, name string, value interface{}) error {
	var err error
	tableName := getMetricsTableName(value)
	if tableName == gaugesTable {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, value.(float64), name)
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	} else {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, value.(int64), name)
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	}

	if err != nil {
		return fmt.Errorf("update metric name '%s', value '%s', table '%v': %w", name, value, tableName, err)
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

func createDatabaseStmts(ctx context.Context, tx *sql.Tx, metricTable string) (map[string]*sql.Stmt, error) {
	existStmtBody, err := tx.PrepareContext(ctx, `SELECT EXISTS (SELECT 1 FROM `+schemaName+"."+metricTable+` WHERE id = $1);`)
	if err != nil {
		return nil, fmt.Errorf("prepare metric exists statemnt: %w", err)
	}

	insertStmtBody, err := tx.PrepareContext(ctx, `INSERT INTO `+schemaName+"."+metricTable+` (id, value) VALUES ($1, $2);`)
	if err != nil {
		return nil, fmt.Errorf("prepare metric insert statemnt: %w", err)
	}

	updateStmtBody, err := tx.PrepareContext(ctx, `UPDATE `+schemaName+"."+metricTable+` SET value = $1 WHERE id = $2;`)
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
