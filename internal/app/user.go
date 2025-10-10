package app

import (
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain"
)

type UserService struct {
	queryRepo   domain.UserQueryRepository
	commandRepo domain.UserCommandRepository
}

func NewUserService(
	queryRepo domain.UserQueryRepository,
	commandRepo domain.UserCommandRepository,
) *UserService {
	return &UserService{
		queryRepo:   queryRepo,
		commandRepo: commandRepo,
	}
}

// --- Commands ---

func (s *UserService) Create(email, displayName string) (*domain.User, error) {
	user := domain.NewUser(email, displayName)
	if err := s.commandRepo.Save(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *UserService) Delete(id uuid.UUID) error {
	user, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	return s.commandRepo.Delete(id)
}

func (s *UserService) VerifyEmail(userID uuid.UUID) error {
	user, err := s.queryRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if user == nil {
		return ErrUserNotFound
	}

	user.VerifyEmail()
	return s.commandRepo.Save(user)
}

// --- Queries ---

func (s *UserService) GetByID(id uuid.UUID) (*domain.User, error) {
	user, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetByEmail(email string) (*domain.User, error) {
	user, err := s.queryRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetAll() ([]*domain.User, error) {
	return s.queryRepo.GetAll()
}
