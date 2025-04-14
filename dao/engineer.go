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

func InsertNewEngineer(ctx context.Context, engineer *domain.Engineer) (string, error) {
	engCol := db.Database.Collection("engineers")

	insertResult, err := engCol.InsertOne(ctx, *engineer)
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

func FindEngineerById(ctx context.Context, engineerId string) (*domain.Engineer, error) {
	engCol := db.Database.Collection("engineers")
	var engineer domain.Engineer

	parsed, err := uuid.Parse(engineerId)
	if err != nil {
		return nil, err
	}

	filter := bson.D{{Key: "_id", Value: parsed}}
	err = engCol.FindOne(ctx, filter).Decode(&engineer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &engineer, nil
}

func FindEngineerByUser(ctx context.Context, userID uuid.UUID) (*domain.Engineer, error) {
	engCol := db.Database.Collection("engineers")
	var engineer domain.Engineer

	filter := bson.D{{Key: "user_id", Value: userID}}
	err := engCol.FindOne(ctx, filter).Decode(&engineer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &engineer, nil
}

func ReadEngineers(ctx context.Context, listParams *domain.ListEngineersParams) ([]*domain.Engineer, error) {
	engCol := db.Database.Collection("engineers")

	var engineers []*domain.Engineer

	paginationOptions := options.Find()
	paginationOptions.SetSkip((listParams.Pagination.Page - 1) * listParams.Pagination.Limit)
	paginationOptions.SetLimit(listParams.Pagination.Limit)
	

	sortOptions := options.Find()
	sortOptions.SetSort(bson.D{{Key: "created_at", Value: -1}})
	
	cur, err := engCol.Find(ctx, listParams.Filter, paginationOptions, sortOptions)
	if err != nil {
		return nil, err
	}

	// Finding multiple documents returns a cursor
	// Iterating through the cursor allows us to decode documents one at a time
	for cur.Next(ctx) {
		// create an engineer into which the single document can be decoded
		var eng domain.Engineer
		err := cur.Decode(&eng)
		if err != nil {
			return nil, err
		}

		engineers = append(engineers, &eng)
	}

	if err := cur.Err(); err != nil {
		return nil, err
	}

	// Close the cursor once finished
	cur.Close(ctx)

	return engineers, nil
}

func UpdateEngineer(ctx context.Context, engineerID string, data *domain.UpdateEngineerPayload) (*domain.Engineer, error)  {
	engCol := db.Database.Collection("engineers")

	parsedEngineerID, err := uuid.Parse(engineerID)
	if err != nil {
		return nil, err
	}

	var udpatedEng *domain.Engineer
	 err = engCol.
	FindOneAndUpdate(ctx, bson.M{"_id": parsedEngineerID}, bson.M{"$set": data}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&udpatedEng)
	if err != nil {
		return nil, err
	}

	return udpatedEng, nil
}

func UpdateEngineerByUser(ctx context.Context, userID uuid.UUID, data *domain.UpdateEngineerPayload) (*domain.Engineer, error)  {
	engCol := db.Database.Collection("engineers")

	var udpatedEng *domain.Engineer
	 err := engCol.
	FindOneAndUpdate(ctx, bson.M{"user_id": userID}, bson.M{"$set": data}, options.FindOneAndUpdate().SetReturnDocument(options.After)).Decode(&udpatedEng)
	if err != nil {
		return nil, err
	}

	return udpatedEng, nil
}

func CountEngineers(ctx context.Context) (int64, error) {
	engCol := db.Database.Collection("engineers")
	var count int64
	count, err := engCol.CountDocuments(ctx, bson.D{})
	if err != nil {
		return 0, err
	}

	return count, nil
}