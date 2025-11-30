package services

import (
	"linkfast/write-api/dtos"
	"linkfast/write-api/models"
	"linkfast/write-api/repositories"
	"linkfast/write-api/utils/consts"
	"log"

	"github.com/jinzhu/copier"
)

type LinkService interface {
	Create(dto dtos.CreateLinkDto) (*models.Links, error)
	GetByID(id int64) (models.Links, error)
	ExistsByID(id int64) (bool, error)
	GetByShotCode(code string) (*models.Links, error)
	ExistsByShotCode(code string) (bool, error)
	Delete(link *models.Links) error
}

type linkService struct {
	repo repositories.LinkRepository
}

func NewUserService(repo repositories.LinkRepository) LinkService {
	return &linkService{
		repo: repo,
	}
}

func (l *linkService) Create(dto dtos.CreateLinkDto) (*models.Links, error) {
	link := new(models.Links)

	if err := copier.Copy(link, dto); err != nil {
		log.Printf("Error copying CreateLinkDto to Links model: %v", err)
		return nil, consts.ErrInternal
	}

	return l.repo.Create(*link)
}

func (l *linkService) GetByID(id int64) (models.Links, error) {
	return l.repo.GetByID(id)
}

func (l *linkService) ExistsByID(id int64) (bool, error) {
	return l.repo.ExistsByID(id)
}

func (l *linkService) GetByShotCode(code string) (*models.Links, error) {
	return l.repo.GetByShotCode(code)
}

func (l *linkService) ExistsByShotCode(code string) (bool, error) {
	return l.repo.ExistsByShotCode(code)
}

func (l *linkService) Delete(link *models.Links) error {
	return l.repo.Delete(link)
}
