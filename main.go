package main

import (
	"log"

	"github.com/go-sql-driver/mysql"
)

func main() {
	cfg := mysql.Config{
		User:                 "root",
		Passwd:               "password",
		Net:                  "tcp",
		Addr:                 "localhost:3306",
		DBName:               "highlights",
		AllowNativePasswords: true,
		ParseTime:            true,
	}

	storage := NewMySQLStorage(cfg)

	db, err := storage.Init()
	if err != nil {
		log.Fatal(err)
	}
	apiServer := NewAPIServer("localhost:3000", db)
	apiServer.Run()
}
