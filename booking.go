package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/net/context"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"
)

type Booking struct {
	Id    string    `bson:"id, json:"id"`
	Name  string    `bson:"name",json:"name"`
	Room  string    `bson:"room",json:"room"`
	Start time.Time `bson:"start",json:"start"`
	End   time.Time `bson:"end",json:"end"`
}

func connectDB() *mongo.Collection {
	mongoURI := "mongodb://localhost:27017"
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	err = client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
	collection := client.Database("Library").Collection("Booking")
	return collection
}

func genUUID() string {
	n := 4
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	s := fmt.Sprintf("%X", b)
	return s
}

func booking(c *gin.Context) {
	log.Println("==> booking")
	ctx := c.Request.Context()

	var booking Booking
	if err := c.Bind(&booking); err != nil {
		return
	}
	log.Printf("%+v\n", booking.Name)

	booking.Id = genUUID()
	collection := connectDB()
	res, err := collection.InsertOne(ctx, booking)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	log.Printf("return %+v\n", res.InsertedID)
	c.JSON(http.StatusOK, booking)
}

func findBooking(c *gin.Context) {
	log.Println("==> Find Booking")

	collection := connectDB()
	ctx := context.TODO()
	options := options.FindOptions{}
	options.Sort = bson.D{{"start", 1}}
	cur, err := collection.Find(ctx, bson.D{}, &options)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	defer cur.Close(ctx)
	var books []Booking
	for cur.Next(ctx) {
		var book Booking
		if err := cur.Decode(&book); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		books = append(books, book)
		log.Printf("%+v\n", book)
	}
	c.JSON(http.StatusOK, books)
}

func findBookingById(c *gin.Context) {
	log.Println("==> Find Booking by Id")
	ctx := c.Request.Context()
	id := c.Param("id")

	collection := connectDB()
	cur, err := collection.Find(ctx, bson.M{"id": id})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	defer cur.Close(ctx)
	var book Booking
	for cur.Next(ctx) {
		if err := cur.Decode(&book); err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		log.Printf("%+v\n", book)
	}
	c.JSON(http.StatusOK, book)
}

func cancelBooking(c *gin.Context) {
	log.Println("==> Cancel Booking")
	ctx := c.Request.Context()
	id := c.Param("id")
	collection := connectDB()
	res, err := collection.DeleteOne(ctx, bson.D{{"id", id}})

	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	if res.DeletedCount == 0 {
		c.Status(http.StatusNotFound)
		return
	}
}

func main() {
	router := gin.Default()
	book := router.Group("/bookings")
	{
		book.POST("/", booking)
		book.GET("/", findBooking)
		book.GET("/:id", findBookingById)
		book.DELETE("/:id", cancelBooking)
	}
	router.Run()
}
