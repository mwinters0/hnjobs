package db

import _ "embed"

//go:embed schema.sql
var newDBSchema string

func NewDB(filepath string) error {
	err := OpenDB(filepath)
	if err != nil {
		return err
	}
	_, err = store.db.Exec(newDBSchema)
	if err != nil {
		return err
	}
	return nil
}

// TODO schema migrations
