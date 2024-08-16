package main

import (
	"ci_cd/internal/db"
	"ci_cd/internal/handlers"
	"context"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	client, err := db.ConnectMongoDB()
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	defer client.Disconnect(context.Background())

	r := mux.NewRouter()

	r.HandleFunc("/tasks", handlers.GetTasks(client)).Methods("GET")
	r.HandleFunc("/tasks", handlers.CreateTask(client)).Methods("POST")
	r.HandleFunc("/tasks/{id}", handlers.GetTask(client)).Methods("GET")
	r.HandleFunc("/tasks/{id}", handlers.UpdateTask(client)).Methods("PUT")
	r.HandleFunc("/tasks/{id}", handlers.DeleteTask(client)).Methods("DELETE")

	log.Println("Server started on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}
