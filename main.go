package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/sojborg/go-todo/internal/auth"
	todoControllers "github.com/sojborg/go-todo/internal/controllers/todoController"
	internalMiddleware "github.com/sojborg/go-todo/internal/middleware"
	"github.com/sojborg/go-todo/internal/routes"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Starting server...")
	fmt.Println("ENV: " + os.Getenv("ENV"))

	if os.Getenv("ENV") != "production" {
		// Load .env file
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatal("Error loading .env file", err)
		}
	}

	ENV := os.Getenv("ENV")
	PORT := os.Getenv("PORT")
	MONGODB_URI := os.Getenv("MONGODB_URI")

	clientOptions := options.Client().ApplyURI(MONGODB_URI)
	client, err := mongo.Connect(context.Background(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	defer client.Disconnect(context.Background())

	fmt.Println("Connected to MongoDB!")

	collection := client.Database("todo").Collection("todos")
	todoControllers.SetCollection(collection)

	r := chi.NewRouter()
	r.Use(middleware.Logger)

	if ENV == "development" {
		r.Use(internalMiddleware.CorsMiddleware)
	}

	auth.NewAuth()

	routes.RegisterRoutes(r)

	if PORT == "" {
		PORT = "8000"
	}

	if os.Getenv("ENV") == "production" {
		fs := http.FileServer(http.Dir("./client/dist"))
		r.Handle("/*", fs)
	}

	http.ListenAndServe(":"+PORT, r)
	fmt.Println("Server started on port", PORT)
}
