package main

import (
	"context"
	"fmt"
	"log"
	"os"

	todoControllers "github.com/sojborg/go-todo/internal/controllers"
	"github.com/sojborg/go-todo/internal/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Starting server...")

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
	fmt.Println("MONGODB_URI: ", MONGODB_URI)
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

	app := fiber.New()

	if ENV == "development" {
		app.Use(func(c *fiber.Ctx) error {
			c.Set("Access-Control-Allow-Origin", "http://localhost:5173")
			c.Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE")
			c.Set("Access-Control-Allow-Headers", "Content-Type")
			if c.Method() == "OPTIONS" {
				return c.SendStatus(fiber.StatusOK)
			}
			return c.Next()
		})
	}

	routes.RegisterRoutes(app)

	if PORT == "" {
		PORT = "8000"
	}

	if os.Getenv("ENV") == "production" {
		app.Static("/", "./client/dist")
	}

	app.Listen(":" + PORT)
	fmt.Println("Server started on port", PORT)
}
