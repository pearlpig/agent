package main

import (
	"agent/cutil"
	"agent/dblogger"
	"agent/dbmovielist"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"

	"log"

	"github.com/PuerkitoBio/goquery"
)

const agentName = "sunrise"
const tid = "68"
const urlDomain = "https://srm.com.tw/"

type movie struct {
	date     string
	name     string
	time     string
	showType string
}

func logExit(err error) {
	err = dblogger.FAIL(agentName, err.Error())
	if err != nil {
		log.Println(err)
	}
	os.Exit(1)
}

func main() {
	official, err := cutil.GetDocument("https://srm.com.tw/%e5%a0%b4%e6%ac%a1%e6%9f%a5%e8%a9%a2/")
	if err != nil {
		log.Println(err)
		logExit(err)
	}

	movies := []movie{}
	movieList := map[string]interface{}{}
	log.Println("length: ", official.Find("div.flex_column_table").Length())
	official.Find("div.flex_column_table").Each(func(_ int, q *goquery.Selection) {
		movieName := q.Find("h2.av-special-heading-tag").Text()
		if movieName == "" {
			return
		}
		log.Println("get movie: ", movieName)
		// typ := "數位"

		now := time.Now()
		dateList := []string{}
		timeTable := q.Find("div.avia-data-table-wrap table")
		timeTable.Find("tr").Each(func(i int, tr *goquery.Selection) {
			if i == 0 {
				tr.Find("td").Each(func(_ int, td *goquery.Selection) {
					r := regexp.MustCompile(`([0-9]{1,2}\/[0-9]{1,2})`)
					tmp := r.FindString(td.Text())
					log.Println(td.Text(), " parse date: ", tmp)
					if tmp == "" {
						log.Println(fmt.Errorf("can't parse date"))
						logExit(fmt.Errorf("can't parse date"))
					}
					timeTarget, err := time.Parse("2006/1/2", fmt.Sprintf("%d", now.Year())+"/"+tmp)
					if err != nil {
						log.Println(err)
						logExit(err)
					}
					diff := now.Sub(timeTarget)
					if diff.Hours() >= 24*365 {
						log.Println("time diff: ", diff.Hours())
						timeTarget = timeTarget.AddDate(1, 0, 0)
					}
					dateList = append(dateList, timeTarget.Format("2006-01-02"))
				})
				return
			}
			tr.Find("td").Each(func(tdIdx int, td *goquery.Selection) {
				nowType := "數位"
				typR := regexp.MustCompile(`(.+[版|廳])`)
				r := regexp.MustCompile(`([0-9]{1,2}\s?\:\s?[0-9]{1,2})`)
				//找第1列字串因為沒有用 <p> 包起來，所以用換行分，如果是什麼版什麼廳就當成 type
				strList := strings.Split(td.Text(), "\n")
				timeList := []string{}
				if r.FindString(strList[0]) != "" {
					tmpTime := strings.ReplaceAll(strList[0], " ", "")
					timeList = append(timeList, tmpTime)
				} else if typR.FindString(strList[0]) != "" {
					nowType = strList[0]
				}

				td.Find("p").Each(func(_ int, p *goquery.Selection) {
					if r.FindString(p.Text()) != "" {
						tmpTime := strings.ReplaceAll(p.Text(), " ", "")
						timeList = append(timeList, tmpTime)
					} else if typR.FindString(p.Text()) != "" {
						if _, ok := movieList[movieName]; !ok {
							movieList[movieName] = map[string]interface{}{}
						}
						dateInfo := movieList[movieName].(map[string]interface{})
						if _, ok := dateInfo[dateList[tdIdx]]; !ok {
							dateInfo[dateList[tdIdx]] = map[string][]string{}
						}
						typeInfo := dateInfo[dateList[tdIdx]].(map[string][]string)
						if _, ok := typeInfo[nowType]; !ok {
							typeInfo[nowType] = []string{}
						}
						timeList = append(typeInfo[nowType], timeList...)
						sort.Slice(timeList, func(a, b int) bool {
							if timeList[a] <= "03:00" && timeList[b] <= "03:00" {
								return timeList[a] < timeList[b]
							} else if timeList[a] <= "03:00" {
								return false
							} else if timeList[b] <= "03:00" {
								return true
							}
							return timeList[a] < timeList[b]
						})
						typeInfo[nowType] = timeList
						dateInfo[dateList[tdIdx]] = typeInfo
						movieList[movieName] = dateInfo
						nowType = p.Text()
						timeList = []string{}
					}
				})

				if len(timeList) == 0 {
					return
				}
				if _, ok := movieList[movieName]; !ok {
					movieList[movieName] = map[string]interface{}{}
				}
				dateInfo := movieList[movieName].(map[string]interface{})
				if _, ok := dateInfo[dateList[tdIdx]]; !ok {
					dateInfo[dateList[tdIdx]] = map[string][]string{}
				}
				typeInfo := dateInfo[dateList[tdIdx]].(map[string][]string)
				if _, ok := typeInfo[nowType]; !ok {
					typeInfo[nowType] = []string{}
				}
				timeList = append(typeInfo[nowType], timeList...)
				sort.Slice(timeList, func(a, b int) bool {
					if timeList[a] <= "03:00" && timeList[b] <= "03:00" {
						return timeList[a] < timeList[b]
					} else if timeList[a] <= "03:00" {
						return false
					} else if timeList[b] <= "03:00" {
						return true
					}
					return timeList[a] < timeList[b]
				})
				typeInfo[nowType] = timeList
				dateInfo[dateList[tdIdx]] = typeInfo
				movieList[movieName] = dateInfo
			})
		})
		for name, v := range movieList {
			dateList := v.(map[string]interface{})
			for date, v := range dateList {
				typList := v.(map[string][]string)
				for typ, timeList := range typList {
					movies = append(movies, movie{
						date:     date,
						name:     name,
						time:     strings.Join(timeList, "|"),
						showType: typ,
					})
				}
			}
		}
	})
	// log.Println(movies)
	// return
	timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
	for _, m := range movies {
		err := dbmovielist.InsertUpdate(m.date, m.name, tid, "日新大戲院", m.showType, m.time, timestamp)
		if err != nil {
			log.Println(err)
			logExit(err)
		}
	}
	for _, m := range movies {
		err := dbmovielist.DeleteOther(tid, m.date, timestamp)
		if err != nil {
			log.Println(err)
			logExit(err)
		}
	}
	err = dbmovielist.DeleteOutdated(tid)
	if err != nil {
		log.Println(err)
		logExit(err)
	}
	err = dblogger.OK(agentName, "OK")
	if err != nil {
		log.Println(err)
		logExit(err)
	}
}
