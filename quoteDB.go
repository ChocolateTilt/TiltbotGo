package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Quote is the field structure for the "Quote" collection
type Quote struct {
	ID        primitive.ObjectID `bson:"_id"`
	CreatedAt time.Time          `bson:"createdAt"`
	Quote     string             `bson:"quote"`
	Quotee    string             `bson:"quotee"`
	Quoter    string             `bson:"quoter"`
}

// QuoteType is a string type that is used to determine the type of quote search
type QuoteType string

var (
	collection *mongo.Collection
	ctx        = context.TODO()
)

// connectMongo opens a connection to the MongoDB URI defined in the .env file
func connectMongo() error {
	mongoURI := os.Getenv("MONGO_URI")
	mongoCollectionName := os.Getenv("MONGO_COLLECTION_NAME")

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	collection = client.Database("TiltBot").Collection(mongoCollectionName)

	return nil
}

// createQuote inserts a new quote into the MongoDB collection
func createQuote(quote Quote) error {
	_, err := collection.InsertOne(ctx, quote)
	if err != nil {
		return fmt.Errorf("problem while creating a quote in the collection: %v", err)
	}
	return nil
}

// Estimate number of documents in the collection. id is only used for "user" type searches.
//
// Accepts: "full", and "user"
func (t QuoteType) quoteCount(id string) int {
	var count int64

	switch t {
	case "full":
		count, _ = collection.EstimatedDocumentCount(ctx)
	case "user":
		userFilter := bson.D{{Key: "quotee", Value: id}}
		count, _ = collection.CountDocuments(ctx, userFilter)
	}
	return int(count)
}

// getQuote returns a quote from the collection based on the type (t) of search. id is only used for "user" type searches.
//
// Accepts: "rand", "latest", and "user"
func (t QuoteType) getQuote(id string) Quote {
	var (
		min         = 1
		emptyFilter = bson.D{}
		quote       Quote
	)

	switch t {
	case "rand":
		var qType QuoteType = "full"
		fullDBMax := qType.quoteCount("")
		rand.New(rand.NewSource(time.Now().UnixNano()))
		randomSkip := rand.Intn(fullDBMax + min - 1)
		opts := options.FindOne().SetSkip(int64(randomSkip))

		doc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			log.Printf("error decoding document: %v\n", err)
		}

		err = bson.Unmarshal(doc, &quote)
		if err != nil {
			log.Printf("error unmarshaling document: %v\n", err)
		}

	case "latest":
		opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

		doc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			log.Printf("Error in utils.GetLatestQuote(): %v\n", err)
		}

		bson.Unmarshal(doc, &quote)
	case "user":
		userDBMax := t.quoteCount(id)
		if userDBMax != 0 {
			userSkip := rand.Intn(userDBMax - min + 1)
			userFilter := bson.D{{Key: "quotee", Value: id}}
			opts := options.FindOne().SetSkip(int64(userSkip))

			doc, err := collection.FindOne(ctx, userFilter, opts).DecodeBytes()
			if err != nil {
				log.Printf("Error in utils.GetQuote(): %v\n", err)
			}

			bson.Unmarshal(doc, &quote)
		}
	default:
		quote.Quote = ""
	}

	return quote
}

// getLeaderboard returns the top 10 quotees from the collection
func getLeaderboard() []bson.M {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$quotee"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		log.Printf("Error in utils.GetLeaderboard(): %v\n", err)
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("Error in utils.GetLeaderboard(): %v\n", err)
	}

	return results
}
