package database

import (
	"context"
	"fmt"
	"log"

	"github.com/cevrimxe/auth-service/config"
	"github.com/jackc/pgx/v4/pgxpool"
)

var DB *pgxpool.Pool

func ConnectDB() *pgxpool.Pool {
	config.LoadEnv()

	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		config.GetEnv("DB_USER"),
		config.GetEnv("DB_PASSWORD"),
		config.GetEnv("DB_HOST"),
		config.GetEnv("DB_PORT"),
		config.GetEnv("DB_NAME"),
	)

	var err error
	DB, err = pgxpool.Connect(context.Background(), connStr)
	if err != nil {
		log.Fatalf("Could not connect to database: %v", err) // log.Fatal, panik yerine hata mesajı verir
	}
	fmt.Println("Database connection established")
	createTable(DB)

	return DB
}

func createTable(db *pgxpool.Pool) {
	createUrlTable := `
	CREATE TABLE IF NOT EXISTS users (
    	id SERIAL PRIMARY KEY,
    	email TEXT NOT NULL UNIQUE,
    	password_hash TEXT NOT NULL,
    	first_name TEXT,
    	last_name TEXT,
    	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    	is_active BOOLEAN DEFAULT TRUE,
    	email_verified BOOLEAN DEFAULT FALSE,
   		role TEXT DEFAULT 'user',
    	reset_token TEXT,
    	reset_token_expiry TIMESTAMP
);
	`

	_, err := db.Exec(context.Background(), createUrlTable)

	if err != nil {
		panic("couldnt create users table")
	}

}
