package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v4"
)

func main() {
	conn, err := pgx.Connect(context.Background(), "postgres://postgres:@localhost:5432")

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening connection %v", err)
	}

	path := filepath.Join("setup.sql")

	c, err := ioutil.ReadFile(path)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening sql file %v", err)
	}

	sql := string(c)

	_, err = conn.Exec(context.Background(), sql)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error executing sql %v", err)
	}

	// server := NewPlayerServer(NewInMemoryPlayerStore())

	server := NewServer(nil)
	log.Fatal(http.ListenAndServe(":5000", server))
}
