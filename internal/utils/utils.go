package utils

import (
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strconv"
	"strings"
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

func ValidateFilters(r *http.Request, required bool, days int) (string, string, string, []interface{}, error) {
	gameType := r.URL.Query().Get("gameType")
	drawDateFrom := r.URL.Query().Get("drawDateFrom")
	drawDateTo := r.URL.Query().Get("drawDateTo")

	if gameType == "" && required {
		return "", "", "", nil, errors.New("gameType is required")
	}

	if drawDateFrom == "" {
		drawDateFrom = time.Now().AddDate(0, 0, -days).Format("2006-01-02")
	}
	if drawDateTo == "" {
		drawDateTo = time.Now().Format("2006-01-02")
	}

	_, errFrom := time.Parse("2006-01-02", drawDateFrom)
	_, errTo := time.Parse("2006-01-02", drawDateTo)

	if errFrom != nil || errTo != nil {
		return "", "", "", nil, errors.New("invalid date format")
	}

	return gameType, drawDateFrom, drawDateTo, []interface{}{}, nil
}

func FormatCombination(numbers []sql.NullInt64) string {
	var combination []string
	for _, num := range numbers {
		if num.Valid {
			combination = append(combination, strconv.FormatInt(num.Int64, 10))
		}
	}
	return strings.Join(combination, "-")
}

func RespondWithJSON(w http.ResponseWriter, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(payload)
}
