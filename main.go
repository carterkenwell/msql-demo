package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()

	router.HandleFunc("/items", getAllItems).Methods("GET")

	router.HandleFunc("/items", createItem).Methods("POST")

	log.Fatal(http.ListenAndServe(":8000", router))
}

type Item struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Quantity int    `json:"quantity"`
}

func getAllItems(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "carter:password@tcp(localhost:3306)/consumables")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM main")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	items := []Item{}

	for rows.Next() {
		var item Item
		err := rows.Scan(&item.ID, &item.Name, &item.Quantity)
		if err != nil {
			log.Fatal(err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		log.Fatal(err)
	}

	// generate da response
	jsonItems, err := json.Marshal(items)
	if err != nil {
		log.Fatal(err)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonItems)
}

func createItem(w http.ResponseWriter, r *http.Request) {
	var item Item
	err := json.NewDecoder(r.Body).Decode(&item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// connect to db
	db, err := sql.Open("mysql", "carter:password@tcp(localhost:3306)/consumables")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	// put http request into db as item
	insertQuery := "INSERT INTO main (name, quantity) VALUES (?, ?)"
	result, err := db.Exec(insertQuery, item.Name, item.Quantity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// get item id
	insertedID, err := result.LastInsertId()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// confirm item is inserted by returning item id
	item.ID = int(insertedID)
	jsonItem, err := json.Marshal(item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// some json shit chatGPT told me to put in here idk
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonItem)
}

// curl command for sending json request to server
// curl -X POST -H "Content-Type: application/json" -d "{\"name\": \"ethernet cable\", \"quantity\": 15}" http://localhost:8000/items
