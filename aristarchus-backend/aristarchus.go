package main

import (
	"fmt"
	"log"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func main() {
	db, err := sql.Open("sqlite3", "../db/books")
	if err != nil {
		log.Fatal(err)
	}
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected to db!")

	var volumes int
	err = db.QueryRow("SELECT COUNT(book_id) FROM books").Scan(&volumes)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %v books in library:\n", volumes)

	rows, err := db.Query("SELECT title, year FROM books ORDER BY title, year")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var title string
		var year int
		err = rows.Scan(&title, &year)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v, %v\n", title, year)
	}
}
