package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type HealthResponse struct {
	Status  string        `json:"status"`
	Error   string        `json:"error,omitempty"`
	Results []interface{} `json:"results,omitempty"`
}

var db *sql.DB

func main() {
	dsn := os.Getenv("PG_CONNECTION_STRING")
	if dsn == "" {
		// Build DSN from component environment variables if provided
		user := os.Getenv("POSTGRES_USER")
		if user == "" {
			user = "postgres"
		}
		pass := os.Getenv("POSTGRES_PASSWORD")
		dbname := os.Getenv("POSTGRES_DB")
		if dbname == "" {
			dbname = "postgres"
		}
		host := os.Getenv("POSTGRES_HOST")
		if host == "" {
			host = "localhost"
		}
		port := os.Getenv("POSTGRES_PORT")
		if port == "" {
			port = "5432"
		}
		sslmode := os.Getenv("POSTGRES_SSLMODE")
		if sslmode == "" {
			sslmode = "disable"
		}

		u := &url.URL{Scheme: "postgres"}
		if pass != "" {
			u.User = url.UserPassword(user, pass)
		} else {
			u.User = url.User(user)
		}
		u.Host = net.JoinHostPort(host, port)
		u.Path = "/" + dbname
		u.RawQuery = "sslmode=" + sslmode
		dsn = u.String()
	}
	log.Printf("Built DSN: %s", dsn)

	tablesToCheck := os.Getenv("PG_HEALTHY_TABLES")
	var tables []string
	if tablesToCheck != "" {
		tables = strings.Split(tablesToCheck, ",")
	}

	exposeQueryResultsValue := os.Getenv("PG_HEALTHY_EXPOSE_QUERY_RESULTS")
	exposeQueryResults := exposeQueryResultsValue == "true" || exposeQueryResultsValue == "1"
	customQueries := os.Getenv("PG_HEALTHY_QUERIES")
	var queries []string
	if customQueries != "" {
		queries = strings.Split(customQueries, ";;")
	}

	var err error
	db, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("DB open failed: %v", err)
	}
	defer db.Close()

	db.SetConnMaxLifetime(2 * time.Minute)
	db.SetMaxOpenConns(2)
	db.SetMaxIdleConns(1)

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		ctx := r.Context()

		if err := db.PingContext(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			_ = json.NewEncoder(w).Encode(HealthResponse{Status: "unhealthy", Error: err.Error()})
			log.Printf("Health check failed (Ping): %v", err)
			return
		}

		// Optional deeper check: verify a table exists or can be queried
		for _, table := range tables {
			err := checkTableExists(ctx, table)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(HealthResponse{Status: "unhealthy", Error: err.Error()})
				log.Printf("Health check failed (Table check): %v", err)
				return
			}
		}

		var results []interface{}
		for _, query := range queries {
			result, err := checkQuery(ctx, query)
			if err != nil {
				w.WriteHeader(http.StatusServiceUnavailable)
				_ = json.NewEncoder(w).Encode(HealthResponse{Status: "unhealthy", Error: err.Error()})
				log.Printf("Health check failed (Query): %v", err)
				return
			}
			results = append(results, result)
		}

		w.WriteHeader(http.StatusOK)
		response := HealthResponse{Status: "ok"}
		if exposeQueryResults {
			response.Results = results
		}
		_ = json.NewEncoder(w).Encode(response)
		log.Printf("Health check OK")
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "2345"
	}

	log.Printf("Listening on :%s ...", port)
	if exposeQueryResults {
		log.Printf("Query results exposed")
	}
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func checkTableExists(ctx context.Context, table string) error {
	stmt, err := db.Prepare("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = $1")
	if err != nil {
		return fmt.Errorf("preparing table check failed: %w", err)
	}
	defer stmt.Close()

	var count int
	err = stmt.QueryRowContext(ctx, table).Scan(&count)
	if err != nil {
		return fmt.Errorf("table check failed: %v", err)
	}
	if count == 0 {
		return fmt.Errorf("table '%s' not found", table)
	}
	return nil
}

func checkQuery(ctx context.Context, query string) (interface{}, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return nil, fmt.Errorf("empty query")
	}

	var result interface{}
	err := db.QueryRowContext(ctx, query).Scan(&result)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	return result, nil
}
