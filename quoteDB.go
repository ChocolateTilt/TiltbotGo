package main

import (
	"context"
	"fmt"
	"log"
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
	rng        = rand.New(rand.NewSource(time.Now().UnixNano()))
	dbMax      int
	dbMaxT     time.Time
)

// connectMongo opens a connection to the MongoDB URI defined in the .env file with a 10 second timeout
func connectMongo() error {
	mongoURI := os.Getenv("MONGO_URI")
	mongoCollectionName := os.Getenv("MONGO_COLLECTION_NAME")

	if mongoURI == "" || mongoCollectionName == "" {
		return fmt.Errorf("mongo URI and collection name not found in .env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	clientOptions := options.Client().ApplyURI(mongoURI).SetServerAPIOptions(serverAPI)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("error connecting to MongoDB: %w", err)
	}

	collection = client.Database("TiltBot").Collection(mongoCollectionName)

	return nil
}

// createQuote inserts a new quote into the MongoDB collection
func createQuote(quote Quote, ctx context.Context) error {
	// Zero out cached last full scan time
	dbMaxT = time.Time{}

	_, err := collection.InsertOne(ctx, quote)
	if err != nil {
		return fmt.Errorf("problem while creating a quote in the collection: %v", err)
	}
	return nil
}

// Estimate number of documents in the collection. id is only used for "user" type searches.
//
// Accepts: "full", and "user"
func (t QuoteType) quoteCount(id string, ctx context.Context) (int, error) {
	var count int64
	var err error

	switch t {
	case "full":
		if time.Since(dbMaxT).Hours() >= 1 || dbMaxT.IsZero() {
			count, err = collection.EstimatedDocumentCount(ctx)
			if err != nil {
				return 0, err
			}
			dbMax = int(count)
			dbMaxT = time.Now()
			log.Println("Cached full quote count at", dbMaxT)
		} else {
			return dbMax, nil
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
// Types: "rand", "latest", and "user"
func (t QuoteType) getQuote(id string, ctx context.Context) (Quote, error) {
	var (
		min         = 1
		emptyFilter = bson.D{}
		quote       Quote
	)

	switch t {
	case "rand":
		var qType QuoteType = "full"
		var err error
		dbMax, err = qType.quoteCount(id, ctx)
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for random quote: %w", err)
		}
		// Update the timestamp
		dbMaxT = time.Now()

		randomSkip := rng.Intn(dbMax + min - 1)
		opts := options.FindOne().SetSkip(int64(randomSkip))

		doc, err := collection.FindOne(ctx, emptyFilter, opts).Raw()
		if err != nil {
			return quote, fmt.Errorf("error decoding random quote: %w", err)
		}

		err = bson.Unmarshal(doc, &quote)
		if err != nil {
			return quote, fmt.Errorf("error unmarshalling random quote: %w", err)
		}

	case "latest":
		opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

		doc, err := collection.FindOne(ctx, emptyFilter, opts).Raw()
		if err != nil {
			return quote, fmt.Errorf("error decoding latest quote: %w", err)
		}

		bson.Unmarshal(doc, &quote)
		if err != nil {
			return quote, fmt.Errorf("error unmarshalling latest quote: %w", err)
		}

	case "user":
		userDBMax, err := t.quoteCount(id, ctx)
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for user quote: %w", err)
		}
		if userDBMax != 0 {
			userSkip := rand.Intn(userDBMax - min + 1)
			userFilter := bson.D{{Key: "quotee", Value: id}}
			opts := options.FindOne().SetSkip(int64(userSkip))

			doc, err := collection.FindOne(ctx, userFilter, opts).Raw()
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
func getLeaderboard(ctx context.Context) (string, error) {
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
