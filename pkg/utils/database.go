package utils

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Tiltbot Database quote structure
type Quote struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"createdAt"`
	Quote     string             `bson:"quote"`
	Quotee    string             `bson:"quotee"`
	Quoter    string             `bson:"quoter"`
}

var collection *mongo.Collection
var ctx = context.TODO()
var filter = bson.D{}
var quote Quote

// Open a connection to MongoDB
func ConnectMongo() {
	clientOptions := options.Client().ApplyURI(Conf.MongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(fmt.Printf("Problem while connecting to the collection: %v", err))
	}

	collection = client.Database("TiltBot").Collection(Conf.Collection)
}

func CreateQuote(quote Quote) error {
	_, insertErr := collection.InsertOne(ctx, quote)
	if insertErr != nil {
		log.Printf("Problem whilte creating a quote in the collection: %v", insertErr)
	}
	return insertErr
}

// Estimate number of documents in collection
func QuoteCount() int {
	count, _ := collection.EstimatedDocumentCount(ctx)
	return int(count)
}

// Gets a random quote document from the collection and unmarshals the BSON to the Quote struct
func GetRandomQuote() Quote {
	dbCount := QuoteCount()
	max := dbCount
	min := 1
	rand.Seed(time.Now().UnixNano())
	randomSkip := rand.Intn(max - min + 1)
	opts := options.FindOne().SetSkip(int64(randomSkip))

	doc, err := collection.FindOne(ctx, filter, opts).DecodeBytes()
	if err != nil {
		log.Printf("Error in utils.GetRandomQuote(): %v\n", err)
	}

	bson.Unmarshal(doc, &quote)

	return quote
}

// Gets the most recent quote from the collection and unmarshals the BSON to the Quote struct
func GetLatestQuote() Quote {
	opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

	doc, err := collection.FindOne(ctx, filter, opts).DecodeBytes()
	if err != nil {
		log.Printf("Error in utils.GetLatestQuote(): %v\n", err)
	}

	bson.Unmarshal(doc, &quote)

	return quote
}
