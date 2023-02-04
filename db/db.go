package db

import (
	"database/sql"
	"log"
	"time"

	"github.com/mattn/go-sqlite3"
)

type EmailEntry struct {
	Id          int64
	Email       string
	ConfirmedAt *time.Time
	OptOut      bool
}

func CreateDB(db *sql.DB) {
	_, err := db.Exec(`
		CREATE TABLE emails (
			id 				INTEGER PRIMARY KEY,
			email 			TEXT UNIQUE,
			confirmed_at 	INTEGER,
			opt_out 		INTEGER
		);
	`)

	if err != nil {
		if sqlError, ok := err.(sqlite3.Error); ok {
			if sqlError.Code != 1 {
				log.Fatal(sqlError)
			}
		} else {
			log.Fatal(err)
		}
	}
}

func InsertEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out)
		VALUES (?, 0, false)
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func UpdateEmail(db *sql.DB, entry EmailEntry) error {
	unixTime := entry.ConfirmedAt.Unix()

	_, err := db.Exec(`
		INSERT INTO emails(email, confirmed_at, opt_out)
		VALUES (?, ?, ?)
		ON CONFLICT DO 
			UPDATE SET confirmed_at=?, opt_out=?
	`, entry.Email, unixTime, entry.OptOut, unixTime, entry.OptOut)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func GetEmail(db *sql.DB, email string) (*EmailEntry, error) {
	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE email = ?
	`, email)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	for rows.Next() {
		return bindEmail(rows)
	}

	return nil, nil
}

func DeleteEmail(db *sql.DB, email string) error {
	_, err := db.Exec(`
		UPDATE emails
		SET opt_out=true
		WHERE email=?
	`, email)

	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

type BatchEmailQueryParams struct {
	Page  int
	Count int
}

func GetEmailBatch(db *sql.DB, params BatchEmailQueryParams) ([]EmailEntry, error) {
	var empty []EmailEntry

	rows, err := db.Query(`
		SELECT id, email, confirmed_at, opt_out
		FROM emails
		WHERE opt_out=false
		ORDER BY id ASC
		LIMIT ? OFFSET ?
	`, params.Count, (params.Page-1)*params.Count)

	if err != nil {
		log.Println(err)
		return empty, err
	}

	defer rows.Close()

	emails := make([]EmailEntry, 0, params.Count)

	for rows.Next() {
		email, err := bindEmail(rows)
		if err != nil {
			return nil, err
		}
		emails = append(emails, *email)
	}

	return emails, nil
}

func bindEmail(row *sql.Rows) (*EmailEntry, error) {
	var id int64
	var email string
	var confirmedAt int64
	var optOut bool

	err := row.Scan(&id, &email, &confirmedAt, &optOut)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	unixTime := time.Unix(confirmedAt, 0)
	return &EmailEntry{Id: id, Email: email, ConfirmedAt: &unixTime, OptOut: optOut}, nil
}
