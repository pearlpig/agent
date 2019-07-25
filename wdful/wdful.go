package main

import (
	"agent/cutil"
	"agent/dblogger"
	"agent/dbmovielist"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const agentName = "wdful"
const wdfulID = "62"
const base = "http://www.wdful.com.tw"
const home = "/Home/TimeTable"

type movie struct {
	date string
	name string
	time string
}

func main() {
	doc, err := cutil.GetDocument(base + home)
	if err != nil {
		logExit(err)
	}
	var dayUrls []string
	doc.Find("table a[href]").Each(func(_ int, d *goquery.Selection) {
		dayUrl, _ := d.Attr("href")
		dayUrls = append(dayUrls, base+dayUrl)
	})
	var movies []movie
	for _, url := range dayUrls {
		var date string
		doc, err := cutil.GetDocument(url)
		if err != nil {
			logExit(err)
		}
		doc.Find(".subtitle_j+.box_j").Each(func(_ int, s *goquery.Selection) {
			r := regexp.MustCompile("日期：(.*)")
			match := r.FindStringSubmatch(s.Text())
			if len(match) != 2 {
				logExit(fmt.Errorf("failed to extract movie date"))
			}
			date = strings.ReplaceAll(match[1], "/", "-")
		})

		var names []string
		doc.Find(".box_j .title_j").Each(func(_ int, s *goquery.Selection) {
			r := regexp.MustCompile("片名： *(.*) *\n")
			match := r.FindStringSubmatch(s.Text())
			if len(match) != 2 {
				logExit(fmt.Errorf("failed to extract movie name"))
			}
			names = append(names, match[1])
		})

		var times []string
		doc.Find(".timebox_j").Each(func(_ int, s *goquery.Selection) {
			var subTimes []string
			s.Find(".time_j").Each(func(_ int, ss *goquery.Selection) {
				if subTime := strings.TrimSpace(ss.Text()); subTime != "" {
					r := regexp.MustCompile("([0-2][0-9]:[0-5][0-9])")
					match := r.FindStringSubmatch(ss.Text())
					if len(match) != 2 {
						logExit(fmt.Errorf("failed to extract movie time"))
					}
					subTimes = append(subTimes, match[1])
				}
			})
			times = append(times, sortAndMergeTime(subTimes))
		})
		for i := 0; i < len(names); i++ {
			movies = append(movies, movie{
				date: date,
				name: names[i],
				time: times[i],
			})
		}
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
	for _, m := range movies {
		err := dbmovielist.InsertUpdate(m.date, m.name, wdfulID, "萬代福影城", "數位二輪", m.time, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	for _, m := range movies {
		err := dbmovielist.DeleteOther(wdfulID, m.date, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	err = dbmovielist.DeleteOutdated(wdfulID)
	if err != nil {
		logExit(err)
	}
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

func sortAndMergeTime(ts []string) string {
	var tTs []time.Time
	for _, t := range ts {
		tT, err := time.Parse("15:04", t)
		if err != nil {
			log.Println(err)
			continue
		}
		tTs = append(tTs, tT)
	}

	var midnightTs []time.Time
	var otherTs []time.Time
	line, _ := time.Parse("15:04", "06:00")
	for _, t := range tTs {
		if t.Before(line) {
			midnightTs = append(midnightTs, t)
		} else {
			otherTs = append(otherTs, t)
		}
	}
	tTs = append(otherTs, midnightTs...)
	for i, t := range tTs {
		ts[i] = t.Format("15:04")
	}
	ret := strings.Join(ts, "|")
	return ret
}
