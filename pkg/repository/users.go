package repository

import (
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/db"
	"github.com/mixedmachine/EfficientLife/user-auth/pkg/models"

	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const UserCollection = "users"

type UsersRepository interface {
	Save(user *models.User) error
	Update(user *models.User) error
	GetById(id string) (user *models.User, err error)
	GetByEmail(email string) (user *models.User, err error)
	GetByName(name string) (user *models.User, err error)
	GetByAdmin(admin bool) (users []*models.User, err error)
	GetAll() (users []*models.User, err error)
	Delete(id string) error
}

type usersRepository struct {
	coll *mongo.Collection
}

func NewUserRepository(conn db.MongoConnection) UsersRepository {
	return &usersRepository{
		conn.DB().Collection(UserCollection),
	}
}

func (r *usersRepository) Save(user *models.User) error {
	err := r.coll.FindOne(
		context.TODO(),
		bson.D{{Key: "email", Value: user.Email}},
	).Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			res, err := r.coll.InsertOne(context.TODO(), user)
			if res != nil {
				log.Printf("Saved user: %v\n", res.InsertedID)
			}
			return err
		}
		return err
	}

	log.Println("User already exists")
	return nil
}

func (r *usersRepository) Update(user *models.User) error {
	res, err := r.coll.UpdateByID(
		context.TODO(),
		user.Id,
		bson.D{{
			Key: "$set",
			Value: bson.D{
				{Key: "email", Value: user.Email},
				{Key: "password", Value: user.Password},
				{Key: "updated_at", Value: user.UpdatedAt},
			},
		}})
	if res != nil {
		log.Printf("Updated user: %v\n", user.Id.Hex())
	}
	return err
}

func (r *usersRepository) GetById(id string) (user *models.User, err error) {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, err
	}
	err = r.coll.FindOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: _id}},
	).Decode(&user)
	return user, err
}

func (r *usersRepository) GetByEmail(email string) (user *models.User, err error) {
	err = r.coll.FindOne(
		context.TODO(),
		bson.D{{Key: "email", Value: email}},
	).Decode(&user)
	return user, err
}

func (r *usersRepository) GetByName(name string) (user *models.User, err error) {
	err = r.coll.FindOne(
		context.TODO(),
		bson.D{{Key: "name", Value: name}},
	).Decode(&user)
	return user, err
}

func (r *usersRepository) GetByAdmin(admin bool) (users []*models.User, err error) {
	cursor, err := r.coll.Find(context.TODO(), bson.D{{Key: "admin", Value: admin}})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &users)
	return users, err
}

func (r *usersRepository) GetAll() (users []*models.User, err error) {
	cursor, err := r.coll.Find(context.TODO(), bson.M{})
	if err != nil {
		return nil, err
	}

	err = cursor.All(context.TODO(), &users)
	return users, err
}

func (r *usersRepository) Delete(id string) error {
	_id, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.coll.DeleteOne(
		context.TODO(),
		bson.D{{Key: "_id", Value: _id}},
	)
	log.Printf("Deleted user: %s\n", id)
	return err
}
