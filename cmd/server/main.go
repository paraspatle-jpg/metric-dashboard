package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"metric-collector/internal/db"
	"metric-collector/internal/models"
	"metric-collector/internal/templates"
)

func main() {
	store, err := db.New("/data/metrics.db")
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer store.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/metrics", func(w http.ResponseWriter, r *http.Request) {
		var m models.Metrics
		if err := json.NewDecoder(r.Body).Decode(&m); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		if err := store.Insert(m); err != nil {
			log.Printf("insert error: %v", err)
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("GET /api/metrics", func(w http.ResponseWriter, r *http.Request) {
		minutes := 60
		if q := r.URL.Query().Get("minutes"); q != "" {
			if v, err := strconv.Atoi(q); err == nil && v > 0 {
				minutes = v
			}
		}
		metrics, err := store.Query(minutes)
		if err != nil {
			log.Printf("query error: %v", err)
			http.Error(w, "query error", http.StatusInternalServerError)
			return
		}
		if metrics == nil {
			metrics = []models.Metrics{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(metrics)
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		data, err := templates.FS.ReadFile("dashboard.html")
		if err != nil {
			http.Error(w, "template not found", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	log.Println("Dashboard server listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
