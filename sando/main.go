package main

import (
	"agent/dblogger"
	"agent/dbmovielist"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/djimenez/iconv-go"
)

const agentName = "sando"
const sandoID = "96"
const base = "http://www.sando.com.tw/"
const mainPage = "Tshow.asp"

type movie struct {
	name string
	date string
	time string
	typ  string
}

func main() {
	client := http.Client{}
	client.Timeout = time.Duration(10) * time.Second
	res, err := client.Get(base + mainPage)
	if err != nil {
		logExit(err)
	}
	if res.StatusCode != 200 {
		logExit(fmt.Errorf("bad reponse status code"))
	}
	utfBody, err := iconv.NewReader(res.Body, "big5", "utf-8")
	if err != nil {
		logExit(err)
	}
	doc, err := goquery.NewDocumentFromReader(utfBody)
	if err != nil {
		logExit(err)
	}
	var moviePages []string
	doc.Find("a > img[src='images/check.gif']").Each(func(_ int, s *goquery.Selection) {
		moviePage, exist := s.Parent().Attr("href")
		if !exist {
			logExit(fmt.Errorf("can not get href for movie page"))
		}
		moviePages = append(moviePages, moviePage)
	})

	var movies []movie
	timeRE := regexp.MustCompile("<b>(.*?)(<font face=\"Arial\" size=3>)(<font.*?>)?\\(.\\)(</font>)?(&nbsp;)?(.*?)</td>") // broken html tag
	nameRE := regexp.MustCompile("(\\*.*\\*)?(.*?)~(.*?)$")
	for _, moviePage := range moviePages {
		res, err := client.Get(base + moviePage)
		if err != nil {
			logExit(err)
		}
		utfBody, err := iconv.NewReader(res.Body, "big5", "utf-8")
		if err != nil {
			logExit(err)
		}
		doc, err := goquery.NewDocumentFromReader(utfBody)
		if err != nil {
			logExit(err)
		}
		titleInfo := doc.Find("b + img").Prev().Text()
		m := nameRE.FindStringSubmatch(titleInfo)
		if len(m) != 4 {
			logExit(fmt.Errorf("can not extract movie name"))
		}
		typ := m[2]
		name := m[3]
		res, err = client.Get(base + moviePage)
		if err != nil {
			logExit(err)
		}
		utfBody, err = iconv.NewReader(res.Body, "big5", "utf-8")
		if err != nil {
			logExit(err)
		}
		bytes, err := ioutil.ReadAll(utfBody)
		if err != nil {
			logExit(err)
		}
		body := string(bytes)
		matches := timeRE.FindAllStringSubmatch(body, -1)
		for _, matche := range matches {
			date, err := formatDate(matche[1])
			if err != nil {
				logExit(err)
			}
			time := formatTime(matche[6])
			movies = append(movies, movie{name, date, time, typ})
		}
	}
	timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
	for _, m := range movies {
		err := dbmovielist.InsertUpdate(m.date, m.name, sandoID, "三多數位3D影城", m.typ, m.time, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	for _, m := range movies {
		err := dbmovielist.DeleteOther(sandoID, m.date, timestamp)
		if err != nil {
			logExit(err)
		}
	}
	err = dbmovielist.DeleteOutdated(sandoID)
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

func formatDate(d string) (string, error) {
	now := time.Now()

	prev := strconv.Itoa(now.AddDate(-1, 0, 0).Year())
	this := strconv.Itoa(now.AddDate(0, 0, 0).Year())
	next := strconv.Itoa(now.AddDate(+1, 0, 0).Year())
	prevDate, err := time.Parse("1/2/2006", d+"/"+prev)
	if err != nil {
		return "", err
	}
	thisDate, err := time.Parse("1/2/2006", d+"/"+this)
	if err != nil {
		return "", err
	}
	nextDate, err := time.Parse("1/2/2006", d+"/"+next)
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
	if dthis < min {
		min = dthis
		minDate = thisDate
	}
	if dnext < min {
		min = dnext
		minDate = nextDate
	}
	return minDate.Format("2006-01-02"), nil
}

func formatTime(ts string) string {
	var times []string
	for _, t := range strings.Split(ts, ",") {
		times = append(times, strings.TrimSpace(t))
	}
	return strings.Join(times, "|")
}
