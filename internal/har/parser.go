package har

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

type Parser struct {
	bufferSize int
}

func NewParser() *Parser {
	return &Parser{
		bufferSize: 64 * 1024, // 64KB buffer
	}
}

func (p *Parser) ParseFile(filepath string) (*HAR, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to open HAR file: %w", err)
	}
	defer file.Close()

	return p.ParseReader(file)
}

func (p *Parser) ParseReader(reader io.Reader) (*HAR, error) {
	bufferedReader := bufio.NewReaderSize(reader, p.bufferSize)
	decoder := json.NewDecoder(bufferedReader)

	var har HAR
	if err := decoder.Decode(&har); err != nil {
		return nil, fmt.Errorf("failed to decode HAR JSON: %w", err)
	}

	return &har, nil
}

func (p *Parser) ParseMultipleFiles(filepaths []string) ([]*HAR, error) {
	hars := make([]*HAR, 0, len(filepaths))

	for _, filepath := range filepaths {
		har, err := p.ParseFile(filepath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w", filepath, err)
		}
		hars = append(hars, har)
	}

	return hars, nil
}

func (p *Parser) ValidateHAR(har *HAR) error {
	if har.Log.Version == "" {
		return fmt.Errorf("missing HAR version")
	}

	if len(har.Log.Entries) == 0 {
		return fmt.Errorf("no entries found in HAR file")
	}

	return nil
}
