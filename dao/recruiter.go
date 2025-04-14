package dao

import (
	"context"
	"errors"
	"angular-talents-backend/db"
	"angular-talents-backend/domain"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func InsertNewRecruiter(ctx context.Context, recruiter *domain.Recruiter) (string, error) {
	recruiterCol := db.Database.Collection("recruiters")

	insertResult, err := recruiterCol.InsertOne(ctx, *recruiter)
	if err != nil {
		return "",  err
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

func FindRecruiterById(ctx context.Context, recruiterId string) (*domain.Recruiter, error) {
	recruiterCol := db.Database.Collection("recruiters")
	var recruiter domain.Recruiter

	parsed, err := uuid.Parse(recruiterId)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: parsed}}
	err = recruiterCol.FindOne(ctx, filter).Decode(&recruiter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &recruiter, nil
}

func FindRecruiterByUser(ctx context.Context, userID uuid.UUID) (*domain.Recruiter, error) {
	recruiterCol := db.Database.Collection("recruiters")
	var recruiter domain.Recruiter

	filter := bson.D{{Key: "user_id", Value: userID}}
	err := recruiterCol.FindOne(ctx, filter).Decode(&recruiter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &recruiter, nil
}

func UpdateRecruiter(ctx context.Context, recruiterID string, data *domain.UpdateRecruiterPayload) (*domain.Recruiter, error)  {
	recruiterCol := db.Database.Collection("recruiters")

	parsedRecruiterID, err := uuid.Parse(recruiterID)
	if err != nil {
		return nil, err
	}

	var updatedRecruiter *domain.Recruiter
	 err = recruiterCol.
	FindOneAndUpdate(ctx, bson.M{"_id": parsedRecruiterID}, bson.M{"$set": data}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedRecruiter)
	if err != nil {
		return nil, err
	}

	return updatedRecruiter, nil
}

func UpdateRecruiterByUser(ctx context.Context, userID uuid.UUID, data *domain.UpdateRecruiterPayload) (*domain.Recruiter, error)  {
	recruiterCol := db.Database.Collection("recruiters")

	var updatedRecruiter *domain.Recruiter
	 err := recruiterCol.
	FindOneAndUpdate(ctx, bson.M{"user_id": userID}, bson.M{"$set": data}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&updatedRecruiter)
	if err != nil {
		return nil, err
	}

	return updatedRecruiter, nil
}