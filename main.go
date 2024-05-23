package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "098098"
	dbname   = "exampledb"
)

type Data struct {
	ID      int
	Name    string
	Product int
}

func findByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		log.Fatalf("Error to get id: %v\n", err)
	}
	w.Write([]byte(dataBase(id)))
}

func ensureTableExists(ctx context.Context, conn *pgx.Conn) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS yourtable (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL,
			product INT NOT NULL
		);
	`

	_, err := conn.Exec(ctx, createTableQuery)
	if err != nil {
		log.Fatalf("Failed to create table: %v\n", err)
	}

	fmt.Println("Table 'exampledb' ensured.")
	return nil
}

func fetchData(ctx context.Context, conn *pgx.Conn, query string, id int) Data {
	var data Data

	rows := conn.QueryRow(ctx, query, id)

	if err := rows.Scan(&data.ID, &data.Name, &data.Product); err != nil {
		log.Fatalf("Failed to scan row: %v\n", err)
	}

	return data
}

func dataBase(id int) string {

	connString := fmt.Sprintf("postgres://%s:%s@%s:%d/%s", user, password, host, port, dbname)

	conn, err := pgx.Connect(context.Background(), connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v\n", err)
	}
	defer conn.Close(context.Background())

	if err := conn.Ping(context.Background()); err != nil {
		log.Fatalf("Failed to ping to database: %v\n", err)
	}

	if err := ensureTableExists(context.Background(), conn); err != nil {
		log.Fatalf("Failed to create connection: %v\n", err)
	}

	query := "SELECT id, name, product FROM yourtable WHERE product = $1"

	data := fetchData(context.Background(), conn, query, id)
	fmt.Printf("Received data: ID=%d, Name=%s, Product=%d\n", data.ID, data.Name, data.Product)

	res := fmt.Sprintf("Received data: ID=%d, Name=%s, Product=%d\n", data.ID, data.Name, data.Product)

	fmt.Println("All data fetched.")

	return res
}

func main() {
	router := http.NewServeMux()

	router.HandleFunc("GET /{id}", findByID)

	server := http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Starting server on port :8080")

	server.ListenAndServe()

}
