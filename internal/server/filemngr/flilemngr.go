package filemngr

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type MetricData struct {
	Name      string `json:"name"`
	ValueType string `json:"type"`
	Value     string `json:"value"`
}

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
	writer  *fileWriter
	scanner *fileScanner
}

func Create() *FileManager {
	return &FileManager{}
}

// FILE HANDLING.

func (fm *FileManager) IsFileOpen() bool {
	return fm.writer != nil && fm.scanner != nil
}

func (fm *FileManager) OpenFile(path string, withTrunc bool) error {
	fm.path = path

	if err := os.MkdirAll(filepath.Dir(fm.path), 0755); err != nil {
		return err
	}

	if err := fm.initWriter(withTrunc); err != nil {
		return err
	}

	if err := fm.initScanner(); err != nil {
		return err
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
	fm.path = ""

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
		return err
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
		return err
	}

	data = append(data, '\n')
	if _, err := fm.write(data); err != nil {
		return err
	}

	return fm.flushWriter()
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
		return err
	}

	fm.scanner = &fileScanner{file, bufio.NewScanner(file)}
	return nil
}

func (fm *FileManager) ScanMetric() (*MetricData, error) {
	var metric MetricData
	if !fm.IsFileOpen() {
		return nil, fmt.Errorf("failed reading metric. file is not open")
	}

	if isScanOk, err := fm.scan(); !isScanOk {
		return nil, err
	}

	data := fm.scannedBytes()
	if err := json.Unmarshal(data, &metric); err != nil {
		return nil, err
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
