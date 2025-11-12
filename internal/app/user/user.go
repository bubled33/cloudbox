package user_service

import (
	"context"

	"github.com/google/uuid"
	"github.com/yourusername/cloud-file-storage/internal/app"
	event_service "github.com/yourusername/cloud-file-storage/internal/app/event"
	file_service "github.com/yourusername/cloud-file-storage/internal/app/file"
	session_service "github.com/yourusername/cloud-file-storage/internal/app/session"
	"github.com/yourusername/cloud-file-storage/internal/domain/user"
)

type UserService struct {
	queryRepo    user.QueryRepository
	commandRepo  user.CommandRepository
	eventService *event_service.EventService
	uow          app.UnitOfWork
	fileService  *file_service.FileService
	sessionSrv   *session_service.SessionService
}

func NewUserService(
	queryRepo user.QueryRepository,
	commandRepo user.CommandRepository,
	eventService *event_service.EventService,
	uow app.UnitOfWork,
	fileService *file_service.FileService,
	sessionSrv *session_service.SessionService,
) *UserService {
	return &UserService{
		queryRepo:    queryRepo,
		commandRepo:  commandRepo,
		eventService: eventService,
		uow:          uow,
		fileService:  fileService,
		sessionSrv:   sessionSrv,
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
	_, _ = s.eventService.Create(ctx, eventType, payload)

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
	_, _ = s.eventService.Create(ctx, eventType, payload)
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
	_, _ = s.eventService.Create(ctx, eventType, payload)
	return nil
}

// UpdateProfile обновляет профиль пользователя (email и displayName)
func (s *UserService) UpdateProfile(ctx context.Context, userID uuid.UUID, rawEmail, rawDisplayName string) (*user.User, error) {
	var updatedUser *user.User
	err := s.uow.Do(ctx, func(ctx context.Context) error {
		u, err := s.queryRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if u == nil {
			return user.ErrNotFound
		}

		if rawEmail != u.Email.String() {
			email, err := user.NewEmail(rawEmail)
			if err != nil {
				return err
			}
			u.Email = email
			u.IsEmailVerified = false
		}

		displayName, err := user.NewDisplayName(rawDisplayName)
		if err != nil {
			return err
		}
		u.Rename(displayName)

		if err := s.commandRepo.Save(ctx, u); err != nil {
			return err
		}

		updatedUser = u
		return nil
	})

	if err != nil {
		return nil, err
	}

	eventType, payload := user.NewUserUpdatedEvent(updatedUser)
	_, _ = s.eventService.Create(ctx, eventType, payload)

	return updatedUser, nil
}

// DeleteAccount удаляет аккаунт пользователя и все связанные данные
func (s *UserService) DeleteAccount(ctx context.Context, userID uuid.UUID) error {
	err := s.uow.Do(ctx, func(ctx context.Context) error {
		u, err := s.queryRepo.GetByID(ctx, userID)
		if err != nil {
			return err
		}
		if u == nil {
			return user.ErrNotFound
		}

		files, err := s.fileService.GetAllByUser(ctx, userID)
		if err != nil {
			return err
		}
		for _, f := range files {
			if f != nil {
				if err := s.fileService.Delete(ctx, f.ID); err != nil {
					return err
				}
			}
		}

		if err := s.sessionSrv.RevokeAllForUser(ctx, userID); err != nil {
			return err
		}

		return s.commandRepo.Delete(ctx, userID)
	})

	if err != nil {
		return err
	}

	eventType, payload := user.NewUserDeletedEvent(userID)
	_, _ = s.eventService.Create(ctx, eventType, payload)

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
