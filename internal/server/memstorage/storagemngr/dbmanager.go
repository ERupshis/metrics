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

	return manager, nil
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

func (m *DataBaseManager) SaveMetricsInStorage(gaugeValues map[string]float64, counterValues map[string]int64) {

}

func (m *DataBaseManager) RestoreDataFromStorage() (map[string]float64, map[string]int64) {
	return nil, nil
}
