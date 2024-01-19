package main

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func TestMain(m *testing.M) {
	code, err := setupRunTeardown(m)
	if err != nil {
		fmt.Println(err)
	}
	os.Exit(code)
}

func setupRunTeardown(m *testing.M) (code int, err error) {
	err = setupTestDatabase()
	if err != nil {
		return 1, err
	}

	defer func() {
		if tempErr := teardownTestDatabase(); tempErr != nil {
			err = tempErr
		}
	}()

	return m.Run(), err
}

func setupTestDatabase() (err error) {
	cmd := exec.Command("sqlite3", "testdb.sqlite", "-init",
		"../db/init_test_database.sql", ".quit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("setupTestDatabase, couldn't set up db: %v", err)
	}
	return nil
}

func teardownTestDatabase() (err error) {
	cmd := exec.Command("sqlite3", "testdb.sqlite", "-init",
		"../db/teardown_test_database.sql", ".quit")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("teardownTestDatabase, issue clearing db: %v", err)
	}

	cmd = exec.Command("rm", "testdb.sqlite")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("teardownTestDatabase, issue removing db file: %v",
			err)
	}

	return nil
}

func makeTestBook() *Book {
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

	return &itts
}

func TestPingDatabase(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pingErr := db.Ping()
	if pingErr != nil {
		t.Errorf("Couldn't ping database: %v", err)
	}
}

func TestDatabaseQuery(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	stmt := "SELECT name FROM people WHERE person_id=?"
	var name string

	if err := db.QueryRow(stmt, 1).Scan(&name); err != nil {
		t.Errorf("Problem querying database: %v", err)
	}

	expected := "R. K. Harrison"
	if name != expected {
		t.Errorf("Wrong name returned, expected \"%v\", got \"%v\"", expected, name)
	}
}

func TestPurchasedDateEmptyString(t *testing.T) {
	dateString := ""
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err == nil {
		t.Error("Empty string did not return error")
	} else {
		var dpe *DateParsingError
		if !errors.As(err, &dpe) {
			t.Errorf("Unexpected error for empty string: %v", err)
		}
	}
}

func TestPurchasedDateYear(t *testing.T) {
	dateString := "2019"
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err != nil {
		t.Errorf("Problem setting PurchasedDate with year only: %v", err)
	}

	returnedString := pd.String()
	if returnedString != dateString {
		t.Errorf("Date not equal after round trip: before was %v, after is %v",
			dateString, returnedString)
	}
}

func TestPurchasedDateMonthYear(t *testing.T) {
	dateString := "May 2019"
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err != nil {
		t.Errorf("Problem setting PurchasedDate with month and year: %v", err)
	}

	returnedString := pd.String()
	if returnedString != dateString {
		t.Errorf("Date not equal after round trip: before was %v, after is %v",
			dateString, returnedString)
	}
}

func TestPurchasedDateDayMonthYear(t *testing.T) {
	dateString := "11 May 2019"
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err != nil {
		t.Errorf("Problem setting PurchasedDate with day, month and year: %v", err)
	}

	returnedString := pd.String()
	if returnedString != dateString {
		t.Errorf("Date not equal after round trip: before was %v, after is %v",
			dateString, returnedString)
	}
}

func TestPurchasedDateInvalidDate(t *testing.T) {
	dateString := "42 May 2019"
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err == nil {
		t.Errorf("No error returned with invalid date %v", dateString)
	} else {
		var dpe *DateParsingError
		if !errors.As(err, &dpe) {
			t.Errorf("Wrong error type raised with invalid date %v: %v",
				dateString, err)
		}
	}
}

func TestPurchasedDateInvalidFormat(t *testing.T) {
	dateString := "11/5/2019"
	var pd PurchasedDate

	err := pd.setDate(dateString)
	if err == nil {
		t.Errorf("No error returned with invalid date format %v", dateString)
	} else {
		var dpe *DateParsingError
		if !errors.As(err, &dpe) {
			t.Errorf("Wrong error type raised with invalid date format %v: %v",
				dateString, err)
		}
	}
}

func TestBookStringMethod(t *testing.T) {
	b := *makeTestBook()

	expected := "Karen H. Jobes and Moisés Silva, Invitation to the Septuagint (2015) [Owned]"

	bkStr := b.String()

	if bkStr != expected {
		t.Errorf("Wrong value returned by String method on Book: expected %v, got %v", expected, bkStr)
	}
}

func TestBookAuthorEditor(t *testing.T) {
	b := *makeTestBook()
	expected := "Karen H. Jobes and Moisés Silva"

	bkAuEd := b.authorEditor()

	if bkAuEd != expected {
		t.Errorf("Wrong value returned by authorEditor method on Book: expected %v, got %v",
			expected, bkAuEd)
	}
}

func TestBookFullTitle(t *testing.T) {
	b := *makeTestBook()
	expected := "Invitation to the Septuagint"

	bkFullTi := b.fullTitle()

	if bkFullTi != expected {
		t.Errorf("Wrong value returned by fullTitle method on Book: expected %v, got %v",
			expected, bkFullTi)
	}
}

// [todo]  test getListOfBookIds function

// [todo]  test FormatNameList function
// [todo]  test NameListFromString function
// [todo]  test GetAuthorsListById function
// [todo]  test GetEditorsListById function
// [todo]  test GetBookById function
// [todo]  test PersonId function
// [todo]  test PublisherId function
// [todo]  test SeriesId function
// [todo]  test CheckBookInDb function

func TestCountAllBooks(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 6

	var volumes int
	volumes, err = countAllBooks(db)
	if err != nil {
		t.Errorf("Could not count books: %v", err)
	}

	if volumes != expected {
		t.Errorf("Wrong number of books: expected %v, got %v", expected, volumes)
	}
}

func TestCountOwnedBooks(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 5

	var owned int
	owned, err = countBooksByStatus(db, "Owned")
	if err != nil {
		t.Errorf("Could not count books: %v", err)
	}

	if owned != expected {
		t.Errorf("Wrong number of owned books: expected %v, got %v", expected, owned)
	}
}

func TestCountWantedBooks(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 1

	var wanted int
	wanted, err = countBooksByStatus(db, "Want")
	if err != nil {
		t.Errorf("Could not count books: %v", err)
	}

	if wanted != expected {
		t.Errorf("Wrong number of owned books: expected %v, got %v", expected, wanted)
	}
}

func TestAddBook(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

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
		t.Errorf("Problem adding new book: %v", err)
	}

	var volumes int
	volumes, err = countAllBooks(db)
	if err != nil {
		t.Errorf("Problem counting books after addition: %v", err)
	}
	expected := 7
	if volumes != expected {
		t.Errorf("Wrong number of books after addition, expected %v, got %v",
			expected, volumes)
	}

	err = deleteBook(db, id)
	if err != nil {
		t.Errorf("Problem deleting added book to reset database: %v", err)
	}
}

func TestAddDuplicateBook(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var iot Book
	iot.author = "R. K. Harrison"
	iot.title = "Introduction to the Old Testament"
	iot.year = 1969
	iot.publisher = "IVP"
	iot.isbn = "0-85111-723-6"
	iot.status = "Owned"

	var iotpd PurchasedDate
	iotpd.setDate("May 2023")
	iot.purchased = iotpd

	_, err = addBook(db, &iot)
	if err == nil {
		t.Error("Adding duplicate book did not result in error")
	} else {
		if _, ok := err.(*AddingDuplicateBookError); !ok {
			t.Errorf("Adding duplicate book wrong error: %v", err)
		}
	}
}

func TestUpdateBookAuthor(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newAuthors string
	newAuthors = "P. G. Wodehouse, J. K. Rowling and Timothy Keller"
	updatedAuthors, err := updateBookAuthor(db, 1, newAuthors)
	if err != nil {
		t.Errorf("Problem updating book author: %v", err)
	}

	if updatedAuthors != newAuthors {
		t.Errorf("Author(s) not properly updated. Updated author(s) should be %v, but is %v",
			newAuthors, updatedAuthors)
	}

	newAuthors = "R. K. Harrison"
	updatedAuthors, err = updateBookAuthor(db, 1, newAuthors)
	if err != nil {
		t.Errorf("Problem reverting updated book author: %v", err)
	}

	if updatedAuthors != newAuthors {
		t.Errorf("Author(s) not properly reverted. Reset author(s) should be %v, but is %v",
			newAuthors, updatedAuthors)
	}
}

func TestUpdateBookEditor(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newEditors string
	newEditors = "James H. Charlesworth, Heinrich von Siebenthal and Francis Brown"
	updatedEditors, err := updateBookEditor(db, 6, newEditors)
	if err != nil {
		t.Errorf("Problem updating book author: %v", err)
	}

	if updatedEditors != newEditors {
		t.Errorf("Editors not properly updated. Updated editors should be %v, but is %v",
			newEditors, updatedEditors)
	}

	newEditors = "N. Gray Sutanto, James Eglinton and Cory C. Brock"
	updatedEditors, err = updateBookEditor(db, 6, newEditors)
	if err != nil {
		t.Errorf("Problem reverting updated book editors: %v", err)
	}

	if updatedEditors != newEditors {
		t.Errorf("Editors not properly reverted. Reset editors should be %v, but were %v",
			newEditors, updatedEditors)
	}
}

// [todo] test modification of person's name
func TestUpdatePersonName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newName string
	newName = "Geoffrey Parker Jr"
	updatedName, err := updatePersonName(db, 3, newName)
	if err != nil {
		t.Errorf("Problem updating person's name: %v", err)
	}
	if updatedName != newName {
		t.Errorf("Name not updated properly, \"%v\" is not \"%v\"",
			updatedName, newName)
	}

	bookId := 4
	queriedName, err := getAuthorsListById(db, bookId)
	if err != nil {
		t.Errorf("Problem getting book #%v's author: %v", bookId, err)
	}

	if len(queriedName) != 1 {
		t.Errorf("Expected a single name, but got %v: %v", len(queriedName), queriedName)
	}
	if queriedName[0] != newName {
		t.Errorf("Person's name not properly updated, \"%v\" is not \"%v\"",
			queriedName[0], newName)
	}

	newName = "Peter J. Gentry"
	updatedName, err = updatePersonName(db, 3, newName)
	if err != nil {
		t.Errorf("Problem reverting person's name: %v", err)
	}
	if updatedName != newName {
		t.Errorf("Name not updated properly, \"%v\" is not \"%v\"",
			updatedName, newName)
	}

	queriedName, err = getAuthorsListById(db, bookId)
	if err != nil {
		t.Errorf("Problem getting book #%v's author: %v", bookId, err)
	}
	if len(queriedName) != 1 {
		t.Errorf("Expected a single name, but got %v: %v", len(queriedName), queriedName)
	}
	if queriedName[0] != newName {
		t.Errorf("Person's name not properly reverted, \"%v\" is not \"%v\"",
			queriedName[0], newName)
	}
}

func TestUpdateBookTitle(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newTitle string = "The Art of Old Testament Studies"
	updatedTitle, err := updateBookTitle(db, 1, newTitle)
	if err != nil {
		t.Errorf("Problem updating book title: %v", err)
	}
	if updatedTitle != newTitle {
		t.Errorf("Title not correctly modified. Should be \"%v\", instead is \"%v\"",
			newTitle, updatedTitle)
	}

	// Reset to proper value for other tests to use an unmodified database
	newTitle = "Introduction to the Old Testament"
	updatedTitle, err = updateBookTitle(db, 1, newTitle)
	if err != nil {
		t.Errorf("Problem reverting book title: %v", err)
	}
	if updatedTitle != newTitle {
		t.Errorf("Title not correctly reverted. Should be \"%v\", instead is \"%v\"",
			newTitle, updatedTitle)
	}
}

func TestUpdateBookTitleEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var emptyTitle string = ""
	updatedTitle, err := updateBookTitle(db, 1, emptyTitle)
	var ete *EmptyTitleError
	if errors.Is(err, ete) {
		t.Errorf("Updating title with empty string returned unexpected error: %v", err)
	}
	if updatedTitle == "" {
		t.Errorf("Title illegally updated to empty string")
	}

	// also check that the book has not been modified
	b, err := getBookById(db, 1)
	if b.title != "Introduction to the Old Testament" {
		t.Errorf("Book title has been wrongly modified to \"%v\"", b.title)
	}
}

func TestUpdateBookSubtitle(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newSubtitle string = "Four views, at least three of them wrong"

	updatedSubtitle, err := updateBookSubtitle(db, 2, newSubtitle)
	if err != nil {
		t.Errorf("Problem updating subtitle: %v", err)
	}
	if updatedSubtitle != newSubtitle {
		t.Errorf("Wrongly updated subtitle: should be \"%v\" but got \"%v\"",
			newSubtitle, updatedSubtitle)
	}

	b, err := getBookById(db, 2)
	if b.subtitle != newSubtitle {
		t.Errorf("Wrongly updated subtitle from book: should be \"%v\" but got \"%v\"",
			newSubtitle, updatedSubtitle)
	}

	// Revert database back to original state
	newSubtitle = "Four Views of God's Emotions and Suffering"
	updatedSubtitle, err = updateBookSubtitle(db, 2, newSubtitle)
	if err != nil {
		t.Errorf("Problem reverting subtitle: %v", err)
	}
	if updatedSubtitle != newSubtitle {
		t.Errorf("Wrongly reverted subtitle: should be \"%v\" but got \"%v\"",
			newSubtitle, updatedSubtitle)
	}

	b, err = getBookById(db, 2)
	if b.subtitle != newSubtitle {
		t.Errorf("Wrongly reverted subtitle from book: should be \"%v\" but got \"%v\"",
			newSubtitle, updatedSubtitle)
	}
}

// Empty subtitle should set null value in database, not an empty string
func TestUpdateBookSubtitleEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newSubtitle string = ""
	updatedSubtitle, err := updateBookSubtitle(db, 2, newSubtitle)
	if err != nil {
		t.Errorf("Problem updating subtitle: %v", err)
	}
	if updatedSubtitle != newSubtitle {
		t.Errorf("Wrongly updated subtitle: should be \"%v\", but got \"%v\"",
			newSubtitle, updatedSubtitle)
	}

	// check for non-null subtitles: error if any found
	sqlStmt := `
      SELECT subtitle
      FROM books
      WHERE book_id = ? AND subtitle IS NOT NULL
    `
	var readSubtitle string
	rows, err := db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readSubtitle); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		t.Errorf("Query returned non-null value \"%v\"", readSubtitle)
	} else {
		if err := rows.Err(); err != nil {
			t.Errorf("rows.Next() failed with non-nil error: %v", err)
		}
	}

	// check for null subtitle: error if none found
	sqlStmt = `
      SELECT subtitle
      FROM books
      WHERE book_id = ? AND subtitle IS NULL
    `
	var readNullSubtitle sql.NullString
	rows, err = db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("Querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readNullSubtitle); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		if readNullSubtitle.Valid {
			t.Errorf("Query returned valid subtitle: \"%v\"", readNullSubtitle.String)
		}
	} else {
		t.Errorf("rows.Next() failed with err: %v", rows.Err())
	}
	// Now we need to explicitly close rows to unlock the database for reversion
	// to original values. We can't wait for the deferred function to take
	// effect.
	rows.Close()

	// Revert database to original state
	var origSubtitle string = "Four Views of God's Emotions and Suffering"
	revertedSubtitle, err := updateBookSubtitle(db, 2, origSubtitle)
	if err != nil {
		t.Errorf("Problem reverting subtitle: %v", err)
	}
	if revertedSubtitle != origSubtitle {
		t.Errorf("Wrongly reverted subtitle: should be \"%v\", but got \"%v\"",
			origSubtitle, revertedSubtitle)
	}
}

func TestUpdateBookYear(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newYear int = 2024
	updatedYear, err := updateBookYear(db, 1, newYear)
	if err != nil {
		t.Errorf("Problem updating book year: %v", err)
	}
	if updatedYear != newYear {
		t.Errorf("Wrongly updated year: Should be %v but got %v", newYear, updatedYear)
	}

	b, err := getBookById(db, 1)
	if b.year != newYear {
		t.Errorf("Book year is wrong in database: should be %v, but is %v",
			newYear, b.year)
	}

	// Revert to restore database state
	var origYear int = 1969
	revertedYear, err := updateBookYear(db, 1, origYear)
	if err != nil {
		t.Errorf("Problem reverting book year: %v", err)
	}
	if revertedYear != origYear {
		t.Errorf("Wrongly reverted year: should be %v but got %v", origYear, revertedYear)
	}

	b, err = getBookById(db, 1)
	if b.year != origYear {
		t.Errorf("Reverted book year is wrong in database: should be %v, but is %v",
			origYear, b.year)
	}
}

func TestUpdateBookEdition(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newEdition int = 5

	updatedEdition, err := updateBookEdition(db, 5, newEdition)
	if err != nil {
		t.Errorf("Problem updating edition: %v", err)
	}
	if updatedEdition != newEdition {
		t.Errorf("Wrongly updated edition: should be \"%v\" but got \"%v\"",
			newEdition, updatedEdition)
	}

	b, err := getBookById(db, 5)
	if b.edition != newEdition {
		t.Errorf("Wrongly updated edition from book: should be \"%v\" but got \"%v\"",
			newEdition, b.edition)
	}

	// Revert database back to original state
	origEdition := 2
	revertedEdition, err := updateBookEdition(db, 5, origEdition)
	if err != nil {
		t.Errorf("Problem reverting edition: %v", err)
	}
	if revertedEdition != origEdition {
		t.Errorf("Wrongly reverted edition: should be \"%v\" but got \"%v\"",
			origEdition, revertedEdition)
	}

	b, err = getBookById(db, 5)
	if b.edition != origEdition {
		t.Errorf("Wrongly reverted edition from book: should be \"%v\" but got \"%v\"",
			origEdition, revertedEdition)
	}
}

// Empty subtitle should set null value in database, not an empty string
func TestUpdateBookEditionZero(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newEdition int = 0
	updatedEdition, err := updateBookEdition(db, 5, newEdition)
	if err != nil {
		t.Errorf("Problem updating edition: %v", err)
	}
	if updatedEdition != newEdition {
		t.Errorf("Wrongly updated edition: should be \"%v\", but got \"%v\"",
			newEdition, updatedEdition)
	}

	// check for non-null edition: error if any found
	sqlStmt := `
      SELECT edition
      FROM books
      WHERE book_id = ? AND edition IS NOT NULL
    `
	var readEdition int
	rows, err := db.Query(sqlStmt, 5)
	if err != nil {
		t.Errorf("querying non-null edition in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readEdition); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		t.Errorf("Query returned non-null value \"%v\"", readEdition)
	} else {
		if err := rows.Err(); err != nil {
			t.Errorf("rows.Next() failed with non-nil error: %v", err)
		}
	}

	// check for null subtitle: error if none found
	sqlStmt = `
      SELECT edition
      FROM books
      WHERE book_id = ? AND edition IS NULL
    `
	var readNullEdition sql.NullInt64
	rows, err = db.Query(sqlStmt, 5)
	if err != nil {
		t.Errorf("Querying null edition in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readNullEdition); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		if readNullEdition.Valid {
			t.Errorf("Query returned valid (non-null) edition: \"%v\"",
				readNullEdition.Int64)
		}
	} else {
		t.Errorf("rows.Next() failed with err: %v", rows.Err())
	}
	// Now we need to explicitly close rows to unlock the database for reversion
	// to original values. We can't wait for the deferred function to take
	// effect.
	rows.Close()

	// Revert database to original state
	var origEdition int = 2
	revertedEdition, err := updateBookEdition(db, 5, origEdition)
	if err != nil {
		t.Errorf("Problem reverting edition: %v", err)
	}
	if revertedEdition != origEdition {
		t.Errorf("Wrongly reverted edition: should be \"%v\", but got \"%v\"",
			origEdition, revertedEdition)
	}
}

// [todo] test modification of book publisher by id (try invalid ids)
func TestUpdateBookPublisherById(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newPublisherId int = 3

	updatedPublisherId, err := updateBookPublisherById(db, 1, newPublisherId)
	if err != nil {
		t.Errorf("Problem updating publisher: %v", err)
	}
	if updatedPublisherId != newPublisherId {
		t.Errorf("Wrongly updated publisher: should be \"%v\" but got \"%v\"",
			newPublisherId, updatedPublisherId)
	}

	b, err := getBookById(db, 1)
	retrievedPublisherId, err := publisherId(db, b.publisher)
	if err != nil {
		t.Errorf("Problem getting publisher ID for publisher \"%v\": %v",
			b.publisher, err)
	}
	if retrievedPublisherId != newPublisherId {
		t.Errorf("Wrongly updated publisher from book: should be \"%v\" but got \"%v\"",
			newPublisherId, retrievedPublisherId)
	}

	// Revert database back to original state
	origPublisherId := 1
	revertedPublisherId, err := updateBookPublisherById(db, 1, origPublisherId)
	if err != nil {
		t.Errorf("Problem reverting publisher: %v", err)
	}
	if revertedPublisherId != origPublisherId {
		t.Errorf("Wrongly reverted publisher: should be \"%v\" but got \"%v\"",
			origPublisherId, revertedPublisherId)
	}

	b, err = getBookById(db, 1)
	restoredPublisherId, err := publisherId(db, b.publisher)
	if restoredPublisherId != origPublisherId {
		t.Errorf("Wrongly reverted publisher from book: should be \"%v\" but got \"%v\"",
			origPublisherId, restoredPublisherId)
	}
}

func TestUpdateBookPublisherByIdInvalid(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var origPublisherId int = 1
	var newPublisherId int = 17

	updatedPublisherId, err := updateBookPublisherById(db, 1, newPublisherId)
	if err == nil {
		t.Errorf("Publisher updated to invalid id #%v without error", newPublisherId)
	} else {
		var invPubIdErr *InvalidPublisherIdError
		if !errors.As(err, &invPubIdErr) {
			t.Errorf("Unexpected error when updating publisher to invalid id: %v", err)
		}
	}

	if updatedPublisherId != origPublisherId {
		t.Errorf("Invalid publisher update: should remain \"%v\" but got \"%v\"",
			origPublisherId, updatedPublisherId)
	}

	b, err := getBookById(db, 1)
	retrievedPublisherId, err := publisherId(db, b.publisher)
	if err != nil {
		t.Errorf("Problem getting publisher ID for publisher \"%v\": %v",
			b.publisher, err)
	}
	if retrievedPublisherId != origPublisherId {
		t.Errorf("Wrongly updated publisher from book: should remain \"%v\" but got \"%v\"",
			origPublisherId, retrievedPublisherId)
	}

	// Revert database back to original state (should have no effect)
	revertedPublisherId, err := updateBookPublisherById(db, 1, origPublisherId)
	if err != nil {
		t.Errorf("Problem reverting publisher: %v", err)
	}
	if revertedPublisherId != origPublisherId {
		t.Errorf("Wrongly reverted publisher: should be \"%v\" but got \"%v\"",
			origPublisherId, revertedPublisherId)
	}

	b, err = getBookById(db, 1)
	restoredPublisherId, err := publisherId(db, b.publisher)
	if restoredPublisherId != origPublisherId {
		t.Errorf("Wrongly reverted publisher from book: should be \"%v\" but got \"%v\"",
			origPublisherId, restoredPublisherId)
	}
}

// [todo] test modification of book publisher by name
func TestUpdateBookPublisherByName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newPublisher string = "Penguin Books"

	updatedPublisher, err := updateBookPublisherByName(db, 1, newPublisher)
	if err != nil {
		t.Errorf("Problem updating publisher: %v", err)
	}
	if updatedPublisher != newPublisher {
		t.Errorf("Wrongly updated publisher: should be \"%v\" but got \"%v\"",
			newPublisher, updatedPublisher)
	}

	b, err := getBookById(db, 1)
	if b.publisher != newPublisher {
		t.Errorf("Wrongly updated publisher from book: should be \"%v\" but got \"%v\"",
			newPublisher, b.publisher)
	}

	// Revert database back to original state
	origPublisher := "IVP"
	revertedPublisher, err := updateBookPublisherByName(db, 1, origPublisher)
	if err != nil {
		t.Errorf("Problem reverting publisher: %v", err)
	}
	if revertedPublisher != origPublisher {
		t.Errorf("Wrongly reverted publisher: should be \"%v\" but got \"%v\"",
			origPublisher, revertedPublisher)
	}

	b, err = getBookById(db, 1)
	if b.publisher != origPublisher {
		t.Errorf("Wrongly reverted publisher from book: should be \"%v\" but got \"%v\"",
			origPublisher, b.publisher)
	}
}

func TestUpdateBookPublisherByNameEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newPublisher string = ""
	origPublisher := "IVP"

	_, err = updateBookPublisherByName(db, 1, newPublisher)
	if err == nil {
		t.Errorf("Did not raise error when setting publisher to empty string")
	}

	b, err := getBookById(db, 1)
	if b.publisher != origPublisher {
		t.Errorf("Wrongly updated publisher from book: should be \"%v\" but got \"%v\"",
			newPublisher, b.publisher)
	}

	// Revert database back to original state
	revertedPublisher, err := updateBookPublisherByName(db, 1, origPublisher)
	if err != nil {
		t.Errorf("Problem reverting publisher: %v", err)
	}
	if revertedPublisher != origPublisher {
		t.Errorf("Wrongly reverted publisher: should be \"%v\" but got \"%v\"",
			origPublisher, revertedPublisher)
	}

	b, err = getBookById(db, 1)
	if b.publisher != origPublisher {
		t.Errorf("Wrongly reverted publisher from book: should be \"%v\" but got \"%v\"",
			origPublisher, b.publisher)
	}
}

// [todo] test modification of publisher name
func TestUpdatePublisherName(t *testing.T) {
    db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origName := "IVP"
	origId, err := publisherId(db, origName)
	if err != nil {
		t.Errorf("Problem retrieving publisher %v ID", origName)
	}

	newName := "NavPress"

	updatedName, err := updatePublisherName(db, origId, newName)
	if err != nil {
		t.Errorf("Problem updating publisher: %v", err)
	}
	if updatedName != newName {
		t.Errorf("Publisher name not updated. Expected %v, got %v",
			newName, updatedName)
	}

	newId, err := publisherId(db, newName)
	if err != nil {
		t.Errorf("Couldn't get id of new publisher name %v: %v", newName, err)
	}
	if newId != origId {
		t.Errorf("Publisher name added instead of updated. Expected id %v but got %v",
			origId, newId)
	}

	// Revert to original state
	updatedName, err = updatePublisherName(db, origId, origName)
	if err != nil {
		t.Errorf("Problem reverting publisher: %v", err)
	}
	if updatedName != origName {
		t.Errorf("Publisher name not reverted. Expected %v, got %v",
			origName, updatedName)
	}
	
}

func TestUpdatePublisherNameEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newName string = ""
	_, err = updatePublisherName(db, 1, newName)
	if err == nil {
		t.Errorf("Empty publisher string did not raise error")
	}
}

func TestUpdatePublisherNameDuplicate(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var newName string = "Hackett"
	_, err = updatePublisherName(db, 1, newName)
	if err == nil {
		t.Errorf("Duplicate publisher name did not raise error")
	}
}

// [todo] test modification of isbn
// [todo] test modification of series by id (including trying invalid ids)
// [todo] test modification of series by name (including empty string for NULL)
// [todo] test modification of series name (including error for empty string)
// [todo] test modification of status function (empty string?)
// [todo] test modification of purchased function (empty for NULL)
// [todo] test deletion of book by ID
// [todo] test deletion of person by ID
// [todo] test deletion of publisher by ID
// [todo] test deletion of series by ID
