package database

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
)

func InitDB() (*pgx.Conn, error) {
	config, err := pgx.ParseConfig(os.Getenv("DB_CONFIG"))
	if err != nil {
		fmt.Println("Error creating configuration: ", err)
		return nil, err
	}

	var connection *pgx.Conn

	for i := 0; i < 10; i++ {
		conn, err := connectDB(config)
		if err != nil {
			fmt.Println("Failed establishing connection to Postgres", err)
		} else {
			fmt.Println("Successfully connected to Postgres")
			connection = conn
			break
		}

		fmt.Printf("Waiting for %0.2f seconds before retrying connection", float32(i)/2)
		time.Sleep(time.Second * time.Duration(i/2))
	}

	return connection, nil
}

func connectDB(config *pgx.ConnConfig) (*pgx.Conn, error) {
	conn, err := pgx.ConnectConfig(context.Background(), config)

	if err != nil {
		return nil, err
	}

	return conn, nil
}
