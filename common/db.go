package common

import (
	"context"
	"os"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
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
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Name                 string             `bson:"name" json:"name"`
	Email                string             `bson:"email" json:"email"`
	RSVPStatus           string             `bson:"rsvpStatus" json:"rsvpStatus"`         // pending, confirmed, declined
	FoodPreference       string             `bson:"foodPreference" json:"foodPreference"` // veg, non-veg
	LikesReading         string             `bson:"likesReading" json:"likesReading"`     // yes, no
	Verified             bool               `bson:"verified" json:"verified"`
	VerificationToken    string             `bson:"verificationToken,omitempty" json:"-"`
	PendingRSVPStatus    string             `bson:"pendingRsvpStatus,omitempty" json:"-"`
	PendingFoodPref      string             `bson:"pendingFoodPref,omitempty" json:"-"`
	PendingLikesReading  string             `bson:"pendingLikesReading,omitempty" json:"-"`
	PendingPhone         string             `bson:"pendingPhone,omitempty" json:"-"`
	Phone                string             `bson:"phone" json:"phone"`
	UniqueCode           string             `bson:"uniqueCode" json:"uniqueCode"`
	LastVerificationSent time.Time          `bson:"lastVerificationSent,omitempty" json:"-"`
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

		// Programmatically ensure indexes exist in background
		go ensureIndexes(client)
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

// ensureIndexes configures database indexes for fast query resolution
func ensureIndexes(client *mongo.Client) {
	dbName := os.Getenv("MONGODB_DB")
	if dbName == "" {
		dbName = DefaultDBName
	}
	db := client.Database(dbName)

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// 1. Indexes for students collection
	studentsColl := db.Collection("students")
	_, _ = studentsColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index(),
		},
		{
			Keys:    bson.D{{Key: "verificationToken", Value: 1}},
			Options: options.Index().SetUnique(true).SetSparse(true),
		},
	})

	// 2. Indexes for media collection
	mediaColl := db.Collection("media")
	_, _ = mediaColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "type", Value: 1}},
			Options: options.Index(),
		},
		{
			Keys:    bson.D{{Key: "createdAt", Value: -1}},
			Options: options.Index(),
		},
	})
}

// Memory Cache Layer for Instant Reads
type MemoryCache struct {
	mu           sync.RWMutex
	students     []Student
	media        []Media
	studentsTime time.Time
	mediaTime    time.Time
}

var (
	globalCache MemoryCache
	cacheTTL    = 5 * time.Second
)

func GetCachedStudents() ([]Student, error) {
	globalCache.mu.RLock()
	if len(globalCache.students) > 0 && time.Since(globalCache.studentsTime) < cacheTTL {
		res := make([]Student, len(globalCache.students))
		copy(res, globalCache.students)
		globalCache.mu.RUnlock()
		return res, nil
	}
	globalCache.mu.RUnlock()

	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	// Double check inside lock
	if len(globalCache.students) > 0 && time.Since(globalCache.studentsTime) < cacheTTL {
		res := make([]Student, len(globalCache.students))
		copy(res, globalCache.students)
		return res, nil
	}

	collection, err := GetCollection("students")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "name", Value: 1}})
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var students []Student
	if err := cursor.All(ctx, &students); err != nil {
		return nil, err
	}

	globalCache.students = students
	globalCache.studentsTime = time.Now()

	res := make([]Student, len(students))
	copy(res, students)
	return res, nil
}

func GetCachedMedia() ([]Media, error) {
	globalCache.mu.RLock()
	if len(globalCache.media) > 0 && time.Since(globalCache.mediaTime) < cacheTTL {
		res := make([]Media, len(globalCache.media))
		copy(res, globalCache.media)
		globalCache.mu.RUnlock()
		return res, nil
	}
	globalCache.mu.RUnlock()

	globalCache.mu.Lock()
	defer globalCache.mu.Unlock()

	if len(globalCache.media) > 0 && time.Since(globalCache.mediaTime) < cacheTTL {
		res := make([]Media, len(globalCache.media))
		copy(res, globalCache.media)
		return res, nil
	}

	collection, err := GetCollection("media")
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := options.Find().SetSort(bson.D{{Key: "createdAt", Value: -1}})
	cursor, err := collection.Find(ctx, bson.D{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var media []Media
	if err := cursor.All(ctx, &media); err != nil {
		return nil, err
	}

	globalCache.media = media
	globalCache.mediaTime = time.Now()

	res := make([]Media, len(media))
	copy(res, media)
	return res, nil
}

func InvalidateStudentsCache() {
	globalCache.mu.Lock()
	globalCache.students = nil
	globalCache.mu.Unlock()
}

func InvalidateMediaCache() {
	globalCache.mu.Lock()
	globalCache.media = nil
	globalCache.mu.Unlock()
}
