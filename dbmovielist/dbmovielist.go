package dbmovielist

import (
	"database/sql"
	"fmt"
	"log"
	"unicode/utf8"

	"cicd.icu/cyberon/config"
	_ "github.com/go-sql-driver/mysql" //
)

var db *sql.DB

func init() {
	host := config.GetString("mysql.host", "localhost")
	database := config.GetString("mysql.database", "movie2")
	username := config.GetString("mysql.username", "VS_Admin")
	password := config.GetString("mysql.password", "(#$JGKhw-902j")
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, host, database))
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}

// InsertUpdate .
func InsertUpdate(date, movie_name, tid, theater_name, show_type, time, source string) error {
	mid, err := getMovieID(movie_name)
	if err != nil {
		return err
	}
	_, err = db.Exec("INSERT INTO time_table (date, mid, movie_name, tid, theater_name, show_type, time, source) VALUES (?,?,?,?,?,?,?,?) ON DUPLICATE KEY UPDATE time=VALUES(time), source=VALUES(source), theater_name=VALUES(theater_name)",
		date, mid, movie_name, tid, theater_name, show_type, time, source)

	return err
}

// DeleteOther .
func DeleteOther(tid, date, source string) error {
	_, err := db.Exec("DELETE FROM time_table WHERE tid = ? AND date = ? AND source != ? ", tid, date, source)
	return err
}

// DeleteOutdated .
func DeleteOutdated(tid string) error {
	_, err := db.Exec("DELETE FROM time_table WHERE tid= ? AND date < DATE_SUB(NOW(),INTERVAL 7 DAY)", tid)
	return err
}

func getMovieID(name string) (string, error) {
	// compare with name
	rows, err := db.Query(fmt.Sprintf("SELECT id, name FROM movie"))
	if err != nil {
		return "", err
	}

	type m struct {
		id           string
		name         string
		english_name string
	}
	ms := make([]m, 1024)
	for rows.Next() {
		var id string
		var name string
		err = rows.Scan(&id, &name)
		if err != nil {
			return "", err
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
	// compare with chinese+english name
	rows, err = db.Query(fmt.Sprintf("SELECT id, name, english_name FROM movie"))
	if err != nil {
		return "", err
	}
	ms = make([]m, 1024)
	for rows.Next() {
		var id string
		var name string
		var english_name string
		err = rows.Scan(&id, &name, &english_name)
		if err != nil {
			return "", err
		}
		ms = append(ms, m{id: id, name: name, english_name: english_name})
	}
	min = 9999
	minName = ""
	minFullName := ""
	minId = ""
	for _, m := range ms {
		if m.name == "" || m.id == "" || m.english_name == "" {
			continue
		}
		cost := LevenshteinDist(m.name+m.english_name, name)
		if cost < min {
			min = cost
			minId = m.id
			minFullName = m.name + m.english_name
			minName = m.name
		}
	}
	cost = float64(min) / float64(utf8.RuneCountInString(name)+utf8.RuneCountInString(minFullName))
	if cost <= 0.4 {
		return minId, nil
	}
	return "", fmt.Errorf("no match movie name in db: " + name)
}
