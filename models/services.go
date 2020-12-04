package models

import "github.com/jinzhu/gorm"

// NewServices ...
func NewServices(connectionInfo string) (*Services, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &Services{
		db:      db,
		User:    NewUserService(db),
		Gallery: NewGalleryService(db),
	}, nil
}

// Services ...
type Services struct {
	Gallery GalleryService
	User    UserService
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
