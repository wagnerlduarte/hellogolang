package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// gin --appPort 8080 --port 3000
func main() {
	mux := configureEndpoints()

	err := http.ListenAndServe(":8080", mux)

	if err != nil {
		log.Fatal(err)
	}
}

func configureEndpoints() http.Handler {

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.GET("/hello", func(c *gin.Context) {
		c.Writer.Write([]byte("Hello World!!!"))
	})

	r.GET("/serie/:id", func(c *gin.Context) {

		id := c.Param("id")

		serie, err := findSerie(id)

		if err != nil {
			log.Fatal(err)
		}
		c.JSON(http.StatusOK, serie)
	})

	r.GET("/series", func(c *gin.Context) {

		page := c.Request.URL.Query().Get("page")
		limit := c.Request.URL.Query().Get("limit")
		genre := c.Request.URL.Query().Get("genre")
		rate := c.Request.URL.Query().Get("rate")

		series, err := listSeries(page, limit, genre, rate)

		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, series)
	})

	return r
}

func mongoConnection() (*mongo.Database, error) {
	// Set client options
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	mongoConnection := os.Getenv("MONGO_CONNECTION")

	client, err := mongo.NewClient(options.Client().ApplyURI(mongoConnection))

	if err != nil {
		log.Fatal(err)
	}

	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)

	// Connect to MongoDB
	err = client.Connect(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")

	return client.Database("minhas-series"), nil
}

type User struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Identifier           string
	Password             string
	PasswordResetToken   string
	PasswordResetExpires string
	CreatedAt            primitive.DateTime
}

type Serie struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name      string             `json:"name" bson:"name,omitempty"`
	Status    string             `json:"status" bson:"status,omitempty"`
	Genre     string             `json:"genre" bson:"genre,omitempty"`
	Comments  string             `json:"comments" bson:"comments"`
	Rate      float64            `json:"rate" bson:"rate"`
	User      primitive.ObjectID `json:"user" bson:"user,omitempty"`
	CreatedAt primitive.DateTime `json:"createdAt" bson:"createdAt"`
}

func findSerie(id string) (*Serie, error) {

	var serie Serie

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	database, _ := mongoConnection()
	collection := database.Collection("series")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid id")
	}

	err = collection.FindOne(ctx, bson.M{"_id": objectId}).Decode(&serie)
	if err != nil {
		return nil, err
	}

	return &serie, nil
}

type SeriesPagination struct {
	Docs   []Serie `json:"docs"`
	Total  int64   `json:"total"`
	Offset int64   `json:"offset"`
	Limit  int64   `json:"limit"`
	Page   int64   `json:"page"`
	Pages  int64   `json:"pages"`
}

func listSeries(page string, limit string, genre string, rate string) (*SeriesPagination, error) {

	var serie Serie
	var series []Serie

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	database, _ := mongoConnection()
	collection := database.Collection("series")

	filter := bson.M{}
	options := new(options.FindOptions)

	var offset int64
	pageParsed, pageParseError := strconv.ParseInt(page, 10, 32)
	limitParsed, limitParseError := strconv.ParseInt(limit, 10, 32)

	if pageParseError == nil && pageParsed >= 1 && limitParseError == nil && limitParsed > 0 {
		offset = (pageParsed - 1) * limitParsed
		options.SetSkip(offset)
		options.SetLimit(limitParsed)
	} else {
		offset = 0
		options.SetSkip(offset)
		options.SetLimit(10)
	}

	if genre != "" {
		filter["genre"] = genre
	}

	if rate != "" {
		pageParam, parseError := strconv.ParseFloat(rate, 64)

		if parseError == nil {
			filter["rate"] = pageParam
		}
	}

	cur, err := collection.Find(ctx, filter, options)

	if err != nil {
		return nil, err
	}

	var count, _ = collection.CountDocuments(ctx, filter)

	for cur.Next(ctx) {
		errorHandlerDecode := cur.Decode(&serie)

		if errorHandlerDecode != nil {
			return nil, errorHandlerDecode
		}

		series = append(series, serie)
	}

	pagination := SeriesPagination{
		Docs:   series,
		Page:   pageParsed,
		Limit:  *options.Limit,
		Total:  count,
		Offset: offset,
		Pages:  int64(math.Ceil(float64(count) / float64(*options.Limit))),
	}

	return &pagination, nil
}
