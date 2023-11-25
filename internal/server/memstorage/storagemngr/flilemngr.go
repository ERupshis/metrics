package storagemngr

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/erupshis/metrics/internal/logger"
)

const (
	initWriterError  = "init writer: %w"
	writeMetricError = "write metric: %w"

	initScannerError = "init scanner: %w"
	scanMetricError  = "scan metric: %w"

	openFileError = "open file: %w"
)

// fileWriter is responsible for writing metric data to a file.
type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

// fileScanner is responsible for scanning metric data from a file.
type fileScanner struct {
	file    *os.File
	scanner *bufio.Scanner
}

// FileManager provides functionality to manage metric storage in a file.
type FileManager struct {
	path    string
	logger  logger.BaseLogger
	writer  *fileWriter
	scanner *fileScanner
}

// CreateFileManager creates a new instance of FileManager with the specified data path and logger.
func CreateFileManager(dataPath string, logger logger.BaseLogger) StorageManager {
	logger.Info("[FileManager::CreateFileManager] create with file path: '%s'", dataPath)
	return &FileManager{path: dataPath, logger: logger}
}

// Close closes the underlying file if open.
func (fm *FileManager) Close() error {
	return nil
}

// CheckConnection checks the file connection status and always returns true for FileManager.
func (fm *FileManager) CheckConnection(_ context.Context) (bool, error) {
	return true, nil
}

// SaveMetricsInStorage saves gauge and counter metric values in the file.
func (fm *FileManager) SaveMetricsInStorage(_ context.Context, gaugeValues map[string]interface{}, counterValues map[string]interface{}) error {
	if !fm.IsFileOpen() {
		if err := fm.OpenFile(fm.path, true); err != nil {
			return fmt.Errorf("cannot open file '%s' to save metrics: %w", fm.path, err)
		}
		defer func() {
			if err := fm.CloseFile(); err != nil {
				fm.logger.Info("[FileManager::SaveMetricsInStorage] failed to close file: %v", err)
			}
		}()
	}

	for name, val := range gaugeValues {
		if err := fm.WriteMetric(name, val); err != nil {
			fm.logger.Info("[FileManager::SaveMetricsInStorage] failed to write gauge metric in file. err: %v", err)
		}
	}

	for name, val := range counterValues {
		if err := fm.WriteMetric(name, val); err != nil {
			fm.logger.Info("[FileManager::SaveMetricsInStorage] failed to write counter metric in file. err: %v", err)
		}
	}

	fm.logger.Info("[FileManager::SaveMetricsInStorage] storage successfully saved in file: %s", fm.path)
	return nil
}

// RestoreDataFromStorage reads metric data from the file and restores it.
func (fm *FileManager) RestoreDataFromStorage(_ context.Context) (map[string]float64, map[string]int64, error) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	if !fm.IsFileOpen() {
		if err := fm.OpenFile(fm.path, false); err != nil {
			return gauges, counters, fmt.Errorf("cannot open file '%s' to read metrics: %w", fm.path, err)
		}
		defer func() {
			if err := fm.CloseFile(); err != nil {
				fm.logger.Info("[FileManager::RestoreDataFromStorage] failed to close file: %v", err)
			}
		}()
	}

	failedToReadMetricsCount := 0
	metric, err := fm.ScanMetric()
	for metric != nil {
		if err != nil {
			fm.logger.Info("[FileManager::RestoreDataFromStorage] failed to scan metric '%s' from file '%s'", metric.Name, fm.path)
			failedToReadMetricsCount++

		} else {
			fm.parseMetric(metric, &gauges, &counters)
		}

		metric, err = fm.ScanMetric()
	}

	fm.logger.Info("[FileManager::restoreDataFromFileIfNeed] storage successfully restored from file: '%s', failed to read metrics: '%d'",
		fm.path, failedToReadMetricsCount)

	err = nil
	if failedToReadMetricsCount > 0 {
		err = fmt.Errorf("some metrics weren't read from file, count: '%d'", failedToReadMetricsCount)
	}

	return gauges, counters, err
}

// parseMetric parses the MetricData and updates the provided gauge and counter maps accordingly.
func (fm *FileManager) parseMetric(metric *MetricData, gauges *map[string]float64, counters *map[string]int64) {
	switch metric.ValueType {
	case "gauge":
		value, err := strconv.ParseFloat(metric.Value, 64)
		if err != nil {
			fm.logger.Info("[FileManager::RestoreDataFromStorage] failed to parse float64 value for '%s'", metric.Name)
			return
		}
		(*gauges)[metric.Name] = value
	case "counter":
		value, err := strconv.ParseInt(metric.Value, 10, 64)
		if err != nil {
			fm.logger.Info("[FileManager::RestoreDataFromStorage] failed to parse int64 value for '%s'", metric.Name)
			return
		}
		(*counters)[metric.Name] = value
	default:
		panic("wrong metric type")
	}
}

// IsFileOpen checks if the file is open.
func (fm *FileManager) IsFileOpen() bool {
	return fm.writer != nil && fm.scanner != nil
}

// OpenFile opens or creates a file for writing or reading metrics.
func (fm *FileManager) OpenFile(path string, withTrunc bool) error {
	fm.path = path

	if err := os.MkdirAll(filepath.Dir(fm.path), 0755); err != nil {
		return fmt.Errorf(openFileError, err)
	}

	if err := fm.initWriter(withTrunc); err != nil {
		return fmt.Errorf(openFileError, err)
	}

	if err := fm.initScanner(); err != nil {
		return fmt.Errorf(openFileError, err)
	}

	return nil
}

// CloseFile closes the file if open.
func (fm *FileManager) CloseFile() error {
	if !fm.IsFileOpen() {
		return nil
	}

	var err = fm.writer.file.Close()

	fm.writer = nil
	fm.scanner = nil

	return err
}

// initWriter initializes the file writer.
func (fm *FileManager) initWriter(withTrunc bool) error {
	var flag int
	flag = os.O_WRONLY | os.O_CREATE
	if withTrunc {
		flag |= os.O_TRUNC
	}

	file, err := os.OpenFile(fm.path, flag, 0666)
	if err != nil {
		return fmt.Errorf(initWriterError, err)
	}

	fm.writer = &fileWriter{file, bufio.NewWriter(file)}
	return nil
}

// WriteMetric writes a metric to the file.
func (fm *FileManager) WriteMetric(name string, value interface{}) error {
	if !fm.IsFileOpen() {
		return fmt.Errorf("failed writing metric. file is not open")
	}

	var data []byte
	var err error

	switch valType := value.(type) {
	case *int64:
		metric := MetricData{
			name,
			"counter",
			strconv.FormatInt(*valType, 10),
		}

		data, err = json.Marshal(&metric)

	case *float64:
		metric := MetricData{
			name,
			"gauge",
			strconv.FormatFloat(*valType, 'f', -1, 64),
		}

		data, err = json.Marshal(&metric)

	default:
		err = fmt.Errorf("unknown value type")
	}

	if err != nil {
		return fmt.Errorf(writeMetricError, err)
	}

	data = append(data, '\n')
	if _, err = fm.write(data); err != nil {
		return fmt.Errorf(writeMetricError, err)
	}

	err = fm.flushWriter()
	if err != nil {
		return fmt.Errorf(writeMetricError, err)
	}
	return nil
}

// write writes data to the file.
func (fm *FileManager) write(data []byte) (int, error) {
	return fm.writer.writer.Write(data)
}

// flushWriter flushes the writer buffer.
func (fm *FileManager) flushWriter() error {
	return fm.writer.writer.Flush()
}

// initScanner initializes the file scanner.
func (fm *FileManager) initScanner() error {
	file, err := os.OpenFile(fm.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf(initScannerError, err)
	}

	fm.scanner = &fileScanner{file, bufio.NewScanner(file)}
	return nil
}

// ScanMetric scans and returns a metric from the file.
func (fm *FileManager) ScanMetric() (*MetricData, error) {
	var metric MetricData
	if !fm.IsFileOpen() {
		return nil, fmt.Errorf(scanMetricError, fmt.Errorf("file is not open"))
	}

	if isScanOk, err := fm.scan(); !isScanOk {
		return nil, fmt.Errorf(scanMetricError, err)
	}

	data := fm.scannedBytes()
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, fmt.Errorf(scanMetricError, err)
	}

	return &metric, nil
}

// scan scans the file for the next line.
func (fm *FileManager) scan() (bool, error) {
	if !fm.scanner.scanner.Scan() {
		if err := fm.scanner.scanner.Err(); err != nil {
			return false, err
		} else {
			return false, nil
		}
	}

	return true, nil
}

// scannedBytes returns the scanned bytes from the scanner.
func (fm *FileManager) scannedBytes() []byte {
	return fm.scanner.scanner.Bytes()
}
