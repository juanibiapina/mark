package app

import (
	"bufio"
	"fmt"
	"os"

	"mark/pkg/util"
)

type ReplaceLine struct {
	Filename   string `json:"filename" jsonschema_description:"The filename of the file to replace a line in"`
	LineNumber int64  `json:"line_number" jsonschema_description:"Line number to replace"`
	Content    string `json:"content" jsonschema_description:"The new content of the line"`
}

func (r ReplaceLine) Invoke() error {
	file, err := os.Open(r.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if r.LineNumber < 1 || int(r.LineNumber) > len(lines) {
		return fmt.Errorf("line number %d out of range", r.LineNumber)
	}

	lines[r.LineNumber-1] = r.Content

	outputFile, err := os.Create(r.Filename)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	writer := bufio.NewWriter(outputFile)
	for _, line := range lines {
		_, err := writer.WriteString(line + "\n")
		if err != nil {
			return err
		}
	}
	return writer.Flush()
}

var ReplaceLineResponseSchema = util.GenerateSchema[ReplaceLine]()
