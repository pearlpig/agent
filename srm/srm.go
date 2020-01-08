package main

import (
	"agent/cutil"
	"agent/dblogger"
	"agent/dbmovielist"

	"log"
	"os"
	"strings"
	"time"
	"bytes"
	"io/ioutil"

	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"github.com/PuerkitoBio/goquery"
)

const agentName = "srm"
const tid = "68"


type movie struct {
	date string
	name string
	time string
}

func main() {
	official, err := cutil.GetDocument("http://www.srm.com.tw/time/time.htm")
	if err != nil {
		logExit(err)
	}

	movies := []movie{}
	dates := []string{}
	qDate := official.Find("table tr").Eq(0)
	qDate.Find("td").Each(func(i int, q *goquery.Selection) {
		if i > 0 {
			str, err := big5ToUtf8(q.Find("font").Eq(0).Text())
			if err != nil {
				logExit(err)
			}
			str = strings.TrimSpace(str)
			date, err := time.Parse("1月2日", str)
			if err != nil {
				logExit(err)
			}
			if date.Month() == time.January && time.Now().Month() == time.December {
				date = date.AddDate(time.Now().Year()+1, 0, 0)
			} else {
				date = date.AddDate(time.Now().Year(), 0, 0)
			}

			dates = append(dates, date.Format("2006-01-02"))
		}
	})
	
	official.Find("table").Each(func(_ int, q *goquery.Selection) {
		name := ""
		q.Find("tr").Eq(1).Find("td").Each(func(i int, q *goquery.Selection) {
			timeList := []string{}
			if i == 0 {
				// movie name has two styles
				if q.Find(".style132").Size() == 0 {
					name = q.Find("h4[class=\"style130\"]").Text()
				} else {
					name = q.Find(".style132 p").Eq(0).Text()
				}

				name, err = big5ToUtf8(name)
				if err != nil {
					logExit(err)
				}
				
				name = strings.TrimSpace(name)
			} else {
				q.Find("p").Each(func(i int, q *goquery.Selection) {
					t, err := big5ToUtf8(q.Text())
					if err != nil {
						logExit(err)
					}
					t = strings.TrimSpace(t)
					if _, err := time.Parse("04:05", t); err == nil {
						timeList = append(timeList, t)
					}
				})
				
				if len(timeList) > 0 {
					movies = append(movies, movie{
						date: dates[i-1],
						name: name,
						time: strings.Join(timeList, "|"),
					})
				}
			}
		})
	})

		timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
		for _, m := range movies {
			err := dbmovielist.InsertUpdate(m.date, m.name, tid, "日新大戲院", "數位", m.time, timestamp)
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