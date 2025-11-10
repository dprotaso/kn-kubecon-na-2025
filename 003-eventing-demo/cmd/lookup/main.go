// nolint
package main

import (
	"context"
	"database/sql"
	"fmt"
	"kndemo"
	"log"
	"os"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"

	"github.com/jackc/pgx/v5"
)

type Receiver struct {
	db *pgx.Conn
}

func main() {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	if dbHost == "" || dbUser == "" || dbPass == "" || dbName == "" {
		log.Fatal("Database environment variables are missing")
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := pgx.Connect(context.TODO(), dsn)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.TODO())

	if err := db.Ping(context.TODO()); err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}

	log.Println("✅ Connected to PostgreSQL")

	r := &Receiver{db}

	// The default client is HTTP.
	c, err := cloudevents.NewClientHTTP()
	if err != nil {
		log.Fatalf("failed to create client, %v", err)
	}
	if err = c.StartReceiver(context.Background(), r.Receive); err != nil {
		log.Fatalf("failed to start receiver: %v", err)
	}
}

func (r *Receiver) Receive(ctx context.Context, e cloudevents.Event) (*cloudevents.Event, error) {
	var payload *kndemo.Payload

	if err := e.DataAs(&payload); err != nil {
		return nil, fmt.Errorf("unable to unmarshal data: %w", err)
	}

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())

	// This ID serves as the unique identifier for the message
	event.SetSubject(e.Subject())
	event.SetSource("lookup")
	event.SetType("lookup.new")

	if payload.ParsedData == nil || payload.ParsedData.EmailAddress == nil {
		event.SetType("lookup.failure")
		event.SetData(cloudevents.ApplicationJSON, map[string]any{"error": "missing email"})
		return &event, nil
	}

	email := payload.ParsedData.EmailAddress

	var c kndemo.Customer
	err := r.db.QueryRow(ctx, `
		SELECT customer_id, company_name, contact_name, contact_email, country, phone
		FROM customers
		WHERE contact_email = $1`, email).Scan(
		&c.CustomerID, &c.CompanyName, &c.ContactName, &c.ContactEmail, &c.Country, &c.Phone)
	if err != nil {
		event.SetType("lookup.failure")
		event.SetData(cloudevents.ApplicationJSON, map[string]any{"error": err.Error()})

		if err == sql.ErrNoRows {
			event.SetType("lookup.notfound")
		}
		return &event, nil
	}

	payload.Customer = &c

	event.SetData(cloudevents.ApplicationJSON, payload)

	return &event, nil
}
