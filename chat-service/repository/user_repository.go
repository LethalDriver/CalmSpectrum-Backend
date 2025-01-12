package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(client *mongo.Client, dbName, collectionName string) *UserRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &UserRepository{collection: collection}
}

// usernameOnly is a minimal struct to decode only the username field.
type usernameOnly struct {
	Username string `bson:"username"`
}

// GetUsernameById retrieves only the username of a user by their ID.
func (repo *UserRepository) GetUsernameById(ctx context.Context, id string) (string, error) {
	var result usernameOnly
	filter := bson.M{"id": id}
	projection := bson.M{"username": 1, "_id": 0} // Include 'username', exclude '_id'

	err := repo.collection.FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return "", err
	}
	return result.Username, nil
}
