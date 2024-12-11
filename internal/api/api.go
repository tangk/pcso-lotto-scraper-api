package api

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"pcso-lotto-scraper-api/internal/utils"
)

var db *sql.DB

func SetDB(database *sql.DB) {
	db = database
}

func GetLottoResults(w http.ResponseWriter, r *http.Request) {
	gameType, drawDateFrom, drawDateTo, args, err := utils.ValidateFilters(r, false, 1)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	query := `
			SELECT game_type, draw_date, jackpot, winners, 
			number1, number2, number3, number4, number5, number6 
			FROM lotto_results WHERE 1=1
		`

	// Filter by gameType
	if gameType != "" {
		query += " AND game_type = ?"
		args = append(args, gameType)
	}

	// Filter by draw date range
	query += " AND draw_date BETWEEN ? AND ?"
	args = append(args, drawDateFrom, drawDateTo)

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

		combination := utils.FormatCombination(numbers)
		results = append(results, map[string]interface{}{
			"gameType":    gameType,
			"drawDate":    drawDate,
			"jackpot":     jackpot,
			"winners":     winners,
			"combination": combination,
		})
	}

	utils.RespondWithJSON(w, results)
}

func GetHeatmap(w http.ResponseWriter, r *http.Request) {
	gameType, drawDateFrom, drawDateTo, args, err := utils.ValidateFilters(r, true, 30)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	args = append(args, gameType, drawDateFrom, drawDateTo)

	query := `
			SELECT number1, number2, number3, number4, number5, number6 
			FROM lotto_results 
			WHERE game_type = ? AND draw_date BETWEEN ? AND ?
		`

	rows, err := db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to fetch results", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

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

	utils.RespondWithJSON(w, heatmap)
}

func GetGameTypes(w http.ResponseWriter, r *http.Request) {
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
