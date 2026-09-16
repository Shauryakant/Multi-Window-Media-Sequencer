package db

import (
	"context"
	"fmt"
	"sync"
	"time"

	"media-sequencer/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type Store interface {
	GetAllWindows(ctx context.Context) ([]models.Window, error)
	GetWindow(ctx context.Context, id string) (*models.Window, error)
	SaveWindow(ctx context.Context, window *models.Window) error
	AppendMedia(ctx context.Context, windowID string, item models.MediaItem) error
	UpdateCycleStartedAt(ctx context.Context, windowID string, t time.Time) error
	SeedWindows(ctx context.Context, windows []models.Window) error
}

type MongoStore struct {
	client     *mongo.Client
	db         *mongo.Database
	collection *mongo.Collection
}

func NewMongoStore(ctx context.Context, uri string, dbName string) (*MongoStore, error) {
	clientOpts := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	db := client.Database(dbName)
	coll := db.Collection("windows")
	return &MongoStore{
		client:     client,
		db:         db,
		collection: coll,
	}, nil
}

func (m *MongoStore) GetAllWindows(ctx context.Context) ([]models.Window, error) {
	cursor, err := m.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var windows []models.Window
	if err := cursor.All(ctx, &windows); err != nil {
		return nil, err
	}
	return windows, nil
}

func (m *MongoStore) GetWindow(ctx context.Context, id string) (*models.Window, error) {
	var window models.Window
	err := m.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&window)
	if err != nil {
		return nil, err
	}
	return &window, nil
}

func (m *MongoStore) SaveWindow(ctx context.Context, window *models.Window) error {
	opts := options.Replace().SetUpsert(true)
	_, err := m.collection.ReplaceOne(ctx, bson.M{"_id": window.ID}, window, opts)
	return err
}

func (m *MongoStore) AppendMedia(ctx context.Context, windowID string, item models.MediaItem) error {
	filter := bson.M{"_id": windowID}
	update := bson.M{"$push": bson.M{"playlist": item}}
	res, err := m.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		return fmt.Errorf("window %s not found", windowID)
	}
	return nil
}

func (m *MongoStore) UpdateCycleStartedAt(ctx context.Context, windowID string, t time.Time) error {
	filter := bson.M{"_id": windowID}
	update := bson.M{"$set": bson.M{"cycle_started_at": t}}
	_, err := m.collection.UpdateOne(ctx, filter, update)
	return err
}

func (m *MongoStore) SeedWindows(ctx context.Context, windows []models.Window) error {
	for _, w := range windows {
		opts := options.Replace().SetUpsert(true)
		_, err := m.collection.ReplaceOne(ctx, bson.M{"_id": w.ID}, w, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

// MemoryStore provides in-memory fallback for local development or testing
type MemoryStore struct {
	mu      sync.RWMutex
	windows map[string]models.Window
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		windows: make(map[string]models.Window),
	}
}

func (s *MemoryStore) GetAllWindows(ctx context.Context) ([]models.Window, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	res := make([]models.Window, 0, len(s.windows))
	for _, w := range s.windows {
		res = append(res, w)
	}
	return res, nil
}

func (s *MemoryStore) GetWindow(ctx context.Context, id string) (*models.Window, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.windows[id]
	if !ok {
		return nil, fmt.Errorf("window %s not found", id)
	}
	return &w, nil
}

func (s *MemoryStore) SaveWindow(ctx context.Context, window *models.Window) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.windows[window.ID] = *window
	return nil
}

func (s *MemoryStore) AppendMedia(ctx context.Context, windowID string, item models.MediaItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.windows[windowID]
	if !ok {
		return fmt.Errorf("window %s not found", windowID)
	}
	w.Playlist = append(w.Playlist, item)
	s.windows[windowID] = w
	return nil
}

func (s *MemoryStore) UpdateCycleStartedAt(ctx context.Context, windowID string, t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	w, ok := s.windows[windowID]
	if !ok {
		return fmt.Errorf("window %s not found", windowID)
	}
	w.CycleStartedAt = t
	s.windows[windowID] = w
	return nil
}

func (s *MemoryStore) SeedWindows(ctx context.Context, windows []models.Window) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, w := range windows {
		s.windows[w.ID] = w
	}
	return nil
}
