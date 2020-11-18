package models

import (
	"errors"

	"github.com/arnoldokoth/lenslocked.com/hash"
	"github.com/arnoldokoth/lenslocked.com/rand"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"

	// initialize the postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// User ...
type User struct {
	gorm.Model
	Name         string `gorm:"type:varchar(50)"`
	EmailAddress string `gorm:"type:varchar(100);not null;unique_index"`
	Password     string `gorm:"-"`
	PasswordHash string `gorm:"not null"`
	Remember     string `gorm:"-"`
	RememberHash string `gorm:"not null;unique_index"`
}

var (
	// ErrNotFound is returned when a resource cannot be found
	// in the database
	ErrNotFound = errors.New("models: resource not found")
	// ErrInvalidID is returned when an invalid ID is provided
	// to the delete method
	ErrInvalidID = errors.New("models: ID provided as invalid")
	// ErrInvalidPassword ...
	ErrInvalidPassword = errors.New("models: invalid password provided")
)

const (
	userPasswordPepper = "5881f867b9078bd1d3ce164cc2466b13c4028ea12df14dfee9a6465e8c0b39ee"
	hmacSecretKey      = "4ed10e653ae1c61f0d842491c00eba6bd0f34fa5702f75abb5a12aaba721c2a9"
)

// UserDB ,,,
type UserDB interface {
	ByID(id uint) (*User, error)
	ByEmail(emailAddress string) (*User, error)
	ByRemember(token string) (*User, error)

	Create(user *User) error
	Update(user *User) error
	Delete(id uint) error

	Close() error

	AutoMigrate() error
	DestructiveReset() error
}

// UserService ,,,
type UserService interface {
	Authenticate(emailAddress, password string) (*User, error)
	UserDB
}

// NewUserService ...
func NewUserService(connectionInfo string) (UserService, error) {
	ug, err := newUserGorm(connectionInfo)
	if err != nil {
		return nil, err
	}

	return &userService{
		UserDB: &userValidator{
			UserDB: ug,
		},
	}, nil
}

// UserService ...
type userService struct {
	UserDB
}

var _ UserService = &userService{}

// Authenticate ...
func (us *userService) Authenticate(emailAddress, password string) (*User, error) {
	foundUser, err := us.ByEmail(emailAddress)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(foundUser.PasswordHash), []byte(password+userPasswordPepper))
	if err != nil {
		switch err {
		case bcrypt.ErrMismatchedHashAndPassword:
			return nil, ErrInvalidPassword
		default:
			return nil, err
		}
	}

	return foundUser, nil
}

type userValidator struct {
	UserDB
}

var _ UserDB = &userValidator{}

func (uv *userValidator) ByID(id uint) (*User, error) {
	if id <= 0 {
		return nil, errors.New("Invalid ID")
	}

	return uv.UserDB.ByID(id)
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	hmac := hash.NewHMAC(hmacSecretKey)
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &userGorm{
		db:   db,
		hmac: hmac,
	}, nil
}

type userGorm struct {
	db   *gorm.DB
	hmac hash.HMAC
}

var _ UserDB = &userGorm{}

// Create will create the provided user
func (ug *userGorm) Create(user *User) error {
	passwordBytes := []byte(user.Password + userPasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	if user.Remember == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.Remember = token
	}

	user.RememberHash = ug.hmac.Hash(user.Remember)

	return ug.db.Create(user).Error
}

// Update ...
func (ug *userGorm) Update(user *User) error {
	if user.Remember != "" {
		user.RememberHash = ug.hmac.Hash(user.Remember)
	}

	return ug.db.Save(user).Error
}

// Delete ...
func (ug *userGorm) Delete(id uint) error {
	if id == 0 {
		return ErrInvalidID
	}

	user := User{
		Model: gorm.Model{
			ID: id,
		},
	}

	return ug.db.Delete(&user).Error
}

// ByID looks up a user using the provided id
func (ug *userGorm) ByID(id uint) (*User, error) {
	var user User
	db := ug.db.Where("id = ?", id)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}

	return &user, err
}

// ByEmail ...
func (ug *userGorm) ByEmail(emailAddress string) (*User, error) {
	var user User
	db := ug.db.Where("email_address = ?", emailAddress)
	err := first(db, &user)
	if err != nil {
		return nil, err
	}

	return &user, err
}

// ByRemember ...
func (ug *userGorm) ByRemember(token string) (*User, error) {
	var user User
	tokenHash := ug.hmac.Hash(token)
	db := ug.db.Where("remember_hash = ?", tokenHash)

	err := first(db, &user)
	if err != nil {
		return nil, err
	}

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
func (ug *userGorm) AutoMigrate() error {
	return ug.db.AutoMigrate(&User{}).Error
}

// DestructiveReset ...
func (ug *userGorm) DestructiveReset() error {
	if err := ug.db.DropTableIfExists(&User{}).Error; err != nil {
		return err
	}

	return ug.AutoMigrate()
}

// Close the UserService database connection
func (ug *userGorm) Close() error {
	return ug.db.Close()
}
