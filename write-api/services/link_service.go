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
	Create(dto dtos.CreateLinkDto) (*models.Links, error, int)
	GetByID(id int64) (*models.Links, error, int)
	ExistsByID(id int64) (bool, error)
	GetByShotCode(code string) (*models.Links, error, int)
	ExistsByShotCode(code string) (bool, error, int)
	Delete(link *models.Links) (error, int)
}

type linkService struct {
	repo repositories.LinkRepository
}

func NewUserService(repo repositories.LinkRepository) LinkService {
	return &linkService{
		repo: repo,
	}
}

func (l *linkService) Create(dto dtos.CreateLinkDto) (*models.Links, error, int) {
	link := models.Links{}

	if err := copier.Copy(dto, &link); err != nil {
		log.Printf("Error the copier CreateLinkDto to Links: %v", err)
		return nil, consts.ErrInternal, 500
	}

	return l.repo.Create(link)
}

func (l *linkService) GetByID(id int64) (*models.Links, error, int) {
	return l.repo.GetByID(id)
}

func (l *linkService) ExistsByID(id int64) (bool, error) {
	return l.repo.ExistsByID(id)
}

func (l *linkService) GetByShotCode(code string) (*models.Links, error, int) {
	return l.repo.GetByShotCode(code)
}

func (l *linkService) ExistsByShotCode(code string) (bool, error, int) {
	return l.repo.ExistsByShotCode(code)
}

func (l *linkService) Delete(link *models.Links) (error, int) {
	return l.repo.Delete(link)
}
