package main

import (
	"bytes"
	"cloud.google.com/go/storage"
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var client *storage.Client

const bucket = "aorta-routes.appspot.com"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	var err error
	// This magically pulls credentials from AppEngine
	client, err = storage.NewClient(context.Background())
	if err != nil {
		log.Fatalf("Can't set up GCS client: %v", err)
	}

	// Versioning is useful in case we get fancier later, like actually having accounts
	http.HandleFunc("/v1/get", get)
	http.HandleFunc("/v1/create", create)
	// TODO List

	log.Printf("Serving on port %v", port)
	http.ListenAndServe(":"+port, nil)
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

	// TODO Do we need to sanitize the input?
	obj := client.Bucket(bucket).Object(fmt.Sprintf("proposals/%v", id))
	r, err := obj.NewReader(context.Background())
	if err != nil {
		http.Error(resp, fmt.Sprintf("new reader failed: %v", err), http.StatusInternalServerError)
		return
	}
	defer r.Close()
	if _, err := io.Copy(resp, r); err != nil {
		http.Error(resp, fmt.Sprintf("reading failed: %v", err), http.StatusInternalServerError)
		return
	}
}

func create(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Access-Control-Allow-Origin", "*")

	rawJSON, mapName, err := validateJSON(req.Body)
	if err != nil {
		log.Printf("Invalid JSON in create call: %v", err)
		http.Error(resp, err.Error(), http.StatusBadRequest)
		return
	}

	checksum := fmt.Sprintf("%x", md5.Sum(rawJSON))
	// If the proposal already exists, overwriting it with the same thing will be idempotent.
	obj := client.Bucket(bucket).Object(fmt.Sprintf("proposals/%v", checksum))
	w := obj.NewWriter(context.Background())
	if _, err := w.Write(rawJSON); err != nil {
		http.Error(resp, fmt.Sprintf("writing failed: %v", err), http.StatusInternalServerError)
		return
	}
	if err := w.Close(); err != nil {
		http.Error(resp, fmt.Sprintf("closing after write failed: %v", err), http.StatusInternalServerError)
		return
	}
	log.Printf("Uploaded new proposal for %v: %v", mapName, checksum)

	// Return the checksum, so the user can share their masterpiece. (They
	// can calculate it anyway with md5sum, but still useful.)
	fmt.Fprintf(resp, "%v", checksum)
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
