package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Product struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		product := Product{
			Id:   "1",
			Name: "Test Product",
		}

		w.WriteHeader(http.StatusOK)

		json, err := json.Marshal(product)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Write(json)
	})

	fmt.Println("server running on port 8080")
	http.ListenAndServe(":8080", mux)
}
