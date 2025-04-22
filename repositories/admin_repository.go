package repositories

import (
	"context"
	"time"

	"angular-talents-backend/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// AdminRepository handles database operations for admin users
type AdminRepository struct {
	collection *mongo.Collection
}

// NewAdminRepository creates a new instance of AdminRepository
func NewAdminRepository(db *mongo.Database) *AdminRepository {
	return &AdminRepository{
		collection: db.Collection("admins"),
	}
}

// FindByID finds an admin by ID
func (r *AdminRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*models.Admin, error) {
	var admin models.Admin
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindByUsername finds an admin by username
func (r *AdminRepository) FindByUsername(ctx context.Context, username string) (*models.Admin, error) {
	var admin models.Admin
	err := r.collection.FindOne(ctx, bson.M{"username": username}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// FindByEmail finds an admin by email
func (r *AdminRepository) FindByEmail(ctx context.Context, email string) (*models.Admin, error) {
	var admin models.Admin
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&admin)
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// Create creates a new admin in the database
func (r *AdminRepository) Create(ctx context.Context, admin *models.Admin) (*models.Admin, error) {
	admin.CreatedAt = time.Now()
	admin.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(ctx, admin)
	if err != nil {
		return nil, err
	}

	admin.ID = result.InsertedID.(primitive.ObjectID)
	return admin, nil
}

// Update updates an existing admin in the database
func (r *AdminRepository) Update(ctx context.Context, admin *models.Admin) error {
	admin.UpdatedAt = time.Now()

	_, err := r.collection.UpdateOne(
		ctx,
		bson.M{"_id": admin.ID},
		bson.M{"$set": admin},
	)
	return err
}

// Delete deletes an admin from the database
func (r *AdminRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

// FindAll finds all admins
func (r *AdminRepository) FindAll(ctx context.Context) ([]*models.Admin, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var admins []*models.Admin
	if err := cursor.All(ctx, &admins); err != nil {
		return nil, err
	}

	return admins, nil
}
