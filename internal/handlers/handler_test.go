package handlers_test

import (

	"ci_cd/internal/models"
	"ci_cd/internal/handlers"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var client *mongo.Client

func setupTestDB(t *testing.T) {
	var err error
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err = mongo.Connect(context.Background(), clientOptions)
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	err = client.Database("taskdb").Drop(context.Background())
	if err != nil {
		t.Fatalf("Failed to drop database: %v", err)
	}
}

func TestGetTasks(t *testing.T) {
	setupTestDB(t)
	defer client.Disconnect(context.Background())

	collection := client.Database("taskdb").Collection("tasks")
	collection.InsertOne(context.Background(), models.Task{Title: "Task 1", Description: "First task", CreatedAt: time.Now()})
	collection.InsertOne(context.Background(), models.Task{Title: "Task 2", Description: "Second task", CreatedAt: time.Now()})

	req, err := http.NewRequest("GET", "/tasks", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handlers.GetTasks(client)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	var tasks []models.Task
	err = json.NewDecoder(rr.Body).Decode(&tasks)
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %v", len(tasks))
	}
}

func TestCreateTask(t *testing.T) {
	setupTestDB(t)
	defer client.Disconnect(context.Background())

	newTask := `{"title":"Test Task","description":"This is a test task"}`
	req, err := http.NewRequest("POST", "/tasks", strings.NewReader(newTask))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := handlers.CreateTask(client)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	// Javobni to'g'ridan-to'g'ri `InsertOne` natijasiga o'xshash formatda o'qiymiz
	var res struct {
		InsertedID primitive.ObjectID `json:"InsertedID"`
	}
	err = json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatal(err)
	}

	// Task haqiqatan qo'shilganligini tekshirish
	var task models.Task
	collection := client.Database("taskdb").Collection("tasks")
	err = collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&task)
	if err != nil {
		t.Fatal("Task not inserted")
	}
}

func TestGetTask(t *testing.T) {
	setupTestDB(t)
	defer client.Disconnect(context.Background())

	collection := client.Database("taskdb").Collection("tasks")
	res, err := collection.InsertOne(context.Background(), models.Task{Title: "Single Task", Description: "Task description", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/tasks/"+res.InsertedID.(primitive.ObjectID).Hex(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/tasks/{id}", handlers.GetTask(client)).Methods("GET")
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	var task models.Task
	err = json.NewDecoder(rr.Body).Decode(&task)
	if err != nil {
		t.Fatal(err)
	}

	if task.Title != "Single Task" {
		t.Errorf("expected task title to be 'Single Task', got '%v'", task.Title)
	}
}

func TestUpdateTask(t *testing.T) {
	setupTestDB(t)
	defer client.Disconnect(context.Background())

	collection := client.Database("taskdb").Collection("tasks")
	res, err := collection.InsertOne(context.Background(), models.Task{Title: "Task to Update", Description: "Update this task", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	updateData := `{"title":"Updated Task","description":"Task has been updated"}`
	req, err := http.NewRequest("PUT", "/tasks/"+res.InsertedID.(primitive.ObjectID).Hex(), strings.NewReader(updateData))
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/tasks/{id}", handlers.UpdateTask(client)).Methods("PUT")
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	var task models.Task
	err = collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&task)
	if err != nil {
		t.Fatal(err)
	}

	if task.Title != "Updated Task" {
		t.Errorf("expected task title to be 'Updated Task', got '%v'", task.Title)
	}
}

func TestDeleteTask(t *testing.T) {
	setupTestDB(t)
	defer client.Disconnect(context.Background())

	collection := client.Database("taskdb").Collection("tasks")
	res, err := collection.InsertOne(context.Background(), models.Task{Title: "Task to Delete", Description: "Delete this task", CreatedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("DELETE", "/tasks/"+res.InsertedID.(primitive.ObjectID).Hex(), nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r := mux.NewRouter()
	r.HandleFunc("/tasks/{id}", handlers.DeleteTask(client)).Methods("DELETE")
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("expected status %v, got %v", http.StatusOK, status)
	}

	err = collection.FindOne(context.Background(), bson.M{"_id": res.InsertedID}).Decode(&models.Task{})
	if err == nil {
		t.Errorf("expected task to be deleted")
	}
}
