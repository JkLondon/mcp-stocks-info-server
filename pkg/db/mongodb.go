package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB представляет собой клиент для работы с MongoDB
type MongoDB struct {
	client     *mongo.Client
	database   *mongo.Database
	collection *mongo.Collection
}

// NewMongoDB создает новый экземпляр клиента MongoDB
func NewMongoDB(uri, database, collection string, timeout time.Duration) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Проверяем соединение с базой данных
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, err
	}

	db := client.Database(database)
	coll := db.Collection(collection)

	return &MongoDB{
		client:     client,
		database:   db,
		collection: coll,
	}, nil
}

// Close закрывает соединение с базой данных
func (m *MongoDB) Close(ctx context.Context) error {
	return m.client.Disconnect(ctx)
}

// GetCollection возвращает коллекцию MongoDB
func (m *MongoDB) GetCollection(collectionName string) *mongo.Collection {
	return m.database.Collection(collectionName)
}

// GetDatabase возвращает базу данных MongoDB
func (m *MongoDB) GetDatabase() *mongo.Database {
	return m.database
}

// GetClient возвращает клиент MongoDB
func (m *MongoDB) GetClient() *mongo.Client {
	return m.client
}
