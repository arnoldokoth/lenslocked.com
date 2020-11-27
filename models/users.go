package models

import (
	"errors"
	"regexp"
	"strings"

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
	hmac := hash.NewHMAC(hmacSecretKey)
	if err != nil {
		return nil, err
	}

	return &userService{
		UserDB: &userValidator{
			hmac:       hmac,
			emailRegex: regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`),
			UserDB:     ug,
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
	hmac       hash.HMAC
	emailRegex *regexp.Regexp
}

var _ UserDB = &userValidator{}

type userValFunc func(*User) error

func runUserValFuncs(user *User, fns ...userValFunc) error {
	for _, fn := range fns {
		if err := fn(user); err != nil {
			return err
		}
	}

	return nil
}

func (uv *userValidator) bcryptPassword(user *User) error {
	if user.Password == "" {
		return nil
	}

	passwordBytes := []byte(user.Password + userPasswordPepper)
	hashedBytes, err := bcrypt.GenerateFromPassword(passwordBytes, bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedBytes)
	user.Password = ""

	return nil
}

func (uv *userValidator) passwordMinLength(user *User) error {
	if user.Password == "" {
		return nil
	}

	if len(user.Password) < 8 {
		return ErrPasswordTooShort
	}

	return nil
}

func (uv *userValidator) passwordRequired(user *User) error {
	if user.Password == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) passwordHashRequired(user *User) error {
	if user.PasswordHash == "" {
		return ErrPasswordRequired
	}

	return nil
}

func (uv *userValidator) hmacRemember(user *User) error {
	if user.Remember == "" {
		return nil
	}

	user.RememberHash = uv.hmac.Hash(user.Remember)
	return nil
}

func (uv *userValidator) rememberHashRequired(user *User) error {
	if user.RememberHash == "" {
		return errors.New("models: remember hash required")
	}

	return nil
}

func (uv *userValidator) setRememberIfUnset(user *User) error {
	if user.Remember != "" {
		return nil
	}

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.Remember = token
	return nil
}

func (uv *userValidator) rememberMinBytes(user *User) error {
	if user.Remember == "" {
		return nil
	}

	n, err := rand.NBytes(user.Remember)
	if err != nil {
		return err
	}

	if n < 32 {
		return errors.New("models: remember token must be at least 32 bytes")
	}

	return nil
}

func (uv *userValidator) idGreaterThanZero(user *User) error {
	if user.ID <= 0 {
		return ErrInvalidID
	}

	return nil
}

func (uv *userValidator) normalizeEmail(user *User) error {
	user.EmailAddress = strings.ToLower(user.EmailAddress)
	user.EmailAddress = strings.TrimSpace(user.EmailAddress)
	return nil
}

func (uv *userValidator) requireEmail(user *User) error {
	if user.EmailAddress == "" {
		return ErrEmailRequired
	}

	return nil
}

func (uv *userValidator) emailFormat(user *User) error {
	if !uv.emailRegex.MatchString(user.EmailAddress) {
		return ErrEmailInvalid
	}

	return nil
}

func (uv *userValidator) emailIsAvailable(user *User) error {
	existing, err := uv.ByEmail(user.EmailAddress)
	if err == ErrNotFound {
		// email is not taken
		return nil
	}
	if err != nil {
		return err
	}
	if user.ID != existing.ID {
		return ErrEmailTaken
	}

	return nil
}

func (uv *userValidator) Create(user *User) error {
	err := runUserValFuncs(user, uv.passwordRequired, uv.passwordMinLength,
		uv.bcryptPassword, uv.passwordHashRequired, uv.setRememberIfUnset,
		uv.rememberMinBytes, uv.hmacRemember, uv.rememberHashRequired, uv.normalizeEmail, uv.requireEmail,
		uv.emailFormat, uv.emailIsAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Create(user)
}

func (uv *userValidator) Update(user *User) error {
	err := runUserValFuncs(user, uv.passwordMinLength, uv.bcryptPassword,
		uv.passwordHashRequired, uv.rememberMinBytes, uv.hmacRemember, uv.rememberHashRequired,
		uv.normalizeEmail, uv.requireEmail, uv.emailFormat, uv.emailIsAvailable)
	if err != nil {
		return err
	}

	return uv.UserDB.Update(user)
}

func (uv *userValidator) Delete(id uint) error {
	var user User
	user.ID = id
	err := runUserValFuncs(&user, uv.idGreaterThanZero)
	if err != nil {
		return err
	}

	return uv.UserDB.Delete(id)
}

func (uv *userValidator) ByID(id uint) (*User, error) {
	if id <= 0 {
		return nil, ErrInvalidID
	}

	return uv.UserDB.ByID(id)
}

func (uv *userValidator) ByEmail(emailAddress string) (*User, error) {
	user := User{EmailAddress: emailAddress}
	err := runUserValFuncs(&user, uv.normalizeEmail)
	if err != nil {
		return nil, err
	}

	return uv.UserDB.ByEmail(user.EmailAddress)
}

func (uv *userValidator) ByRemember(token string) (*User, error) {
	user := User{Remember: token}
	err := runUserValFuncs(&user, uv.hmacRemember)
	if err != nil {
		return nil, err
	}

	return uv.ByRemember(user.RememberHash)
}

func newUserGorm(connectionInfo string) (*userGorm, error) {
	db, err := gorm.Open("postgres", connectionInfo)
	if err != nil {
		return nil, err
	}

	db.LogMode(true)

	return &userGorm{
		db: db,
	}, nil
}

type userGorm struct {
	db *gorm.DB
}

var _ UserDB = &userGorm{}

// Create will create the provided user
func (ug *userGorm) Create(user *User) error {
	return ug.db.Create(user).Error
}

// Update ...
func (ug *userGorm) Update(user *User) error {
	return ug.db.Save(user).Error
}

// Delete ...
func (ug *userGorm) Delete(id uint) error {
	user := User{Model: gorm.Model{ID: id}}

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
func (ug *userGorm) ByRemember(rememberHash string) (*User, error) {
	var user User
	db := ug.db.Where("remember_hash = ?", rememberHash)

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
