package utils

import (
	"database/sql"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestParseDate(t *testing.T) {
	tests := []struct {
		dateStr string
		want    time.Time
		wantErr bool
	}{
		{"01/02/2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"1/2/2006", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"2006-01-02", time.Date(2006, 1, 2, 0, 0, 0, 0, time.UTC), false},
		{"invalid", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.dateStr, func(t *testing.T) {
			got, err := ParseDate(tt.dateStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseDate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !got.Equal(tt.want) {
				t.Errorf("ParseDate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReadCSV(t *testing.T) {
	tests := []struct {
		filePath string
		want     [][]string
		wantErr  bool
	}{
		{"valid.csv", [][]string{{"1", "2", "3"}, {"4", "5", "6"}}, false},
	}

	for _, tt := range tests {
		file, err := os.Create(tt.filePath)
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()

		writer := csv.NewWriter(file)
		writer.Comma = '#'
		writer.WriteAll(tt.want)
		writer.Flush()

		t.Run(tt.filePath, func(t *testing.T) {
			got, err := ReadCSV(tt.filePath)
			if (err != nil) != tt.wantErr {
				t.Errorf("ReadCSV() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !equal(got, tt.want) {
				t.Errorf("ReadCSV() = %v, want %v", got, tt.want)
			}
		})

		os.Remove(tt.filePath)
	}
}

func TestValidateFilters(t *testing.T) {
	tests := []struct {
		query    string
		wantType string
		wantFrom string
		wantTo   string
		wantErr  bool
	}{
		{"gameType=lotto&drawDateFrom=2006-01-02&drawDateTo=2006-02-02", "lotto", "2006-01-02", "2006-02-02", false},
		{"gameType=lotto", "lotto", time.Now().AddDate(0, 0, -30).Format("2006-01-02"), time.Now().Format("2006-01-02"), false},
		{"drawDateFrom=2006-01-02&drawDateTo=2006-02-02", "", "", "", true},
		{"gameType=lotto&drawDateFrom=invalid&drawDateTo=invalid", "", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "/?"+tt.query, nil)
			gotType, gotFrom, gotTo, _, err := ValidateFilters(req, true, 30)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFilters() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotType != tt.wantType {
				t.Errorf("ValidateFilters() gotType = %v, want %v", gotType, tt.wantType)
			}
			if gotFrom != tt.wantFrom {
				t.Errorf("ValidateFilters() gotFrom = %v, want %v", gotFrom, tt.wantFrom)
			}
			if gotTo != tt.wantTo {
				t.Errorf("ValidateFilters() gotTo = %v, want %v", gotTo, tt.wantTo)
			}
		})
	}
}

func TestFormatCombination(t *testing.T) {
	tests := []struct {
		numbers []sql.NullInt64
		want    string
	}{
		{[]sql.NullInt64{{Int64: 1, Valid: true}, {Int64: 2, Valid: true}, {Int64: 3, Valid: true}}, "1-2-3"},
		{[]sql.NullInt64{{Int64: 1, Valid: true}, {Int64: 0, Valid: false}, {Int64: 3, Valid: true}}, "1-3"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := FormatCombination(tt.numbers); got != tt.want {
				t.Errorf("FormatCombination() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRespondWithJSON(t *testing.T) {
	rr := httptest.NewRecorder()
	payload := map[string]string{"message": "hello"}

	RespondWithJSON(rr, payload)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("RespondWithJSON() status code = %v, want %v", status, http.StatusOK)
	}

	expected := `{"message":"hello"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("RespondWithJSON() body = %v, want %v", rr.Body.String(), expected)
	}
}

func equal(a, b [][]string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			if a[i][j] != b[i][j] {
				return false
			}
		}
	}
	return true
}
