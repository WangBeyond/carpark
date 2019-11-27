package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"
)

type Server struct {
	db    *sql.DB
	cache atomic.Value
}

func main() {
	server := &Server{}
	server.db = dbConn()

	http.HandleFunc("/carparks/update", server.UpdateAvailability)
	http.HandleFunc("/carparks/nearest", server.GetNearest)
	http.ListenAndServe(":8080", nil)
}

func dbConn() (db *sql.DB) {
	dbDriver := "mysql"
	dbUser := os.Getenv("MYSQL_USER")
	dbPass := os.Getenv("MYSQL_PASSWORD")
	dbName := os.Getenv("MYSQL_DB")
	dbHost := os.Getenv("MYSQL_HOST")
	db, err := sql.Open(dbDriver, dbUser+":"+dbPass+"@"+dbHost+"/"+dbName)
	if err != nil {
		panic(err.Error())
	}
	log.Println("connected to ", dbUser+":"+dbPass+"@"+dbHost+"/"+dbName)
	return db
}
