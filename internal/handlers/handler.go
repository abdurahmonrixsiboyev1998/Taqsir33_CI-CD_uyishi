package handlers

import (
	"ci_cd/internal/models"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetTasks(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var tasks []models.Task
		collection := client.Database("taskdb").Collection("tasks")
		cur, err := collection.Find(context.Background(), bson.M{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer cur.Close(context.Background())

		for cur.Next(context.Background()) {
			var task models.Task
			err := cur.Decode(&task)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tasks = append(tasks, task)
		}
		json.NewEncoder(w).Encode(tasks)
	}
}

func CreateTask(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var task models.Task
		_ = json.NewDecoder(r.Body).Decode(&task)
		task.CreatedAt = time.Now()

		collection := client.Database("taskdb").Collection("tasks")
		res, err := collection.InsertOne(context.Background(), task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(res)
	}
}

func GetTask(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		var task models.Task
		collection := client.Database("taskdb").Collection("tasks")
		err := collection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(task)
	}
}

func UpdateTask(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		var task models.Task
		_ = json.NewDecoder(r.Body).Decode(&task)
		collection := client.Database("taskdb").Collection("tasks")
		update := bson.M{
			"$set": bson.M{
				"title":       task.Title,
				"description": task.Description,
				"status":      task.Status,
			},
		}

		_, err := collection.UpdateOne(context.Background(), models.Task{ID: id}, update)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(task)
	}
}

func DeleteTask(client *mongo.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		params := mux.Vars(r)
		id, _ := primitive.ObjectIDFromHex(params["id"])
		collection := client.Database("taskdb").Collection("tasks")
		_, err := collection.DeleteOne(context.Background(), models.Task{ID: id})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(bson.M{"message": "Task deleted successfully"})
	}
}
