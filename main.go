package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "098098"
	dbname   = "exampledb"
)

type Data struct {
	ID   int
	Name string
}

func ensureTableExists(ctx context.Context, pool *pgxpool.Pool) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS yourtable (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			product INT NOT NULL
		);
	`

	_, err := pool.Exec(ctx, createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	fmt.Println("Table 'exampledb' ensured.")
	return nil
}

func fetchData(ctx context.Context, pool *pgxpool.Pool, query string, wg *sync.WaitGroup, ch chan<- Data) {
	defer wg.Done()

	rows, err := pool.Query(ctx, query)
	if err != nil {
		log.Fatalf("Failed to execute query: %v\n", err)
	}
	defer rows.Close()

	for rows.Next() {
		var data Data
		if err := rows.Scan(&data.ID, &data.Name); err != nil {
			log.Fatalf("Failed to scan row: %v\n", err)
		}
		ch <- data
	}
	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v\n", err)
	}
}

func main() {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)

	pool, err := pgxpool.New(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer pool.Close()

	if err := pool.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping to database: %v\n", err)
	}

	if err := ensureTableExists(context.Background(), pool); err != nil {
		log.Fatalf("Failed to create connection pool: %v\n", err)
	}

	ch := make(chan Data)
	var wg sync.WaitGroup

	queries := []string{
		"SELECT id, name FROM yourtable WHERE product = 1",
		"SELECT id, name FROM yourtable WHERE product = 2",
		"SELECT id, name FROM yourtable WHERE product = 3",
	}

	for _, query := range queries {
		wg.Add(1)
		go fetchData(context.Background(), pool, query, &wg, ch)
	}

	go func() {
		wg.Wait()
		close(ch)
	}()

	for data := range ch {
		fmt.Printf("Received data: ID=%d, Name=%s\n", data.ID, data.Name)
	}

	fmt.Println("All data fetched.")
}
