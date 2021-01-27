package models

import (
	"strings"

	"github.com/jinzhu/gorm"
)

// Gallery ...
type Gallery struct {
	gorm.Model
	UserID uint    `gorm:"not_null;index"`
	Title  string  `gorm:"not_null"`
	Images []Image `gorm:"-"`
}

// ImagesSplitN ...
func (g *Gallery) ImagesSplitN(n int) [][]Image {
	ret := make([][]Image, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]Image, 0)
	}

	for i, img := range g.Images {
		bucket := i % n
		ret[bucket] = append(ret[bucket], img)
	}

	return ret
}

// GalleryDB ...
type GalleryDB interface {
	Create(gallery *Gallery) error
	Update(gallery *Gallery) error
	Delete(id uint) error

	ByID(id uint) (*Gallery, error)
	ByUserID(id uint) ([]Gallery, error)
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

func (gv *galleryValidator) Update(gallery *Gallery) error {
	err := runGalleryValFuncs(gallery, gv.requireTitle, gv.requireUserID)
	if err != nil {
		return err
	}

	return gv.GalleryDB.Update(gallery)
}

func (gv *galleryValidator) Delete(id uint) error {
	if id <= 0 {
		return ErrInvalidID
	}

	return gv.GalleryDB.Delete(id)
}

var _ GalleryDB = &galleryGorm{}

type galleryGorm struct {
	db *gorm.DB
}

func (gg *galleryGorm) Create(gallery *Gallery) error {
	return gg.db.Create(gallery).Error
}

func (gg *galleryGorm) Update(gallery *Gallery) error {
	return gg.db.Save(gallery).Error
}

func (gg *galleryGorm) Delete(id uint) error {
	gallery := Gallery{Model: gorm.Model{ID: id}}
	return gg.db.Delete(&gallery).Error
}

func (gg *galleryGorm) ByID(id uint) (*Gallery, error) {
	var gallery Gallery
	db := gg.db.Where("id = ?", id)
	err := first(db, &gallery)
	return &gallery, err
}

func (gg *galleryGorm) ByUserID(id uint) ([]Gallery, error) {
	var galleries []Gallery
	err := gg.db.Where("user_id = ?", id).Find(&galleries).Error
	if err != nil {
		return nil, err
	}

	return galleries, nil
}
