package authservices

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log/slog"
	rd "math/rand"
	"strings"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	"github.com/devathh/hive/backend/internal/common/auth"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/google/uuid"
)

var (
	ErrInternalServer        = errors.New("internal server error")
	ErrInvalidCode           = errors.New("invalid code")
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrInvalidToken          = errors.New("invalid token")
	ErrUserAlreadyRegistered = errors.New("user already registered")
)

type JWTManager interface {
	Generate(userID uuid.UUID) (string, error)
	Validate(tokenString string) (uuid.UUID, error)
}

type SMTPClient interface {
	SendCode(ctx context.Context, code int, to []string) error
	SendPassword(ctx context.Context, password string, to []string) error
}

type AuthService struct {
	cfg             *config.Config
	log             *slog.Logger
	userPersistence user.UserPersistenceRepository
	userCache       user.UserCacheRepository
	jwt             JWTManager
	smtpClient      SMTPClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	userPersistence user.UserPersistenceRepository,
	userCache user.UserCacheRepository,
	jwt JWTManager,
	smtpClient SMTPClient,
) *AuthService {
	return &AuthService{
		cfg:             cfg,
		log:             log,
		userPersistence: userPersistence,
		userCache:       userCache,
		jwt:             jwt,
		smtpClient:      smtpClient,
	}
}

// Prepare is a method that prepare user
// to further confirmation
func (as *AuthService) Prepare(ctx context.Context, req *requests.PrepareUser) (*responses.User, error) {
	email := user.Email(req.Email).Norm()
	if err := email.Validate(); err != nil {
		return nil, err
	}

	if _, err := as.userPersistence.GetByEmail(ctx, email); err == nil {
		return nil, ErrUserAlreadyRegistered
	} else if !errors.Is(err, user.ErrUserNotFound) {
		as.log.Error("failed to get user by email", slog.String("error", err.Error()))
		return nil, ErrInternalServer
	}

	newPassword, isGenerated := as.getPassword(req.Password)
	if isGenerated {
		req.Password = newPassword
	}

	user, err := user.New(
		req.Firstname,
		req.Lastname,
		user.Email(req.Email),
		req.Password,
	)
	if err != nil {
		as.log.Warn("trying to create invalid user", slog.String("error", err.Error()))
		return nil, err
	}

	code := rd.Intn(899999) + 100000
	if err := as.userCache.Prepare(ctx, user, newPassword, code); err != nil {
		as.log.Error("failed to set prepared user into cache", slog.String("error", err.Error()), slog.Int("code", code))
		return nil, ErrInternalServer
	}

	if err := as.sendCode(ctx, user.ConfirmedEmail(), code); err != nil {
		as.log.Error("failed to send code to user", slog.String("error", err.Error()))
		return nil, ErrInternalServer
	}

	return as.userToDTO(user), nil
}

// Confirm is a method that finding prepared user in cache
// and save his into db, also generate the jwt token
func (as *AuthService) Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error) {
	preparedUser, generatedPassword, err := as.userCache.GetPrepared(ctx, req.Code, req.UserID)
	if err != nil {
		as.log.Warn("trying to confirm unexisted user", slog.String("error", err.Error()))
		return nil, ErrInvalidCode
	}

	if generatedPassword != "" {
		if err := as.sendPassword(ctx, preparedUser.ConfirmedEmail(), generatedPassword); err != nil {
			as.log.Error("failed to send password to user", slog.String("error", err.Error()))
			return nil, ErrInternalServer
		}
	}

	savedUser, err := as.userPersistence.Save(ctx, preparedUser)
	if err != nil {
		safeErr, ok := as.handleRepoError(err)
		if !ok {
			as.log.Error("failed to handle repo error", slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	token, err := as.jwt.Generate(savedUser.ID())
	if err != nil {
		as.log.Error("failed to generate jwt token of user", slog.String("error", err.Error()))
		return nil, ErrInternalServer
	}

	return &responses.Token{
		Token: token,
	}, nil
}

// Login is a method that find user in db
// and compares the passwords, if it equal - generate jwt token
func (as *AuthService) Login(ctx context.Context, req *requests.Login) (*responses.Token, error) {
	email := user.Email(req.Email)
	if err := email.Validate(); err != nil {
		return nil, err
	}

	user, err := as.userPersistence.GetByEmail(ctx, email)
	if err != nil {
		safeErr, ok := as.handleRepoError(err)
		if !ok {
			as.log.Error("failed to get user by email", slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	if !user.PasswordHash().IsEqual(req.Password) {
		as.log.Warn("trying to login user with incorrect password")
		return nil, ErrInvalidCredentials
	}

	token, err := as.jwt.Generate(user.ID())
	if err != nil {
		as.log.Error("failed to generate jwt token of user", slog.String("error", err.Error()))
		return nil, ErrInternalServer
	}

	return &responses.Token{
		Token: token,
	}, nil
}

// Update is a method that updates the field of user
// and safe him into db
func (as *AuthService) Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, ErrInvalidToken
	}

	updUser, err := as.userPersistence.GetByID(ctx, userID)
	if err != nil {
		safeErr, ok := as.handleRepoError(err)
		if !ok {
			as.log.Error("failed to get user by id", slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	if err := as.applyUserChanges(updUser, req); err != nil {
		return nil, err
	}

	updatedUser, err := as.userPersistence.Update(ctx, updUser)
	if err != nil {
		safeErr, ok := as.handleRepoError(err)
		if !ok {
			as.log.Error("failed to update user by id", slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	return as.userToDTO(updatedUser), nil
}

// GetSelfUser is a method that find user in db (or cache)
// and returning him to user
func (as *AuthService) GetSelfUser(ctx context.Context) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, ErrInvalidToken
	}

	if cachedUser, err := as.userCache.Get(ctx, userID); err == nil {
		return as.userToDTO(cachedUser), nil
	} else if !errors.Is(err, user.ErrUserNotFound) {
		as.log.Warn("failed to get user from cache", slog.String("error", err.Error()))
	}

	user, err := as.userPersistence.GetByID(ctx, userID)
	if err != nil {
		safeErr, ok := as.handleRepoError(err)
		if !ok {
			as.log.Error("failed to get user by id", slog.String("error", err.Error()))
		}

		return nil, safeErr
	}

	if err := as.userCache.Set(ctx, user); err != nil {
		as.log.Warn("failed to set user into cache", slog.String("error", err.Error()))
	}

	return as.userToDTO(user), nil
}

func (as *AuthService) applyUserChanges(user *user.User, req *requests.UpdateUser) error {
	if req.Firstname != nil {
		if err := user.ChangeFirstname(*req.Firstname); err != nil {
			return err
		}
	}
	if req.Lastname != nil {
		if err := user.ChangeLastname(*req.Lastname); err != nil {
			return err
		}
	}
	if req.Password != nil {
		if err := user.ChangePassword(*req.Password); err != nil {
			return err
		}
	}

	if req.OnVacation != nil {
		if *req.OnVacation {
			return user.OnRest()
		}

		return user.OffRest()
	}

	return nil
}

func (as *AuthService) userToDTO(user *user.User) *responses.User {
	return &responses.User{
		ID:         user.ID(),
		Firstname:  user.Firstname(),
		Lastname:   user.Lastname(),
		Email:      user.ConfirmedEmail().String(),
		PaidUntil:  user.PaidUntil().Unix(),
		OnVacation: user.OnVacation(),
		TgChatID:   user.TgChatID(),
		Notification: responses.Notification{
			TelegramEnable: user.Notification().IsTelegramEnable(),
		},
		CreatedAt: user.CreatedAt().Unix(),
		UpdatedAt: user.UpdatedAt().Unix(),
	}
}

func (as *AuthService) sendPassword(ctx context.Context, email user.Email, password string) error {
	if err := as.smtpClient.SendPassword(ctx, password, []string{email.String()}); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) sendCode(ctx context.Context, email user.Email, code int) error {
	if err := as.smtpClient.SendCode(ctx, code, []string{email.String()}); err != nil {
		return err
	}

	return nil
}

// getPassword is a function that
// generate new password if it doesn't exist
func (as *AuthService) getPassword(password string) (string, bool) {
	password = strings.TrimSpace(password)
	if password != "" {
		return "", false
	}

	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", false
	}

	return hex.EncodeToString(buf), true
}

func (as *AuthService) handleRepoError(err error) (error, bool) {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return err, true
	case errors.Is(err, user.ErrUserNotFound),
		errors.Is(err, user.ErrUserAlreadyExists):
		return err, true
	}

	return ErrInternalServer, false
}

func (as *AuthService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(auth.KeyUserID)
	if raw == nil {
		return uuid.Nil, ErrInvalidToken
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, ErrInvalidToken
}
