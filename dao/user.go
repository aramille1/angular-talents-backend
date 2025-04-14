package dao

import (
	"context"
	"errors"
	"reverse-job-board/db"
	"reverse-job-board/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertNewUser(ctx context.Context, user *domain.User) (string, error) {
	userCol := db.Database.Collection("users")

	insertResult, err := userCol.InsertOne(ctx, user)
	if err != nil {
		return "", err
	}

	rawId, ok := insertResult.InsertedID.(primitive.Binary)
	if !ok {
		return "", errors.New("_id not of type primitive.Binary")
	}

	insertedId, err := uuid.FromBytes(rawId.Data)
	if err != nil {
		return "", err
	}

	return insertedId.String(), nil
}

func FindUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	userCol := db.Database.Collection("users")

	var user domain.User

	filter := bson.D{{Key: "email", Value: email}}

	err := userCol.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func FindUserById(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	userCol := db.Database.Collection("users")

	var user domain.User

	filter := bson.D{{Key: "_id", Value: userID}}

	err := userCol.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

func UpdateUserVerifiedStatus(ctx context.Context, userID string) (error)  {
	userCol := db.Database.Collection("users")

	parsedUserID, err := uuid.Parse(userID)
	if err != nil {
		return err
	}

	type UpdateUserPayload struct {
		Verified bool	`bson:"verified,omitempty"`
	}
	
	data := &UpdateUserPayload{ Verified: true}
	 userCol.FindOneAndUpdate(ctx, bson.M{"_id": parsedUserID}, bson.M{"$set": data}, options.FindOneAndUpdate())
	return nil
}
