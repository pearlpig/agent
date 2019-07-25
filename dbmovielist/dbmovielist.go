package dbmovielist

import (
	"database/sql"
	"fmt"
	"log"
	"unicode/utf8"

	_ "github.com/go-sql-driver/mysql" //
)

var db *sql.DB

func init() {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "VS_Admin", "(#$JGKhw-902j", "localhost", "movie2"))
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}

// InsertUpdate .
func InsertUpdate(date, movie_name, tid, theater_name, show_type, time, source string) error {
	mid, err := GetMovieID(movie_name)
	if err != nil {
		return err
	}

	_, err = db.Exec(fmt.Sprintf(
		"INSERT INTO time_table (date, mid, movie_name, tid, theater_name, show_type, time, source)"+
			" VALUES (\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\",\"%s\")"+
			" ON DUPLICATE KEY UPDATE time=VALUES(time), source=VALUES(source), theater_name=VALUES(theater_name)",
		date, mid, movie_name, tid, theater_name, show_type, time, source))

	return err
}

// DeleteOther .
func DeleteOther(tid, date, source string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM time_table WHERE tid =\"%s\" AND date = \"%s\" AND source != \"%s\" ",
		tid, date, source))
	return err
}

// DeleteOutdated .
func DeleteOutdated(tid string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM time_table WHERE tid= \"%s\" AND date < DATE_SUB(NOW(),INTERVAL 7 DAY)", tid))
	return err
}

func GetMovieID(name string) (string, error) {
	rows, err := db.Query(fmt.Sprintf("SELECT id, name FROM movie"))
	if err != nil {
		log.Fatal(err)
	}

	type m struct {
		id   string
		name string
	}
	ms := make([]m, 1024)
	for rows.Next() {
		var id string
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			log.Fatal(err)
		}
		ms = append(ms, m{id: id, name: name})
	}
	min := 9999
	minName := ""
	minId := ""
	for _, m := range ms {
		if m.name == "" || m.id == "" {
			continue
		}
		cost := LevenshteinDist(m.name, name)
		if cost < min {
			min = cost
			minId = m.id
			minName = m.name
		}
	}
	cost := float64(min) / float64(utf8.RuneCountInString(name)+utf8.RuneCountInString(minName))
	if cost <= 0.4 {
		return minId, nil
	}
	return "", fmt.Errorf("no match movie name in db")
}
