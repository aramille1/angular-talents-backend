package domain

import (
	"context"
	"crypto/md5"
	"errors"
	"math/rand"
	"os"
	"reverse-job-board/db"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID               uuid.UUID `bson:"_id,required"`
	Email            string    `bson:"email,required"`
	Password         string    `bson:"password,required"`
	Verified         bool      `bson:"verified,omitempty"`
	VerificationCode int       `bson:"verificationCode,omitempty"`
}

type BodyData struct {
	Email string `json:"email" validate:"required,email"`
}

type SignUpData struct {
	BodyData
	Password       string `json:"password" validate:"required,min=8,max=20"`
	RecaptchaToken string `json:"recaptchaToken" validate:"required"`
	// Honeypot field won't be included here since it's meant to be hidden
}

type LoginData struct {
	BodyData
	Password string `json:"password" validate:"required"`
}

type JwtCustomClaims struct {
	UserID uuid.UUID `json:"userId"`
	jwt.StandardClaims
}

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func (d *SignUpData) NewUser() (*User, error) {
	newUser := &User{
		Email:            d.Email,
		Verified:         false,
		VerificationCode: rand.Intn(10000000),
	}

	err := newUser.generateID()
	if err != nil {
		return nil, err
	}

	err = newUser.hashPassword(d.Password)
	if err != nil {
		return nil, err
	}

	return newUser, nil
}

func (u *User) hashPassword(givenPassword string) error {
	bytes, err := bcrypt.GenerateFromPassword([]byte(givenPassword), 14)
	if err != nil {
		return err
	}
	u.Password = string(bytes)
	return nil
}

func (u *User) generateID() error {
	userHash := md5.Sum([]byte(u.Email))
	userID, err := uuid.FromBytes(userHash[:])
	if err != nil {
		return err
	}
	u.ID = userID
	return nil
}

func (u *User) Validate(ctx context.Context) error {
	alreadyCreated, err := u.checkAlreadyCreated(ctx)
	if err != nil {
		return err
	}

	if alreadyCreated {
		return errors.New("user already created")
	}

	return nil
}

func (u *User) checkAlreadyCreated(ctx context.Context) (bool, error) {
	userCol := db.Database.Collection("users")
	filter := bson.D{{Key: "email", Value: u.Email}}
	count, err := userCol.CountDocuments(ctx, filter)
	if err != nil {
		return false, err
	}

	return count != 0, nil
}

func (ld *LoginData) VerifyLogin(ctx context.Context) (*User, error) {
	user, err := ld.checkExists(ctx)
	if err != nil {
		return nil, err
	}

	err = ld.verifyPassword(user.Password)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, "userID", user.ID)
	return user, nil
}

func (ld *LoginData) verifyPassword(storedPassword string) error {
	err := bcrypt.CompareHashAndPassword([]byte(storedPassword), []byte(ld.Password))
	if err != nil {
		return err
	}

	return nil
}

func GenerateJWT(userID uuid.UUID) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &JwtCustomClaims{
		UserID: userID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func ValidateToken(signedToken string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(
		signedToken,
		&JwtCustomClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return []byte(jwtSecret), nil
		},
	)

	if err != nil {
		return uuid.Nil, err
	}

	if !token.Valid {
		return uuid.Nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*JwtCustomClaims)

	if !ok || !token.Valid {
		return uuid.Nil, errors.New("failed to extract claims")
	}

	if claims.StandardClaims.ExpiresAt < time.Now().Local().Unix() {
		return uuid.Nil, errors.New("token expired")
	}

	return claims.UserID, nil
}

func (ld *LoginData) checkExists(ctx context.Context) (*User, error) {
	var user User
	useCol := db.Database.Collection("users")
	filter := bson.D{{Key: "email", Value: ld.Email}}
	err := useCol.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}
