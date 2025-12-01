package repositories

import (
	"context"
	"errors"
	models "linkfast/url-projector/model"
	"linkfast/url-projector/utils/consts"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type LinkRepository interface {
	GetByCode(ctx context.Context, code string) (models.Link, error)
	GetById(ctx context.Context, id int64) (models.Link, error)
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
