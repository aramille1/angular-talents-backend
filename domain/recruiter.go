package domain

import (
	"context"
	"crypto/md5"
	"errors"
	"net/http"
	"angular-talents-backend/db"
	"angular-talents-backend/internal"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type Recruiter struct {
	ID uuid.UUID 			`bson:"_id,required"`
	UserID uuid.UUID		`bson:"user_id,required"`
	Firstname string 		`bson:"first_name,required"`
	Lastname string			`bson:"last_name,required"`
	Company string			`bson:"company,omitempty"`
	Role string				`bson:"role,required"`
	Logo string				`bson:"logo,required"`
	Bio string				`bson:"bio,required"`
	LinkedIn string			`bson:"linkedin,required"`
	Website string			`bson:"website,omitempty"`
	IsMember bool			`bson:"is_member,required"`
}

type CreateRecruiterPayload struct {
	FirstName string 	`json:"firstName" validate:"required,alpha"`
	LastName string		`json:"lastName"  validate:"required,alpha"`
	Company string		`json:"company"  validate:"required"`
	Role string			`json:"role"  validate:"required"`
	Logo string			`json:"logo"  validate:"required"`
	Bio string			`json:"bio"  validate:"required"`
	LinkedIn string		`json:"linkedIn"  validate:"required,url"`
	Website string		`json:"website,omitempty"  validate:"omitempty,url"`
}

type UpdateRecruiterPayload struct {
	FirstName string 	`bson:"first_name,omitempty" json:"firstName" validate:"omitempty,alpha"`
	LastName string		`bson:"last_name,omitempty" json:"lastName"  validate:"omitempty,alpha"`
	Company string 		`bson:"company,omitempty" json:"company"  validate:"omitempty"`
	Bio string			`bson:"bio,omitempty" json:"bio" validate:"omitempty"`
	Logo string			`bson:"logo,omitempty" json:"logo" validate:"omitempty"`
	Role string			`bson:"role,omitempty" json:"role"  validate:"omitempty"`
	Website string		`bson:"website,omitempty" json:"website,omitempty"  validate:"omitempty,url"`
}

func (p *CreateRecruiterPayload) NewRecruiter(ctx context.Context) (*Recruiter, error) {
	userID := ctx.Value("userID").(uuid.UUID)
	userHash := md5.Sum([]byte(userID.String()))
	recruiterID, err := uuid.FromBytes(userHash[:]);
	if err != nil {
		return nil, internal.NewError(http.StatusInternalServerError, "recruiter.generate_id", "failed to generate id for new recruiter", err.Error())
	}


	 return &Recruiter{
		ID: recruiterID,
		UserID: userID,
		Firstname: p.FirstName,
		Lastname: p.LastName,
		Bio: p.Bio,
		Logo: p.Logo,
		Company: p.Company,
		Role: p.Role,
		Website: p.Website,
		LinkedIn: p.LinkedIn,
		IsMember: false,
	}, nil
}

func (u *UpdateRecruiterPayload) Validate(ctx context.Context, userID uuid.UUID, recruiterID string) error {
	v := validator.New()
	err := v.Struct(u)
	if err != nil {
		return err
	}

	parsedRecruiterID, err := uuid.Parse(recruiterID)
	if err != nil {
		return err
	}

	ok, err := u.checkUserOwnsRecruiter(ctx, userID, parsedRecruiterID )
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("can't update engineer belonging to other user")
	}

	return nil
}

func (r *UpdateRecruiterPayload) checkUserOwnsRecruiter(ctx context.Context, userID, recruiterID uuid.UUID) (bool, error) {
	engCol := db.Database.Collection("recruiters")

	var recruiter Recruiter

	filter := bson.D{{Key: "_id", Value: recruiterID}}
	err := engCol.FindOne(ctx, filter).Decode(&recruiter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("Recruiter not found")
		}
		return false, err
	}

	if recruiter.UserID != userID{
		return false, nil
	}

	return true, nil
}
