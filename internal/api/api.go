package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"pcso-lotto-scraper-api/internal/utils"
)

func GetLottoResults(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameType := r.URL.Query().Get("gameType")
		drawDateFrom := r.URL.Query().Get("drawDateFrom")
		drawDateTo := r.URL.Query().Get("drawDateTo")

		query := `
			SELECT game_type, draw_date, jackpot, winners, 
			number1, number2, number3, number4, number5, number6 
			FROM lotto_results WHERE 1=1
		`
		var args []interface{}

		// Filter by game type
		if gameType != "" {
			query += " AND game_type = ?"
			args = append(args, gameType)
		}

		// Validate and filter by draw date range
		if drawDateFrom != "" || drawDateTo != "" {
			if drawDateFrom == "" {
				drawDateFrom = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			}

			if drawDateTo == "" {
				drawDateTo = time.Now().Format("2006-01-02")
			}

			// Ensure valid date formats
			parseDateFrom, errFrom := utils.ParseDate(drawDateFrom)
			parseDateTo, errTo := utils.ParseDate(drawDateTo)

			if errFrom != nil || errTo != nil {
				http.Error(w, "Invalid date format. Use YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			query += " AND draw_date BETWEEN ? AND ?"
			args = append(args, parseDateFrom, parseDateTo)
		} else {
			// Default to the today and yesterday
			query += " AND draw_date >= ?"
			args = append(args, time.Now().AddDate(0, 0, -1).Format("2006-01-02"))
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var results []map[string]interface{}

		for rows.Next() {
			var gameType, drawDate string
			var jackpot float64
			var winners int
			numbers := make([]sql.NullInt64, 6)

			err := rows.Scan(&gameType, &drawDate, &jackpot, &winners, &numbers[0], &numbers[1], &numbers[2], &numbers[3], &numbers[4], &numbers[5])
			if err != nil {
				http.Error(w, "Failed to scan results", http.StatusInternalServerError)
				return
			}

			var combination []string
			for _, num := range numbers {
				if num.Valid {
					combination = append(combination, strconv.FormatInt(num.Int64, 10))
				}
			}

			results = append(results, map[string]interface{}{
				"gameType":    gameType,
				"drawDate":    drawDate,
				"jackpot":     jackpot,
				"winners":     winners,
				"combination": strings.Join(combination, "-"),
			})
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(results)
	}
}

func GetHeatmap(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		gameType := r.URL.Query().Get("gameType")
		drawDateFrom := r.URL.Query().Get("drawDateFrom")
		drawDateTo := r.URL.Query().Get("drawDateTo")

		query := `
			SELECT number1, number2, number3, number4, number5, number6
			FROM lotto_results WHERE 1=1
		`
		var args []interface{}

		// Validate required params
		if gameType == "" {
			http.Error(w, "gameType is required", http.StatusBadRequest)
			return
		} else {
			query += " AND game_type = ?"
			args = append(args, gameType)
		}

		// Validate and filter by draw date range
		if drawDateFrom != "" || drawDateTo != "" {
			if drawDateFrom == "" {
				drawDateFrom = time.Now().AddDate(0, 0, -1).Format("2006-01-02")
			}

			if drawDateTo == "" {
				drawDateTo = time.Now().Format("2006-01-02")
			}

			// Ensure valid date formats
			parseDateFrom, errFrom := utils.ParseDate(drawDateFrom)
			parseDateTo, errTo := utils.ParseDate(drawDateTo)

			if errFrom != nil || errTo != nil {
				http.Error(w, "Invalid date format. Use YYYY-MM-DD.", http.StatusBadRequest)
				return
			}

			query += " AND draw_date BETWEEN ? AND ?"
			args = append(args, parseDateFrom, parseDateTo)
		} else {
			// Default to the today and yesterday
			query += " AND draw_date >= ?"
			args = append(args, time.Now().AddDate(0, 0, -30).Format("2006-01-02"))
		}

		rows, err := db.Query(query, args...)
		if err != nil {
			http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		// Initialize a heatmap for counting number occurrences
		heatmap := make(map[int]int)

		for rows.Next() {
			var numbers [6]sql.NullInt64
			err := rows.Scan(&numbers[0], &numbers[1], &numbers[2], &numbers[3], &numbers[4], &numbers[5])
			if err != nil {
				http.Error(w, "Failed to scan results", http.StatusInternalServerError)
				return
			}

			// Count number occurrences
			for _, num := range numbers {
				if num.Valid {
					heatmap[int(num.Int64)]++
				}
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(heatmap)
	}
}

func GetGameTypes(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `
            SELECT game_type
            FROM lotto_results
            GROUP BY game_type
            HAVING COUNT(*) > 10
		`

		rows, err := db.Query(query)
		if err != nil {
			http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var gameTypes []string

		for rows.Next() {
			var gameType string
			err := rows.Scan(&gameType)
			if err != nil {
				http.Error(w, "Failed to scan results", http.StatusInternalServerError)
				return
			}

			gameTypes = append(gameTypes, gameType)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(gameTypes)
	}
}
