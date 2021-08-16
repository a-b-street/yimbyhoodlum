package main

import (
	"bytes"
	"database/sql"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
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
	rawJSON, mapName, err := validateJSON(req.Body)
	if err != nil {
		// TODO Also log errors
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	// In the unlikely event of UUID collision, should the retry loop live here?
	id := uuid.New().String()
	moderated := 0
	time := time.Now().Unix()
	if _, err := server.createStmt.Exec(id, mapName, rawJSON, moderated, time); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the UUID, so the user can share their masterpiece
	fmt.Println(resp, "%v", id)
}

// Verifies the input is MapEdits JSON. Returns the raw JSON and extracts the map name.
func validateJSON(input io.Reader) ([]byte, string, error) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(input)
	if err != nil {
		return nil, "", err
	}

	var edits mapEdits
	if err := json.NewDecoder(&buffer).Decode(&edits); err != nil {
		return nil, "", err
	}

	return buffer.Bytes(), edits.name.path(), nil
}

type mapEdits struct {
	name mapName `json:"map_name"`
	// We're not going to validate the other fields
}

type mapName struct {
	city    cityName
	mapName string `json:"map"`
}

type cityName struct {
	country string
	city    string
}

func (n *mapName) path() string {
	return fmt.Sprintf("%v/%v/%v", n.city.country, n.city.city, n.mapName)
}
