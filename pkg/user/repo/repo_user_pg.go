package repo

import (
	"database/sql"
	"redditclone/pkg/hash"
	"redditclone/pkg/user"
)

type UserRepo struct {
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{DB: db}
}

func (repo *UserRepo) CheckUser(login, pass string) (*user.User, error) {
	u := &user.User{}
	row := repo.DB.QueryRow("SELECT id, login, password, created_at FROM users WHERE login = $1", login)
	err := row.Scan(&u.ID, &u.Login, &u.Password, &u.Created)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}

	hashed := u.Password
	if !hash.CheckPassHash(pass, hashed) {
		return nil, user.ErrBadPass
	}

	return u, nil
}

func (repo *UserRepo) Create(pass, login string) (int, error) {
	var lastID int
	err := repo.DB.QueryRow(`INSERT INTO users ("login", "password")
	VALUES ($1, $2) ON CONFLICT (login) DO NOTHING RETURNING id;`, login, pass).Scan(&lastID)
	if err != nil {
		if err == sql.ErrNoRows {
			return -1, user.ErrUserExist
		} else {
			return -1, err
		}
	}
	return lastID, nil
}
