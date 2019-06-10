package main

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"agent/dblogger"

	"github.com/PuerkitoBio/goquery"
	_ "github.com/go-sql-driver/mysql"
)

type program struct {
	date string
	time string
	name string
}

const agentID = "5"
const momokidsID = "67"

func main() {
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "VS_Admin", "(#$JGKhw-902j", "10.0.0.92", "channel_timetable"))
	if err != nil {
		err = dblogger.FAIL(agentID, err.Error())
		if err != nil {
			log.Println(err)
		}
		return
	}

	doc, err := getDocument("https://www.momokids.com.tw/program-time.php?lmenuid=8")
	if err != nil {
		err = dblogger.FAIL(agentID, err.Error())
		if err != nil {
			log.Println(err)
		}
		return
	}
	doc.Find(".carousel-item").Each(func(_ int, day *goquery.Selection) {
		// "06/12 星期三" to "2019-06-12"
		date, err := formatDate(day.Find(".tit .t0").Text())
		if err != nil {
			err = dblogger.FAIL(agentID, err.Error())
			if err != nil {
				log.Println(err)
			}
			os.Exit(1)
		}
		programs := make([]program, 0)
		day.Find(".program-time > div > div").Each(func(_ int, s *goquery.Selection) {
			time := strings.TrimSpace(s.Find(".col-2.col-md-3.col-lg-2.time.text-center").Text()) + ":00" //  "23:00" to "23:00:00"
			name := strings.TrimSpace(s.Find(".col-6.col-md-6.col-lg-7").Text())
			programs = append(programs, program{date, time, name})
		})

		if len(programs) == 0 {
			err = dblogger.FAIL(agentID, "failed to get programs")
			if err != nil {
				log.Println(err)
			}
			os.Exit(1)
		} else {
			timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
			for _, p := range programs {
				_, err = db.Exec(insertUpdateStmt(momokidsID, p.name, p.date, p.time, timestamp))
				if err != nil {
					err = dblogger.FAIL(agentID, err.Error())
					if err != nil {
						log.Println(err)
					}
					os.Exit(1)
				}
			}
			_, err := db.Exec(deleteOtherStmt(momokidsID, date, timestamp))
			if err != nil {
				err = dblogger.FAIL(agentID, err.Error())
				if err != nil {
					log.Println(err)
				}
				os.Exit(1)
			}
		}
	})
	_, err = db.Exec(deleteOutdatedStmt(momokidsID))
	if err != nil {
		err = dblogger.FAIL(agentID, err.Error())
		if err != nil {
			log.Println(err)
		}
		return
	}
	err = dblogger.OK(agentID, "OK")
	if err != nil {
		log.Println(err)
	}
}

func insertUpdateStmt(channelID, programName, date, time, source string) string {
	return fmt.Sprintf("INSERT INTO timetable (channel_id, program_name, date, time, source) VALUES(\"%s\",\"%s\",\"%s\",\"%s\",\"%s\")"+
		" ON DUPLICATE KEY UPDATE program_name=VALUES(program_name),source=VALUES(source)", channelID, programName, date, time, source)
}

func deleteOtherStmt(channelID, date, source string) string {
	return fmt.Sprintf("DELETE FROM timetable where channel_id = \"%s\" AND date = \"%s\" AND source != \"%s\" ", channelID, date, source)
}

func deleteOutdatedStmt(channelID string) string {
	return fmt.Sprintf("DELETE FROM timetable WHERE channel_id = \"%s\" AND date < DATE_SUB(NOW(),INTERVAL 7 DAY)", channelID)
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

func getDocument(url string) (*goquery.Document, error) {
	client := http.Client{}
	client.Timeout = time.Duration(10) * time.Second
	res, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("status code error: %d %s", res.StatusCode, res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}
