package models

import (
	"errors"

	"github.com/jinzhu/gorm"
	// initialize the postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// User ...
type User struct {
	gorm.Model
	Name         string `gorm:"type:varchar(50)"`
	EmailAddress string `gorm:"type:varchar(100);not null;unique_index"`
}

var (
	// ErrNotFound is returned when a resource cannot be found
	// in the database
	ErrNotFound = errors.New("models: resource not found")
	// ErrInvalidID is returned when an invalid ID is provided
	// to the delete method
	ErrInvalidID = errors.New("models: ID provided as invalid")
)

// NewUserService ...
func NewUserService(connectionInfo string) (*UserService, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &UserService{
		db: db,
	}, nil
}

// UserService ...
type UserService struct {
	db *gorm.DB
}

// Create will create the provided user
func (us *UserService) Create(user *User) error {
	return us.db.Create(user).Error
}

// Update ...
func (us *UserService) Update(user *User) error {
	return us.db.Save(user).Error
}

// Delete ...
func (us *UserService) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{
		Model: gorm.Model{
			ID: id,
		},
	}

	return us.db.Delete(&user).Error
}

// ByID looks up a user using the provided id
func (us *UserService) ByID(id uint) (*User, error) {
	var user User
	db := us.db.Where("id = ?", id)
	err := first(db, &user)

	return &user, err
}

// ByEmail ...
func (us *UserService) ByEmail(emailAddress string) (*User, error) {
	var user User
	db := us.db.Where("email_address = ?", emailAddress)
	err := first(db, &user)

	return &user, err
}

func first(db *gorm.DB, dst interface{}) error {
	err := db.First(dst).Error
	if err == gorm.ErrRecordNotFound {
		return ErrNotFound
	}

	return err
}

// AutoMigrate creates the defined models in the models package
func (us *UserService) AutoMigrate() error {
	return us.db.AutoMigrate(&User{}).Error
}

// DestructiveReset ...
func (us *UserService) DestructiveReset() error {
	if err := us.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}

	return us.AutoMigrate()
}

// Close the UserService database connection
func (us *UserService) Close() error {
	return us.db.Close()
}
