package main

import "testing"

func TestGetDBConnectorPostgresDriver(t *testing.T) {
	config := &config{
		driver:    "postgres",
		host:      "host",
		port:      5432,
		user:      "user",
		password:  "password",
		database:  "database",
		directory: "directory",
		table:     "table",
	}

	conn, err := getDBConnector(config)

	if err != nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if conn == nil {
		t.Fatalf("Expected connector, got nil")
	}
}

func TestGetDBConnectorInvalidConfig(t *testing.T) {
	config := &config{}

	conn, err := getDBConnector(config)

	if err == nil {
		t.Fatalf("Expected nil error, got %v", err)
	}

	if conn != nil {
		t.Fatalf("Expected nil connector, got %v", conn)
	}
}

func TestGetDBConnectorInvalidDriver(t *testing.T) {
	config := &config{
		driver:    "invalid",
		host:      "host",
		port:      5432,
		user:      "user",
		password:  "password",
		database:  "database",
		directory: "directory",
		table:     "table",
	}

	conn, err := getDBConnector(config)

	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	if conn != nil {
		t.Fatalf("Expected nil connector, got %v", conn)
	}
}
