package database

import "database/sql"

func GetToken(db *sql.DB, tokenType string) (string, error) {
	var token string
	sql := "SELECT token FROM tokens WHERE id = ?"
	row := db.QueryRow(sql, tokenType).Scan(&token)
	return token, row
}

func SetToken(db *sql.DB, tokenType string, token string) error {
	sql := "INSERT INTO tokens (id, token) VALUES (?, ?) ON CONFLICT (id) DO UPDATE SET token = ?"
	_, err := db.Exec(sql, tokenType, token, token)
	return err
}
