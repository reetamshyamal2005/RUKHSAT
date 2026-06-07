package common

import (
	"context"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	clientInstance *mongo.Client
	clientErr      error
	mongoOnce      sync.Once
)

const DefaultDBName = "rukhsat"

// Student represents a graduating senior
type Student struct {
	ID                primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name              string             `bson:"name" json:"name"`
	Email             string             `bson:"email" json:"email"`
	RSVPStatus        string             `bson:"rsvpStatus" json:"rsvpStatus"`         // pending, confirmed, declined
	FoodPreference    string             `bson:"foodPreference" json:"foodPreference"` // veg, non-veg
	Verified          bool               `bson:"verified" json:"verified"`
	VerificationToken string             `bson:"verificationToken,omitempty" json:"-"`
	PendingRSVPStatus   string             `bson:"pendingRsvpStatus,omitempty" json:"-"`
	PendingFoodPref     string             `bson:"pendingFoodPref,omitempty" json:"-"`
}

// Media represents a photo or video upload metadata
type Media struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	URL         string             `bson:"url" json:"url"`
	Title       string             `bson:"title" json:"title"`
	Type        string             `bson:"type" json:"type"`         // photo, video
	Category    string             `bson:"category" json:"category"` // campus, classroom, fests, hostel, messages
	Description string             `bson:"description" json:"description"`
	Duration    string             `bson:"duration" json:"duration"`
	CreatedAt   time.Time          `bson:"createdAt" json:"createdAt"`
}

// GetMongoClient returns the shared mongo.Client instance using a thread-safe singleton pattern
func GetMongoClient() (*mongo.Client, error) {
	mongoOnce.Do(func() {
		uri := os.Getenv("MONGODB_URI")
		if uri == "" {
			uri = "mongodb://localhost:27017"
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(uri)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			clientErr = err
			return
		}

		clientInstance = client
	})

	return clientInstance, clientErr
}

// GetCollection returns a helper mongo.Collection instance
func GetCollection(collectionName string) (*mongo.Collection, error) {
	client, err := GetMongoClient()
	if err != nil {
		return nil, err
	}

	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = DefaultDBName
	}

	return client.Database(dbName).Collection(collectionName), nil
}
