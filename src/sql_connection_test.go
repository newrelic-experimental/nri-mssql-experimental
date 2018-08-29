package main

import (
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
)

// createMockSQL creates a Test SQLConnection. Must Close con when done
func createMockSQL(t *testing.T) (con *SQLConnection, mock sqlmock.Sqlmock) {
	mockDB, mock, err := sqlmock.New()
	if err != nil {
		t.Errorf("Unexpected error while mocking: %s", err.Error())
		t.FailNow()
	}

	con = &SQLConnection{
		connection: sqlx.NewDb(mockDB, "sqlmock"),
	}

	return
}

func Test_SQLConnection_Close(t *testing.T) {
	conn, mock := createMockSQL(t)

	mock.ExpectClose().WillReturnError(errors.New("error"))
	conn.Close()

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("close expectation was not met: %s", err.Error())
	}
}

func Test_SQLConnection_Query(t *testing.T) {
	conn, mock := createMockSQL(t)
	defer conn.Close()

	// Temp data structure to store data into
	temp := []struct {
		One int `db:"one"`
		Two int `db:"two"`
	}{}

	// dummy query to run
	query := "select one, two from everywhere"

	rows := sqlmock.NewRows([]string{"one", "two"}).AddRow(1, 2)
	mock.ExpectQuery(query).WillReturnRows(rows)

	if err := conn.Query(&temp, query); err != nil {
		t.Errorf("Unexpected error: %s", err.Error())
		t.FailNow()
	}

	if length := len(temp); length != 1 {
		t.Errorf("Expected 1 element got %d", length)
		t.FailNow()
	}

	if temp[0].One != 1 || temp[0].Two != 2 {
		t.Error("Query did not marshal correctly")
	}
}

func Test_createConnectionURL(t *testing.T) {
	testCases := []struct {
		name string
		arg  *argumentList
		want string
	}{
		{
			"Port No SSL",
			&argumentList{
				Username:  "user",
				Password:  "pass",
				Hostname:  "localhost",
				EnableSSL: false,
				Port:      "1443",
				Timeout:   "30",
			},
			"sqlserver://user:pass@localhost:1443?dial+timeout=30",
		},
		{
			"Instance No SSL",
			&argumentList{
				Username:  "user",
				Password:  "pass",
				Hostname:  "localhost",
				EnableSSL: false,
				Instance:  "SQLExpress",
				Timeout:   "30",
			},
			"sqlserver://user:pass@localhost/SQLExpress?dial+timeout=30",
		},
		{
			"Instance SSL Trust",
			&argumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: true,
				Instance:               "SQLExpress",
				Timeout:                "30",
			},
			"sqlserver://user:pass@localhost/SQLExpress?TrustServerCertificate=true&dial+timeout=30&encrypt=true",
		},
		{
			"Instance SSL Certificate",
			&argumentList{
				Username:               "user",
				Password:               "pass",
				Hostname:               "localhost",
				EnableSSL:              true,
				TrustServerCertificate: false,
				CertificateLocation:    "file.ca",
				Instance:               "SQLExpress",
				Timeout:                "30",
			},
			"sqlserver://user:pass@localhost/SQLExpress?TrustServerCertificate=false&certificate=file.ca&dial+timeout=30&encrypt=true",
		},
	}

	for _, tc := range testCases {
		args = *tc.arg
		if out := createConnectionURL(); out != tc.want {
			t.Errorf("Test Case %s Failed: Expected '%s' got '%s'", tc.name, tc.want, out)
		}
	}
}
