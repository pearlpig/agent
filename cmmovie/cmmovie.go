package main

import (
	"agent/cutil"
	"agent/dblogger"
	"agent/dbmovielist"

	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const agentName = "cmmovie"
const tid1 = "86"
const tid2 = "81"

type movie struct {
	date string
	name string
	time string
}

func main() {
	official, err := cutil.GetDocument("https://www.cm-movie.com.tw/category/time/")
	if err != nil {
		logExit(err)
	}

	official.Find("figcaption a").Each(func(_ int, q *goquery.Selection) {
		href, ok := q.Attr("href")
		if ok != true {
			logExit(fmt.Errorf("href not found"))
		}

		doc, err := cutil.GetDocument(href)
		if err != nil {
			logExit(err)
		}

		names := make(map[string]string)
		times := make(map[string]([]string))
		doc.Find("[title=\"今日戲院\"] + div td").Each(func(_ int, q *goquery.Selection) {
			q.Find("b").Each(func(_ int, q *goquery.Selection) {
				name := q.Text()
				r := regexp.MustCompile("^(.*?)\\(.*?\\)+")
				m := r.FindAllStringSubmatch(name, -1)
				if len(m) != 1 || len(m[0]) != 2 {
					logExit(fmt.Errorf("fail to extract time"))
				}
				name = m[0][1]
				names[string([]rune(name)[:1])] = name
			})
			q.Find("h4").Each(func(_ int, q *goquery.Selection) {
				if t := strings.TrimSpace(q.Text()); t != "" {
					name := string([]rune(t)[:1])
					time := string([]rune(t)[1:])
					times[names[name]] = append(times[names[name]], strings.TrimSpace(time))
				}
			})
		})

		dates := parseDates(doc.Find("#content header h1").Text())
		for _, date := range dates {
			timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
			for name, time := range times {
				err := dbmovielist.InsertUpdate(
					date.Format("2006-01-02"),
					name,
					tid1,
					"今日戲院",
					"數位",
					strings.Join(time, "|"),
					timestamp,
				)

				if err != nil {
					logExit(err)
				}
			}
			err := dbmovielist.DeleteOther(tid1, date.Format("2006-01-02"), timestamp)
			if err != nil {
				logExit(err)
			}
		}

		err = dbmovielist.DeleteOutdated(tid1)
		if err != nil {
			logExit(err)
		}

		//------------------------------------------------------------------------------------------

		names = make(map[string]string)
		times = make(map[string]([]string))
		doc.Find("[title=\"全美戲院\"] + div td").Each(func(_ int, q *goquery.Selection) {
			q.Find("b").Each(func(_ int, q *goquery.Selection) {
				name := q.Text()
				r := regexp.MustCompile("^(.*?)\\(.*?\\)+")
				m := r.FindAllStringSubmatch(name, -1)
				if len(m) != 1 || len(m[0]) != 2 {
					logExit(fmt.Errorf("fail to extract time"))
				}
				name = m[0][1]
				names[string([]rune(name)[:1])] = name
			})
			q.Find("h4").Each(func(_ int, q *goquery.Selection) {
				if t := strings.TrimSpace(q.Text()); t != "" {
					name := string([]rune(t)[:1])
					time := string([]rune(t)[1:])
					times[names[name]] = append(times[names[name]], strings.TrimSpace(time))
				}
			})
		})

		dates = parseDates(doc.Find("#content header h1").Text())
		for _, date := range dates {
			timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
			for name, time := range times {
				err := dbmovielist.InsertUpdate(
					date.Format("2006-01-02"),
					name,
					tid2,
					"全美戲院",
					"數位",
					strings.Join(time, "|"),
					timestamp,
				)

				if err != nil {
					logExit(err)
				}
			}
			err := dbmovielist.DeleteOther(tid2, date.Format("2006-01-02"), timestamp)
			if err != nil {
				logExit(err)
			}
		}

		err = dbmovielist.DeleteOutdated(tid2)
		if err != nil {
			logExit(err)
		}
	})

	err = dblogger.OK(agentName, "OK")
	if err != nil {
		logExit(err)
	}
}

func logExit(err error) {
	err = dblogger.FAIL(agentName, err.Error())
	if err != nil {
		log.Println(err)
	}
	os.Exit(1)
}

func parseDates(source string) []time.Time {
	r := regexp.MustCompile("(.*?)~(.*?)\\(")
	m := r.FindAllStringSubmatch(source, -1)
	if len(m) != 1 || len(m[0]) != 3 {
		logExit(fmt.Errorf("fail to extract date"))
	}
	start, err := time.Parse("1月2日", m[0][1])
	if err != nil {
		logExit(fmt.Errorf("fail to extract date"))
	}
	end, err := time.Parse("1月2日", m[0][2])
	if err != nil {
		logExit(fmt.Errorf("fail to extract date"))
	}
	var dates []time.Time
	current := start
	for ; !current.After(end); current = current.AddDate(0, 0, 1) {
		dates = append(dates, guessYear(current))
	}
	return dates
}

func guessYear(t time.Time) time.Time {
	date := t.Format("01/02")
	now := time.Now()

	prev := strconv.Itoa(now.AddDate(-1, 0, 0).Year())
	this := strconv.Itoa(now.AddDate(+0, 0, 0).Year())
	next := strconv.Itoa(now.AddDate(+1, 0, 0).Year())

	prevDate, _ := time.Parse("01/02/2006", date+"/"+prev)
	thisDate, _ := time.Parse("01/02/2006", date+"/"+this)
	nextDate, _ := time.Parse("01/02/2006", date+"/"+next)

	dprev := math.Abs(now.Sub(prevDate).Seconds())
	dthis := math.Abs(now.Sub(thisDate).Seconds())
	dnext := math.Abs(now.Sub(nextDate).Seconds())

	min := math.MaxFloat64
	var minDate time.Time
	if dprev < min {
		min = dprev
		minDate = prevDate
	}
	if dnext < min {
		min = dnext
		minDate = nextDate
	}
	if dthis < min {
		min = dthis
		minDate = thisDate
	}
	return minDate
}
