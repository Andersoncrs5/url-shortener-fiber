package repositories

import (
	"context"
	"errors"
	models "linkfast/url-projector/model"
	"linkfast/url-projector/utils/consts"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type LinkRepository interface {
	GetByCode(ctx context.Context, code string) (models.Link, error)
	GetById(ctx context.Context, id int64) (models.Link, error)
	Delete(ctx context.Context, id int64) error
	Create(ctx context.Context, link models.Link) (models.Link, error)
	Upsert(ctx context.Context, cdcLink *models.Link) (*models.Link, error)
}

type linkRepository struct {
	collection *mongo.Collection
}

func NewLinkRepository(db *mongo.Database) LinkRepository {
	return &linkRepository{
		collection: db.Collection("links"),
	}
}

func (l *linkRepository) Upsert(ctx context.Context, cdcLink *models.Link) (*models.Link, error) {

	filter := bson.M{"_id": cdcLink.ID}
	mongoLink := models.Link{
		ID:         cdcLink.ID,
		SHORT_CODE: cdcLink.SHORT_CODE,
		LONG_URL:   cdcLink.LONG_URL,
		CreatedAt:  cdcLink.CreatedAt,
		ExpiresAt:  cdcLink.ExpiresAt,
	}

	opts := options.Replace().SetUpsert(true)

	_, err := l.collection.ReplaceOne(ctx, filter, mongoLink, opts)
	if err != nil {
		return nil, consts.ErrInternal
	}

	return &mongoLink, nil
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

func (l *linkRepository) Delete(ctx context.Context, id int64) error {
	filter := bson.M{"_id": id}
	result, err := l.collection.DeleteOne(ctx, filter)

	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return consts.ErrRecordNotFound
	}

	return nil
}

func (l *linkRepository) Create(ctx context.Context, link models.Link) (models.Link, error) {
	if _, err := l.collection.InsertOne(ctx, link); err != nil {
		return link, consts.ErrInternal
	}

	return link, nil
}
