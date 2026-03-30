package repo

import (
	"database/sql"
	"fmt"
	"redditclone/pkg/user"
	"regexp"
	"testing"
	"time"

	sqlmock "gopkg.in/DATA-DOG/go-sqlmock.v1"
)

func TestRepoCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock DB: %s", err)
	}
	defer db.Close()

	repo := &UserRepo{
		DB: db,
	}
	login := "test"
	password := "testpass"
	q := regexp.QuoteMeta(`INSERT INTO users ("login", "password")
	VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id;`)

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(q).WithArgs(login, password).WillReturnRows(rows)
	id, err := repo.Create(password, login)
	if err != nil {
		t.Errorf("unexpected err: %s", err)
		return
	}
	if id != 1 {
		t.Errorf("bad id: want %d, have %d", id, 1)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	//err db
	mock.ExpectQuery(q).WithArgs(login, password).WillReturnError(fmt.Errorf("err db"))
	_, err = repo.Create(password, login)
	if err == nil {
		t.Errorf("expected error, got nil")
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}

	//err user exist
	mock.ExpectQuery(q).WithArgs(login, password).WillReturnError(sql.ErrNoRows)
	id, err = repo.Create(password, login)
	if err != user.ErrUserExist {
		t.Errorf("expected user alredy exist, got %v", err)
		return
	}
	if id != -1 {
		t.Errorf("expected -1 id on err, got %d", id)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}

func TestRepoCheckUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock DB: %s", err)
	}
	defer db.Close()

	repo := &UserRepo{
		DB: db,
	}
	login := "test"
	password := "testpass"
	created := time.Now()
	q := regexp.QuoteMeta(`SELECT * FROM users WHERE login = $1`)
	rows := sqlmock.NewRows([]string{"id", "login", "password", "created"}).
		AddRow(1, login, password, created)
	mock.ExpectQuery(q).WithArgs(login).WillReturnRows(rows)

	_, err = repo.CheckUser(login, password)
	// err hashed pass
	if err != user.ErrBadPass {
		t.Errorf("expected bad pass err, got: %s", err)
		return
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("there were unfulfilled expectations: %s", err)
		return
	}
}
