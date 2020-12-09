package main

import (
	"agent/cutil"
	"agent/dblogger"
	"agent/dbmovielist"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const agentName = "umovie"
const umovieID = "87"
const base = "https://www.u-movie.com.tw/cinema/"
const url = base + "page.php?page_type=now&ver=tw&portal=cinema"

type movieTime struct {
	name string
	date string
	time string
}

func main() {
	var movieTimes []movieTime
	doc, err := cutil.GetDocument(url)
	if err != nil {
		logExit(err)
	}
	doc.Find("a.btn-info").Each(func(_ int, s *goquery.Selection) {
		h, exist := s.Attr("href")
		if !exist {
			logExit(fmt.Errorf("href not exist"))
		}
		doc2, err := cutil.GetDocument(base + h)
		if err != nil {
			logExit(err)
		}

		titleSel := doc2.Find("h3.text-large")
		if titleSel == nil {
			logExit(fmt.Errorf("failed to find title block"))
		}
		name := strings.TrimSpace(titleSel.Text())
		imgSel := titleSel.Find("img")
		if imgSel != nil {
			altStr, isExist := imgSel.Attr("alt")
			if isExist && strings.TrimSpace(altStr) != "" {
				name = strings.TrimSpace(altStr)
			}
		}
		date := ""
		titleSel.Parent().Find("div").Each(func(_ int, s *goquery.Selection) {
			clas, exist := s.Attr("class")
			if !exist {
				logExit(fmt.Errorf("class not exist"))
			}
			if clas == "row justify-content-center" { // date
				s.Find(".date").Contents().EachWithBreak(func(_ int, s *goquery.Selection) bool {
					if goquery.NodeName(s) == "#text" {
						date = strings.TrimSpace(s.Text())
						return false
					}
					return true
				})
			} else if clas == "ghost-button" { // time
				s.Contents().EachWithBreak(func(_ int, s *goquery.Selection) bool {
					if goquery.NodeName(s) == "#text" {
						time := strings.TrimSpace(s.Text())
						movieTimes = append(movieTimes, movieTime{name: name, date: date, time: time})
						return false
					}
					return true
				})
			}
		})
	})
	// group by movie and date
	m := make(map[string]movieTime)
	for _, mt := range movieTimes {
		key := mt.name + mt.date
		value, ok := m[key]
		if ok {
			value.time = value.time + "|" + mt.time
			m[key] = value
		} else {
			m[key] = mt
		}
	}
	movieTimes = movieTimes[:0]
	for _, v := range m {
		movieTimes = append(movieTimes, v)
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
	for _, m := range movieTimes {
		err := dbmovielist.InsertUpdate(m.date, m.name, umovieID, "高雄環球數位3D影城", "數位", m.time, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	for _, m := range movieTimes {
		err := dbmovielist.DeleteOther(umovieID, m.date, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	err = dbmovielist.DeleteOutdated(umovieID)
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
