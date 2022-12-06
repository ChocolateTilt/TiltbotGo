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

var (
	collection  *mongo.Collection
	ctx         = context.TODO()
	emptyFilter = bson.D{}
	quote       Quote
	min         = 1
)

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
func QuoteCount(t string, id string) int {
	var count int64
	if t == "full" {
		count, _ = collection.EstimatedDocumentCount(ctx)
	} else if t == "user" {
		userFilter := bson.D{{Key: "quotee", Value: id}}
		count, _ = collection.CountDocuments(ctx, userFilter)
	}
	return int(count)
}

// Gets a quote from the collection based on the type (t) of search.
//
// Accepts: "rand", "latest", and "user"
func GetQuote(t string, id string) Quote {
	rand.Seed(time.Now().UnixNano())

	if t == "rand" {
		fullDBMax := QuoteCount("full", "")
		randomSkip := rand.Intn(fullDBMax - min + 1)
		opts := options.FindOne().SetSkip(int64(randomSkip))

		dbRandDoc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			log.Printf("Error in utils.GetQuote(): %v\n", err)
		}

		bson.Unmarshal(dbRandDoc, &quote)

	} else if t == "latest" {
		opts := options.FindOne().SetSort(bson.D{{Key: "createdAt", Value: -1}})

		latestDoc, err := collection.FindOne(ctx, emptyFilter, opts).DecodeBytes()
		if err != nil {
			log.Printf("Error in utils.GetLatestQuote(): %v\n", err)
		}

		bson.Unmarshal(latestDoc, &quote)

	} else if t == "user" {
		userDBMax := QuoteCount("user", id)
		if userDBMax != 0 {
			userSkip := rand.Intn(userDBMax - min + 1)
			userFilter := bson.D{{Key: "quotee", Value: id}}
			opts := options.FindOne().SetSkip(int64(userSkip))

			userRandDoc, err := collection.FindOne(ctx, userFilter, opts).DecodeBytes()
			if err != nil {
				log.Printf("Error in utils.GetQuote(): %v\n", err)
			}

			bson.Unmarshal(userRandDoc, &quote)
		} else {
			quote.Quote = ""
		}
	}

	return quote
}
