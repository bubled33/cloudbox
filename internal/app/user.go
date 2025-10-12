package app

import (
	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserService struct {
	queryRepo    user.QueryRepository
	commandRepo  user.CommandRepository
	eventService EventService
}

func NewUserService(
	queryRepo user.QueryRepository,
	commandRepo user.CommandRepository,
	eventService EventService,
) *UserService {
	return &UserService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
	}
}

// --- Commands ---

func (s *UserService) Create(rawEmail, rawDisplayName string) (*user.User, error) {
	email, err := user.NewEmail(rawEmail)
	if err != nil {
		return nil, err
	}

	displayName, err := user.NewDisplayName(rawDisplayName)
	if err != nil {
		return nil, err
	}

	u := user.NewUser(email, displayName)

	if err := s.commandRepo.Save(u); err != nil {
		return nil, err
	}

	eventType, payload := user.NewUserCreatedEvent(u)
	_, _ = s.eventService.Create(eventType, payload)

	return u, nil
}

func (s *UserService) Delete(id uuid.UUID) error {
	u, err := s.queryRepo.GetByID(id)
	if err != nil {
		return err
	}
	if u == nil {
		return user.ErrNotFound
	}

	if err := s.commandRepo.Delete(id); err != nil {
		return err
	}

	eventType, payload := user.NewUserDeletedEvent(id)
	_, _ = s.eventService.Create(eventType, payload)

	return nil
}

func (s *UserService) VerifyEmail(userID uuid.UUID) error {
	u, err := s.queryRepo.GetByID(userID)
	if err != nil {
		return err
	}
	if u == nil {
		return user.ErrNotFound
	}

	u.VerifyEmail()

	if err := s.commandRepo.Save(u); err != nil {
		return err
	}

	eventType, payload := user.NewUserEmailVerifiedEvent(userID)
	_, _ = s.eventService.Create(eventType, payload)

	return nil
}

// --- Queries ---

func (s *UserService) GetByID(id uuid.UUID) (*user.User, error) {
	u, err := s.queryRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (s *UserService) GetByEmail(email string) (*user.User, error) {
	u, err := s.queryRepo.GetByEmail(email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (s *UserService) GetAll() ([]*user.User, error) {
	return s.queryRepo.GetAll()
}
