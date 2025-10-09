package app

import (
	"errors"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

var ErrUserNotFound = errors.New("user not found")

type UserService struct {
	userRepo domain.UserRepository
}

func NewUserService(userRepo domain.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) Create(email, displayName string) (*domain.User, error) {
	user := domain.NewUser(email, displayName)
	if err := s.userRepo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) GetByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetByEmail(email string) (*domain.User, error) {
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetAll() ([]*domain.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) Delete(id uuid.UUID) error {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.userRepo.Delete(id)
}

func (s *UserService) VerifyEmail(userID uuid.UUID) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.VerifyEmail()
	return s.userRepo.Save(user)
}
