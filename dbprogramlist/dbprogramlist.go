package dbprogramlist

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" //
)

var db *sql.DB

func init() {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "VS_Admin", "(#$JGKhw-902j", "localhost", "channel_timetable"))
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}

// InsertUpdate .
func InsertUpdate(channelID, programName, date, time, source string) error {
	_, err := db.Exec(fmt.Sprintf("INSERT INTO timetable (channel_id, program_name, date, time, source) VALUES(\"%s\",\"%s\",\"%s\",\"%s\",\"%s\")"+
		" ON DUPLICATE KEY UPDATE program_name=VALUES(program_name),source=VALUES(source)", channelID, programName, date, time, source))
	return err
}

// DeleteOther .
func DeleteOther(channelID, date, source string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM timetable where channel_id = \"%s\" AND date = \"%s\" AND source != \"%s\" ", channelID, date, source))
	return err
}

// DeleteOutdated .
func DeleteOutdated(channelID string) error {
	_, err := db.Exec(fmt.Sprintf("DELETE FROM timetable WHERE channel_id = \"%s\" AND date < DATE_SUB(NOW(),INTERVAL 7 DAY)", channelID))
	return err
}
