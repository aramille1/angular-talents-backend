package internal

import (
	"context"
	"angular-talents-backend/db"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
)

type ExistingRecruiterProfileError struct {}

func (m *ExistingRecruiterProfileError) Error() string {
	return "recruiter profile already created for this user"
}

type ExistingEngineerProfileError struct {}

func (m *ExistingEngineerProfileError) Error() string {
	return "engineer profile already created for this user"
}

func Validate(ctx context.Context, id uuid.UUID) error {
	err := checkAlreadyCreated(ctx, id)
	if err != nil {
		return err
	}

	return nil
}

func checkAlreadyCreated(ctx context.Context, id uuid.UUID) (error) {
	engCol := db.Database.Collection("engineers")
	recruiterCol := db.Database.Collection("recruiters")

	filter := bson.D{{Key: "_id", Value: id}}
	countEng, err := engCol.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if countEng != 0 {
		return &ExistingEngineerProfileError{}
	}
	
	countRecruiters, err := recruiterCol.CountDocuments(ctx, filter)
	if err != nil {
		return err
	}

	if countRecruiters != 0 {
		return &ExistingRecruiterProfileError{}
	}

	return nil
}