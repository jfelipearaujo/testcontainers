package main

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Product struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

var connStr string = os.Getenv("DATABASE_URL")

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /products", createProduct)
	mux.HandleFunc("GET /products/{id}", getProduct)

	slog.Info("server running on port 8080")
	http.ListenAndServe(":8080", mux)
}

func createProduct(w http.ResponseWriter, r *http.Request) {
	var product Product
	err := json.NewDecoder(r.Body).Decode(&product)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		slog.Error("failed to decode the request body", "error", err)
		return
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to open the database connection", "error", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to ping the database", "conn_str", connStr, "error", err)
		return
	}

	_, err = db.Exec("INSERT INTO products (id, name) VALUES ($1, $2)", product.Id, product.Name)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to create the product", "error", err)
		return
	}

	w.WriteHeader(http.StatusCreated)

	json, err := json.Marshal(product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to marshal the response body", "error", err)
		return
	}

	w.Write(json)
}

func getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to open the database connection", "conn_str", connStr, "error", err)
		return
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to ping the database", "error", err)
		return
	}

	row := db.QueryRow("SELECT id, name FROM products WHERE id = $1", id)
	var product Product
	err = row.Scan(&product.Id, &product.Name)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		slog.Error("failed to get the product", "error", err)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json, err := json.Marshal(product)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to marshal the response body", "error", err)
		return
	}

	w.Write(json)
}
