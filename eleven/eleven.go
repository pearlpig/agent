package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"agent/cutil"
	"agent/dblogger"
	"agent/dbprogramlist"
)

const agentID = "2"
const elevenID = "1176"

const elevenSports1Fmt = "https://apis.v-saas.com:9502/content/api/getContentEpg?contentId=1&dateStart=%s&dateStop=%s&lang=zh-CHT"

type apiData struct {
	Data []map[string][]program `json:"data"`
}

type program struct {
	Name  string `json:"name"`
	Start string `json:"start"`
}

func main() {
	bytes, err := cutil.Curl(urlGen())
	if err != nil {
		logExit(err)
	}
	var data apiData
	err = json.Unmarshal(bytes, &data)
	if err != nil {
		logExit(err)
	}
	for _, dateObj := range data.Data {
		for date, ps := range dateObj {
			timestamp := time.Now().Format("2006-01-02 15:04:05.0000")
			for _, p := range ps {
				err = dbprogramlist.InsertUpdate(elevenID, p.Name, date, p.Start+":00", timestamp)
				if err != nil {
					logExit(err)
				}
			}
			err = dbprogramlist.DeleteOther(elevenID, date, timestamp)
			if err != nil {
				logExit(err)
			}
		}
	}
	err = dbprogramlist.DeleteOutdated(elevenID)
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

func urlGen() string {
	lastWeek := time.Now().Add(-24 * time.Hour).Format("2006-01-02")
	nextWeek := time.Now().Add(+24 * time.Hour).Format("2006-01-02")
	return fmt.Sprintf(elevenSports1Fmt, lastWeek, nextWeek)
}
