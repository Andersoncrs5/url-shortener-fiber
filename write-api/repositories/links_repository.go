package repositories

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"linkfast/write-api/dtos"
	"linkfast/write-api/models"
	"linkfast/write-api/utils/consts"
	"strings"

	"github.com/godruoyi/go-snowflake"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

type LinkRepository interface {
	Create(dto dtos.CreateLinkDto) (*models.Links, error)
	GetByID(id int64) (*models.Links, error)
	GetByShotCode(code string) (*models.Links, error)
}

type linkRepository struct {
	db *gorm.DB
}

func NewLinkRepository(db *gorm.DB) LinkRepository {
	return &linkRepository{
		db: db,
	}
}

func (l *linkRepository) Create(dto dtos.CreateLinkDto) (*models.Links, error) {
	snowflake.SetMachineID(1)
	link := models.Links{}

	if err := copier.Copy(&link, &dto); err != nil {
		return nil, consts.ErrInternal
	}

	link.ID = int64(snowflake.ID())

	base, err := parseToBase64(link.ID)
	if err != nil {
		return nil, consts.ErrInternal
	}

	link.SHORT_CODE = base

	if err_db := l.db.Create(&link); err_db != nil {
		if errors.Is(err_db.Error, gorm.ErrRecordNotFound) {
			return nil, consts.ErrRecordNotFound
		}

		return nil, consts.ErrInternalDB
	}

	return &link, nil
}

func (l *linkRepository) GetByID(id int64) (*models.Links, error) {
	link := models.Links{}
	err := l.db.First(&link, id)

	if err.Error != nil {
		return nil, fmt.Errorf("Error the get link by id %v", err)
	}

	return &link, nil
}

func (l *linkRepository) GetByShotCode(code string) (*models.Links, error) {
	link := models.Links{}

	err := l.db.Where("short_code = ?", code).First(&link)

	if err.Error != nil {
		return nil, fmt.Errorf("Error the get link by id %v", err)
	}

	return &link, nil
}

func parseToBase64(id int64) (string, error) {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, id)
	if err != nil {
		return "", fmt.Errorf("Error the to write int64 to bytes: %v", err)
	}

	encodedID := base64.RawURLEncoding.EncodeToString(buf.Bytes())

	return strings.ReplaceAll(encodedID, "=", ""), nil
}
