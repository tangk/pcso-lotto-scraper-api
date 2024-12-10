package database

import (
	"database/sql"
	"log"
	"strconv"
	"strings"

	"pcso-lotto-scraper-api/internal/utils"

	_ "github.com/mattn/go-sqlite3"
)

const createTableQuery = `
CREATE TABLE IF NOT EXISTS lotto_results (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    game_type TEXT,
    draw_date DATE,
    jackpot DECIMAL,
    winners INTEGER,
    number1 INTEGER,
    number2 INTEGER,
    number3 INTEGER,
    number4 INTEGER,
    number5 INTEGER,
    number6 INTEGER,
    UNIQUE(game_type, draw_date)
);

CREATE TABLE IF NOT EXISTS meta (
    key TEXT PRIMARY KEY,
    value TEXT
);
`

func InitDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./lotto.db")
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the database has been seeded
	var seeded string
	err = db.QueryRow("SELECT value FROM meta WHERE key = 'seeded'").Scan(&seeded)
	if err == sql.ErrNoRows {
		log.Println("Seeding database...")
		// Seed the database with data from the CSV file
		records, err := utils.ReadCSV("lotto_export.csv")
		if err != nil {
			log.Fatal(err)
		}

		var results []map[string]string

		for _, record := range records {
			if len(record) < 5 {
				log.Println("Invalid record:", record)
				continue
			}

			results = append(results, map[string]string{
				"gameType":    record[0],
				"combination": record[1],
				"drawDate":    record[2],
				"jackpot":     record[3],
				"winners":     record[4],
			})
		}

		if err != nil {
			log.Fatal(err)
		}

		err = InsertLottoResults(db, results)
		if err != nil {
			log.Fatal(err)
		}

		// Mark the database as seeded
		_, err = db.Exec("INSERT INTO meta (key, value) VALUES ('seeded', 'true')")
		if err != nil {
			log.Fatal(err)
		}
	} else if err != nil {
		log.Fatal(err)
	}

	return db
}

func InsertLottoResults(db *sql.DB, results []map[string]string) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	stmt, err := tx.Prepare(`
        INSERT INTO lotto_results (game_type, draw_date, jackpot, winners, number1, number2, number3, number4, number5, number6)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
        ON CONFLICT(game_type, draw_date) DO UPDATE SET
            jackpot = excluded.jackpot,
            winners = excluded.winners,
            number1 = excluded.number1,
            number2 = excluded.number2,
            number3 = excluded.number3,
            number4 = excluded.number4,
            number5 = excluded.number5,
            number6 = excluded.number6
    `)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, result := range results {
		drawDate, err := utils.ParseDate(result["drawDate"])
		if err != nil {
			log.Println("Failed to parse draw date:", err)
			continue
		}

		// Remove commas from jackpot string before parsing
		jackpotStr := strings.ReplaceAll(result["jackpot"], ",", "")
		jackpot, err := strconv.ParseFloat(jackpotStr, 64)
		if err != nil {
			log.Println("Failed to parse jackpot:", err)
			continue
		}

		winners, err := strconv.Atoi(result["winners"])
		if err != nil {
			log.Println("Failed to parse winners:", err)
			continue
		}

		// Split combination into individual numbers
		numbers := strings.Split(result["combination"], "-")
		var numValues [6]*int
		for i := 0; i < 6; i++ {
			if i < len(numbers) {
				num, err := strconv.Atoi(numbers[i])
				if err != nil {
					log.Println("Failed to parse combination number:", result["drawDate"], result["gameType"], err)
					continue
				}
				numValues[i] = &num
			} else {
				numValues[i] = nil
			}
		}

		_, err = stmt.Exec(
			result["gameType"], drawDate, jackpot, winners,
			numValues[0], numValues[1], numValues[2], numValues[3], numValues[4], numValues[5],
		)
		if err != nil {
			log.Println("Failed to insert record:", result["drawDate"], result["gameType"], err)
			continue
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
