package authservices

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/devathh/hive/backend/internal/common/application/dtos/requests"
	"github.com/devathh/hive/backend/internal/common/application/dtos/responses"
	contextvalue "github.com/devathh/hive/backend/internal/common/context"
	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/registration"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/tgbinding"
	"github.com/devathh/hive/backend/internal/modules/auth/domain/user"
	"github.com/devathh/hive/backend/pkg/errs"
	"github.com/google/uuid"
)

const (
	ResetPassword = `
Ваша ссылка для сброса пароля: %s?token=%s
`
)

type JWTManager interface {
	Generate(userID uuid.UUID, ttl time.Duration) (string, error)
	GenerateResetToken(userID uuid.UUID, ttl time.Duration) (string, error)
	Validate(tokenString string) (uuid.UUID, error)
	ValidateReset(resetToken string) (uuid.UUID, error)
}

type SMTPClient interface {
	Notify(ctx context.Context, body string, subject string, to []string) error
	NotifyHTML(ctx context.Context, body string, subject string, to []string) error
}

type AuthService struct {
	cfg                     *config.Config
	log                     *slog.Logger
	userPersistence         user.UserPersistenceRepository
	userCache               user.UserCacheRepository
	preparedCache           user.PreparedCacheRepository
	registrationPersistence registration.RegistrationPersistenceRepository
	tgBindingCache          tgbinding.TgBindingCacheRepository
	jwt                     JWTManager
	smtpClient              SMTPClient
}

func New(
	cfg *config.Config,
	log *slog.Logger,
	userPersistence user.UserPersistenceRepository,
	userCache user.UserCacheRepository,
	preparedCache user.PreparedCacheRepository,
	registrationPersistence registration.RegistrationPersistenceRepository,
	tgBindingCache tgbinding.TgBindingCacheRepository,
	jwt JWTManager,
	smtpClient SMTPClient,
) *AuthService {
	return &AuthService{
		cfg:                     cfg,
		log:                     log,
		userPersistence:         userPersistence,
		userCache:               userCache,
		preparedCache:           preparedCache,
		registrationPersistence: registrationPersistence,
		tgBindingCache:          tgBindingCache,
		jwt:                     jwt,
		smtpClient:              smtpClient,
	}
}

// Prepare is a method that prepare user
// to further confirmation
func (as *AuthService) Prepare(ctx context.Context, req *requests.PrepareUser) (*responses.User, error) {
	fp, err := as.getFingerprint(ctx)
	if err != nil {
		return nil, err
	}

	if err := as.checkFingerprint(ctx, fp); err != nil {
		return nil, err
	}

	email, err := as.ensureEmailIsNew(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	generated, isGenerated := as.getPasswordOrGenerate(req.Password)
	if isGenerated {
		req.Password = generated
	}

	user, err := as.buildNewUser(req, email)
	if err != nil {
		return nil, err
	}

	code, err := as.storePreparedUserCache(ctx, user, generated)
	if err != nil {
		return nil, err
	}

	go as.sendConfirmationCode(context.Background(), user.ConfirmedEmail(), code)
	return as.userToDTO(user), nil
}

func (as *AuthService) ResendCode(ctx context.Context, req *requests.ResendCode) (*responses.CodeResent, error) {
	email, err := as.validateEmail(req.Email)
	if err != nil {
		return nil, err
	}

	preparedUser, code, err := as.getPreparedUser(ctx, email)
	if err != nil {
		return nil, err
	}

	if err := as.checkLastSendTime(ctx, preparedUser.ConfirmedEmail()); err != nil {
		return nil, err
	}

	newSendTime := time.Now().Add(as.cfg.Service.SendingDuration).UTC()
	if err := as.storeLastSendTime(ctx, newSendTime, preparedUser.ConfirmedEmail()); err != nil {
		return nil, err
	}

	go as.sendConfirmationCode(context.Background(), email, code)
	return &responses.CodeResent{
		NextSendAt: newSendTime.Unix(),
	}, nil
}

// Confirm is a method that finding prepared user in cache
// and save his into db, also generate the jwt token
func (as *AuthService) Confirm(ctx context.Context, req *requests.ConfirmUser) (*responses.Token, error) {
	fp, err := as.getFingerprint(ctx)
	if err != nil {
		return nil, err
	}

	email, err := as.validateEmail(req.Email)
	if err != nil {
		return nil, err
	}

	preparedUser, err := as.resolvePreparedUser(ctx, req.Code, email)
	if err != nil {
		return nil, err
	}

	if err = as.storePreparedUserDB(ctx, preparedUser); err != nil {
		return nil, err
	}

	if err := as.clearPreparedCache(ctx, preparedUser.ConfirmedEmail()); err != nil {
		as.log.Warn("failed to clear prepared user's cache",
			slog.String("error", err.Error()),
		)
	}

	if err := as.saveFingerprint(ctx, fp, preparedUser.ID()); err != nil {
		as.log.Warn("failed to save fingerprint",
			slog.String("error", err.Error()),
		)
	}

	return as.generateToken(preparedUser.ID())
}

func (as *AuthService) ForgetPassword(ctx context.Context, req *requests.ForgetPassword) error {
	email, err := as.validateEmail(req.Email)
	if err != nil {
		return err
	}

	user, err := as.getUserByEmailFromPersistence(ctx, email)
	if err != nil {
		return err
	}

	token, err := as.generateChangeToken(user)
	if err != nil {
		return err
	}

	go as.sendChangeURL(context.Background(), user.ConfirmedEmail(), token)
	return nil
}

func (as *AuthService) ChangePassword(ctx context.Context, req *requests.ChangePassword) error {
	userID, err := as.validateResetToken(req.ResetToken)
	if err != nil {
		return err
	}

	user, err := as.getUserFromPersistence(ctx, userID)
	if err != nil {
		return err
	}

	if err := as.changePassword(user, req.NewPassword); err != nil {
		return err
	}

	if err = as.updateUserStore(ctx, user); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) UnbindTg(ctx context.Context) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	user, err := as.getUserFromPersistence(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := as.unbindTg(user); err != nil {
		return nil, err
	}

	if err = as.updateUserStore(ctx, user); err != nil {
		return nil, err
	}

	if err := as.clearUserCache(ctx, user.ID()); err != nil {
		as.log.Warn("failed to clear user cache",
			slog.String("error", err.Error()),
		)
	}

	return as.userToDTO(user), nil
}

func (as *AuthService) unbindTg(updUser *user.User) error {
	if err := updUser.ChangeNotification(user.NewNotification(false)); err != nil {
		return errs.NewErrorWithKind(err, errs.KindInternal)
	}

	if err := updUser.ChangeTgChatID(0); err != nil {
		return errs.NewErrorWithKind(err, errs.KindInternal)
	}

	return nil
}

func (as *AuthService) changePassword(user *user.User, newPassword string) error {
	if err := user.ChangePassword(newPassword); err != nil {
		return errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return nil
}

func (as *AuthService) sendChangeURL(ctx context.Context, email user.Email, token string) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Montserrat, Arial, sans-serif; background-color: #E6E6E6; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; background-color: #FAFAFA; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
        .header { text-align: center; padding: 10px 0; border-bottom: 2px solid #FFB61A; }
        .header h1 { color: #1A1A1A; margin: 0; font-weight: 600; }
        .content { padding: 20px 0; text-align: center; }
        .content p { font-size: 16px; color: #1A1A1A; }
        .button { display: inline-block; padding: 12px 24px; background-color: #FFB61A; color: #1A1A1A; text-decoration: none; border-radius: 8px; font-weight: bold; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Восстановление пароля</h1>
        </div>
        <div class="content">
            <p>Для сброса пароля нажмите на кнопку ниже:</p>
            <a href="%s?token=%s" class="button">Сбросить пароль</a>
            <p>Если вы не запрашивали восстановление, проигнорируйте это письмо.</p>
        </div>
    </div>
</body>
</html>
`, as.cfg.Service.ChangePasswordURL, token)

	if err := as.smtpClient.NotifyHTML(ctxTimeout, htmlBody, "Восстановление пароля", []string{email.String()}); err != nil {
		as.log.Error("failed to send reset url for user",
			slog.String("error", err.Error()),
		)
	}
}

func (as *AuthService) generateChangeToken(user *user.User) (string, error) {
	token, err := as.jwt.GenerateResetToken(user.ID(), as.cfg.Service.ChangePasswordTTL)
	if err != nil {
		as.log.Error("failed to generate jwt token",
			slog.String("error", err.Error()),
			slog.String("user_id", user.ID().String()),
		)

		return "", errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return token, nil
}

func (as *AuthService) validateResetToken(resetToken string) (uuid.UUID, error) {
	userID, err := as.jwt.ValidateReset(resetToken)
	if err != nil {
		as.log.Error("failed to validate reset token",
			slog.String("error", err.Error()),
		)

		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindRequest)
	}

	return userID, nil
}

func (as *AuthService) getUserByEmailFromPersistence(ctx context.Context, email user.Email) (*user.User, error) {
	user, err := as.userPersistence.GetByEmail(ctx, email)
	if err != nil {
		as.log.Error("failed to get user from persistence by email",
			slog.String("error", err.Error()),
			slog.String("email", email.String()),
		)

		return nil, as.handleRepoError(err)
	}

	return user, nil
}

// Login is a method that find user in db
// and compares the passwords, if it equal - generate jwt token
func (as *AuthService) Login(ctx context.Context, req *requests.Login) (*responses.Token, error) {
	email, err := as.validateEmail(req.Email)
	if err != nil {
		return nil, err
	}

	user, err := as.authenticateUser(ctx, email, req.Password)
	if err != nil {
		return nil, err
	}

	return as.generateToken(user.ID())
}

// Update is a method that updates the field of user
// and safe him into db
func (as *AuthService) Update(ctx context.Context, req *requests.UpdateUser) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	updUser, err := as.getUserFromPersistence(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := as.applyUserChanges(updUser, req); err != nil {
		return nil, err
	}

	if err = as.updateUserStore(ctx, updUser); err != nil {
		return nil, err
	}

	if err := as.clearUserCache(ctx, updUser.ID()); err != nil {
		as.log.Warn("failed to clear user's cache", slog.String("error", err.Error()))
	}

	return as.userToDTO(updUser), nil
}

// GetSelfUser is a method that find user in db (or cache)
// and returning him to user
func (as *AuthService) GetSelfUser(ctx context.Context) (*responses.User, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	if cachedUser, ok := as.getUserFromCache(ctx, userID); ok {
		return as.userToDTO(cachedUser), nil
	}

	user, err := as.getUserFromPersistence(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := as.storeUserCache(ctx, user); err != nil {
		as.log.Warn("failed to store user cache", slog.String("error", err.Error()))
	}

	return as.userToDTO(user), nil
}

func (as *AuthService) GetUserByID(ctx context.Context, userID uuid.UUID) (*responses.User, error) {
	if cachedUser, ok := as.getUserFromCache(ctx, userID); ok {
		return as.userToDTO(cachedUser), nil
	}

	user, err := as.getUserFromPersistence(ctx, userID)
	if err != nil {
		return nil, err
	}

	if err := as.storeUserCache(ctx, user); err != nil {
		as.log.Warn("failed to store user cache", slog.String("error", err.Error()))
	}

	return as.userToDTO(user), nil
}

func (as *AuthService) GetTgBindingURL(ctx context.Context) (*responses.TgBinding, error) {
	userID, err := as.getUserID(ctx)
	if err != nil {
		return nil, err
	}

	tgBinding, err := as.createTgBinding(userID)
	if err != nil {
		return nil, err
	}

	if err := as.storeBindingCache(ctx, tgBinding); err != nil {
		return nil, err
	}

	return as.bindingToDTO(tgBinding), nil
}

func (as *AuthService) BindTG(ctx context.Context, req *requests.BindTG) (*responses.User, error) {
	tgBinding, err := as.getBindingFromCache(ctx, req.BindingID)
	if err != nil {
		return nil, err
	}

	if err := as.clearTgBindingCache(ctx, tgBinding.ID()); err != nil {
		return nil, err
	}

	updUser, err := as.getUserFromPersistence(ctx, tgBinding.UserID())
	if err != nil {
		return nil, err
	}

	if err := as.applyChangeTG(updUser, req.TgChatID); err != nil {
		return nil, err
	}

	if err := as.updateUserStore(ctx, updUser); err != nil {
		return nil, err
	}

	if err := as.clearUserCache(ctx, updUser.ID()); err != nil {
		as.log.Warn("failed to clear user's cache", slog.String("error", err.Error()))
	}

	return as.userToDTO(updUser), nil
}

func (as *AuthService) Lookup(ctx context.Context, req *requests.Lookup) (*responses.LookupUsers, error) {
	if err := as.validateIDs(req.IDs); err != nil {
		return nil, err
	}

	users, err := as.lookupUsersStore(ctx, req.IDs)
	if err != nil {
		return nil, err
	}

	return as.usersToDTO(users), nil
}

func (as *AuthService) usersToDTO(users []*user.User) *responses.LookupUsers {
	response := responses.LookupUsers{
		List: make([]responses.LookupUser, len(users)),
	}
	for i, user := range users {
		response.List[i] = responses.LookupUser{
			ID:             user.ID(),
			Firstname:      user.Firstname(),
			Lastname:       user.Lastname(),
			ConfirmedEmail: user.ConfirmedEmail().String(),
		}
	}

	return &response
}

func (as *AuthService) lookupUsersStore(ctx context.Context, ids []uuid.UUID) ([]*user.User, error) {
	users, err := as.userPersistence.Lookup(ctx, ids)
	if err != nil {
		as.log.Error("failed to lookup users from db",
			slog.String("error", err.Error()),
		)

		return nil, as.handleRepoError(err)
	}

	return users, nil
}

func (as *AuthService) validateIDs(ids []uuid.UUID) error {
	n := len(ids)
	if n == 0 || n > 200 {
		return errs.NewErrorWithKind(errs.ErrInvalidLenIDs, errs.KindRequest)
	}

	return nil
}

func (as *AuthService) storeLastSendTime(ctx context.Context, at time.Time, email user.Email) error {
	if err := as.preparedCache.SetLastSend(ctx, email, at); err != nil {
		as.log.Error("failed to store last send time",
			slog.String("error", err.Error()),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) checkLastSendTime(ctx context.Context, email user.Email) error {
	lastSendAt, err := as.preparedCache.GetLastSend(ctx, email)
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return nil
		}

		as.log.Error("failed to get last send time",
			slog.String("error", err.Error()),
		)

		return errs.NewErrorWithKind(errs.ErrTooMuchResents, errs.KindConflict)
	}

	now := time.Now().UTC()
	duration := now.Sub(lastSendAt)
	if duration < as.cfg.Service.SendingDuration {
		return errs.NewErrorWithKind(errs.ErrTooMuchResents, errs.KindConflict)
	}

	return nil
}

func (as *AuthService) clearTgBindingCache(ctx context.Context, bindingID uuid.UUID) error {
	if err := as.tgBindingCache.Del(ctx, bindingID); err != nil {
		as.log.Error("failed to clear tg binding's cache",
			slog.String("error", err.Error()),
			slog.String("binding_id", bindingID.String()),
		)

		return as.handleRepoError(err)
	}
	return nil
}

func (as *AuthService) getBindingFromCache(ctx context.Context, bindingID uuid.UUID) (*tgbinding.TgBinding, error) {
	tgBinding, err := as.tgBindingCache.Get(ctx, bindingID)
	if err != nil {
		as.log.Error("failed to get binding from cache",
			slog.String("error", err.Error()),
			slog.String("binding_id", bindingID.String()),
		)

		return nil, as.handleRepoError(err)
	}

	return tgBinding, nil
}

func (as *AuthService) bindingToDTO(tgBinding *tgbinding.TgBinding) *responses.TgBinding {
	return &responses.TgBinding{
		ID:        tgBinding.ID(),
		UserID:    tgBinding.UserID(),
		BotURL:    tgBinding.BotURL().String(),
		CreatedAt: tgBinding.CreatedAt().Unix(),
	}
}

func (as *AuthService) storeBindingCache(ctx context.Context, tgBinding *tgbinding.TgBinding) error {
	if err := as.tgBindingCache.Set(ctx, tgBinding); err != nil {
		as.log.Error("failed to save tg binding's cache",
			slog.String("binding_id", tgBinding.ID().String()),
			slog.String("error", err.Error()),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) createTgBinding(userID uuid.UUID) (*tgbinding.TgBinding, error) {
	binding, err := tgbinding.New(
		userID,
		as.cfg.Service.TgBotURL,
	)
	if err != nil {
		as.log.Error("failed to create tg binding",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return nil, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return binding, nil
}

func (as *AuthService) ClearUserCache(ctx context.Context, userID uuid.UUID) error {
	if err := as.clearUserCache(ctx, userID); err != nil {
		as.log.Error("failed to clear user's cache",
			slog.String("error", err.Error()),
			slog.String("user_id", userID.String()),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) applyChangeTG(user *user.User, tgChatID int64) error {
	if err := user.ChangeTgChatID(tgChatID); err != nil {
		return errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return nil
}

func (as *AuthService) ensureEmailIsNew(ctx context.Context, rawEmail string) (user.Email, error) {
	email, err := as.validateEmail(rawEmail)
	if err != nil {
		return "", err
	}

	// Checks for existing user
	if _, err := as.userPersistence.GetByEmail(ctx, email); err == nil {
		return "", errs.NewErrorWithKind(user.ErrUserAlreadyExists, errs.KindConflict)
	} else if !errors.Is(err, user.ErrUserNotFound) {
		as.log.Error("failed to get user by email", slog.String("error", err.Error()))
		return "", errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return email, nil
}

func (as *AuthService) buildNewUser(req *requests.PrepareUser, email user.Email) (*user.User, error) {
	user, err := user.New(
		req.Firstname,
		req.Lastname,
		email,
		user.PhoneNumber(req.PhoneNumber),
		req.Password,
	)
	if err != nil {
		as.log.Warn("trying to create invalid user", slog.String("error", err.Error()))
		return nil, errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return user, nil
}

func (as *AuthService) storePreparedUserCache(ctx context.Context, user *user.User, generated string) (int, error) {
	code, err := as.generateSecureCode()
	if err != nil {
		as.log.Error("failed to generate secure code",
			slog.String("error", err.Error()),
		)

		return 0, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	if err := as.preparedCache.Prepare(ctx, user, generated, code); err != nil {
		as.log.Error("failed to set prepared user into cache",
			slog.String("user_id", user.ID().String()),
			slog.String("error", err.Error()),
		)

		return 0, as.handleRepoError(err)
	}

	return code, nil
}

func (as *AuthService) sendConfirmationCode(ctx context.Context, email user.Email, code int) {
	ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := as.sendCode(ctxTimeout, email, code); err != nil {
		as.log.Error("failed to send code to user", slog.String("error", err.Error()))
	}
}

func (as *AuthService) sendCode(ctx context.Context, email user.Email, code int) error {
	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Montserrat, Arial, sans-serif; background-color: #E6E6E6; margin: 0; padding: 0; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; background-color: #FAFAFA; border-radius: 12px; box-shadow: 0 4px 16px rgba(0,0,0,0.12); }
        .header { text-align: center; padding: 10px 0; border-bottom: 2px solid #FFB61A; }
        .header h1 { color: #1A1A1A; margin: 0; font-weight: 600; }
        .content { padding: 20px 0; text-align: center; }
        .content p { font-size: 16px; color: #1A1A1A; }
        .code { display: inline-block; padding: 15px 30px; font-size: 32px; font-weight: bold; color: #1A1A1A; background-color: #FFB61A; border-radius: 8px; letter-spacing: 4px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Подтверждение регистрации</h1>
        </div>
        <div class="content">
            <div class="code">%d</div>
            <p>Если вы не запрашивали код, проигнорируйте это письмо.</p>
        </div>
    </div>
</body>
</html>
`, code)

	subject := fmt.Sprintf("Подтверждение регистрации: %d", code)
	if err := as.smtpClient.NotifyHTML(ctx, htmlBody, subject, []string{email.String()}); err != nil {
		return err
	}

	return nil
}

func (as *AuthService) userToDTO(user *user.User) *responses.User {
	return &responses.User{
		ID:          user.ID(),
		Firstname:   user.Firstname(),
		Lastname:    user.Lastname(),
		Email:       user.ConfirmedEmail().String(),
		PhoneNumber: user.PhoneNumber().String(),
		PaidUntil:   user.PaidUntil().Unix(),
		OnVacation:  user.OnVacation(),
		TgChatID:    user.TgChatID(),
		Notification: responses.Notification{
			TelegramEnable: user.Notification().IsTelegramEnable(),
		},
		TrialUsed: user.TrialUsed(),
		CreatedAt: user.CreatedAt().Unix(),
		UpdatedAt: user.UpdatedAt().Unix(),
	}
}

func (as *AuthService) validateEmail(raw string) (user.Email, error) {
	email := user.Email(raw).Norm()
	if err := email.Validate(); err != nil {
		return "", errs.NewErrorWithKind(err, errs.KindRequest)
	}

	return email, nil
}

func (as *AuthService) getPreparedUser(ctx context.Context, email user.Email) (*user.User, int, error) {
	preparedUser, _, code, err := as.preparedCache.GetPrepared(ctx, email)
	if err != nil {
		as.log.Warn("trying to get unexisted prepared user", slog.String("error", err.Error()))
		return nil, 0, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return preparedUser, code, nil
}

func (as *AuthService) resolvePreparedUser(ctx context.Context, requestedCode int, email user.Email) (*user.User, error) {
	preparedUser, generatedPassword, code, err := as.preparedCache.GetPrepared(ctx, email)
	if err != nil {
		as.log.Warn("trying to confirm unexisted user", slog.String("error", err.Error()))
		return nil, errs.NewErrorWithKind(errs.ErrInvalidCode, errs.KindRequest)
	}

	if requestedCode != code {
		return nil, errs.NewErrorWithKind(errs.ErrInvalidCode, errs.KindNotFound)
	}

	if generatedPassword != "" {
		if err := as.sendPassword(ctx, preparedUser.ConfirmedEmail(), generatedPassword); err != nil {
			as.log.Error("failed to send password to user", slog.String("error", err.Error()))
			return nil, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
		}
	}

	return preparedUser, nil
}

func (as *AuthService) storePreparedUserDB(ctx context.Context, user *user.User) error {
	if err := as.userPersistence.Save(ctx, user); err != nil {
		as.log.Error("failed to save user into db",
			slog.String("user_id", user.ID().String()),
			slog.String("error", err.Error()),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) clearPreparedCache(ctx context.Context, email user.Email) error {
	return as.preparedCache.DelPrepared(ctx, email)
}

func (as *AuthService) saveFingerprint(ctx context.Context, fp string, userID uuid.UUID) error {
	registration, err := registration.New(userID, fp)
	if err != nil {
		return err
	}

	if err := as.registrationPersistence.Save(ctx, registration); err != nil {
		as.log.Error("failed to save registration into db",
			slog.String("error", err.Error()),
			slog.String("fingerprint", fp),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) checkFingerprint(ctx context.Context, fp string) error {
	count, err := as.registrationPersistence.Count(ctx, fp)
	if err != nil {
		as.log.Error("failed to get count of registrations by fp",
			slog.String("fingerprint", fp),
			slog.String("error", err.Error()),
		)

		return as.handleRepoError(err)
	}

	if count >= int64(as.cfg.Service.MaxRegistrations) {
		return errs.NewErrorWithKind(errs.ErrTooMuchRegistrations, errs.KindConflict)
	}

	return nil
}

func (as *AuthService) getFingerprint(ctx context.Context) (string, error) {
	raw := ctx.Value(contextvalue.KeyFingerprint)
	if raw == nil {
		return "", errs.NewErrorWithKind(errs.ErrInvalidFingerprint, errs.KindPermissionDenied)
	}

	if fingerprint, ok := raw.(string); ok {
		return fingerprint, nil
	}

	return "", errs.NewErrorWithKind(errs.ErrInvalidFingerprint, errs.KindPermissionDenied)
}

func (as *AuthService) generateToken(userID uuid.UUID) (*responses.Token, error) {
	token, err := as.jwt.Generate(userID, as.cfg.Authorization.JWT.TokenTTL)
	if err != nil {
		as.log.Error("failed to generate jwt token of user", slog.String("error", err.Error()))
		return nil, errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
	}

	return &responses.Token{
		Token: token,
	}, nil
}

func (as *AuthService) authenticateUser(ctx context.Context, email user.Email, password string) (*user.User, error) {
	user, err := as.userPersistence.GetByEmail(ctx, email)
	if err != nil {
		as.log.Error("failed to get user by email from db",
			slog.String("error", err.Error()),
		)

		return nil, as.handleRepoError(err)
	}

	if !user.PasswordHash().IsEqual(password) {
		as.log.Warn("trying to login user with incorrect password")
		return nil, errs.NewErrorWithKind(errs.ErrInvalidCredentials, errs.KindUnauthorized)
	}

	return user, nil
}

func (as *AuthService) getUserID(ctx context.Context) (uuid.UUID, error) {
	raw := ctx.Value(contextvalue.KeyUserID)
	if raw == nil {
		return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
	}

	if id, ok := raw.(uuid.UUID); ok {
		return id, nil
	}

	return uuid.Nil, errs.NewErrorWithKind(errs.ErrInvalidToken, errs.KindUnauthorized)
}

func (as *AuthService) getUserFromPersistence(ctx context.Context, id uuid.UUID) (*user.User, error) {
	user, err := as.userPersistence.GetByID(ctx, id)
	if err != nil {
		as.log.Error("failed to get user from db by id",
			slog.String("user_id", id.String()),
			slog.String("error", err.Error()),
		)

		return nil, as.handleRepoError(err)
	}

	return user, nil
}

func (as *AuthService) applyUserChanges(updUser *user.User, req *requests.UpdateUser) error {
	if req.Firstname != nil {
		if err := updUser.ChangeFirstname(*req.Firstname); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.Lastname != nil {
		if err := updUser.ChangeLastname(*req.Lastname); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.PhoneNumber != nil {
		if err := updUser.ChangePhoneNumber(user.PhoneNumber(*req.PhoneNumber)); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}
	if req.Notification.TelegramEnable != nil {
		if err := updUser.ChangeNotification(user.NewNotification(*req.Notification.TelegramEnable)); err != nil {
			return errs.NewErrorWithKind(err, errs.KindRequest)
		}
	}

	if req.OnVacation != nil {
		if *req.OnVacation {
			if err := updUser.OnRest(); err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}
		} else {
			if err := updUser.OffRest(); err != nil {
				return errs.NewErrorWithKind(err, errs.KindRequest)
			}
		}
	}

	return nil
}

func (as *AuthService) updateUserStore(ctx context.Context, user *user.User) error {
	if err := as.userPersistence.Update(ctx, user); err != nil {
		as.log.Error("failed to update user into db",
			slog.String("user_id", user.ID().String()),
			slog.String("error", err.Error()),
		)

		return as.handleRepoError(err)
	}

	return nil
}

func (as *AuthService) clearUserCache(ctx context.Context, id uuid.UUID) error {
	return as.userCache.Del(ctx, id)
}

func (as *AuthService) getUserFromCache(ctx context.Context, id uuid.UUID) (*user.User, bool) {
	cachedUser, err := as.userCache.Get(ctx, id)
	if err != nil {
		if !errors.Is(err, user.ErrUserNotFound) {
			as.log.Warn("failed to get user from cache",
				slog.String("error", err.Error()),
				slog.String("user_id", id.String()),
			)
		}

		return nil, false
	}

	as.log.Debug("user has been taken from cache", slog.String("user_id", cachedUser.ID().String()))
	return cachedUser, true
}

func (as *AuthService) storeUserCache(ctx context.Context, user *user.User) error {
	return as.userCache.Set(ctx, user)
}

func (as *AuthService) generateSecureCode() (int, error) {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return 0, err
	}
	return 100000 + int(uint32(b[0])<<16|uint32(b[1])<<8|uint32(b[2]))%900000, nil
}

func (as *AuthService) sendPassword(ctx context.Context, email user.Email, password string) error {
	if err := as.smtpClient.Notify(ctx, fmt.Sprintf("ваш сгенерированный пароль: %s", password), "Ваш пароль", []string{email.String()}); err != nil {
		return err
	}

	return nil
}

// getPassword is a function that
// generate new password if it doesn't exist
func (as *AuthService) getPasswordOrGenerate(password string) (string, bool) {
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

func (as *AuthService) handleRepoError(err error) error {
	switch {
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return errs.NewErrorWithKind(err, errs.KindTimeout)
	case errors.Is(err, user.ErrUserNotFound),
		errors.Is(err, tgbinding.ErrTgBindingNotFound):
		return errs.NewErrorWithKind(err, errs.KindNotFound)
	case errors.Is(err, user.ErrUserAlreadyExists):
		return errs.NewErrorWithKind(err, errs.KindConflict)
	}

	return errs.NewErrorWithKind(errs.ErrInternalServer, errs.KindInternal)
}
