package repositories

import (
	"context"
	"errors"
	"linkfast/read-api/models"
	"linkfast/read-api/utils/consts"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LinkRepository interface {
	GetByCode(ctx context.Context, code string) (models.Link, error)
	ExistsByShortCode(ctx context.Context, code string) (bool, error)
	GetById(ctx context.Context, id int64) (models.Link, error)
	ExistsByID(ctx context.Context, id int64) (bool, error)
}

type linkRepository struct {
	collection *mongo.Collection
}

func NewLinkRepository(db *mongo.Database) LinkRepository {
	return &linkRepository{
		collection: db.Collection("links"),
	}
}

func (l *linkRepository) GetByCode(ctx context.Context, code string) (models.Link, error) {
	var link models.Link

	filter := bson.M{"short_code": code}

	if err := l.collection.FindOne(ctx, filter).Decode(&link); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return link, consts.ErrRecordNotFound
		}

		return link, consts.ErrInternal
	}

	return link, nil
}

func (l *linkRepository) GetById(ctx context.Context, id int64) (models.Link, error) {
	var link models.Link

	filter := bson.M{"_id": id}

	if err := l.collection.FindOne(ctx, filter).Decode(&link); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return link, consts.ErrRecordNotFound
		}

		return link, consts.ErrInternal
	}

	return link, nil
}

func (l *linkRepository) ExistsByID(ctx context.Context, id int64) (bool, error) {
	filter := bson.M{"_id": id}

	count, err := l.collection.CountDocuments(ctx, filter)

	if err != nil {
		log.Printf("error counting records by ID %d: %v", id, err)
		return false, consts.ErrInternal
	}

	return count > 0, nil
}

func (l *linkRepository) ExistsByShortCode(ctx context.Context, code string) (bool, error) {
	filter := bson.M{"short_code": code}

	count, err := l.collection.CountDocuments(ctx, filter)

	if err != nil {
		log.Printf("error counting records by ShortCode %s: %v", code, err)
		return false, consts.ErrInternal
	}

	return count > 0, nil
}
