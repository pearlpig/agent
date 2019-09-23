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

const insertStmt = "INSERT INTO `GoLog` (agent_name, isOK, message) VALUES (?, ?, ?)"
const TRUE = 1
const FALSE = 0

// OK .
func OK(agentName, msg string) error {
	_, err := db.Exec(insertStmt, agentName, TRUE, msg)
	return err
}

// FAIL .
func FAIL(agentName, msg string) error {
	_, err := db.Exec(insertStmt, agentName, FALSE, msg)
	return err
}
