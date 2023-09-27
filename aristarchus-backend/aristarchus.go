package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type PurchasedDate struct {
	year  int
	month time.Month
	day   int
}

func (pd PurchasedDate) String() string {
	if pd.year == 0 {
		return ""
	}
	if pd.month == 0 {
		return fmt.Sprintf("%v", pd.year)
	}
	if pd.day == 0 {
		return fmt.Sprintf("%v %v", pd.month, pd.year)
	}
	return fmt.Sprintf("%v %v %v", pd.day, pd.month, pd.year)
}

func (pd *PurchasedDate) setDate(s string) error {
    params := strings.Split(s, " ")
	switch len(params) {
	case 0:
	    return fmt.Errorf("setDate: Can't convert date %v", s)
	case 1:
	    dateString := "2006"
		t, err := time.Parse(dateString, s)
		if err != nil {
			return fmt.Errorf("setDate: Problem parsing year %v, %v", s, err)
		}
		pd.year = t.Year()
	case 2:
	    dateString := "January 2006"
		t, err := time.Parse(dateString, s)
		if err != nil {
			return fmt.Errorf("setDate: Problem parsing month-year date %v, %v",
			    s, err)
		}
		pd.year = t.Year()
		pd.month = t.Month()
	case 3:
	    dateString := "2 January 2006"
		t, err := time.Parse(dateString, s)
		if err != nil {
			return fmt.Errorf("setDate: Problem parsing day-month-year date %v, %v",
				s, err)
		}
		pd.year = t.Year()
		pd.month = t.Month()
		pd.day = t.Day()
	}
	return nil
}

type Book struct {
	id        int
	author    string
	editor    string
	title     string
	subtitle  string
	year      int
	edition   int
	publisher string
	isbn      string
	series    string
	status    string
	purchased PurchasedDate
}

func (b Book) String() string {
	return fmt.Sprintf("%v, %v (%v) [%v]", b.authorEditor(), b.fullTitle(),
		b.year, b.status)
}

func (b Book) authorEditor() string {
	if len(b.author) > 0 {
		return fmt.Sprintf("%v", b.author)
	} else if len(b.editor) > 0 {
		return fmt.Sprintf("%v (ed.)", b.editor)
	} else {
		return fmt.Sprintf("[No author]")
	}
}

func (b Book) fullTitle() string {
	if len(b.subtitle) > 0 {
		return fmt.Sprintf("%v: %v", b.title, b.subtitle)
	} else {
		return fmt.Sprintf("%v", b.title)
	}
}

func countAllBooks(db *sql.DB) (int, error) {
	var bookCount int
	err := db.QueryRow("SELECT COUNT(book_id) FROM books").Scan(&bookCount)
	if err != nil {
		return 0, err
	}
	return bookCount, nil
}

func countBooksByStatus(db *sql.DB, status string) (int, error) {
	var bookCount int
	err := db.QueryRow("SELECT COUNT(book_id) FROM books WHERE status = ?",
		status).Scan(&bookCount)
	if err != nil {
		return 0, err
	}
	return bookCount, nil
}

func getListOfBookIDs(db *sql.DB) ([]int, error) {
	var idList []int
	rows, err := db.Query("SELECT book_id FROM books ORDER BY book_id")
	if err != nil {
		return idList, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		idList = append(idList, id)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return idList, nil
}

func formatNameList(names []string) string {
	switch len(names) {
	case 0:
		return ""
	case 1:
		return names[0]
	case 2:
		return names[0] + " and " + names[1]
	default:
		var s string = ""
		for i := 0; i < len(names)-2; i++ {
			s += names[i]
			s += ", "
		}
		s += names[len(names)-2]
		s += " and "
		s += names[len(names)-1]
		return s
	}
}

func getAuthorsListById(db *sql.DB, id int) ([]string, error) {
	var authors []string
	sqlStmt := `
          SELECT people.name
          FROM people
          INNER JOIN book_author
            ON book_author.author_id = people.person_id
          WHERE book_author.book_id = ?`
	authorRows, err := db.Query(sqlStmt, id)
	// I think we need to handle no rows case as meaning no authors, not an error!
	if err != nil {
		if err == sql.ErrNoRows {
			return authors, fmt.Errorf("getAuthorsListById %d: No such book", id)
		}
		return authors, fmt.Errorf("getAuthorsListById %d: %v", id, err)
	}
	defer authorRows.Close()
	for authorRows.Next() {
		var authorName string
		err = authorRows.Scan(&authorName)
		if err != nil {
			return authors, fmt.Errorf("getAuthorsListById %d, %v", id, err)
		}
		authors = append(authors, authorName)
	}
	return authors, nil
}

func getEditorsListById(db *sql.DB, id int) ([]string, error) {
	var editors []string
	sqlStmt := `
          SELECT people.name
          FROM people
          INNER JOIN book_editor
            ON book_editor.editor_id = people.person_id
          WHERE book_editor.book_id = ?`
	editorRows, err := db.Query(sqlStmt, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return editors, fmt.Errorf("getEditorsListById %d: No such book", id)
		}
		return editors, fmt.Errorf("getEditorsListById %d: %v", id, err)
	}
	defer editorRows.Close()
	for editorRows.Next() {
		var editorName string
		err = editorRows.Scan(&editorName)
		// I think we need to handle no rows case as meaning no editors!!
		if err != nil {
			return editors, fmt.Errorf("getEditorsListById %d, %v", id, err)
		}
		editors = append(editors, editorName)
	}
	return editors, nil
}

func getBookById(db *sql.DB, id int) (Book, error) {
	var b Book
	b.id = id

	// [todo] need to handle purchased date
	var subtitle sql.NullString
	var seriesName sql.NullString
	var edition sql.NullInt64
	var purDate sql.NullString

	sqlStmt := `
            SELECT title, subtitle, year, edition, publishers.name, isbn,
            series.series_name, status, purchased_date
            FROM books
            INNER JOIN publishers
              ON books.publisher_id = publishers.publisher_id
            LEFT JOIN series
              ON books.series_id = series.series_id
            WHERE book_id = ?`
	row := db.QueryRow(sqlStmt, id)
	if err := row.Scan(&b.title, &subtitle, &b.year, &edition,
		&b.publisher, &b.isbn, &seriesName, &b.status, &purDate); err != nil {
		if err == sql.ErrNoRows {
			return b, fmt.Errorf("getBookById %d: No such book", id)
		}
		return b, fmt.Errorf("getBookById %d: %v", id, err)
	}

	if subtitle.Valid {
		b.subtitle = subtitle.String
	}
	if seriesName.Valid {
		b.series = seriesName.String
	}
	if edition.Valid {
		b.edition = int(edition.Int64)
	}
	if purDate.Valid {
		b.purchased.setDate(purDate.String)
	}

	var authorList []string
	authorList, err := getAuthorsListById(db, id)
	if err != nil {
		log.Fatal(err)
	}
	b.author = formatNameList(authorList)

	var editorList []string
	editorList, err = getEditorsListById(db, id)
	if err != nil {
		log.Fatal(err)
	}
	b.editor = formatNameList(editorList)

	return b, nil
}

func printBookList(db *sql.DB) ([]Book, error) {
	idList, err := getListOfBookIDs(db)
	if err != nil {
		return nil, err
	}

	var bookList []Book

	fmt.Println("Books in library are:")

	for _, id := range idList {
		var book Book
		book, err = getBookById(db, id)
		if err != nil {
			return nil, err
		}

		fmt.Println(book)
		bookList = append(bookList, book)

	}

	return bookList, nil
}

func main() {
	// set up database connection
	db, err := sql.Open("sqlite3", "../db/books.sqlite")
	if err != nil {
		log.Fatal(err)
	}
	pingErr := db.Ping()
	if pingErr != nil {
		log.Fatal(pingErr)
	}
	fmt.Println("Connected to db!")

	// Count how many books are in library (a single line query)
	var volumes int
	volumes, err = countAllBooks(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Found %v books in library:\n", volumes)

	// Count how many books are owned and how many wanted in library
	var owned, wanted int
	owned, err = countBooksByStatus(db, "Owned")
	if err != nil {
		log.Fatal(err)
	}
	wanted, err = countBooksByStatus(db, "Want")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%v are owned, %v are wanted. Combined makes %v.\n", owned,
		wanted, owned+wanted)

	fmt.Println("\nNow get books using functions:")

	_, err = printBookList(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("End of library\n")

	// Get the ID of a particular book, in this case, Kingdom though Covenant
	var id int
	err = db.QueryRow("SELECT book_id FROM books WHERE title IS 'Kingdom through Covenant'").Scan(&id)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("ID of Kingdom through Covenant: %v\n", id)

	// Get the author(s) of a book, method 1: use the book ID
	sqlStmt := `
          SELECT people.name
          FROM people
          INNER JOIN book_author
            ON book_author.author_id = people.person_id
          WHERE book_author.book_id = ?`
	authorRows, err := db.Query(sqlStmt, 5)
	if err != nil {
		log.Fatal(err)
	}
	defer authorRows.Close()
	fmt.Print("Authors of book ID #5 are: ")
	for authorRows.Next() {
		var authorName string
		err = authorRows.Scan(&authorName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v, ", authorName)
	}
	fmt.Print("\n")

	// Get the author(s) of a book, method 2: use book name
	sqlStmt = `
          SELECT people.name
          FROM people
          INNER JOIN book_author
            ON book_author.author_id = people.person_id
          INNER JOIN books
            ON book_author.book_id = books.book_id
          WHERE books.title = ?`
	authorRows, err = db.Query(sqlStmt, "Kingdom through Covenant")
	if err != nil {
		log.Fatal(err)
	}
	defer authorRows.Close()
	fmt.Print("Authors of Kingdom through Covenant are: ")
	for authorRows.Next() {
		var authorName string
		err = authorRows.Scan(&authorName)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%v, ", authorName)
	}
	fmt.Print("\n")

	fmt.Println("Testing use of PurchasedDate type")
	var myPD PurchasedDate
	fmt.Printf("Init value is %v\n", &myPD)
	myPD.setDate("2019")
	fmt.Printf("Value with year is %v\n", &myPD)
	myPD.setDate("May 2019")
	fmt.Printf("Value with year and month is %v\n", &myPD)
	myPD.setDate("11 May 2019")
	fmt.Printf("Value with year, month and day is %v\n", &myPD)
}
