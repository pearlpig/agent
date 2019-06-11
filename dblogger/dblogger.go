package dblogger

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql" //
)

var db *sql.DB

func init() {
	_db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", "VS_Admin", "(#$JGKhw-902j", "54.248.245.251", "agent"))
	if err != nil {
		log.Fatal(err)
	}
	db = _db
}

const insertStmt = "INSERT INTO `GoLog` (agent_id, isOK, message) VALUES (\"%s\", %s, \"%s\")"

// OK .
func OK(id, msg string) error {
	_, err := db.Exec(fmt.Sprintf(insertStmt, id, "TRUE", msg))
	return err
}

// FAIL .
func FAIL(id, msg string) error {
	_, err := db.Exec(fmt.Sprintf(insertStmt, id, "FALSE", msg))
	return err
}
