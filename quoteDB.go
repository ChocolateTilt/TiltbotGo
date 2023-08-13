package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"strings"
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
		return fmt.Errorf("error connecting to MongoDB: %w", err)
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
func (t QuoteType) quoteCount(id string) (int, error) {
	var count int64
	var err error

	switch t {
	case "full":
		count, err = collection.EstimatedDocumentCount(ctx)
		if err != nil {
			return 0, err
		}
	case "user":
		userFilter := bson.D{{Key: "quotee", Value: id}}
		count, err = collection.CountDocuments(ctx, userFilter)
		if err != nil {
			return 0, fmt.Errorf("error counting documents for user quote: %w", err)
		}
	}
	return int(count), err
}

// getQuote returns a quote from the collection based on the type (t) of search. id is only used for "user" type searches.
//
// Accepts: "rand", "latest", and "user"
func (t QuoteType) getQuote(id string) (Quote, error) {
	var (
		min         = 1
		emptyFilter = bson.D{}
		quote       Quote
	)

	switch t {
	case "rand":
		var qType QuoteType = "full"
		fullDBMax, err := qType.quoteCount("")
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for random quote: %w", err)
		}
		rand.New(rand.NewSource(time.Now().UnixNano()))
		randomSkip := rand.Intn(fullDBMax + min - 1)
		opts := options.FindOne().SetSkip(int64(randomSkip))

		doc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			return quote, fmt.Errorf("error decoding random quote: %w", err)
		}

		err = bson.Unmarshal(doc, &quote)
		if err != nil {
			return quote, fmt.Errorf("error unmarshalling random quote: %w", err)
		}

	case "latest":
		opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

		doc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			return quote, fmt.Errorf("error decoding latest quote: %w", err)
		}

		bson.Unmarshal(doc, &quote)
		if err != nil {
			return quote, fmt.Errorf("error unmarshalling latest quote: %w", err)
		}

	case "user":
		userDBMax, err := t.quoteCount(id)
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for user quote: %w", err)
		}
		if userDBMax != 0 {
			userSkip := rand.Intn(userDBMax - min + 1)
			userFilter := bson.D{{Key: "quotee", Value: id}}
			opts := options.FindOne().SetSkip(int64(userSkip))

			doc, err := collection.FindOne(ctx, userFilter, opts).DecodeBytes()
			if err != nil {
				return quote, fmt.Errorf("error decoding user quote: %w", err)
			}

			err = bson.Unmarshal(doc, &quote)
			if err != nil {
				return quote, fmt.Errorf("error unmarshalling user quote: %w", err)
			}

		}
	default:
		quote.Quote = ""
	}

	return quote, nil
}

// getLeaderboard returns the top 10 quotees from the collection
func getLeaderboard() (string, error) {
	var leaderboard []string

	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.D{{Key: "_id", Value: "$quotee"}, {Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}}}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
		{{Key: "$limit", Value: 10}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return "", fmt.Errorf("error aggregating documents for leaderboard: %w", err)
	}

	var results []bson.M
	if err = cursor.All(ctx, &results); err != nil {
		return "", fmt.Errorf("error decoding documents for leaderboard: %w", err)
	}

	for i, v := range results {
		leaderboard = append(leaderboard, fmt.Sprintf("`%v:`%v: %v\n", i+1, v["_id"], v["count"]))
	}

	cleanLB := strings.Join(leaderboard, "\n")

	return cleanLB, nil
}
