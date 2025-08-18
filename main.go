package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var studentCollection *mongo.Collection

type Student struct {
	ID    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name  string             `bson:"name" json:"name"`
	Age   int                `bson:"age" json:"age"`
	Email string             `bson:"email" json:"email"`
}

func main() {
	// MongoDB connection (only use timeout here)
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	studentCollection = client.Database("studentdb").Collection("students")
	log.Println("✅ Connected to MongoDB and using 'students' collection")

	// Gin router
	r := gin.Default()

	// Create student
	r.POST("/students", func(c *gin.Context) {
		var student Student
		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		res, err := studentCollection.InsertOne(ctx, student)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, res)
	})

	// Get all students
	r.GET("/students", func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		cur, err := studentCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer cur.Close(ctx)

		var students []Student
		for cur.Next(ctx) {
			var student Student
			if err := cur.Decode(&student); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			students = append(students, student)
		}
		c.JSON(http.StatusOK, students)
	})

	// Update student by ID
	r.PUT("/students/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var student Student
		if err := c.ShouldBindJSON(&student); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		update := bson.M{"$set": bson.M{"name": student.Name, "age": student.Age, "email": student.Email}}
		_, err = studentCollection.UpdateOne(ctx, bson.M{"_id": objID}, update)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Student updated"})
	})

	// Delete student by ID
	r.DELETE("/students/:id", func(c *gin.Context) {
		idParam := c.Param("id")
		objID, err := primitive.ObjectIDFromHex(idParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = studentCollection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Student deleted"})
	})

	// Start server
	r.Run(":8080")
}
