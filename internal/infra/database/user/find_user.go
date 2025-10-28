package user

import (
	"context"

	"github.com/auction-goexpert/internal/entity"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(database *mongo.Database) *UserRepository {
	return &UserRepository{
		Collection: database.Collection("users"),
	}
}

func (ur *UserRepository) FindUserById(ctx context.Context, id string) (*entity.User, error) {
	filter := bson.M{"_id": id}

	var userEntityMongo entity.UserEntityMongo
	err := ur.Collection.FindOne(ctx, filter).Decode(&userEntityMongo)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &entity.User{
		Id:   userEntityMongo.Id,
		Name: userEntityMongo.Name,
	}, nil
}
