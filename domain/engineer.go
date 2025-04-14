package domain

import (
	"context"
	"crypto/md5"
	"errors"
	"net/url"
	"angular-talents-backend/db"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PartialEngineer struct {
	ID uuid.UUID 			`bson:"_id,required"`
	UserID uuid.UUID		`bson:"user_id,required"`
	Tagline string			`bson:"tagline,required"`
	City string				`bson:"city,required"`
	State string			`bson:"state,omitempty"`
	Country string			`bson:"country,required"`
	Avatar string			`bson: "avatar,required"`
	Bio string				`bson:"bio,required"`
	SearchStatus string		`bson:"search_status,required"`
	RoleType []string		`bson:"role_type,required"`
	RoleLevel []string		`bson:"role_level,required"`
}

type Engineer struct {
	ID uuid.UUID 			`bson:"_id,required"`
	UserID uuid.UUID		`bson:"user_id,required"`
	Firstname string 		`bson:"first_name,required"`
	Lastname string			`bson:"last_name,required"`
	Tagline string			`bson:"tagline,required"`
	City string				`bson:"city,required"`
	State string			`bson:"state,omitempty"`
	Country string			`bson:"country,required"`
	Avatar string			`bson: "avatar,required"`
	Bio string				`bson:"bio,required"`
	SearchStatus string		`bson:"search_status,required"`
	RoleType []string		`bson:"role_type,required"`
	RoleLevel []string		`bson:"role_level,required"`
	Website string			`bson:"website,omitempty"`
	Github string			`bson:"github,required"`
	Twitter string			`bson:"twitter,omitempty"`
	LinkedIn string			`bson:"linkedin,required"`
	StackOverflow string	`bson:"stackoverflow,omitempty"`
}

type CreateEngineerPayload struct {
	FirstName string 	`json:"firstName" validate:"required,alpha"`
	LastName string		`json:"lastName"  validate:"required,alpha"`
	Tagline string		`json:"tagline"  validate:"required"`
	City string			`json:"city"  validate:"required"`
	State string		`json:"state,omitempty"`
	Country string		`json:"country"  validate:"required"`
	Avatar string			`json: "avatar"  validate:"required"`
	Bio string			`json:"bio"  validate:"required"`
	SearchStatus string	`json:"searchStatus"  validate:"required,oneof=actively_looking open not_interested invisible"`
	RoleType []string	`json:"roleType"  validate:"required,dive,oneof=contract_part_time contract_full_time employee_part_time employee_full_time"`
	RoleLevel []string	`json:"roleLevel"  validate:"required,dive,oneof=junior mid_level senior principal_staff c_level"`
	Website string		`json:"website,omitempty"  validate:"omitempty,url"`
	Github string		`json:"github"  validate:"required,url"`
	Twitter string		`json:"twitter,omitempty"  validate:"omitempty,url"`
	LinkedIn string		`json:"linkedIn"  validate:"required,url"`
	StackOverflow string`json:"stackOverflow,omitempty"  validate:"omitempty,url"`
}

type UpdateEngineerPayload struct {
	FirstName string 	`bson:"first_name,omitempty" json:"firstName" validate:"omitempty,alpha"`
	LastName string		`bson:"last_name,omitempty" json:"lastName"  validate:"omitempty,alpha"`
	Tagline string		`bson:"tagline,omitempty" json:"tagline"  validate:"omitempty"`
	City string			`bson:"city,omitempty" json:"city"  validate:"omitempty,alpha"`
	State string		`bson:"state,omitempty" json:"state,omitempty" validate:"omitempty"`
	Country string		`bson:"country,omitempty" json:"country"  validate:"omitempty,alpha"`
	Avatar string			`bson: "avatar,omitempty" json:"avatar"  validate:"omitempty,alpha"`
	Bio string			`bson:"bio,omitempty" json:"bio" validate:"omitempty"`
	SearchStatus string	`bson:"search_status,omitempty" json:"searchStatus"  validate:"omitempty,oneof=actively_looking open not_interested invisible"`
	RoleType []string	`bson:"role_type,omitempty" json:"roleType"  validate:"omitempty,dive,oneof=contract_part_time contract_full_time employee_part_time employee_full_time"`
	RoleLevel []string	`bson:"role_level,omitempty" json:"roleLevel"  validate:"omitempty,dive,oneof=junior mid_level senior principal_staff c_level"`
	Website string		`bson:"website,omitempty" json:"website,omitempty"  validate:"omitempty,url"`
	Twitter string		`bson:"twitter,omitempty" json:"twitter,omitempty"  validate:"omitempty,url"`
	StackOverflow string`bson:"stackoverflow,omitempty" json:"stackOverflow,omitempty"  validate:"omitempty,url"`
}

type ReadEngineerPayload struct {
	EngineerID string 	`json:"engineerId" validate:"required,min=36"`
}

type ListEngineersPagination struct {
	Page int64 		`json:"page" bson:"page"`
	Limit int64 	`json:"limit" bson:"limit"`
}

type ListEngineersFilter struct {
	Country string 		`json:"country"  bson:"country,omitempty"`
	SearchStatus string `json:"searchStatus" bson:"searchStatus,omitempty"`
	RoleLevel string 	`json:"roleLevel" bson:"roleLevel,omitempty"`
	RoleType string 	`json:"roleType" bson:"roleType,omitempty"`
}

type ListEngineersParams struct {
	Pagination *ListEngineersPagination
	Filter *ListEngineersFilter
}

func (e *Engineer) NewPartialEngineer() (*PartialEngineer) {
	return &PartialEngineer{
		ID: e.ID,
		UserID: e.UserID,
		Tagline: e.Tagline,
		City: e.City,
		State: e.State,
		Country: e.Country,
		Avatar: e.Avatar,
		Bio: e.Bio,
		SearchStatus: e.SearchStatus,
		RoleType: e.RoleType,
		RoleLevel: e.RoleLevel,
	}
}

func (p *CreateEngineerPayload) NewEngineer(ctx context.Context) (*Engineer, error) {
	userID := ctx.Value("userID").(uuid.UUID)
	userHash := md5.Sum([]byte(userID.String()))
	engineerID, err := uuid.FromBytes(userHash[:]);
	if err != nil {
		return nil, err
	}

	 return &Engineer{
		ID: engineerID,
		UserID: userID,
		Firstname: p.FirstName,
		Lastname: p.LastName,
		Tagline: p.Tagline,
		City: p.City,
		State: p.State,
		Country: p.Country,
		Avatar: p.Avatar,
		Bio: p.Bio,
		SearchStatus: p.SearchStatus,
		RoleType: p.RoleType,
		RoleLevel: p.RoleLevel,
		Website: p.Website,
		Github: p.Github,
		Twitter: p.Twitter,
		LinkedIn: p.LinkedIn,
		StackOverflow: p.StackOverflow,
	}, nil
}

func (u *UpdateEngineerPayload) Validate(ctx context.Context, userID uuid.UUID, engineerID string) error {
	v := validator.New()
	err := v.Struct(u)
	if err != nil {
		return err
	}

	parsedEngineerID, err := uuid.Parse(engineerID)
	if err != nil {
		return err
	}

	ok, err := u.checkUserOwnsEngineer(ctx, userID, parsedEngineerID )
	if err != nil {
		return err
	}

	if !ok {
		return errors.New("can't update engineer belonging to other user")
	}

	return nil
}

func (e *UpdateEngineerPayload) checkUserOwnsEngineer(ctx context.Context, userID, engineerID uuid.UUID) (bool, error) {
	engCol := db.Database.Collection("engineers")

	var engineer Engineer

	filter := bson.D{{Key: "_id", Value: engineerID}}
	err := engCol.FindOne(ctx, filter).Decode(&engineer)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, errors.New("Engineer not found")
		}
		return false, err
	}

	if engineer.UserID != userID{
		return false, nil
	}

	return true, nil
}

func NewListEngineerParams(isMember bool, q url.Values) (*ListEngineersParams, error) {
	params := &ListEngineersParams{
		Pagination: &ListEngineersPagination{
			Page: 1,
			Limit: 10,
		},
		Filter:  &ListEngineersFilter{},
	}

	if !isMember {
		return params, nil
	}

	if q.Get("page") != "" {
		page, err := strconv.ParseInt(q.Get("page"), 10, 64)
		if err != nil {
			return nil, err
		}
		params.Pagination.Page = page
	}

	if q.Get("limit") != "" {
		limit, err := strconv.ParseInt(q.Get("limit"), 10, 64)
		if err != nil {
			return nil, err
		}
		params.Pagination.Limit = limit
	}

	if q.Get("country") != "" {
		params.Filter.Country = q.Get("country")
	}

	if q.Get("searchStatus") != "" {
		params.Filter.SearchStatus = q.Get("searchStatus")
	}

	if q.Get("roleLevel") != "" {
		params.Filter.RoleLevel = q.Get("roleLevel")
	}

	if q.Get("roleType") != "" {
		params.Filter.RoleType = q.Get("roleType")
	}

	return params, nil
}
