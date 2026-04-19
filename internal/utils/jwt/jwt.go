package jwtUtil

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateToken 生成 JWT 令牌，有效期 24 小时
func GenerateToken(userID uint, tokenSalt string) (string, time.Time, error) {
	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": userID,
		"exp": expiresAt.Unix(),
	})
	tokenString, err := token.SignedString([]byte(tokenSalt))
	if err != nil {
		return "", time.Time{}, err
	}
	return tokenString, expiresAt, nil
}

// ParseToken 解析并验证 JWT 令牌，返回用户 ID
func ParseToken(tokenString string, tokenSalt string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(tokenSalt), nil
	})
	if err != nil {
		return 0, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, errors.New("invalid token")
	}

	sub, ok := claims["sub"]
	if !ok {
		return 0, errors.New("missing sub claim")
	}

	// JWT 数字类型解码为 float64，需转为 uint
	subFloat, ok := sub.(float64)
	if !ok {
		return 0, errors.New("invalid sub claim")
	}

	return uint(subFloat), nil
}
