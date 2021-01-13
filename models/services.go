package models

import "github.com/jinzhu/gorm"

// ServicesConfig ...
type ServicesConfig func(*Services) error

// WithGorm ...
func WithGorm(dialect, connectionInfo string) ServicesConfig {
	return func(s *Services) error {
		db, err := gorm.Open(dialect, connectionInfo)
		if err != nil {
			return err
		}
		s.db = db
		return nil
	}
}

// WithLogMode ...
func WithLogMode(mode bool) ServicesConfig {
	return func(s *Services) error {
		s.db.LogMode(mode)
		return nil
	}
}

// WithUser ...
func WithUser(hmacKey, pepper string) ServicesConfig {
	return func(s *Services) error {
		s.User = NewUserService(s.db, hmacKey, pepper)
		return nil
	}
}

// WithGallery ...
func WithGallery() ServicesConfig {
	return func(s *Services) error {
		s.Gallery = NewGalleryService(s.db)
		return nil
	}
}

// WithImage ...
func WithImage() ServicesConfig {
	return func(s *Services) error {
		s.Image = NewImageService()
		return nil
	}
}

// NewServices ...
func NewServices(cfgs ...ServicesConfig) (*Services, error) {
	var s Services
	for _, cfg := range cfgs {
		if err := cfg(&s); err != nil {
			return nil, err
		}
	}

	return &s, nil
}

// Services ...
type Services struct {
	Gallery GalleryService
	User    UserService
	Image   ImageService
	db      *gorm.DB
}

// AutoMigrate creates the defined models in the models package
func (s *Services) AutoMigrate() error {
	return s.db.AutoMigrate(&User{}, &Gallery{}).Error
}

// DestructiveReset drops all tables and recreates them
func (s *Services) DestructiveReset() error {
	err := s.db.DropTableIfExists(&User{}, &Gallery{}).Error
	if err != nil {
		return err
	}

	return s.AutoMigrate()
}

// Close the database connection
func (s *Services) Close() error {
	return s.db.Close()
}
