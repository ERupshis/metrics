// Package storagemngr implements the StorageManager interface and provides functionality
// for managing metric data storage in a PostgreSQL database. It includes methods for saving metrics,
// restoring data, checking connection status, and closing the database connection.
package storagemngr

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/erupshis/metrics/internal/logger"
	"github.com/erupshis/metrics/internal/retryer"
	"github.com/erupshis/metrics/internal/server/config"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

// Constants defining schema and table names.
const (
	schemaName    = "metrics"
	gaugesTable   = "gauges"
	countersTable = "counters"

	insertStmt = "insert"
	updateStmt = "update"
	existStmt  = "exist"

	createDatabaseError = "create db: %w"
	saveMetricError     = "save metric: %w"
	saveMetricsError    = "save metrics in db: %w"
	restoreMetricsError = "restore metrics from db: %w"
	restoreDataError    = "restore data from db response: %w"

	logSaveMetricsInStorageStart     = "[DataBaseManager:SaveMetricsInStorage] start transaction"
	logSaveMetricsInStorageCompleted = "[DataBaseManager:SaveMetricsInStorage] transaction completed"
)

// DatabaseErrorsToRetry is a list of database errors that are considered retryable.
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

// DataBaseManager is a struct implementing the StorageManager interface
// for managing metric data storage in a PostgreSQL database.
type DataBaseManager struct {
	database *sql.DB
	log      logger.BaseLogger
}

// CreateDataBaseManager creates a new instance of DataBaseManager, initializes the database, and performs migrations.
func CreateDataBaseManager(ctx context.Context, cfg *config.Config, log logger.BaseLogger) (StorageManager, error) {
	log.Info("[storagemngr:CreateDataBaseManager] Open database with settings: '%s'", cfg.DataBaseDSN)
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

// Close closes the underlying SQL database connection.
func (m *DataBaseManager) Close() error {
	return m.database.Close()
}

// CheckConnection checks the connection to the SQL database and returns true if successful.
func (m *DataBaseManager) CheckConnection(ctx context.Context) (bool, error) {
	exec := func(context context.Context) error {
		return m.database.PingContext(context)
	}
	err := retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	if err != nil {
		return false, fmt.Errorf("check connection: %w", err)
	}
	return true, nil
}

// SaveMetricsInStorage saves gauge and counter metric values in the PostgreSQL database.
func (m *DataBaseManager) SaveMetricsInStorage(ctx context.Context, gaugesValues map[string]interface{}, countersValues map[string]interface{}) error {
	// m.log.Info(logSaveMetricsInStorageStart)

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

	// m.log.Info(logSaveMetricsInStorageCompleted)
	return nil
}

// RestoreDataFromStorage retrieves and restores stored metric data from the PostgreSQL database.
func (m *DataBaseManager) RestoreDataFromStorage(ctx context.Context) (map[string]float64, map[string]int64, error) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	m.log.Info("[DataBaseManager:RestoreDataFromStorage] start transaction")
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

// restoreDataInMap populates the provided mapDest with data from the specified database table.
func (m *DataBaseManager) restoreDataInMap(ctx context.Context, tx *sql.Tx, tableName string, mapDest interface{}) error {
	var err error
	var rows *sql.Rows
	if rows != nil {
		// get rid of static check problem
		err = rows.Err()

	}

	if err != nil {
		return err
	}

	query := func(context context.Context) error {
		sqlSelect, _, err := sq.Select("*").From(schemaName + "." + tableName).ToSql()
		if err != nil {
			return fmt.Errorf("restore metrics: %w", err)
		}

		rows, err = tx.QueryContext(ctx, sqlSelect)
		if rows != nil {
			// get rid of static check problem
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

// saveMetrics saves the provided metricsValues in the specified database table using a transaction.
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

// saveMetric saves a single metric with the specified name and value using database statements.
func (m *DataBaseManager) saveMetric(ctx context.Context, stmts map[string]*sql.Stmt, name string, value interface{}) error {
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

// checkMetricExists checks if a metric with the specified name already exists in the database.
func (m *DataBaseManager) checkMetricExists(ctx context.Context, stmt *sql.Stmt, name string, _ interface{}) (bool, error) {
	var exists bool
	var err error

	query := func(context context.Context) error {
		return stmt.QueryRowContext(ctx, name).Scan(&exists)
	}
	err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, query)
	if err != nil {
		err = fmt.Errorf("exists metric check: %w", err)
	}

	return exists, err
}

// insertMetric inserts a new metric with the specified name and value into the database.
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

// updateMetric updates an existing metric with the specified name and value in the database.
func (m *DataBaseManager) updateMetric(ctx context.Context, stmt *sql.Stmt, name string, value interface{}) error {
	var err error
	tableName := getMetricsTableName(value)
	if tableName == gaugesTable {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, *value.(*float64), name)
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	} else {
		exec := func(context context.Context) error {
			_, err = stmt.ExecContext(context, *value.(*int64), name)
			return err
		}
		err = retryer.RetryCallWithTimeout(ctx, m.log, nil, DatabaseErrorsToRetry, exec)
	}

	if err != nil {
		return fmt.Errorf("update metric name '%s', value '%s', table '%v': %w", name, value, tableName, err)
	}
	return nil
}

// getMetricsTableName determines the database table name based on the type of the metric value.
func getMetricsTableName(value interface{}) string {
	switch value.(type) {
	case *int64:
		return countersTable
	case *float64:
		return gaugesTable
	default:
		panic("unknown value type")
	}
}

// createDatabaseStmts prepares and returns a map of SQL statements for the specified metric table.
func createDatabaseStmts(ctx context.Context, tx *sql.Tx, metricTable string) (map[string]*sql.Stmt, error) {
	existStmtBody, err := createExistsStmt(ctx, tx, metricTable)
	if err != nil {
		return nil, fmt.Errorf("prepare metric exists statement: %w", err)
	}

	insertStmtBody, err := createInsertStmt(ctx, tx, metricTable)
	if err != nil {
		return nil, fmt.Errorf("prepare metric insert statement: %w", err)
	}

	updateStmtBody, err := createUpdateStmt(ctx, tx, metricTable)
	if err != nil {
		return nil, fmt.Errorf("prepare metric update statement: %w", err)
	}

	return map[string]*sql.Stmt{
		existStmt:  existStmtBody,
		insertStmt: insertStmtBody,
		updateStmt: updateStmtBody,
	}, nil
}

// createExistsStmt prepares and returns an SQL statement to check if a metric exists in the database.
func createExistsStmt(ctx context.Context, tx *sql.Tx, metricTable string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sqlExist, _, err := psql.Select("1").From(schemaName + "." + metricTable).Where("id = ?").ToSql()
	if err != nil {
		return nil, fmt.Errorf("squirrel sql statement: %w", err)

	}
	return tx.PrepareContext(ctx, "SELECT EXISTS("+sqlExist+")")
}

// createInsertStmt prepares and returns an SQL statement to insert a new metric into the database.
func createInsertStmt(ctx context.Context, tx *sql.Tx, metricTable string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sqlInsert, _, err := psql.Insert(schemaName+"."+metricTable).Columns("id", "value").Values("(?, ?)", "val").ToSql()
	if err != nil {
		return nil, fmt.Errorf("squirrel sql statement: %w", err)

	}
	return tx.PrepareContext(ctx, sqlInsert)
}

// createUpdateStmt prepares and returns an SQL statement to update an existing metric in the database.
func createUpdateStmt(ctx context.Context, tx *sql.Tx, metricTable string) (*sql.Stmt, error) {
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	sqlUpdate, _, err := psql.Update(schemaName+"."+metricTable).Set("value", "?").Where("id = ?").ToSql()
	if err != nil {
		return nil, fmt.Errorf("squirrel sql statement: %w", err)

	}
	return tx.PrepareContext(ctx, sqlUpdate)
}

// closeDatabaseStmts closes all the SQL statements in the provided map.
func closeDatabaseStmts(stmts map[string]*sql.Stmt) {
	for _, stmt := range stmts {
		stmt.Close()
	}
}
