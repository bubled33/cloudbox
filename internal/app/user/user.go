package user_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserService struct {
	queryRepo    user.QueryRepository
	commandRepo  user.CommandRepository
	eventService event_service.EventService
	uow          app.UnitOfWork
}

func NewUserService(
	queryRepo user.QueryRepository,
	commandRepo user.CommandRepository,
	eventService event_service.EventService,
	uow app.UnitOfWork,
) *UserService {
	return &UserService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
		uow:          uow,
	}
}

func (s *UserService) Create(ctx context.Context, rawEmail, rawDisplayName string) (*user.User, error) {
	var createdUser *user.User
	err := s.uow.Do(ctx, func(ctx context.Context) error {
		email, err := user.NewEmail(rawEmail)
		if err != nil {
			return err
		}

		displayName, err := user.NewDisplayName(rawDisplayName)
		if err != nil {
			return err
		}

		u := user.NewUser(email, displayName)

		if err := s.commandRepo.Save(ctx, u); err != nil {
			return err
		}

		createdUser = u
		return nil
	})

	if err != nil {
		return nil, err
	}

	eventType, payload := user.NewUserCreatedEvent(createdUser)
	_, _ = s.eventService.Create(eventType, payload)

	return createdUser, nil
}

func (s *UserService) Delete(ctx context.Context, id uuid.UUID) error {
	var u *user.User
	err := s.uow.Do(ctx, func(ctx context.Context) error {
		var err error
		u, err = s.queryRepo.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if u == nil {
			return user.ErrNotFound
		}

		return s.commandRepo.Delete(ctx, id)
	})

	if err != nil {
		return err
	}

	eventType, payload := user.NewUserDeletedEvent(id)
	_, _ = s.eventService.Create(eventType, payload)
	return nil
}

func (s *UserService) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	err := s.uow.Do(ctx, func(ctx context.Context) error {
		u, err := s.queryRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if u == nil {
			return user.ErrNotFound
		}

		u.VerifyEmail()
		return s.commandRepo.Save(ctx, u)
	})

	if err != nil {
		return err
	}

	eventType, payload := user.NewUserEmailVerifiedEvent(userID)
	_, _ = s.eventService.Create(eventType, payload)
	return nil
}

func (s *UserService) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	u, err := s.queryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	u, err := s.queryRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if u == nil {
		return nil, user.ErrNotFound
	}
	return u, nil
}

func (s *UserService) GetAll(ctx context.Context) ([]*user.User, error) {
	return s.queryRepo.GetAll(ctx)
}
