package models

import (
	"database/sql"

	"github.com/yimm/rfid-api/config"
)

type User struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"-"`
	IsActive bool   `json:"-"`
	Role     string `json:"role"`
}

func GetUserByEmail(email string) (*User, error) {
	user := &User{}
	query := `SELECT id, name, username, email, password, is_active FROM users WHERE email = ?`
	err := config.DB.QueryRow(query, email).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByUsername(username string) (*User, error) {
	user := &User{}
	query := `SELECT id, name, username, email, password, is_active FROM users WHERE username = ?`
	err := config.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserByID(id int64) (*User, error) {
	user := &User{}
	query := `SELECT id, name, username, email, password, is_active FROM users WHERE id = ?`
	err := config.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Name, &user.Username, &user.Email, &user.Password, &user.IsActive,
	)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func GetUserRoleSimple(userID int64) string {
	var role string
	query := `
		SELECT r.name
		FROM model_has_roles mhr
		JOIN roles r ON r.id = mhr.role_id
		WHERE mhr.model_type = 'App\\Models\\User' AND mhr.model_id = ?
		LIMIT 1
	`
	err := config.DB.QueryRow(query, userID).Scan(&role)
	if err == sql.ErrNoRows {
		return "petugas"
	}
	if err != nil {
		return "petugas"
	}
	return role
}