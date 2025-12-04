package services

import (
	"context"
	"linkfast/read-api/models"
	"linkfast/read-api/repositories"
)

type LinkService interface {
	GetByCode(ctx context.Context, code string) (models.Link, error)
	ExistsByShortCode(ctx context.Context, code string) (bool, error)
	GetById(ctx context.Context, id int64) (models.Link, error)
	ExistsByID(ctx context.Context, id int64) (bool, error)
}

type linkService struct {
	repo repositories.LinkRepository
}

func NewLinkService(repo repositories.LinkRepository) LinkService {
	return &linkService{
		repo: repo,
	}
}

func (l *linkService) GetByCode(ctx context.Context, code string) (models.Link, error) {
	return l.repo.GetByCode(ctx, code)
}

func (l *linkService) ExistsByShortCode(ctx context.Context, code string) (bool, error) {
	return l.repo.ExistsByShortCode(ctx, code)
}

func (l *linkService) GetById(ctx context.Context, id int64) (models.Link, error) {
	return l.repo.GetById(ctx, id)
}

func (l *linkService) ExistsByID(ctx context.Context, id int64) (bool, error) {
	return l.repo.ExistsByID(ctx, id)
}
