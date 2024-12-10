package utils

import (
	"encoding/csv"
	"os"
	"time"
)

func ParseDate(dateStr string) (time.Time, error) {
	formats := []string{"01/02/2006", "1/2/2006", "2006-01-02"}
	var err error
	var date time.Time
	for _, format := range formats {
		date, err = time.Parse(format, dateStr)
		if err == nil {
			return date, nil
		}
	}
	return time.Time{}, err
}

func ReadCSV(filePath string) ([][]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = '#'
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	return records, nil
}
