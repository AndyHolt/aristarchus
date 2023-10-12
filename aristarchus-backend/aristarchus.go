package main

import (
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

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

type DBInterface interface {
	Exec(query string, args ...any) (sql.Result, error)
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
}

func countAllBooks(db DBInterface) (int, error) {
	var bookCount int
	err := db.QueryRow("SELECT COUNT(book_id) FROM books").Scan(&bookCount)
	if err != nil {
		return 0, err
	}
	return bookCount, nil
}

func countBooksByStatus(db DBInterface, status string) (int, error) {
	var bookCount int
	err := db.QueryRow("SELECT COUNT(book_id) FROM books WHERE status = ?",
		status).Scan(&bookCount)
	if err != nil {
		return 0, err
	}
	return bookCount, nil
}

func getListOfBookIDs(db DBInterface) ([]int, error) {
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

func nameListFromString(nameString string) []string {
	if len(nameString) == 0 {
		var retval []string
		return retval
	}

	splitAnd := strings.Split(nameString, " and ")

	if len(splitAnd) == 1 {
		return splitAnd
	}
	splitComma := strings.Split(splitAnd[0], ", ")

	var nameList []string
	for _, name := range splitComma {
		nameList = append(nameList, name)
	}
	nameList = append(nameList, splitAnd[1])

	return nameList
}

func getAuthorsListById(db DBInterface, id int) ([]string, error) {
	var authors []string
	sqlStmt := `
          SELECT people.name
          FROM people
          INNER JOIN book_author
            ON book_author.author_id = people.person_id
          WHERE book_author.book_id = ?`
	authorRows, err := db.Query(sqlStmt, id)
	// [review] I think we need to handle no rows case as meaning no authors, not an error!
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

func getEditorsListById(db DBInterface, id int) ([]string, error) {
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

func getBookById(db DBInterface, id int) (Book, error) {
	var b Book
	b.id = id

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

func printBookList(db DBInterface) ([]Book, error) {
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

func personId(db DBInterface, person string) (int, error) {
	var id int
	if err := db.QueryRow("SELECT person_id FROM people WHERE name = ?",
		person).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			result, err := db.Exec("INSERT INTO people (name) VALUES (?)", person)
			if err != nil {
				return 0, fmt.Errorf("personId, %v", err)
			}
			liid, err := result.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("personId, %v", err)
			}
			id = int(liid)
		} else {
			return 0, fmt.Errorf("personId, %v", err)
		}
	}
	return id, nil
}

func publisherId(db DBInterface, publisher string) (int, error) {
	var id int
	if err := db.QueryRow("SELECT publisher_id FROM publishers WHERE name = ?",
		publisher).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			result, err := db.Exec("INSERT INTO publishers (name) VALUES (?)",
				publisher)
			if err != nil {
				return 0, fmt.Errorf("publisherId, %v", err)
			}
			liid, err := result.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("publisherId, %v", err)
			}
			id = int(liid)
		} else {
			return 0, fmt.Errorf("publisherId, %v", err)
		}
	}
	return id, nil
}

func seriesId(db DBInterface, series string) (int, error) {
	var id int
	if err := db.QueryRow("SELECT series_id FROM series WHERE series_name = ?",
		series).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			result, err := db.Exec("INSERT INTO series (series_name) VALUES (?)",
				series)
			if err != nil {
				return 0, fmt.Errorf("seriesId, %v", err)
			}
			liid, err := result.LastInsertId()
			if err != nil {
				return 0, fmt.Errorf("seriesId, %v", err)
			}
			id = int(liid)
		} else {
			return 0, fmt.Errorf("seriesId, %v", err)
		}
	}
	return id, nil
}

type AddingDuplicateBookError struct {
	book *Book
	id   int
}

func (e *AddingDuplicateBookError) Error() string {
	return fmt.Sprintf("Book \"%v\" already in database, id #%v",
		e.book.title,
		e.id)
}

func checkBookInDb(db DBInterface, b *Book) (int, error) {
	// [todo] update checkBookInDb to use isbn, if available.

	var id int
	var authorList, editorList []string
	var authorForCheck, editorForCheck string

	authorList = nameListFromString(b.author)
	if len(authorList) != 0 {
		authorForCheck = authorList[0]
	}

	editorList = nameListFromString(b.editor)
	if len(editorList) != 0 {
		editorForCheck = editorList[0]
	}

	sqlStmt := `
        SELECT books.book_id
        FROM books
        INNER JOIN book_author
          ON books.book_id = book_author.book_id
        INNER JOIN people
          ON book_author.author_id = people.person_id
        WHERE people.name = ?
          AND books.title = ?
        UNION
        SELECT books.book_id
        FROM books
        INNER JOIN book_editor
          ON books.book_id = book_editor.book_id
        INNER JOIN people
          ON book_editor.editor_id = people.person_id
        WHERE people.name = ?
          AND books.title = ?
`

	if scanErr := db.QueryRow(sqlStmt,
		authorForCheck,
		b.title,
		editorForCheck,
		b.title).Scan(&id); scanErr != nil {
		if scanErr == sql.ErrNoRows {
			return 0, nil
		} else {
			return 0, fmt.Errorf("checkBookInDb, SQL scan error, %v", scanErr)
		}
	} else {
		return id, &AddingDuplicateBookError{b, id}
	}
}

func addBook(db *sql.DB, b *Book) (int, error) {
	// check if book is already in database
	id, bookInDbErr := checkBookInDb(db, b)
	if bookInDbErr != nil {
		if _, ok := bookInDbErr.(*AddingDuplicateBookError); ok {
			return id, bookInDbErr
		} else {
			return id, fmt.Errorf("addBook, %v", bookInDbErr)
		}
	}

	// handle people
	var authorList, editorList []string
	authorList = nameListFromString(b.author)
	editorList = nameListFromString(b.editor)

	// Create lists of author ids from the author lists
	var authorIdList, editorIdList []int
	for _, authorName := range authorList {
		authorId, err := personId(db, authorName)
		if err != nil {
			return 0, fmt.Errorf("addBook, %v", err)
		}
		authorIdList = append(authorIdList, authorId)
	}
	for _, editorName := range editorList {
		editorId, err := personId(db, editorName)
		if err != nil {
			return 0, fmt.Errorf("addBook, %v", err)
		}
		editorIdList = append(editorIdList, editorId)
	}

	// handle publisher
	pubId, err := publisherId(db, b.publisher)
	if err != nil {
		return 0, fmt.Errorf("addBook, issue with publisher, %v", err)
	}

	// handle series
	var serId int
	if len(b.series) != 0 {
		serId, err = seriesId(db, b.series)
		if err != nil {
			return 0, fmt.Errorf("addBook, issue with series, %v", err)
		}
	} else {
		serId = -1
	}

	// insert book -- at this point, use a transaction to ensure author/editor
	// info is included for every book in the database.
	tx, err := db.Begin()
	if err != nil {
		return 0, fmt.Errorf("addBook, Couldn't start sql transaction: %v", err)
	}
	defer tx.Rollback()

	var bookId int

	if serId != -1 {
		result, err := tx.Exec(`INSERT INTO books (title, subtitle, year, edition,
                            publisher_id, isbn, series_id, status,
                            purchased_date) VALUES (?, ?, ?, ?, ?, ?, ?,
                            ?, ?)`,
			b.title, b.subtitle, b.year, b.edition, pubId, b.isbn,
			serId, b.status, b.purchased.String())

		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
		liid, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
		bookId = int(liid)
	} else {
		result, err := tx.Exec(`INSERT INTO books (title, subtitle, year, edition,
                            publisher_id, isbn, status, purchased_date) VALUES
			                (?, ?, ?, ?, ?, ?, ?, ?)`,
			b.title, b.subtitle, b.year, b.edition, pubId, b.isbn,
			b.status, b.purchased.String())

		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
		liid, err := result.LastInsertId()
		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
		bookId = int(liid)
	}

	// handle book_author
	for _, authId := range authorIdList {
		_, err = tx.Exec("INSERT INTO book_author VALUES (?, ?)", bookId,
			authId)
		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
	}

	// handle book_editor
	for _, edId := range editorIdList {
		_, err = tx.Exec("INSERT INTO book_editor VALUES (?, ?)", bookId, edId)
		if err != nil {
			return 0, fmt.Errorf("addBook: %v", err)
		}
	}

	err = tx.Commit()
	if err != nil {
		return 0, fmt.Errorf("addBook, issue adding book: %v", err)
	}

	return bookId, nil
}

func updateBookAuthor(db *sql.DB, id int, authorString string) (string, error) {
	newAuthorsList := nameListFromString(authorString)
	oldAuthorsList, err := getAuthorsListById(db, id)
	if err != nil {
		return "", err
	}

	var authorsToAdd, authorsToDelete []string

	for _, author := range newAuthorsList {
		if !slices.Contains(oldAuthorsList, author) {
			authorsToAdd = append(authorsToAdd, author)
		}
	}

	for _, author := range oldAuthorsList {
		if !slices.Contains(newAuthorsList, author) {
			authorsToDelete = append(authorsToDelete, author)
		}
	}

	// start a transaction to make the edit of authors atomic
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("updateBookAuthor, Couldn't start sql transaction: %v", err)
	}
	defer tx.Rollback()

	for _, author := range authorsToAdd {
		personId, personIdErr := personId(tx, author)
		if personIdErr != nil {
			return "", fmt.Errorf("updateBookAuthor: %v", personIdErr)
		}

		_, err := tx.Exec("INSERT INTO book_author (book_id, author_id) VALUES (?, ?)",
			id, personId)
		if err != nil {
			return "", fmt.Errorf("updateBookAuthor: %v", err)
		}
	}

	for _, author := range authorsToDelete {
		personId, personIdErr := personId(tx, author)
		if personIdErr != nil {
			return "", fmt.Errorf("updateBookAuthor:, %v", personIdErr)
		}

		_, err := tx.Exec("DELETE FROM book_author WHERE book_id = ? AND author_id = ?",
			id, personId)
		if err != nil {
			return "", fmt.Errorf("updateBookAuthor: %v, err")
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("updateBookAuthor, issue updating authors: %v", err)
	}

	updatedAuthorList, err := getAuthorsListById(db, id)
	if err != nil {
		return "", fmt.Errorf("updateBookAuthor, Couldn't fetch updated authors: %v", err)
	}
	updatedAuthor := formatNameList(updatedAuthorList)

	return updatedAuthor, nil
}

func updateBookEditor(db *sql.DB, id int, editorString string) (string, error) {
	newEditorsList := nameListFromString(editorString)
	oldEditorsList, err := getEditorsListById(db, id)
	if err != nil {
		return "", err
	}

	var editorsToAdd, editorsToDelete []string

	for _, editor := range newEditorsList {
		if !slices.Contains(oldEditorsList, editor) {
			editorsToAdd = append(editorsToAdd, editor)
		}
	}

	for _, editor := range oldEditorsList {
		if !slices.Contains(newEditorsList, editor) {
			editorsToDelete = append(editorsToDelete, editor)
		}
	}

	// start a transaction to make the edit of authors atomic
	tx, err := db.Begin()
	if err != nil {
		return "", fmt.Errorf("updateBookEditor, Couldn't start sql transaction: %v", err)
	}
	defer tx.Rollback()

	for _, editor := range editorsToAdd {
		personId, personIdErr := personId(tx, editor)
		if personIdErr != nil {
			return "", fmt.Errorf("updateBookEditor: %v", personIdErr)
		}

		_, err := tx.Exec("INSERT INTO book_editor (book_id, editor_id) VALUES (?, ?)",
			id, personId)
		if err != nil {
			return "", fmt.Errorf("updateBookEditor: %v", err)
		}
	}

	for _, editor := range editorsToDelete {
		personId, personIdErr := personId(tx, editor)
		if personIdErr != nil {
			return "", fmt.Errorf("updateBookEditor:, %v", personIdErr)
		}

		_, err := tx.Exec("DELETE FROM book_editor WHERE book_id = ? AND editor_id = ?",
			id, personId)
		if err != nil {
			return "", fmt.Errorf("updateBookEditor: %v, err")
		}
	}

	err = tx.Commit()
	if err != nil {
		return "", fmt.Errorf("updateBookEditor, issue updating editors: %v", err)
	}

	updatedEditorList, err := getEditorsListById(db, id)
	if err != nil {
		return "", fmt.Errorf("updateBookEditor, Couldn't fetch updated editors: %v", err)
	}
	updatedEditor := formatNameList(updatedEditorList)

	return updatedEditor, nil
}

func updatePersonName(db DBInterface, id int, newName string) (string, error) {
	sqlStmt := `
      UPDATE people
      SET name = ?
      WHERE person_id = ?
      `

	_, err := db.Exec(sqlStmt, newName, id)
	if err != nil {
		return "", fmt.Errorf("updatePersonName, Couldn't update person #%v to %v: %v",
			id, newName, err)
	}

	var updatedName string
	if err := db.QueryRow("SELECT name FROM people WHERE person_id = ?",
		id).Scan(&updatedName); err != nil {
		return "", fmt.Errorf("updatePersonName, Couldn't get updated name: %v", err)
	}
	if updatedName != newName {
		return "", fmt.Errorf("updatePersonName, Updated name \"%v\" is not desired new name \"%v\".")
	}

	return updatedName, nil
}

func updateBookTitle(db DBInterface, id int, title string) (string, error) {
	sqlStmt := `
      UPDATE books
      SET title = ?
      WHERE book_id = ?
      `

	_, err := db.Exec(sqlStmt, title, id)
	if err != nil {
		return "", fmt.Errorf("updateBookTitle, Couldn't update book #%v title to %v: %v",
			id, title, err)
	}

	var updatedTitle string
	if err := db.QueryRow("SELECT title FROM books WHERE book_id = ?",
		id).Scan(&updatedTitle); err != nil {
		return "", fmt.Errorf("updateBookTitle, Couldn't get updated title: %v", err)
	}
	if updatedTitle != title {
		return "", fmt.Errorf("updateBookTitle: Updated title \"%v\" does not match requested title \"%v\"")
	}

	return updatedTitle, nil
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
	defer db.Close()

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

	// 	type Book struct {
	// 	id        int
	// 	author    string
	// 	editor    string
	// 	title     string
	// 	subtitle  string
	// 	year      int
	// 	edition   int
	// 	publisher string
	// 	isbn      string
	// 	series    string
	// 	status    string
	// 	purchased PurchasedDate
	// }
	var itts Book
	itts.author = "Karen H. Jobes and Moisés Silva"
	itts.title = "Invitation to the Septuagint"
	itts.year = 2015
	itts.edition = 2
	itts.publisher = "Baker Academic"
	itts.isbn = "978-0-8010-3649-1"
	itts.status = "Owned"

	var ittspd PurchasedDate
	ittspd.setDate("December 2021")
	itts.purchased = ittspd

	id, err := addBook(db, &itts)
	if err != nil {
		if _, ok := err.(*AddingDuplicateBookError); ok {
			fmt.Println(err)
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("Added new book with id %v\n", id)
	}
	volumes, err = countAllBooks(db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Now there are %v books in library:\n", volumes)

	// var gpe Book
	// gpe.editor = "Chad Meister and James K. Dew Jr."
	// gpe.title = "God and the Problem of Evil"
	// gpe.subtitle = "Five Views"
	// gpe.year = 2017
	// gpe.publisher = "IVP Academic"
	// gpe.isbn = "978-0-8308-4024-3"
	// gpe.series = "Spectrum Multiview Books"
	// gpe.status = "Owned"

	// var gpepd PurchasedDate
	// gpepd.setDate("March 2023")
	// gpe.purchased = gpepd

	// id, err = addBook(db, &gpe)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("Added new book with id %v\n", id)

	// 	type Book struct {
	// 	id        int
	// 	author    string
	// 	editor    string
	// 	title     string
	// 	subtitle  string
	// 	year      int
	// 	edition   int
	// 	publisher string
	// 	isbn      string
	// 	series    string
	// 	status    string
	// 	purchased PurchasedDate
	// }

	// var tag Book
	// tag.editor = "Simon Gathercole"
	// tag.title = "The Apocryphal Gospels"
	// tag.year = 2021
	// tag.publisher = "Penguin"
	// tag.isbn = "978-0-241-34055-4"
	// tag.series = "Penguin Classics"
	// tag.status = "Owned"

	// var tagpd PurchasedDate
	// tagpd.setDate("March 2023")
	// tag.purchased = tagpd

	// id, err = addBook(db, &tag)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// fmt.Printf("Added new book with id %v\n", id)

	// id, err := personId(db, "Peter J. Gentry")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("ID of Peter J. Gentry is %v\n", id)

	// id, err = personId(db, "Karen H. Jobes")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("ID of Karen H. Jobes is %v\n", id)

	// [todo] Add functions to edit each attribute of the book
	//   [done] Modify author(s) of a book function
	fmt.Printf("\n*** Testing modification of book author ***\n")
	newAuthors, err := updateBookAuthor(db, 7, "P. G. Wodehouse, J. K. Rowling and Timothy Keller")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Authors of \"Invitation to the Septuagint\" are now %v\n",
		newAuthors)

	newAuthors, err = updateBookAuthor(db, 7, "Karen H. Jobes and Moisés Silva")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Authors of \"Invitation to the Septuagint\" are now %v\n",
		newAuthors)

	//   [done] Modify editor(s) of a book function
	fmt.Printf("\n*** Testing modification of book editor ***\n")
	newEditors, err := updateBookEditor(db, 7, "Anselm and P. G. Wodehouse")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Editors of \"Invitation to the Septuagint\" are now %v\n",
		newEditors)

	newEditors, err = updateBookEditor(db, 7, "")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Editors of \"Invitation to the Septuagint\" are now %v\n",
		newEditors)

	//   [done] Modify a person's name function
	fmt.Printf("\n*** Testing modification of person's name ***\n")
	booksByAuthorIdSql := `
      SELECT author_id, name, title
      FROM books
      INNER JOIN book_author
        ON books.book_id = book_author.book_id
      INNER JOIN people
        ON book_author.author_id = people.person_id
      WHERE author_id = ?
      `

	rows, err := db.Query(booksByAuthorIdSql, 3)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Printf("Books by author with person_id #3:\n")
	for rows.Next() {
		var authorId int
		var authorName string
		var bookTitle string
		if scanErr := rows.Scan(&authorId, &authorName, &bookTitle); scanErr != nil {
			log.Fatal(err)
		}
		fmt.Printf("person_id: %v, name: %v, book title: %v\n", authorId,
			authorName, bookTitle)
	}

	updatePersonName(db, 3, "Geoffrey Parker Jr.")

	rows, err = db.Query(booksByAuthorIdSql, 3)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Printf("Books by author with person_id #3:\n")
	for rows.Next() {
		var authorId int
		var authorName string
		var bookTitle string
		if scanErr := rows.Scan(&authorId, &authorName, &bookTitle); scanErr != nil {
			log.Fatal(err)
		}
		fmt.Printf("person_id: %v, name: %v, book title: %v\n", authorId,
			authorName, bookTitle)
	}

	updatePersonName(db, 3, "Peter J. Gentry")

	rows, err = db.Query(booksByAuthorIdSql, 3)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Printf("Books by author with person_id #3:\n")
	for rows.Next() {
		var authorId int
		var authorName string
		var bookTitle string
		if scanErr := rows.Scan(&authorId, &authorName, &bookTitle); scanErr != nil {
			log.Fatal(err)
		}
		fmt.Printf("person_id: %v, name: %v, book title: %v\n", authorId,
			authorName, bookTitle)
	}

	//   [done] Modify title function
	fmt.Printf("\n*** Testing modification of book title ***\n")

	var aBook Book
	aBook, err = getBookById(db, 1)
	fmt.Printf("Prior to modification, book #1 is: %v\n", aBook)

	newTitle, err := updateBookTitle(db, 1, "The Art of Old Testament Studies")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Book title changed to \"%v\"\n", newTitle)
	aBook, err = getBookById(db, 1)
	fmt.Printf("After modification, book #1 is: %v\n", aBook)

	newTitle, err = updateBookTitle(db, 1, "Introduction to the Old Testament")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Book title changed back to \"%v\"\n", newTitle)
	aBook, err = getBookById(db, 1)
	fmt.Printf("After re-modification, book #1 is: %v\n", aBook)

	//   [todo] Modify subtitle function
	//   [todo] Modify year function
	//   [todo] Modify edition function
	//   [todo] Modify publisher function
	//   [todo] Modify publisher name function
	//   [todo] Modify isbn function
	//   [todo] Modify series function
	//   [todo] Modify series name function
	//   [todo] Modify status function
	//   [todo] Modify purchased function

	// [todo] Add functions to delete from database
	//   [todo] Delete book by ID function
	//   [todo] Delete person by ID function
	//   [todo] Delete publisher by ID function
	//   [todo] Delete series by ID function
}
