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
		t.Errorf("Wrong value returned by String method on Book: expected %v, got %v",
			expected, bkStr)
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

func TestGetListOfBookIDs(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var expectedIDs = []int{1, 2, 3, 4, 5, 6}

	returnedIDs, err := getListOfBookIDs(db)
	if err != nil {
		t.Errorf("Problem getting book IDs list from DB: %v", err)
	}

	// Cannot simply compare slices with == or !=, as slices can only be
	// compared to nil. So need to check lengths

	if len(returnedIDs) != len(expectedIDs) {
		t.Errorf(
			"getListOfBookIDs returned unexpected value (different lengths).\n"+
				"Expected %v, got %v",
			expectedIDs, returnedIDs,
		)
	}
	for i := range expectedIDs {
		if returnedIDs[i] != expectedIDs[i] {
			t.Errorf(
				"getListOfBookIDs returned unexpected value "+
					"(different value at index %v). Expected %v, got %v",
				i, expectedIDs, returnedIDs,
			)
		}
	}
}

func TestFormatNameListEmpty(t *testing.T) {
	expected := ""

	var nameList = []string{}

	returned := formatNameList(nameList)

	if returned != expected {
		t.Errorf(
			"formatNameList returned unexpected value for input \"%v\". "+
				"Expected \"%v\" but got \"%v\"",
			nameList, expected, returned,
		)
	}
}

func TestFormatNameListSingle(t *testing.T) {
	expected := "Peter J. Gentry"

	var nameList = []string{"Peter J. Gentry"}

	returned := formatNameList(nameList)

	if returned != expected {
		t.Errorf(
			"formatNameList returned unexpected value for input \"%v\".\n"+
				"Expected \"%v\" but got \"%v\"",
			nameList, expected, returned,
		)
	}
}

func TestFormatNameListDouble(t *testing.T) {
	expected := "Peter J. Gentry and Stephen J. Wellum"

	var nameList = []string{"Peter J. Gentry", "Stephen J. Wellum"}

	returned := formatNameList(nameList)

	if returned != expected {
		t.Errorf(
			"formatNameList returned unexpected value for input \"%v\".\n"+
				"Expected \"%v\" but got \"%v\"",
			nameList, expected, returned,
		)
	}
}

func TestFormatNameListTriple(t *testing.T) {
	expected := "Peter J. Gentry, Stephen J. Wellum and Thomas R. Schreiner"

	var nameList = []string{
		"Peter J. Gentry",
		"Stephen J. Wellum",
		"Thomas R. Schreiner",
	}

	returned := formatNameList(nameList)

	if returned != expected {
		t.Errorf(
			"formatNameList returned unexpected value for input \"%v\".\n"+
				"Expected \"%v\" but got \"%v\"",
			nameList, expected, returned,
		)
	}
}

func TestFormatNameListQuadruple(t *testing.T) {
	expected := "Peter J. Gentry, Stephen J. Wellum, Thomas R. Schreiner and Michael A. G. Haykin"

	var nameList = []string{
		"Peter J. Gentry",
		"Stephen J. Wellum",
		"Thomas R. Schreiner",
		"Michael A. G. Haykin",
	}

	returned := formatNameList(nameList)

	if returned != expected {
		t.Errorf(
			"formatNameList returned unexpected value for input \"%v\".\n"+
				"Expected \"%v\" but got \"%v\"",
			nameList, expected, returned,
		)
	}
}

func TestNameListFromStringEmpty(t *testing.T) {
	var expected = []string{}

	nameString := ""

	returned := nameListFromString(nameString)

	if len(returned) != len(expected) {
		t.Errorf(
			"nameListFromString returned unexpected value (slice length).\n"+
				"Expected %v,\nbut got %v",
			expected, returned,
		)
	}

	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"nameListFromString returned unexpected value for slice element %v.\n"+
					"Expected \"%v\", but got \"%v\".\nExpected slice %v,\nbut got slice %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestNameListFromStringSingle(t *testing.T) {
	var expected = []string{"Peter J. Gentry"}

	nameString := "Peter J. Gentry"

	returned := nameListFromString(nameString)

	if len(returned) != len(expected) {
		t.Errorf(
			"nameListFromString returned unexpected value (slice length).\n"+
				"Expected %v,\nbut got %v",
			expected, returned,
		)
	}

	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"nameListFromString returned unexpected value for slice element %v.\n"+
					"Expected \"%v\", but got \"%v\".\nExpected slice %v,\nbut got slice %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestNameListFromStringDouble(t *testing.T) {
	var expected = []string{"Peter J. Gentry", "Stephen J. Wellum"}

	nameString := "Peter J. Gentry and Stephen J. Wellum"

	returned := nameListFromString(nameString)

	if len(returned) != len(expected) {
		t.Errorf(
			"nameListFromString returned unexpected value (slice length).\n"+
				"Expected %v,\nbut got %v",
			expected, returned,
		)
	}

	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"nameListFromString returned unexpected value for slice element %v.\n"+
					"Expected \"%v\", but got \"%v\".\nExpected slice %v,\nbut got slice %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestNameListFromStringTriple(t *testing.T) {
	var expected = []string{
		"Peter J. Gentry",
		"Stephen J. Wellum",
		"Thomas R. Schreiner",
	}

	nameString := "Peter J. Gentry, Stephen J. Wellum and Thomas R. Schreiner"

	returned := nameListFromString(nameString)

	if len(returned) != len(expected) {
		t.Errorf(
			"nameListFromString returned unexpected value (slice length).\n"+
				"Expected %v,\nbut got %v",
			expected, returned,
		)
	}

	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"nameListFromString returned unexpected value for slice element %v.\n"+
					"Expected \"%v\", but got \"%v\".\n"+
					"Expected slice %v,\nbut got slice %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestNameListFromStringQuadruple(t *testing.T) {
	var expected = []string{
		"Peter J. Gentry",
		"Stephen J. Wellum",
		"Thomas R. Schreiner",
		"Michael A. G. Haykin",
	}

	nameString := "Peter J. Gentry, Stephen J. Wellum, " +
		"Thomas R. Schreiner and Michael A. G. Haykin"

	returned := nameListFromString(nameString)

	if len(returned) != len(expected) {
		t.Errorf(
			"nameListFromString returned unexpected value (slice length).\n"+
				"Expected %v,\nbut got %v",
			expected, returned,
		)
	}

	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"nameListFromString returned unexpected value for slice element %v.\n"+
					"Expected \"%v\", but got \"%v\".\nExpected slice %v,\nbut got slice %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestGetAuthorsListById(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var expected = []string{"Peter J. Gentry", "Stephen J. Wellum"}

	returned, err := getAuthorsListById(db, 5)
	if err != nil {
		t.Errorf("Could not get authors list for book id #%v: %v", 5, err)
	}
	if len(returned) != len(expected) {
		t.Errorf(
			"getAuthorsListById returned unexpected value (slice length)\n"+
				"Expected slice of length %v, but got %v.\n"+
				"Expected slice is %v, returned slice is %v",
			len(expected), len(returned), expected, returned,
		)
	}
	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"getAuthorsById returned unexpected value for slice element %v\n"+
					"Expected \"%v\", got \"%v\"\n"+
					"Expected slice is: %v\n"+
					"Returned slice is: %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestGetAuthorsListByIdEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var expected = []string{}

	returned, err := getAuthorsListById(db, 2)
	if err != nil {
		t.Errorf("Error getting authors list: %v", err)
	}
	if len(returned) != len(expected) {
		t.Errorf(
			"getAuthorsListById returned unexpected value (slice length)\n"+
				"Expected slice of length %v, but got %v.\n"+
				"Expected slice is %v, returned slice is %v",
			len(expected), len(returned), expected, returned,
		)
	}
	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"getAuthorsById returned unexpected value for slice element %v\n"+
					"Expected \"%v\", got \"%v\"\n"+
					"Expected slice is: %v\n"+
					"Returned slice is: %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestGetAuthorsListByIdInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	returned, err := getAuthorsListById(db, 17)
	if err == nil {
		t.Errorf("Requesting invalid ID did not return error")
	} else {
		var ibie *InvalidBookIdError
		if !errors.As(err, &ibie) {
			t.Errorf("Requesting invalid ID returned unexpected error: %v", err)
		}
	}
	if len(returned) != 0 {
		t.Errorf("Non-empty slice returned: %v", returned)
	}
}

func TestGetEditorsListById(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var expected = []string{"Robert J. Matz", "A. Chadwick Thornhill"}

	returned, err := getEditorsListById(db, 2)
	if err != nil {
		t.Errorf("Could not get editors list for book id #%v: %v", 2, err)
	}
	if len(returned) != len(expected) {
		t.Errorf(
			"getEditorsListById returned unexpected value (slice length)\n"+
				"Expected slice of length %v, but got %v.\n"+
				"Expected slice is %v, returned slice is %v",
			len(expected), len(returned), expected, returned,
		)
	}
	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"getEditorsById returned unexpected value for slice element %v\n"+
					"Expected \"%v\", got \"%v\"\n"+
					"Expected slice is: %v\n"+
					"Returned slice is: %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestGetEditorsListByIdEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var expected = []string{}

	returned, err := getEditorsListById(db, 5)
	if err != nil {
		t.Errorf("Error getting editors list: %v", err)
	}
	if len(returned) != len(expected) {
		t.Errorf(
			"getEditorsListById returned unexpected value (slice length)\n"+
				"Expected slice of length %v, but got %v.\n"+
				"Expected slice is %v, returned slice is %v",
			len(expected), len(returned), expected, returned,
		)
	}
	for i := range expected {
		if returned[i] != expected[i] {
			t.Errorf(
				"getEditorsById returned unexpected value for slice element %v\n"+
					"Expected \"%v\", got \"%v\"\n"+
					"Expected slice is: %v\n"+
					"Returned slice is: %v",
				i, expected[i], returned[i], expected, returned,
			)
		}
	}
}

func TestGetEditorsListByIdInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	returned, err := getEditorsListById(db, 17)
	if err == nil {
		t.Errorf("Requesting invalid ID did not return error")
	} else {
		var ibie *InvalidBookIdError
		if !errors.As(err, &ibie) {
			t.Errorf("Requesting invalid ID returned unexpected error: %v", err)
		}
	}
	if len(returned) != 0 {
		t.Errorf("Non-empty slice returned: %v", returned)
	}
}

func TestBookIDValid(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 1

	bookValid, err := BookIDValid(db, id)
	if err != nil {
		t.Errorf("BookIDValid returned error: %v", err)
	}
	if !bookValid {
		t.Errorf("BookIDValid returned %v for valid ID #%v", bookValid, id)
	}
}

func TestBookIDValidInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 17

	bookValid, err := BookIDValid(db, id)
	if err != nil {
		t.Errorf("BookIDValid with invalid ID #%v returned unexpected error: %v", id, err)
	}
	if bookValid {
		t.Errorf("BookIDValid returned %v for invalid ID #%v", bookValid, id)
	}
}

func TestGetBookById(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	var purDate PurchasedDate
	err = purDate.setDate("May 2023")
	if err != nil {
		t.Errorf("Problem setting date of PurchasedDate: %v", err)
	}
	var expected = Book{
		id:        1,
		author:    "R. K. Harrison",
		title:     "Introduction to the Old Testament",
		year:      1969,
		publisher: "IVP",
		isbn:      "0-85111-723-6",
		status:    "Owned",
		purchased: purDate,
	}

	id := 1

	returned, err := getBookById(db, id)
	if err != nil {
		t.Errorf("getBookById returned error: %v", err)
	}
	if returned != expected {
		t.Errorf(
			"getBookById returned unexpected book. Expected %v, got %v.",
			expected,
			returned,
		)
	}
}

func TestGetBookByIdInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 17
	returned, err := getBookById(db, id)
	if err == nil {
		t.Errorf("getBookById returned nil error for invalid id #%v", id)
	} else {
		var invlBookIdErr *InvalidBookIdError
		if !errors.As(err, &invlBookIdErr) {
			t.Errorf(
				"getBookById returned unexpected error for invalid id #%v: %v",
				id,
				err,
			)
		}
	}
	var expected = Book{}
	if returned != expected {
		t.Errorf(
			"getBookById returned unexpected value for invalid id #%v:\n"+
				"Expected %v, got %v",
			id,
			expected,
			returned,
		)
	}

}

func TestPersonName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 1
	expected := "R. K. Harrison"

	result, err := personName(db, id)
	if err != nil {
		t.Errorf(
			"Unexpected error getting person name for id #%v: %v",
			id,
			err,
		)
	}
	if result != expected {
		t.Errorf(
			"Unexpected value from personName for id #%v. Expected %v, got %v",
			id,
			expected,
			result,
		)
	}
}

func TestPersonNameInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 26
	expected := ""

	result, err := personName(db, id)
	if err == nil {
		t.Errorf(
			"personName did not return error for invalid id #%v",
			id,
		)
	} else {
		var invlPersIdErr *InvalidPersonIdError
		if !errors.As(err, &invlPersIdErr) {
			t.Errorf(
				"personName returned unexpected error for invalid ID #%v: %v",
				id,
				err,
			)
		}
	}
	if result != expected {
		t.Errorf(
			"Unexpected value from personName for id #%v. Expected %v, got %v",
			id,
			expected,
			result,
		)
	}
}

func TestPersonId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "R. K. Harrison"
	expectedID := 1

	returnedID, err := personId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of person \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"personId returned unexpected value for person \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestPersonIdNewPerson(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "Thomas R. Schreiner"
	expectedID := 12

	returnedID, err := personId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of person \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"personId returned unexpected value for person \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}

	// revert database to original setting
	err = deletePerson(db, returnedID)
	if err != nil {
		t.Errorf(
			"Unexpected error when deleting person (to restore DB state): %v",
			err,
		)
	}
}

func TestPersonIdEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := ""
	expectedID := 0

	returnedID, err := personId(db, name)
	if err == nil {
		t.Errorf("Calling personID with empty name string did not return error")
	}
	if returnedID != expectedID {
		t.Errorf(
			"personId returned unexpected value for person \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestBooksByPersonId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	person := "Peter J. Gentry"
	persId, err := personId(db, person)
	if err != nil {
		t.Errorf("Problem retrieving ID for person \"%v\": %v", person, err)
	}

	expected := []int{4, 5}

	result, err := booksByPersonId(db, persId)
	if err != nil {
		t.Errorf(
			"Problem retrieving books of person #%v %v: %v",
			persId,
			person,
			err,
		)
	}

	if len(result) != len(expected) {
		t.Errorf(
			"booksByPersonId returned unexpected value (slice length).\n"+
				"Expected slice of length %v, got length %v\n"+
				"Expected slice: %v, got: %v",
			len(expected),
			len(result),
			expected,
			result,
		)
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf(
				"booksByPersonId returned unexpected value (slice element %v).\n"+
					"Expected element: %v, got: %v\n"+
					"Expected slice: %v, got: %v",
				i,
				b,
				result[i],
				expected,
				result,
			)
		}
	}
}

func TestBooksByPersonIdInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	persId := 73
	result, err := booksByPersonId(db, persId)
	if err == nil {
		t.Errorf("booksByPersonId did not return error for invalid id #%v", persId)
	} else {
		var invlPersIdErr *InvalidPersonIdError
		if !errors.As(err, &invlPersIdErr) {
			t.Errorf(
				"booksByPersonId returned unexpected error for invalid id #%v: %v",
				persId,
				err,
			)
		}
	}
	if result != nil {
		t.Errorf("booksByPersonId returned non-nil value for invalid id: %v", result)
	}
}

func TestPublisherName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pubId := 1
	expected := "IVP"

	result, err := publisherName(db, pubId)
	if err != nil {
		t.Errorf("Unexpected error from publisherName: %v", err)
	}
	if result != expected {
		t.Errorf(
			"publisherName returned unexpected value. Expected %v, got %v.",
			expected,
			result,
		)
	}
}

func TestPublisherNameInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pubId := 7

	result, err := publisherName(db, pubId)
	if err == nil {
		t.Errorf("publisherName did not return error for invalid ID #%v", pubId)
	} else {
		var invlPubIdErr *InvalidPublisherIdError
		if !errors.As(err, &invlPubIdErr) {
			t.Errorf(
				"publisherName returned unexpected error for invalid ID #%v: %v",
				pubId,
				err,
			)
		}
	}
	if result != "" {
		t.Errorf(
			"publisherName returned unexpected result for invalid ID #%v: %v",
			pubId,
			result,
		)
	}
}

func TestPublisherBooks(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pubId := 3
	expected := []int{4, 5, 6}

	result, err := publisherBooks(db, pubId)
	if err != nil {
		t.Errorf("publisherBooks, Unexpected error: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf(
			"publisherBooks returned unexpected result (slice length)\n"+
				"Expected slice of length %v, but got length %v\n"+
				"Expected slice: %v, but got: %v",
			len(expected),
			len(result),
			expected,
			result,
		)
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf(
				"publisherBooks returned unexpected result (slice element %v)\n"+
					"Expected element: %v, got element: %v\n"+
					"Expected slice: %v, got slice: %v",
				i,
				b,
				result[i],
				expected,
				result,
			)
		}
	}
}

func TestPublisherBooksInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pubId := 7

	result, err := publisherBooks(db, pubId)
	if err == nil {
		t.Errorf("publisherBooks did not return error for invalid ID #%v", pubId)
	} else {
		var invlPubIdErr *InvalidPublisherIdError
		if !errors.As(err, &invlPubIdErr) {
			t.Errorf(
				"publisherBooks returned unexpected error for invalid ID #%v: %v",
				pubId,
				err,
			)
		}
	}
	if result != nil {
		t.Errorf(
			"publisherBooks returned unexpected result for invalid ID #%v: %v",
			pubId,
			result,
		)
	}
}

func TestPublisherId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "IVP"
	expectedID := 1

	returnedID, err := publisherId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of publisher \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"publisherId returned unexpected value for publisher \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestPublisherIdNewPublisher(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "Penguin Books"
	expectedID := 4

	returnedID, err := publisherId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of publisher \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"publisherId returned unexpected value for publisher \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}

	// revert database to original setting
	err = deletePublisher(db, returnedID)
	if err != nil {
		t.Errorf(
			"Unexpected error when deleting publisher (to restore DB state): %v",
			err,
		)
	}
}

func TestPublisherIdEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := ""
	expectedID := 0

	returnedID, err := publisherId(db, name)
	if err == nil {
		t.Errorf("Calling publisherID with empty name string did not return error")
	}
	if returnedID != expectedID {
		t.Errorf(
			"publisherId returned unexpected value for publisher \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestSeriesId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "Spectrum Multiview Books"
	expectedID := 1

	returnedID, err := seriesId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of series \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"seriesId returned unexpected value for series \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestSeriesIdNewSeries(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := "Penguin Classics"
	expectedID := 2

	returnedID, err := seriesId(db, name)
	if err != nil {
		t.Errorf("Unexpected error when getting ID of series \"%v\": %v", name, err)
	}
	if returnedID != expectedID {
		t.Errorf(
			"seriesId returned unexpected value for series \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}

	// revert database to original setting
	err = deleteSeries(db, returnedID)
	if err != nil {
		t.Errorf(
			"Unexpected error when deleting series (to restore DB state): %v",
			err,
		)
	}
}

func TestSeriesIdEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	name := ""
	expectedID := 0

	returnedID, err := seriesId(db, name)
	if err == nil {
		t.Errorf("Calling seriesID with empty name string did not return error")
	}
	if returnedID != expectedID {
		t.Errorf(
			"seriesId returned unexpected value for series \"%v\"\n"+
				"Expected %v, got %v",
			name,
			expectedID,
			returnedID,
		)
	}
}

func TestSeriesBooks(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	serId := 1
	expected := []int{2}

	result, err := seriesBooks(db, serId)
	if err != nil {
		t.Errorf("seriesBooks, Unexpected error: %v", err)
	}

	if len(result) != len(expected) {
		t.Errorf(
			"seriesBooks returned unexpected result (slice length)\n"+
				"Expected slice of length %v, but got length %v\n"+
				"Expected slice: %v, but got: %v",
			len(expected),
			len(result),
			expected,
			result,
		)
	}
	for i, b := range expected {
		if result[i] != b {
			t.Errorf(
				"seriesBooks returned unexpected result (slice element %v)\n"+
					"Expected element: %v, got element: %v\n"+
					"Expected slice: %v, got slice: %v",
				i,
				b,
				result[i],
				expected,
				result,
			)
		}
	}
}

func TestSeriesBooksInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	serId := 5

	result, err := seriesBooks(db, serId)
	if err == nil {
		t.Errorf("seriesBooks did not return error for invalid ID #%v", serId)
	} else {
		var invlSerIdErr *InvalidSeriesIdError
		if !errors.As(err, &invlSerIdErr) {
			t.Errorf(
				"seriesBooks returned unexpected error for invalid ID #%v: %v",
				serId,
				err,
			)
		}
	}
	if result != nil {
		t.Errorf(
			"seriesBooks returned unexpected result for invalid ID #%v: %v",
			serId,
			result,
		)
	}
}

func TestSeriesName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	serId := 1
	expected := "Spectrum Multiview Books"

	result, err := seriesName(db, serId)
	if err != nil {
		t.Errorf("Unexpected error from seriesName: %v", err)
	}
	if result != expected {
		t.Errorf(
			"seriesName returned unexpected value. Expected %v, got %v.",
			expected,
			result,
		)
	}
}

func TestSeriesNameInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	serId := 5

	result, err := seriesName(db, serId)
	if err == nil {
		t.Errorf("seriesName did not return error for invalid ID #%v", serId)
	} else {
		var invlSerIdErr *InvalidSeriesIdError
		if !errors.As(err, &invlSerIdErr) {
			t.Errorf(
				"seriesName returned unexpected error for invalid ID #%v: %v",
				serId,
				err,
			)
		}
	}
	if result != "" {
		t.Errorf(
			"seriesBooks returned unexpected result for invalid ID #%v: %v",
			serId,
			result,
		)
	}
}

func TestCheckBookInDb(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 1

	var pd PurchasedDate
	err = pd.setDate("May 2023")
	if err != nil {
		t.Errorf("Problem setting date of PurchasedDate: %v", err)
	}

	var b = Book{
		id:        1,
		author:    "R. K. Harrison",
		title:     "Introduction to the Old Testament",
		year:      1969,
		publisher: "IVP",
		isbn:      "0-85111-723-6",
		status:    "Owned",
		purchased: pd,
	}

	result, err := checkBookInDb(db, &b)
	if err != nil {
		t.Errorf("Unexpected error while checking for book: %v", err)
	}
	if result != expected {
		t.Errorf(
			"Unexpected return value from checkBookInDb\n"+
				"Expected %v, got %v",
			expected, result,
		)
	}
}

// Ensure that adding a book with the same author, title, etc, but a different
// ID value (not actually describing the book) finds the book. In other words,
// ensure that checking for the book in the database depends only on information
// about the book, not about the database.
func TestCheckBookInDbDifferentId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 1

	var pd PurchasedDate
	err = pd.setDate("January 2024")
	if err != nil {
		t.Errorf("Problem setting date of PurchasedDate: %v", err)
	}

	var b = Book{
		id:        27,
		author:    "R. K. Harrison",
		title:     "Introduction to the Old Testament",
		year:      1969,
		publisher: "IVP",
		isbn:      "0-85111-723-6",
		status:    "Owned",
		purchased: pd,
	}

	result, err := checkBookInDb(db, &b)
	if err != nil {
		t.Errorf("Unexpected error while checking for book: %v", err)
	}
	if result != expected {
		t.Errorf(
			"Unexpected return value from checkBookInDb\n"+
				"Expected %v, got %v",
			expected, result,
		)
	}
}

func TestCheckBookInDbUnknownBook(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	expected := 0

	var pd PurchasedDate
	err = pd.setDate("January 2023")
	if err != nil {
		t.Errorf("Problem setting date of PurchasedDate: %v", err)
	}

	var b = Book{
		id:        48,
		author:    "Christopher Watkin",
		title:     "Biblical Critical Theory",
		year:      2022,
		publisher: "Zondervan",
		isbn:      "978-0-310-12872-4",
		status:    "Owned",
		purchased: pd,
	}

	result, err := checkBookInDb(db, &b)
	if err != nil {
		t.Errorf("Unexpected error while checking for book: %v", err)
	}
	if result != expected {
		t.Errorf(
			"Unexpected return value from checkBookInDb\n"+
				"Expected %v, got %v",
			expected, result,
		)
	}
}

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
		t.Errorf(
			"Title not correctly reverted. Should be \"%v\", instead is \"%v\"",
			newTitle,
			updatedTitle,
		)
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

func TestUpdateBookIsbn(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origIsbn := "0-85111-723-6"
	newIsbn := "978-1408855652"

	updatedIsbn, err := updateBookIsbn(db, 1, newIsbn)
	if err != nil {
		t.Errorf("Problem updating ISBN: %v", err)
	}
	if updatedIsbn != newIsbn {
		t.Errorf("ISBN not updated. Expected %v, got %v", newIsbn, updatedIsbn)
	}

	b, err := getBookById(db, 1)
	if err != nil {
		t.Errorf("Problem reading book from database: %v", err)
	}
	if b.isbn != newIsbn {
		t.Errorf("ISBN not properly updated. Expected %v, got %v", newIsbn, b.isbn)
	}

	// Revert to original state
	revertedIsbn, err := updateBookIsbn(db, 1, origIsbn)
	if err != nil {
		t.Errorf("Problem reverting ISBN: %v", err)
	}
	if revertedIsbn != origIsbn {
		t.Errorf("ISBN not reverted. Expected %v, got %v",
			origIsbn, revertedIsbn)
	}
}

func TestUpdateBookSeriesById(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origId := 1
	origName := "Spectrum Multiview Books"

	// Add an extra series
	seriesName := "New Studies in Biblical Theology"
	newId, err := seriesId(db, seriesName)
	if err != nil {
		t.Errorf("Problem creating new series: %v", err)
	}

	updatedId, err := updateBookSeriesById(db, 2, newId)
	if err != nil {
		t.Errorf("Problem updating series id: %v", err)
	}
	if updatedId != newId {
		t.Errorf("Series id not updated. Expected %v, got %v", newId, updatedId)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != seriesName {
		t.Errorf("Series not properly updated. Expected %v, got %v", seriesName,
			b.series)
	}

	// revert
	revertedId, err := updateBookSeriesById(db, 2, origId)
	if err != nil {
		t.Errorf("Problem updating series id: %v", err)
	}
	if revertedId != origId {
		t.Errorf("Series id not reverted. Expected %v, got %v", origId, revertedId)
	}

	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != origName {
		t.Errorf("Series not properly updated. Expected %v, got %v", origName,
			b.series)
	}

	// clean up by deleting added series
	err = deleteSeries(db, newId)
	if err != nil {
		t.Errorf("Problem removing series to revert database: %v", err)
	}
}

func TestUpdateBookSeriesByIdNull(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origId := 1
	origName := "Spectrum Multiview Books"

	newId := 0
	seriesName := ""

	updatedId, err := updateBookSeriesById(db, 2, newId)
	if err != nil {
		t.Errorf("Problem setting null series id: %v", err)
	}
	if updatedId != newId {
		t.Errorf("Series id not correctly updated. Expected %v, got %v", newId, updatedId)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != seriesName {
		t.Errorf("Series not properly updated. Expected %v, got %v", seriesName,
			b.series)
	}

	// check that database value is NULL, not simply set to id zero and empty
	// string name

	// check for non-null subtitles: error if any found
	sqlStmt := `
      SELECT series_id
      FROM books
      WHERE book_id = ? AND series_id IS NOT NULL
    `
	var readSeriesId sql.NullInt64
	rows, err := db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("Error querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readSeriesId); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		t.Errorf("Query returned non-null value \"%v\"", readSeriesId)
	} else {
		if err := rows.Err(); err != nil {
			t.Errorf("rows.Next() failed with non-nil error: %v", err)
		}
	}

	// check for null subtitle: error if none found
	sqlStmt = `
      SELECT series_id
      FROM books
      WHERE book_id = ? AND series_id IS NULL
    `
	rows, err = db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("Querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readSeriesId); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		if readSeriesId.Valid {
			t.Errorf("Query returned valid series id: \"%v\"", readSeriesId.Int64)
		}
	} else {
		t.Errorf("rows.Next() failed with err: %v", rows.Err())
	}
	// Now we need to explicitly close rows to unlock the database for reversion
	// to original values. We can't wait for the deferred function to take
	// effect.
	rows.Close()

	// revert
	revertedId, err := updateBookSeriesById(db, 2, origId)
	if err != nil {
		t.Errorf("Problem updating series id: %v", err)
	}
	if revertedId != origId {
		t.Errorf("Series id not reverted. Expected %v, got %v", origId, revertedId)
	}

	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != origName {
		t.Errorf("Series not properly updated. Expected %v, got %v", origName,
			b.series)
	}
}

func TestUpdateBookSeriesByIdInvalid(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origId := 1
	origName := "Spectrum Multiview Books"

	// Add an extra series
	invalidId := 2

	updatedId, err := updateBookSeriesById(db, 2, invalidId)
	if err != nil {
		var invSerId *InvalidSeriesIdError
		if !errors.As(err, &invSerId) {
			t.Errorf("Unexpected error updating with invalid series id: %v", err)
		}
	} else {
		t.Errorf("Updating series to invalid ID did not generate error")
	}
	if updatedId != origId {
		t.Errorf("Series id wrongly updated. Expected %v, got %v", origId, updatedId)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != origName {
		t.Errorf("Series wrongly updated. Expected %v, got %v", origName, b.series)
	}

	// revert
	revertedId, err := updateBookSeriesById(db, 2, origId)
	if err != nil {
		t.Errorf("Problem updating series id: %v", err)
	}
	if revertedId != origId {
		t.Errorf("Series id not reverted. Expected %v, got %v", origId, revertedId)
	}

	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != origName {
		t.Errorf("Series not properly updated. Expected %v, got %v", origName,
			b.series)
	}
}

func TestUpdateBookSeriesByName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origName := "Spectrum Multiview Books"

	newName := "New Studies in Biblical Theology"

	updatedName, err := updateBookSeriesByName(db, 2, newName)
	if err != nil {
		t.Errorf("Problem updating book series: %v", err)
	}

	if updatedName != newName {
		t.Errorf("Series not correctly updated. Expected \"%v\", got \"%v\"",
			newName, updatedName)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Could not get book with ID #%v: %v", 2, err)
	}
	if b.series != newName {
		t.Errorf("Series not correctly updated in database. Expected \"%v\", got \"%v\"", newName, b.series)
	}

	// revert to original value
	revertedName, err := updateBookSeriesByName(db, 2, origName)
	if err != nil {
		t.Errorf("Problem reverting book series: %v", err)
	}
	if revertedName != origName {
		t.Errorf("Series not correctly reverted. Expected \"%v\", got \"%v\"",
			origName, revertedName)
	}
	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Could not get book with ID #%v: %v", 2, err)
	}
	if b.series != origName {
		t.Errorf("Series not correctly reverted in database. Expected \"%v\", got \"%v\"",
			origName, b.series)
	}

	newSeriesId, err := seriesId(db, newName)
	if err != nil {
		t.Errorf("Could not get ID to delete series \"%v\": %v", newName, err)
	}
	deleteSeries(db, newSeriesId)
}

func TestUpdateBookSeriesByNameEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origId := 1
	origName := "Spectrum Multiview Books"

	newName := ""

	updatedName, err := updateBookSeriesByName(db, 2, newName)
	if err != nil {
		t.Errorf("Problem setting empty series name: %v", err)
	}
	if updatedName != newName {
		t.Errorf("Series name not correctly updated. Expected \"%v\", got \"%v\"", newName, updatedName)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != newName {
		t.Errorf("Book series not properly updated in DB. Expected \"%v\", got \"%v\"",
			newName, b.series)
	}

	// check that database value is NULL, not simply set to id zero and empty
	// string name

	// check for non-null series: error if any found
	sqlStmt := `
      SELECT series_id
      FROM books
      WHERE book_id = ? AND series_id IS NOT NULL
    `
	var readSeriesId sql.NullInt64
	rows, err := db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("Error querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readSeriesId); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		t.Errorf("Query returned non-null value \"%v\"", readSeriesId)
	} else {
		if err := rows.Err(); err != nil {
			t.Errorf("rows.Next() failed with non-nil error: %v", err)
		}
	}

	// check for null subtitle: error if none found
	sqlStmt = `
      SELECT series_id
      FROM books
      WHERE book_id = ? AND series_id IS NULL
    `
	rows, err = db.Query(sqlStmt, 2)
	if err != nil {
		t.Errorf("Querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readSeriesId); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		if readSeriesId.Valid {
			t.Errorf("Query returned valid series id: \"%v\"", readSeriesId.Int64)
		}
	} else {
		t.Errorf("rows.Next() failed with err: %v", rows.Err())
	}
	// Now we need to immediately close rows to unlock the database for reversion
	// to original values. We can't wait for the deferred function to take
	// effect.
	rows.Close()

	// revert
	revertedId, err := updateBookSeriesById(db, 2, origId)
	if err != nil {
		t.Errorf("Problem updating series id: %v", err)
	}
	if revertedId != origId {
		t.Errorf("Series id not reverted. Expected %v, got %v", origId, revertedId)
	}

	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Problem getting book from database: %v", err)
	}
	if b.series != origName {
		t.Errorf("Series not properly updated. Expected %v, got %v", origName,
			b.series)
	}
}

func TestUpdateSeriesName(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origName := "Spectrum Multiview Books"
	newName := "New Studies in Biblical Theology"

	updatedName, err := updateSeriesName(db, 1, newName)
	if err != nil {
		t.Errorf("Problem updating series name: %v", err)
	}

	if updatedName != newName {
		t.Errorf("Series name not correctly updated. Expected \"%v\", got \"%v\"",
			newName, updatedName)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Could not get book with ID #%v: %v", 2, err)
	}
	if b.series != newName {
		t.Errorf("Series not correctly updated in database. Expected \"%v\", got \"%v\"", newName, b.series)
	}

	// revert to original value
	revertedName, err := updateSeriesName(db, 1, origName)
	if err != nil {
		t.Errorf("Problem reverting book series: %v", err)
	}
	if revertedName != origName {
		t.Errorf("Series not correctly reverted. Expected \"%v\", got \"%v\"",
			origName, revertedName)
	}
	b, err = getBookById(db, 2)
	if err != nil {
		t.Errorf("Could not get book with ID #%v: %v", 2, err)
	}
	if b.series != origName {
		t.Errorf("Series name not correctly reverted in database. Expected \"%v\", got \"%v\"",
			origName, b.series)
	}
}

func TestUpdateSeriesNameEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origName := "Spectrum Multiview Books"
	newName := ""

	updatedName, err := updateSeriesName(db, 1, newName)
	if err == nil {
		t.Errorf("Setting series name to empty string did not cause error")
	}

	if updatedName != origName {
		t.Errorf("Series name incorrectly updated. Expected \"%v\", got \"%v\"",
			origName, updatedName)
	}

	b, err := getBookById(db, 2)
	if err != nil {
		t.Errorf("Could not get book with ID #%v: %v", 2, err)
	}
	if b.series != origName {
		t.Errorf("Series incorrectly updated in database. Expected \"%v\", got \"%v\"", origName, b.series)
	}
}

func TestUpdateBookStatus(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origStatus := "Owned"
	newStatus := "Want"

	updatedStatus, err := updateBookStatus(db, 1, newStatus)
	if err != nil {
		t.Errorf("Could not update book status: %v", err)
	}
	if updatedStatus != newStatus {
		t.Errorf("UpdateBookStatus returned unexpected value. Expected \"%v\", got \"%v\"",
			newStatus, updatedStatus)
	}
	b, err := getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.status != newStatus {
		t.Errorf("Book status not properly updated in database: expected \"%v\", got \"%v\"",
			newStatus, b.status)
	}

	// Revert database to original values
	revertedStatus, err := updateBookStatus(db, 1, origStatus)
	if err != nil {
		t.Errorf("Could not revert book status: %v", err)
	}
	if revertedStatus != origStatus {
		t.Errorf("UpdateBookStatus returned unexpected value. Expected \"%v\", got \"%v\"",
			origStatus, revertedStatus)
	}
	b, err = getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.status != origStatus {
		t.Errorf("Book status not properly updated in database: expected \"%v\", got \"%v\"",
			origStatus, b.status)
	}
}

func TestUpdateBookStatusEmptyString(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origStatus := "Owned"
	newStatus := ""

	_, err = updateBookStatus(db, 1, newStatus)
	if err == nil {
		t.Errorf("Book status empty string did not return error")
	}

	b, err := getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.status != origStatus {
		t.Errorf("Book status wrongly updated in database: expected \"%v\", got \"%v\"",
			origStatus, b.status)
	}
}

func TestUpdateBookPurchaseDate(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origDate := "May 2023"
	newDate := "19 April 2021"

	var origPD PurchasedDate
	var newPD PurchasedDate

	err = origPD.setDate(origDate)
	if err != nil {
		t.Errorf("Problem setting date with value \"%v\": %v", origDate, err)
	}
	err = newPD.setDate(newDate)
	if err != nil {
		t.Errorf("Problem setting date with value \"%v\": %v", newDate, err)
	}

	updatedPD, err := updateBookPurchaseDate(db, 1, newPD)
	if err != nil {
		t.Errorf("Could not update purchase date: %v", err)
	}
	if updatedPD != newPD {
		t.Errorf("updateBookPurchaseDate returned unexpected value. Expected \"%v\", got \"%v\"",
			newPD, updatedPD)
	}

	b, err := getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.purchased != newPD {
		t.Errorf("Purchase date not properly updated in DB. Expected \"%v\" but got \"%v\"",
			newPD, b.purchased)
	}

	// Revert database to default state
	revertedPD, err := updateBookPurchaseDate(db, 1, origPD)
	if err != nil {
		t.Errorf("Could not revert purchase date: %v", err)
	}
	if revertedPD != origPD {
		t.Errorf("updateBookPurchaseDate returned unexpected value. Expected \"%v\", got \"%v\"",
			origPD, revertedPD)
	}

	b, err = getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.purchased != origPD {
		t.Errorf("Purchase date not properly reverted in DB. Expected \"%v\" but got \"%v\"",
			origPD, b.purchased)
	}
}

func TestUpdateBookPurchaseDateNull(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	origDate := "May 2023"
	var origPD PurchasedDate
	err = origPD.setDate(origDate)
	if err != nil {
		t.Errorf("Could not set date value \"%v\": %v", origDate, err)
	}

	var newPD PurchasedDate

	updatedPD, err := updateBookPurchaseDate(db, 1, newPD)
	if err != nil {
		t.Errorf("Could not update purchased date: %v", err)
	}
	if updatedPD != newPD {
		t.Errorf("updateBookPurchaseDate returned unexpected value. Expected \"%v\", got \"%v\"",
			newPD, updatedPD)
	}

	b, err := getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.purchased != newPD {
		t.Errorf("Purchase date did not update properly in DB. Expected \"%v\", got \"%v\"",
			newPD, b.purchased)
	}

	// check for non-null purchased date: error if any found
	sqlStmt := `
      SELECT purchased_date
      FROM books
      WHERE book_id = ? AND purchased_date IS NOT NULL
    `
	var readPurchasedDate sql.NullString
	rows, err := db.Query(sqlStmt, 1)
	if err != nil {
		t.Errorf("Error querying purchased date in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readPurchasedDate); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		t.Errorf("Query returned non-null value \"%v\"", readPurchasedDate)
	} else {
		if err := rows.Err(); err != nil {
			t.Errorf("rows.Next() failed with non-nil error: %v", err)
		}
	}

	// check for null purchased date: error if none found
	sqlStmt = `
      SELECT purchased_date
      FROM books
      WHERE book_id = ? AND purchased_date IS NULL
    `
	rows, err = db.Query(sqlStmt, 1)
	if err != nil {
		t.Errorf("Querying subtitle in database: %v", err)
	}
	defer rows.Close()
	if rows.Next() {
		if err := rows.Scan(&readPurchasedDate); err != nil {
			t.Errorf("Issue scanning row: %v", err)
		}
		if readPurchasedDate.Valid {
			t.Errorf("Query returned valid (non-null) purchased date: \"%v\"", readPurchasedDate.String)
		}
	} else {
		t.Errorf("rows.Next() failed with err: %v", rows.Err())
	}
	// Now we need to immediately close rows to unlock the database for reversion
	// to original values. We can't wait for the deferred function to take
	// effect.
	rows.Close()

	// Now need to revert database to original state
	revertedPD, err := updateBookPurchaseDate(db, 1, origPD)
	if err != nil {
		t.Errorf("Could not revert purchase date: %v", err)
	}
	if revertedPD != origPD {
		t.Errorf("updateBookPurchaseDate returned unexpected value. Expected \"%v\", got \"%v\"",
			origPD, revertedPD)
	}

	b, err = getBookById(db, 1)
	if err != nil {
		t.Errorf("Could not retrieve book from database: %v", err)
	}
	if b.purchased != origPD {
		t.Errorf(
			"Purchase date not properly reverted in DB. "+
				"Expected \"%v\" but got \"%v\"",
			origPD,
			b.purchased,
		)
	}

}

func TestDeleteBook(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	newBook := makeTestBook()
	newBook.series = "Studies in Septuagint and Sausages"

	id, err := addBook(db, newBook)
	if err != nil {
		t.Errorf("Issue adding book to test deletion: %v", err)
	}
	addedBook, err := getBookById(db, id)
	if err != nil {
		t.Errorf("Issue retrieving book: %v", err)
	}
	newBook.id = id
	if addedBook != *newBook {
		t.Errorf(
			"Added book does not match book added.\n"+
				"Expected: %v, but got %v",
			newBook,
			addedBook,
		)

	}

	checkId, err := checkBookInDb(db, newBook)
	if err != nil {
		t.Errorf("Problem checking for book in DB: %v", err)
	}
	if checkId != id {
		t.Errorf("checkBookInDb returned wrong id: Expected %v, got %v", id, checkId)
	}

	err = deleteBook(db, id)
	if err != nil {
		t.Errorf("Issue deleting book: %v", err)
	}

	checkId, err = checkBookInDb(db, newBook)
	if err != nil {
		t.Errorf("Problem checking for book in DB: %v", err)
	}
	if checkId != 0 {
		t.Errorf(
			"checkBookInDb returned valid id after book deletion: "+
				"Expected %v, got %v",
			id,
			checkId,
		)
	}

	// ensure cleanup of people, publishers, series if no books remain
	checkPeopleSql := `SELECT COUNT(*)
        FROM people
        WHERE name = ?`

	var people = []string{"Karen H. Jobes", "Moisés Silva"}
	var count int

	for _, p := range people {
		if err := db.QueryRow(checkPeopleSql, p).Scan(&count); err != nil {
			t.Errorf("Problem querying people in DB: %v", err)
		}
		if count != 0 {
			t.Errorf(
				"Person \"%v\" still in database after deletion of sole book",
				p,
			)
		}
	}

	checkPublisherSql := `SELECT COUNT(*)
	    FROM publishers
        WHERE name = ?`

	publisher := "Baker Academic"
	if err := db.QueryRow(checkPublisherSql, publisher).Scan(&count); err != nil {
		t.Errorf("Problem querying publisher in DB: %v", err)
	}
	if count != 0 {
		t.Errorf(
			"Publisher \"%v\" still in database after deletion of sole book",
			publisher,
		)
	}

	checkSeriesSql := `SELECT COUNT(*)
        FROM series
        WHERE series_name = ?`

	series := "Studies in Septuagint and Sausages"
	if err := db.QueryRow(checkSeriesSql, series).Scan(&count); err != nil {
		t.Errorf("Problem querying series in DB: %v", err)
	}
	if count != 0 {
		t.Errorf(
			"Series \"%v\" still in database after deletion of sole book",
			series,
		)
	}
}

func TestDeleteBookInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 43
	err = deleteBook(db, id)
	if err == nil {
		t.Errorf("Deleting invalid book id #%v did not return error", id)
	} else {
		var invlBookIdErr *InvalidBookIdError
		if !errors.As(err, &invlBookIdErr) {
			t.Errorf(
				"Deleting invalid book id #%v returned unexpected error "+
					"(not InvalidBookIdError): %v",
				id,
				err,
			)
		}
	}
}

func TestDeletePerson(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	newPerson := "Francis Turretin"
	id, err := personId(db, newPerson)
	if err != nil {
		t.Errorf("Could not retrieve ID for new person %v: %v", newPerson, err)
	}

	err = deletePerson(db, id)
	if err != nil {
		t.Errorf("Problem deleting newly added person: %v ", err)
	}

	persName, err := personName(db, id)
	if err == nil {
		t.Errorf(
			"personName did not raise error after deletion of person ID #%v: %v",
			id,
			err,
		)
	} else {
		var invlPersIdErr *InvalidPersonIdError
		if !errors.As(err, &invlPersIdErr) {
			t.Errorf(
				"Unexpected error raised by personName after deletion of"+
					" person with ID #%v: %v",
				id,
				err,
			)
		}
	}
	if persName != "" {
		t.Errorf("personName returned non-empty string after deletion of person")
	}
}

func TestDeletePersonInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	id := 26

	err = deletePerson(db, id)
	if err == nil {
		t.Errorf(
			"deletePerson did not return error for invalid id #%v",
			id,
		)
	} else {
		var invlPersIdErr *InvalidPersonIdError
		if !errors.As(err, &invlPersIdErr) {
			t.Errorf(
				"deletePerson returned unexpected error for invalid ID #%v: %v",
				id,
				err,
			)
		}
	}
}

func TestDeletePersonInUse(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	person := "Peter J. Gentry"
	persId, err := personId(db, person)
	if err != nil {
		t.Errorf("Problem getting ID for %v: %v", person, err)
	}

	err = deletePerson(db, persId)
	if err == nil {
		t.Errorf(
			"deletePerson did not return error for in use person #%v %v",
			persId,
			person,
		)
	} else {
		var persInUseErr *PersonInUseError
		if !errors.As(err, &persInUseErr) {
			t.Errorf(
				"deletePerson returned unexpected error for in use person #%v %v: %v",
				persId,
				person,
				err,
			)
		}
	}
}

func TestDeletePublisher(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	publisher := "Penguin Books"
	pubId, err := publisherId(db, publisher)
	if err != nil {
		t.Errorf("publisherId returned unexpected error: %v", err)
	}

	err = deletePublisher(db, pubId)
	if err != nil {
		t.Errorf("deletePublisher gave unexpected error: %v", err)
	}

	name, err := publisherName(db, pubId)
	if err == nil {
		t.Errorf(
			"Publisher still in DB after deletion, publisherName returned no error",
		)
	} else {
		var invlPubIdErr *InvalidPublisherIdError
		if !errors.As(err, &invlPubIdErr) {
			t.Errorf(
				"publisherName returned unexpected error after deletion of publisher: %v",
				err,
			)
		}
	}
	if name != "" {
		t.Errorf(
			"publisherName returned unexpected result after "+
				"deletion of publisher: %v",
			name,
		)
	}
}

func TestDeletePublisherInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	pubId := 7

	err = deletePublisher(db, pubId)
	if err == nil {
		t.Errorf("deletePublisher did not return error for invalid ID #%v", pubId)
	} else {
		var invlPubIdErr *InvalidPublisherIdError
		if !errors.As(err, &invlPubIdErr) {
			t.Errorf(
				"deletePublisher returned unexpected error for invalid ID #%v: %v",
				pubId,
				err,
			)
		}
	}
}

func TestDeletePublisherInUse(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	publisher := "IVP"
	pubId, err := publisherId(db, publisher)
	if err != nil {
		t.Errorf("Problem getting ID for %v: %v", publisher, err)
	}

	err = deletePublisher(db, pubId)
	if err == nil {
		t.Errorf(
			"deletePublisher did not return error for in use publisher #%v %v",
			pubId,
			publisher,
		)
	} else {
		var pubInUseErr *PublisherInUseError
		if !errors.As(err, &pubInUseErr) {
			t.Errorf(
				"deletePublisher returned unexpected error for in use publisher #%v %v: %v",
				pubId,
				publisher,
				err,
			)
		}
	}
}

func TestDeleteSeries(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	series := "Cambridge Texts in the History of Political Thought"
	serId, err := seriesId(db, series)
	if err != nil {
		t.Errorf("seriesId returned unexpected error: %v", err)
	}

	err = deleteSeries(db, serId)
	if err != nil {
		t.Errorf("deleteSeries gave unexpected error: %v", err)
	}

	name, err := seriesName(db, serId)
	if err == nil {
		t.Errorf(
			"Series still in DB after deletion, seriesName returned no error",
		)
	} else {
		var invlSerIdErr *InvalidSeriesIdError
		if !errors.As(err, &invlSerIdErr) {
			t.Errorf(
				"seriesName returned unexpected error after deletion of series: %v",
				err,
			)
		}
	}
	if name != "" {
		t.Errorf(
			"seriesName returned unexpected result after "+
				"deletion of series: %v",
			name,
		)
	}
}

func TestDeleteSeriesInvalidId(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	serId := 5

	err = deleteSeries(db, serId)
	if err == nil {
		t.Errorf("deleteSeries did not return error for invalid ID #%v", serId)
	} else {
		var invlSerIdErr *InvalidSeriesIdError
		if !errors.As(err, &invlSerIdErr) {
			t.Errorf(
				"deleteSeries returned unexpected error for invalid ID #%v: %v",
				serId,
				err,
			)
		}
	}
}

func TestDeleteSeriesInUse(t *testing.T) {
	db, err := sql.Open("sqlite3", "testdb.sqlite")
	if err != nil {
		t.Errorf("Problem opening database: %v", err)
	}
	defer db.Close()

	series := "Spectrum Multiview Books"
	serId, err := seriesId(db, series)
	if err != nil {
		t.Errorf("Problem getting ID for %v: %v", series, err)
	}

	err = deleteSeries(db, serId)
	if err == nil {
		t.Errorf(
			"deleteSeries did not return error for in use series #%v %v",
			serId,
			series,
		)
	} else {
		var serInUseErr *SeriesInUseError
		if !errors.As(err, &serInUseErr) {
			t.Errorf(
				"deleteSeries returned unexpected error for in use series #%v %v: %v",
				serId,
				series,
				err,
			)
		}
	}
}
