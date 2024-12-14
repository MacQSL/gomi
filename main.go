package main

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"flag"
	"fmt"
	"github.com/lib/pq"
	"log"
	"os"
	"path/filepath"
	"strconv"
)

type migration struct {
	name string
	sql  string
}

type config struct {
	host      string
	port      int
	user      string
	password  string
	database  string
	driver    string
	directory string
	table     string
}

// Main function to run pending migrations
func main() {
	config := ParseFlags()

	// Connect to the database using the connector
	log.Println("Phase 1: Connecting to the database...")

	db, err := ConnectDB(config)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Phase 1: Database connected.")

	// Get the new migrations from the directory
	log.Println("Phase 2: Getting new migrations...")

	migrations, err := GetNewMigrations(db, config.table, config.directory)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Phase 2: Found", len(migrations), "new migration(s).")

	// Run the migrations against the database
	log.Println("Phase 3: Running migrations...")

	err = RunMigrations(db, config.table, migrations)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Phase 3: Migrations complete.")
}

// Parse the command line flags and return a config struct
func ParseFlags() *config {
	hostPtr := flag.String("host", "localhost", "Database host")
	portPtr := flag.Int("port", 5432, "Database port")
	userPtr := flag.String("user", "gomi", "Database username")
	passwordPtr := flag.String("password", "gomi", "Database password")
	databasePtr := flag.String("database", "gomi", "Database name")
	driverPtr := flag.String("driver", "postgres", "Database SQL driver")
	directoryPtr := flag.String("directory", "./migrations", "Directory containing migration files")
	tablePtr := flag.String("table", "_migration", "Table to store migration history")

	flag.Parse()

	return &config{
		host:      *hostPtr,
		port:      *portPtr,
		user:      *userPtr,
		password:  *passwordPtr,
		database:  *databasePtr,
		driver:    *driverPtr,
		directory: *directoryPtr,
		table:     *tablePtr,
	}
}

// Get a database connector for the given config
func GetDBConnector(config *config) (driver.Connector, error) {
	if config.driver == "postgres" {
		dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.host, strconv.Itoa(config.port), config.user, config.password, config.database)
		return pq.NewConnector(dsn)
	}

	return nil, fmt.Errorf("Error database driver not supported: '%s'", config.driver)
}

// Connect to the database using the given connector
func ConnectDB(config *config) (*sql.DB, error) {
	connector, err := GetDBConnector(config)

	if err != nil {
		log.Fatal(err)
	}

	db := sql.OpenDB(connector)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("Error pinging database: %s", err)
	}

	return db, nil
}

// Run the migrations against the database and record them in the tracking table
func RunMigrations(db *sql.DB, table string, migrations []migration) error {
	transaction, err := db.BeginTx(context.TODO(), nil)

	if err != nil {
		return fmt.Errorf("Error starting transaction: %s", err)
	}

	// Apply each migration and record it in the tracking table
	for _, migration := range migrations {
		log.Println("Running migration:", migration.name)

		// Execute the migration SQL
		_, err := transaction.Exec(migration.sql)

		if err != nil {
			txErr := transaction.Rollback()
			return fmt.Errorf("Error executing migration: %s", errors.Join(err, txErr))
		}

		// Record the migration in the tracking table
		_, err = transaction.Exec(fmt.Sprintf(`INSERT INTO public.%s (name) VALUES ($1);`, table), migration.name)

		if err != nil {
			txErr := transaction.Rollback()
			return fmt.Errorf("Error inserting migration record: %s", errors.Join(err, txErr))
		}
	}

	return transaction.Commit()
}

func GetAppliedMigrations(db *sql.DB, table string) (map[string]bool, error) {
	_, err := db.Exec(fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS public.%s (
      id SERIAL PRIMARY KEY,
      name VARCHAR(255) NOT NULL,
      applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL);
    `, table))

	if err != nil {
		return nil, fmt.Errorf("Error creating migration tracking table: %s", err)
	}

	rows, err := db.Query(fmt.Sprintf(`SELECT name FROM public.%s;`, table))

	if err != nil {
		return nil, fmt.Errorf("Error getting applied migrations: %s", err)
	}

	defer rows.Close()

	appliedMigrations := make(map[string]bool)
	for rows.Next() {
		var name string
		err := rows.Scan(&name)

		if err != nil {
			return nil, fmt.Errorf("Error scanning applied migration: %s", err)
		}

		appliedMigrations[name] = true
	}

	return appliedMigrations, nil
}

// Reads SQL migration files from a directory and returns a list of non-applied migrations
func GetNewMigrations(db *sql.DB, table string, directory string) ([]migration, error) {
	entries, err := os.ReadDir(directory)

	if err != nil {
		return nil, fmt.Errorf("Error reading directory: %s", err)
	}

	appliedMigrations, err := GetAppliedMigrations(db, table)

	if err != nil {
		return nil, err
	}

	var migrations []migration
	for _, entry := range entries {
		if appliedMigrations[entry.Name()] {
			continue // Skip previously applied migrations
		}

		if entry.IsDir() {
			continue // Skip directories
		}

		filePath := filepath.Join(directory, entry.Name())
		content, err := os.ReadFile(filePath)

		if err != nil {
			return nil, fmt.Errorf("Error reading file: '%s': %w", filePath, err)
		}

		migrations = append(migrations, migration{name: entry.Name(), sql: string(content)})
	}

	return migrations, nil
}
