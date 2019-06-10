package main

import (
	"errors"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"agent/cutil"
	"agent/dblogger"
	"agent/dbprogramlist"

	"github.com/PuerkitoBio/goquery"
)

type program struct {
	date string
	time string
	name string
}

const agentID = "1"
const momokidsID = "67"

func main() {

	doc, err := cutil.GetDocument("https://www.momokids.com.tw/program-time.php?lmenuid=8")
	if err != nil {
		logExit(err)
	}

	doc.Find(".carousel-item").Each(func(_ int, day *goquery.Selection) {
		// "06/12 星期三" to "2019-06-12"
		date, err := formatDate(day.Find(".tit .t0").Text())
		if err != nil {
			logExit(err)
		}
		programs := make([]program, 0)
		day.Find(".program-time > div > div").Each(func(_ int, s *goquery.Selection) {
			time := strings.TrimSpace(s.Find(".col-2.col-md-3.col-lg-2.time.text-center").Text()) + ":00" // "23:00" to "23:00:00"
			name := strings.TrimSpace(s.Find(".col-6.col-md-6.col-lg-7").Text())
			programs = append(programs, program{date, time, name})
		})

		if len(programs) == 0 {
			logExit(errors.New("failed to get programs"))
		} else {
			timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
			for _, p := range programs {
				err = dbprogramlist.InsertUpdate(momokidsID, p.name, p.date, p.time, timestamp)
				if err != nil {
					logExit(err)
				}
			}
			err = dbprogramlist.DeleteOther(momokidsID, date, timestamp)
			if err != nil {
				logExit(err)
			}
		}
	})

	err = dbprogramlist.DeleteOutdated(momokidsID)
	if err != nil {
		logExit(err)
	}
	err = dblogger.OK(agentID, "OK")
	if err != nil {
		logExit(err)
	}
}

func logExit(err error) {
	err = dblogger.FAIL(agentID, err.Error())
	if err != nil {
		log.Println(err)
	}
	os.Exit(1)
}

func formatDate(s string) (string, error) {
	now := time.Now()
	date := strings.Fields(s)[0]

	prev := strconv.Itoa(time.Now().AddDate(-1, 0, 0).Year())
	this := strconv.Itoa(time.Now().AddDate(+0, 0, 0).Year())
	next := strconv.Itoa(time.Now().AddDate(+1, 0, 0).Year())

	prevDate, err := time.Parse("01/02/2006", date+"/"+prev)
	thisDate, err := time.Parse("01/02/2006", date+"/"+this)
	nextDate, err := time.Parse("01/02/2006", date+"/"+next)
	if err != nil {
		return "", err
	}

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
	return minDate.Format("2006-01-02"), nil
}
