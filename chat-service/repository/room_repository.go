package repository

import (
	"context"
	"fmt"

	"example.com/chat_app/chat_service/structs"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)



type MongoChatRoomRepository struct {
	collection *mongo.Collection
}

func NewMongoChatRoomRepository(client *mongo.Client, dbName, collectionName string) *MongoChatRoomRepository {
	collection := client.Database(dbName).Collection(collectionName)
	return &MongoChatRoomRepository{collection: collection}
}

func (repo *MongoChatRoomRepository) GetRoom(ctx context.Context, id string) (*structs.ChatRoomEntity, error) {
	var room structs.ChatRoomEntity
	filter := bson.M{"id": id}
	err := repo.collection.FindOne(ctx, filter).Decode(&room)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (repo *MongoChatRoomRepository) CreateRoom(ctx context.Context) (*structs.ChatRoomEntity, error) {
	// Create a new room if it does not exist
	newRoom := &structs.ChatRoomEntity{
		Id:       uuid.NewString(),
		Messages: []structs.Message{},
		Users:    []structs.UserPermissions{},
	}

	_, err := repo.collection.InsertOne(ctx, newRoom)
	if err != nil {
		return nil, fmt.Errorf("error inserting new room: %w", err)
	}
	return newRoom, nil
}

func (repo *MongoChatRoomRepository) DeleteRoom(ctx context.Context, id string) error {
	filter := bson.D{{Key: "id", Value: id}}
	_, err := repo.collection.DeleteOne(ctx, filter)
	return err
}

func (repo *MongoChatRoomRepository) AddMessageToRoom(ctx context.Context, roomId string, message *structs.Message) error {
	filter := bson.M{"id": roomId}
	update := bson.M{
		"$push": bson.M{
			"messages": message,
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *MongoChatRoomRepository) InsertUserIntoRoom(ctx context.Context, roomId string, user structs.UserPermissions) error {
	filter := bson.M{"id": roomId}
	update := bson.M{
		"$addToSet": bson.M{
			"users": user,
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *MongoChatRoomRepository) ChangeUserRole(ctx context.Context, roomId string, userId string, role structs.Role) error {
	filter := bson.M{"id": roomId, "users.userId": userId}
	update := bson.M{
		"$set": bson.M{
			"users.$.role": role,
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *MongoChatRoomRepository) DeleteUserFromRoom(ctx context.Context, roomId string, userId string) error {
	filter := bson.M{"id": roomId}
	update := bson.M{
		"$pull": bson.M{
			"users": bson.M{"userId": userId},
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *MongoChatRoomRepository) GetUsersPermissions(ctx context.Context, roomId string, userId string) (*structs.UserPermissions, error) {
	// Define the aggregation pipeline
	pipeline := mongo.Pipeline{
		// Match the room with the specified ID
		bson.D{{Key: "$match", Value: bson.D{{Key: "id", Value: roomId}}}},
		// Unwind the users array
		bson.D{{Key: "$unwind", Value: "$users"}},
		// Match the specific user
		bson.D{{Key: "$match", Value: bson.D{{Key: "users.userId", Value: userId}}}},
		// Project the user permissions
		bson.D{{Key: "$project", Value: bson.D{{Key: "userPermissions", Value: "$users"}}}},
	}

	// Execute the aggregation pipeline
	cursor, err := repo.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if !cursor.Next(ctx) {
		return nil, mongo.ErrNoDocuments
	}

	var result struct {
		UserPermissions structs.UserPermissions `bson:"userPermissions"`
	}
	if err := cursor.Decode(&result); err != nil {
		return nil, err
	}

	return &result.UserPermissions, nil
}

func (repo *MongoChatRoomRepository) InsertSeenBy(ctx context.Context, roomId string, messageId string, userId string) error {
	filter := bson.M{"id": roomId, "messages.id": messageId}
	update := bson.M{
		"$addToSet": bson.M{
			"messages.$.seenBy": userId,
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}

func (repo *MongoChatRoomRepository) DeleteMessage(ctx context.Context, roomId string, messageId string) error {
	filter := bson.M{"id": roomId, "messages.id": messageId}
	update := bson.M{
		"$pull": bson.M{
			"messages": bson.M{"id": messageId},
		},
	}
	_, err := repo.collection.UpdateOne(ctx, filter, update)
	return err
}
