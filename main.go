package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

func main() {
	config := mysql.NewConfig()
	config.User = os.Getenv("MYSQL_USER")
	config.Passwd = os.Getenv("MYSQL_PWD")
	config.Net = "unix"
	config.Addr = os.Getenv("MYSQL_SOCKET")
	config.DBName = os.Getenv("MYSQL_DATABASE")
	config.Timeout = 5 * time.Second
	config.AllowCleartextPasswords = true

	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		setHeaders(w)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	mux.HandleFunc("/db-health", func(w http.ResponseWriter, r *http.Request) {
		setHeaders(w)

		if r.Method != http.MethodGet {
			w.Header().Set("Allow", http.MethodGet)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		var result int
		err := db.QueryRowContext(ctx, "SELECT 1").Scan(&result)
		if err != nil {
			log.Printf("DB connection failed: %v", err)
			w.WriteHeader(http.StatusServiceUnavailable)
			json.NewEncoder(w).Encode(map[string]string{
				"status":  "error",
				"message": "DBに接続できませんでした",
			})
			return
		}

		json.NewEncoder(w).Encode(map[string]any{
			"status": "ok",
			"result": result,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Listening on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}

func setHeaders(w http.ResponseWriter) {
	w.Header().Set(
		"Access-Control-Allow-Origin",
		"https://hackathon-frontend-seven-beta.vercel.app",
	)
	w.Header().Set("Content-Type", "application/json")
}
