package main

import (
	"database/sql"
	_ "embed"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

var (
	//go:embed proposals.sql
	createTable string

	server *serverType
)

type serverType struct {
	db         *sql.DB
	getStmt    *sql.Stmt
	createStmt *sql.Stmt
}

func main() {
	var err error
	server, err = initDB()
	if err != nil {
		log.Fatalf("Can't set up DB: %v", err)
	}

	http.HandleFunc("/get", get)
	http.HandleFunc("/create", create)
	// TODO List

	log.Print("Serving on localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func initDB() (*serverType, error) {
	db, err := sql.Open("sqlite3", "dev_proposals.db")
	if err != nil {
		return nil, err
	}

	// Create the table, if needed
	log.Print("Setting up DB")
	_, err = db.Exec(createTable)
	if err != nil {
		return nil, err
	}

	// Prepare statements guard against SQL injection
	getStmt, err := db.Prepare("SELECT json FROM proposals WHERE id = ?")
	if err != nil {
		return nil, err
	}

	createStmt, err := db.Prepare("INSERT INTO proposals (id, map_name, json, moderated, time) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		return nil, err
	}

	return &serverType{db, getStmt, createStmt}, nil
}

func get(resp http.ResponseWriter, req *http.Request) {
	// TODO Fail gracefully if the GET param is missing/blank
	id := req.URL.Query().Get("id")[0]
	var json []byte
	if err := server.getStmt.QueryRow(id).Scan(&json); err != nil {
		// TODO Also log errors
		http.Error(resp, err.Error(), http.StatusNotFound)
		return
	}
	// TODO Check error
	resp.Write(json)
}

func create(resp http.ResponseWriter, req *http.Request) {
}
