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

// [todo] test modification of book title (test empty string)
// [todo] test modification of book subtitle (ensure empty string results in NULL)
// [todo] test modification of book year
// [todo] test modification of book edition (including null)
// [todo] test modification of book publisher by id (try invalid ids)
// [todo] test modification of book publisher by name
// [todo] test modification of publisher name
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
