package repositories

import "go.mongodb.org/mongo-driver/mongo"

type LinkRepository interface {
}

type linkRepository struct {
	collection *mongo.Collection
}

func NewLinkRepository(db *mongo.Database) LinkRepository {
	return &linkRepository{
		collection: db.Collection("links"),
	}
}

func (l *linkRepository) GetByCode(code string) {

}
