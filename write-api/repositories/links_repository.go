package repositories

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"linkfast/write-api/models"
	"linkfast/write-api/utils/consts"
	"log"
	"strings"

	"github.com/godruoyi/go-snowflake"
	"gorm.io/gorm"
)

func init() {
	snowflake.SetMachineID(1)
	log.Println("Snowflake Machine ID set to 1.")
}

type LinkRepository interface {
	Create(link models.Links) (*models.Links, error)
	GetByID(id int64) (*models.Links, error)
	GetByShotCode(code string) (*models.Links, error)
	ExistsByShotCode(code string) (bool, error)
	Delete(link *models.Links) error
	ExistsByID(id int64) (bool, error)
}

type linkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepository{
		db: db,
	}
}

func (l *linkRepository) Create(link models.Links) (*models.Links, error) {
	link.ID = int64(snowflake.ID())

	base, err := parseToBase64(link.ID)
	if err != nil {
		return nil, consts.ErrInternal
	}

	link.SHORT_CODE = base

	if err_db := l.db.Create(&link); err_db != nil {
		return nil, consts.ErrInternalDB
	}

	return &link, nil
}

func (l *linkRepository) GetByID(id int64) (*models.Links, error) {
	link := models.Links{}
	result := l.db.First(&link, id)

	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, consts.ErrRecordNotFound
		}

		return nil, consts.ErrInternalDB
	}

	return &link, nil
}

func (l *linkRepository) ExistsByID(id int64) (bool, error) {
	var count int64

	result := l.db.Where(&models.Links{ID: id}).Count(&count)

	if result.Error != nil {
		log.Fatalf("error counting records: %v", result.Error)
		return false, consts.ErrInternalDB
	}

	return count > 0, nil
}

func (l *linkRepository) GetByShotCode(code string) (*models.Links, error) {
	link := models.Links{}

	result := l.db.Where("short_code = ?", code).First(&link)

	if result.Error != nil {

		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, consts.ErrRecordNotFound
		}

		return nil, consts.ErrInternalDB
	}

	return &link, nil
}

func (l *linkRepository) ExistsByShotCode(code string) (bool, error) {
	var count int64

	result := l.db.Model(&models.Links{}).Where("short_code = ?", code).Count(&count)

	if result.Error != nil {
		log.Fatalf("error counting records: %v", result.Error)
		return false, consts.ErrInternalDB
	}

	return count > 0, nil
}

func (l *linkRepository) Delete(link *models.Links) error {
	result := l.db.Delete(link)

	if result.Error != nil {
		log.Fatalf("Error the delete link %v", result.Error.Error())
		return consts.ErrInternalDB
	}

	if result.RowsAffected == 0 {
		return consts.ErrRecordNotFound
	}

	return nil
}

func parseToBase64(id int64) (string, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, id)
	if err != nil {
		log.Fatalf("Error the to write int64 to bytes: %v", err)
		return "", consts.ErrInternal
	}

	encodedID := base64.RawURLEncoding.EncodeToString(buf.Bytes())

	return strings.ReplaceAll(encodedID, "=", ""), nil
}
