package db_test

import (
	"ci_cd/internal/db"
	"context"
	"testing"
)

func TestConnectMongoDB(t *testing.T) {
	client, err := db.ConnectMongoDB()
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer client.Disconnect(context.Background())

	err = client.Ping(context.Background(), nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}
}
