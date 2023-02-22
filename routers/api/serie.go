package api

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func FindSerie(userId string, id string) (*Serie, error) {

	var serie Serie

	ctx, _ := context.WithTimeout(context.Background(), time.Second*10)

	database, _ := mongoConnection()
	collection := database.Collection("series")

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		log.Println("Invalid id")
	}

	user, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println("Invalid userId")
	}

	err = collection.FindOne(ctx, bson.M{"_id": objectId, "user": user}).Decode(&serie)
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

func ListSeries(userId string, page string, limit string, genre string, rate string) (*SeriesPagination, error) {

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

	user, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		log.Println("Invalid userId")
	}

	filter["user"] = user

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
