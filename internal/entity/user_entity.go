package entity

import (
	"context"

	"github.com/google/uuid"
)

type User struct {
	Id   string
	Name string
}

type UserEntityMongo struct {
	Id   string `bson:"_id"`
	Name string `bson:"name"`
}

type UserRepositoryInterface interface {
	FindUserById(ctx context.Context, id string) (*User, error)
}

func CreateUser(name string) (*User, error) {
	user := &User{
		Id:   uuid.New().String(),
		Name: name,
	}

	return user, nil
}
