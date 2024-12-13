package main

import (
	"bufio"
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/lib/pq"
	"log"
	"os"
)

type config struct {
	host      string
	port      string
	user      string
	password  string
	database  string
	driver    string
	directory string
}

// NewConfig creates a new config struct from environment variables
func newConfig() *config {
	return &config{
		host:      os.Getenv("DB_HOST"),       // ex: localhost
		port:      os.Getenv("DB_PORT"),       // ex: 5432
		user:      os.Getenv("DB_USER"),       // ex: postgres
		password:  os.Getenv("DB_PASSWORD"),   // ex: password
		database:  os.Getenv("DB_NAME"),       // ex: mydb
		driver:    os.Getenv("DB_DRIVER"),     // ex: postgres
		directory: os.Getenv("MIGRATION_DIR"), // ex: ./migrations
		//table:     os.Getenv("MIGRATION_TABLE"), // ex: _migration
	}
}

// Main function to run migrations
func main() {
	ctx := context.Background()
	config := newConfig()
	db := connectDB(config)

	migrations, err := getMigrationsSQL(config.directory)

	if err != nil {
		log.Fatal(err)
	}

	if len(migrations) == 0 {
		return
	}

	transaction, err := db.BeginTx(ctx, nil)

	if err != nil {
		log.Fatal(err)
	}

	for _, migration := range migrations {
		_, err := transaction.Exec(migration)

		if err != nil {
			err := transaction.Rollback()
			log.Fatal(err)
			return
		}
	}

}

// connectDB connects to the database
func connectDB(config *config) *sql.DB {
	var err error
	var connector driver.Connector

	if config.driver == "postgres" {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", config.host, config.port, config.user, config.password, config.database)
		connector, err = pq.NewConnector(dsn)
	}

	if err != nil {
		panic(err)
	}

	db := sql.OpenDB(connector)

	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	return db
}

// createMigrationsTable creates the migrations table
func createMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
    CREATE TABLE IF NOT EXISTS _migration (
      id SERIAL PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      batch INTEGER NOT NULL,
      applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
    );
    `)

	return err
}

// getCompletedMigrations returns a list of completed migrations from the database table
func getCompletedMigrations(db *sql.DB) ([]string, error) {
	return nil, nil
}

// getMigrationsSQL reads the SQL files from the directory and returns a list of migrations as strings
func getMigrationsSQL(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	var migrations []string

	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		var migration string

		file, err := os.Open(directory + "/" + entry.Name())

		if err != nil {
			return nil, err
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			migration += scanner.Text()
		}

		if err := scanner.Err(); err != nil {
			return nil, err
		}

		migrations = append(migrations, migration)

		file.Close()
	}

	return migrations, nil
}
