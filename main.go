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
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
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
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var err error
	server, err = initDB(os.Getenv("MYSQL_URI"))
	if err != nil {
		log.Fatalf("Can't set up DB: %v", err)
	}

	// Versioning is useful in case we get fancier later, like actually having accounts
	http.HandleFunc("/v1/get", get)
	http.HandleFunc("/v1/create", create)
	// TODO List

	log.Printf("Serving on port %v", port)
	http.ListenAndServe(":"+port, nil)
}

func initDB(mysqlURI string) (*serverType, error) {
	if mysqlURI == "" {
		log.Fatalf("You must set the MYSQL_URI env variable")
	}

	log.Printf("Connecting to %v", mysqlURI)
	db, err := sql.Open("mysql", mysqlURI)
	if err != nil {
		return nil, err
	}

	// Create the table, if needed
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
	// TODO Actually be careful with CORS
	resp.Header().Set("Access-Control-Allow-Origin", "*")

	values, ok := req.URL.Query()["id"]
	if !ok || len(values[0]) < 1 {
		http.Error(resp, "missing ID param", http.StatusBadRequest)
		return
	}
	id := values[0]
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
	resp.Header().Set("Access-Control-Allow-Origin", "*")

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
	log.Printf("Uploaded new proposal for %v: %v", mapName, id)

	// Return the UUID, so the user can share their masterpiece
	fmt.Fprintf(resp, "%v", id)
}

// Verifies the input is MapEdits JSON. Returns the raw JSON and extracts the map name.
func validateJSON(input io.Reader) ([]byte, string, error) {
	var buffer bytes.Buffer
	_, err := buffer.ReadFrom(input)
	if err != nil {
		return nil, "", err
	}
	rawJson := buffer.Bytes()

	// Extract the map name from the JSON
	var edits mapEdits
	if err := json.Unmarshal(rawJson, &edits); err != nil {
		return nil, "", err
	}

	return rawJson, edits.Name.path(), nil
}

type mapEdits struct {
	Name mapName `json:"map_name"`
	// We're not going to validate the other fields
}

type mapName struct {
	City    cityName
	MapName string `json:"map"`
}

type cityName struct {
	Country string
	City    string
}

func (n *mapName) path() string {
	return fmt.Sprintf("%v/%v/%v", n.City.Country, n.City.City, n.MapName)
}
