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
	"sort"
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"github.com/PuerkitoBio/goquery"
)

const agentName = "carnival"
const tid = "78"
const urlDomain = "http://www.3d-movies.tw"

type movie struct {
	date string
	name string
	time string
	showType string
}

func main() {
	official, err := cutil.GetDocument("http://www.3d-movies.tw/showtimes.asp")
	if err != nil {
		logExit(err)
	}

	movies := []movie{}
	official.Find("#myDate option").Each(func(_ int, q *goquery.Selection) {
		date, err := time.Parse("2006/01/02", q.Text())
		if err != nil {
			return
		}

		val, ok := q.Attr("value")
		if ok != true {
			logExit(fmt.Errorf("value not found"))
		}
		href := urlDomain + val
		doc, err := cutil.GetDocument(href)
		if err != nil {
			logExit(err)
		}

		doc.Find("[background=\"images/st2.jpg\"]").Each(func(_ int, q *goquery.Selection) {
			name, err := big5ToUtf8(q.Find(".s010").Text())
			if err != nil {
				logExit(err)
			}
			typ := "數位"
			r := regexp.MustCompile(`(.*)\((.*)\)`)
			if matchs := r.FindStringSubmatch(name); len(matchs) == 3 {
				name = matchs[1]
				typ = matchs[2]
			}

			timeList := []string{}
			q.Find(".time_color").Each(func(_ int, q *goquery.Selection) {
				if t := strings.TrimSpace(q.Text()); t != "" {
					r = regexp.MustCompile(`([0-9]{2}\:[0-9]{2})`)
					tmp := r.FindAllString(t, -1)
					if len(tmp) <= 0 {
						logExit(fmt.Errorf("failed to extract movie time"))
					}
					for _, t := range tmp {
						timeList = append(timeList, t)
					}
				}
			})

			sort.Slice(timeList, func(i, j int)bool{return timeList[i] < timeList[j]})
			movies = append(movies, movie{
				date: date.Format("2006-01-02"),
				name: name,
				time: strings.Join(timeList, "|"),
				showType: typ,
			})
		})
	})

	timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
	for _, m := range movies {
		err := dbmovielist.InsertUpdate(m.date, m.name, tid, "嘉年華影城", m.showType, m.time, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	for _, m := range movies {
		err := dbmovielist.DeleteOther(tid, m.date, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	err = dbmovielist.DeleteOutdated(tid)
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

func big5ToUtf8(s string) (string, error) {
	I := bytes.NewReader([]byte(s))
	O := transform.NewReader(I, traditionalchinese.Big5.NewDecoder())
	d, e := ioutil.ReadAll(O)
	if e != nil {
		return "", e
	}
	return string(d), nil
}