package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// Gallery ...
type Gallery struct {
	gorm.Model
	UserID uint   `gorm:"not_null;index"`
	Title  string `gorm:"not_null"`
}

// GalleryDB ...
type GalleryDB interface {
	Create(gallery *Gallery) error
}

// GalleryService ...
type GalleryService interface {
	GalleryDB
}

// NewGalleryService ...
func NewGalleryService(db *gorm.DB) GalleryService {
	return &galleryService{
		GalleryDB: &galleryValidator{
			&galleryGorm{db},
		},
	}
}

type galleryService struct {
	GalleryDB
}

type galleryValFunc func(*Gallery) error

func runGalleryValFuncs(gallery *Gallery, fns ...galleryValFunc) error {
	for _, fn := range fns {
		err := fn(gallery)
		if err != nil {
			return err
		}
	}

	return nil
}

type galleryValidator struct {
	GalleryDB
}

func (gv *galleryValidator) requireTitle(gallery *Gallery) error {
	if strings.TrimSpace(gallery.Title) == "" {
		return ErrTitleRequired
	}

	return nil
}

func (gv *galleryValidator) requireUserID(gallery *Gallery) error {
	if gallery.UserID <= 0 {
		return ErrUserIDRequired
	}

	return nil
}

func (gv *galleryValidator) Create(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery, gv.requireTitle, gv.requireUserID)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Create(gallery)
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}
