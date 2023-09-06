package storagemngr

import (
	"bufio"
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

type fileWriter struct {
	file   *os.File
	writer *bufio.Writer
}

type fileScanner struct {
	file    *os.File
	scanner *bufio.Scanner
}

type FileManager struct {
	path    string
	logger  logger.BaseLogger
	writer  *fileWriter
	scanner *fileScanner
}

func CreateFileManager(dataPath string, logger logger.BaseLogger) StorageManager {
	logger.Info("[FileManager::CreateFileManager] create with file path: '%s'", dataPath)
	return &FileManager{path: dataPath, logger: logger}
}

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// INTERFACE FOR STORAGE.

func (fm *FileManager) Close() error {
	return nil
}

func (fm *FileManager) CheckConnection() (bool, error) {
	return true, nil
}

func (fm *FileManager) SaveMetricsInStorage(gaugeValues map[string]interface{}, counterValues map[string]interface{}) error {
	if !fm.IsFileOpen() {
		if err := fm.OpenFile(fm.path, true); err != nil {
			return fmt.Errorf("cannot open file '%s' to save metrics: %w", fm.path, err)
		}
		defer fm.CloseFile()
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

func (fm *FileManager) RestoreDataFromStorage() (map[string]float64, map[string]int64, error) {
	gauges := map[string]float64{}
	counters := map[string]int64{}

	if !fm.IsFileOpen() {
		if err := fm.OpenFile(fm.path, false); err != nil {
			return gauges, counters, fmt.Errorf("cannot open file '%s' to read metrics: %w", fm.path, err)
		}
		defer fm.CloseFile()
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

///////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// FILE HANDLING.

func (fm *FileManager) IsFileOpen() bool {
	return fm.writer != nil && fm.scanner != nil
}

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

func (fm *FileManager) CloseFile() error {
	if !fm.IsFileOpen() {
		return nil
	}

	var err = fm.writer.file.Close()

	fm.writer = nil
	fm.scanner = nil

	return err
}

// WRITER.

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

func (fm *FileManager) WriteMetric(name string, value interface{}) error {
	if !fm.IsFileOpen() {
		return fmt.Errorf("failed writing metric. file is not open")
	}

	var data []byte
	var err error

	switch valType := value.(type) {
	case int64:
		metric := MetricData{
			name,
			"counter",
			strconv.FormatInt(valType, 10),
		}

		data, err = json.Marshal(&metric)

	case float64:
		metric := MetricData{
			name,
			"gauge",
			strconv.FormatFloat(valType, 'f', -1, 64),
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

func (fm *FileManager) write(data []byte) (int, error) {
	return fm.writer.writer.Write(data)
}

func (fm *FileManager) flushWriter() error {
	return fm.writer.writer.Flush()
}

// SCANNER.

func (fm *FileManager) initScanner() error {
	file, err := os.OpenFile(fm.path, os.O_RDONLY|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf(initScannerError, err)
	}

	fm.scanner = &fileScanner{file, bufio.NewScanner(file)}
	return nil
}

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

func (fm *FileManager) scannedBytes() []byte {
	return fm.scanner.scanner.Bytes()
}
