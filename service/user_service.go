package service

import (
	"context"
	"errors"
	"fmt"

	"example.com/myproject/entity"
	"example.com/myproject/repository"

	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrNoUser = errors.New("user doesn't exist")
	ErrWrongPassword = errors.New("incorrect password")
	ErrUserExists = errors.New("user already exists")
)

type UserService struct {
	repo      repository.UserRepository
	jwt *AuthService
}

func NewUserService(repo repository.UserRepository, jwt *AuthService) *UserService {
	return &UserService{
		repo: repo,
		jwt: jwt,
	}
}

type RegistrationRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func (s *UserService) RegisterUser(ctx context.Context, r RegistrationRequest) (string, error) {

	err := s.validateRegistrationRequest(r)
	if err != nil {
		return "", fmt.Errorf("registration request invalid: %w", err)
	}
	exists := s.checkIfUserExists(ctx, r.Username)
	if exists {
		return "", ErrUserExists
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(r.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed hashing password: %w", err)
	}
	user := entity.NewUserEntity(r.Username, r.Email, string(hashedPassword))
	err = s.repo.Save(ctx, user)
	if err != nil {
		return "", fmt.Errorf("failed saving user %q to the database: %w", user.Username, err)
	}

	token, err := s.jwt.GenerateToken(user.Id, user.Username)
	if err != nil {
		return "", fmt.Errorf("failed generating jwt token: %w", err)
	}
	return token, nil
}

func (s *UserService) LoginUser(ctx context.Context, r LoginRequest) (string, error) {
	user, err := s.repo.GetByUsername(ctx, r.Username)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return "", ErrNoUser
		}
		return "", errors.New("unknown error while logging in")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(r.Password))
	if err != nil {
		return "", ErrWrongPassword
	}

	token, err := s.jwt.GenerateToken(user.Id, user.Username)
	if err != nil {
		return "", fmt.Errorf("failed generating jwt token: %w", err)
	}

	return token, nil 
}

func (s *UserService) validateRegistrationRequest(r RegistrationRequest) error {
	err := ValidateEmail(r.Email)
	if err != nil {
		return err
	}
	err = ValidateUsername(r.Username)
	if err != nil {
		return err
	}
	err = ValidatePassword(r.Password)
	if err != nil {
		return err
	}
	return nil
}

func (s *UserService) checkIfUserExists(ctx context.Context, username string) bool {
	_, err := s.repo.GetByUsername(ctx, username)
	return err != mongo.ErrNoDocuments
}