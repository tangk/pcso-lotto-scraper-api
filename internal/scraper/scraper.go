package scraper

import (
	"strings"
	"time"

	"github.com/gocolly/colly"
)

type LottoResult struct {
	GameType    string
	Combination string
	DrawDate    string
	Jackpot     string
	Winners     string
}

var gameTypes = map[string]string{
	"6/58":     "1",
	"6/55":     "2",
	"6/49":     "3",
	"6/45":     "4",
	"6/42":     "5",
	"6D":       "6",
	"4D":       "7",
	"SWERTRES": "8",
	"EZ2":      "9",
}

func FetchDaily() ([]LottoResult, error) {
	c := colly.NewCollector()

	var results []LottoResult

	c.OnHTML("table#cphContainer_cpContent_GridView1 tbody tr", func(e *colly.HTMLElement) {
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

func FetchToLatest(fromDate time.Time) ([]LottoResult, error) {
	c := colly.NewCollector()

	var viewState string
	var viewGenerator string
	var eventValidation string

	c.OnHTML("input#__VIEWSTATE", func(e *colly.HTMLElement) {
		viewState = e.Attr("value")
	})

	c.OnHTML("input#__VIEWSTATEGENERATOR", func(e *colly.HTMLElement) {
		viewGenerator = e.Attr("value")
	})

	c.OnHTML("input#__EVENTVALIDATION", func(e *colly.HTMLElement) {
		eventValidation = e.Attr("value")
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("Content-Type", "application/x-www-form-urlencoded")
		r.Headers.Set("Referer", "https://www.pcso.gov.ph/SearchLottoResult.aspx")
	})

	var results []LottoResult

	c.OnHTML("table#cphContainer_cpContent_GridView1 tbody tr", func(e *colly.HTMLElement) {
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

	now := time.Now()

	formData := map[string]string{
		"__EVENTTARGET":        "",
		"__EVENTARGUMENT":      "",
		"__VIEWSTATE":          viewState,
		"__VIEWSTATEGENERATOR": viewGenerator,
		"__EVENTVALIDATION":    eventValidation,
		"ctl00$ctl00$cphContainer$cpContent$ddlStartMonth": fromDate.Format("January"),
		"ctl00$ctl00$cphContainer$cpContent$ddlStartDate":  fromDate.Format("2"),
		"ctl00$ctl00$cphContainer$cpContent$ddlStartYear":  fromDate.Format("2006"),
		"ctl00$ctl00$cphContainer$cpContent$ddlEndDay":     now.Format("2"),
		"ctl00$ctl00$cphContainer$cpContent$ddlEndMonth":   now.Format("January"),
		"ctl00$ctl00$cphContainer$cpContent$ddlEndYear":    now.Format("2006"),
		"ctl00$ctl00$cphContainer$cpContent$ddlSelectGame": "0",
		"ctl00$ctl00$cphContainer$cpContent$btnSearch":     "Search+Lotto",
	}

	err = c.Post("https://www.pcso.gov.ph/SearchLottoResult.aspx", formData)
	if err != nil {
		return nil, err
	}

	return results, nil
}
