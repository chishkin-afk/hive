package jwt

import (
	"crypto/rsa"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/devathh/hive/backend/internal/common/infrastructure/config"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
)

type JWTManager struct {
	private *rsa.PrivateKey
	public  *rsa.PublicKey
	cfg     *config.Config
}

func (jm *JWTManager) Generate(userID uuid.UUID) (string, error) {
	now := time.Now().UTC()
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, &customClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "hivecrm",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(jm.cfg.Authorization.JWT.TokenTTL)),
			Subject:   "hiveuser",
		},
	})

	signedToken, err := token.SignedString(jm.private)
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func (jm *JWTManager) Validate(tokenString string) (uuid.UUID, error) {
	token, err := jwt.ParseWithClaims(tokenString, &customClaims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, ErrInvalidToken
		}

		return jm.public, nil
	})
	if err != nil {
		return uuid.Nil, err
	}

	if claims, ok := token.Claims.(*customClaims); ok {
		return claims.UserID, nil
	}

	return uuid.Nil, ErrInvalidToken
}

func New(cfg *config.Config) (*JWTManager, error) {
	private, err := loadPrivate(cfg.Authorization.JWT.PrivateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	public, err := loadPublic(cfg.Authorization.JWT.PublicKeyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return &JWTManager{
		private: private,
		public:  public,
		cfg:     cfg,
	}, nil
}

func loadPublic(path string) (*rsa.PublicKey, error) {
	bytes, err := readFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(bytes)
}

func loadPrivate(path string) (*rsa.PrivateKey, error) {
	bytes, err := readFile(path)
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPrivateKeyFromPEM(bytes)
}

func readFile(path string) ([]byte, error) {
	path = filepath.Clean(path)
	if _, err := filepath.Abs(path); err != nil {
		return nil, err
	}

	return os.ReadFile(path)
}
