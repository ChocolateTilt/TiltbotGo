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

	ctx, cancel := ctxWithTimeout()
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
func quoteCount(id, qType string, ctx context.Context) (int, error) {
	var count int64
	var err error

	switch qType {
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
// Types: "rand", "latest", "latestUser", and "user"
func getQuote(id, t string, ctx context.Context) (Quote, error) {
	var (
		min    = 1
		quote  Quote
		filter interface{}
		opts   *options.FindOneOptions
	)

	switch t {
	case "rand":
		dbMax, err := quoteCount(id, "full", ctx)
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for random quote: %w", err)
		}
		// Update the timestamp
		dbMaxT = time.Now()

		randomSkip := rng.Intn(dbMax + min - 1)
		opts = options.FindOne().SetSkip(int64(randomSkip))

	case "latest", "latestUser":
		opts = options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})
		if t == "latestUser" {
			filter = bson.D{{Key: "quotee", Value: id}}
		}

	case "user":
		userDBMax, err := quoteCount(id, "user", ctx)
		if err != nil {
			return quote, fmt.Errorf("error getting quote count for user quote: %w", err)
		}
		if userDBMax != 0 {
			userSkip := rand.Intn(userDBMax - min + 1)
			filter = bson.D{{Key: "quotee", Value: id}}
			opts = options.FindOne().SetSkip(int64(userSkip))
		}

	default:
		quote.Quote = ""
		return quote, fmt.Errorf("invalid quote type: %v", t)
	}

	if filter == nil {
		filter = bson.D{}
	}

	doc, err := collection.FindOne(ctx, filter, opts).Raw()
	if err != nil {
		return quote, fmt.Errorf("error decoding quote: %w", err)
	}

	err = bson.Unmarshal(doc, &quote)
	if err != nil {
		return quote, fmt.Errorf("error unmarshalling quote: %w", err)
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
