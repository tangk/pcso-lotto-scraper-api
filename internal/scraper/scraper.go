package scraper

import (
	"strings"

	"github.com/gocolly/colly"
)

type LottoResult struct {
	GameType    string
	Combination string
	DrawDate    string
	Jackpot     string
	Winners     string
}

func FetchLottoResults() ([]LottoResult, error) {
	c := colly.NewCollector()

	var results []LottoResult

	c.OnHTML("table#cphContainer_cpContent_GridView1 tbody tr", func(e *colly.HTMLElement) {
		// Skip header row
		if e.Index == 0 {
			return
		}

		result := LottoResult{
			GameType:    strings.TrimSpace(e.ChildText("td:nth-child(1)")),
			Combination: strings.TrimSpace(e.ChildText("td:nth-child(2)")),
			DrawDate:    strings.TrimSpace(e.ChildText("td:nth-child(3)")),
			Jackpot:     strings.TrimSpace(e.ChildText("td:nth-child(4)")),
			Winners:     strings.TrimSpace(e.ChildText("td:nth-child(5)")),
		}

		results = append(results, result)
	})

	err := c.Visit("https://www.pcso.gov.ph/SearchLottoResult.aspx")
	if err != nil {
		return nil, err
	}

	return results, nil
}
