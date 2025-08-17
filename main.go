package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Student struct (MongoDB model)
type Student struct {
	ID    string `json:"id" bson:"_id,omitempty"`
	Name  string `json:"name" bson:"name"`
	Age   int    `json:"age" bson:"age"`
	Grade string `json:"grade" bson:"grade"`
}

var studentCollection *mongo.Collection

func initMongo() {
	// Connect to MongoDB
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Ping MongoDB
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("MongoDB not reachable:", err)
	}

	// Select DB and collection
	studentCollection = client.Database("schoolDB").Collection("students")
	log.Println("✅ Connected to MongoDB and using 'students' collection")
}

func main() {
	// Initialize MongoDB
	initMongo()

	// Initialize Gin router
	router := gin.Default()

	// Create Student
	router.POST("/students", func(c *gin.Context) {
		var student Student
		if err := c.BindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err := studentCollection.InsertOne(ctx, student)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert student"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Student added successfully"})
	})

	// Get All Students
	router.GET("/students", func(c *gin.Context) {
		var students []Student
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cursor, err := studentCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch students"})
			return
		}
		defer cursor.Close(ctx)

		if err := cursor.All(ctx, &students); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode students"})
			return
		}

		c.JSON(http.StatusOK, students)
	})

	// Run server
	router.Run(":8080")
}
