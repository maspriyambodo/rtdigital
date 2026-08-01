package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidTOTPCode = errors.New("invalid TOTP code")
	totpEncoding       = base32.StdEncoding.WithPadding(base32.NoPadding)
)

func GenerateTOTPSecret() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("generate TOTP secret: %w", err)
	}
	return totpEncoding.EncodeToString(bytes), nil
}

func GenerateTOTPURI(secret, accountName, issuer string) string {
	label := url.PathEscape(issuer) + ":" + url.PathEscape(accountName)
	return fmt.Sprintf(
		"otpauth://totp/%s?secret=%s&issuer=%s&algorithm=SHA1&digits=6&period=30",
		label,
		url.QueryEscape(secret),
		url.QueryEscape(issuer),
	)
}

func GenerateTOTPCode(secret string, at time.Time) (string, error) {
	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return "", err
	}
	return generateTOTPCode(key, uint64(at.UTC().Unix()/30)), nil
}

func VerifyTOTP(code, secret string, at time.Time) (bool, error) {
	code = strings.TrimSpace(code)
	if len(code) != 6 {
		return false, ErrInvalidTOTPCode
	}
	if _, err := strconv.Atoi(code); err != nil {
		return false, ErrInvalidTOTPCode
	}

	key, err := decodeTOTPSecret(secret)
	if err != nil {
		return false, err
	}

	counter := at.UTC().Unix() / 30
	for _, offset := range []int64{-1, 0, 1} {
		if counter+offset >= 0 && hmac.Equal([]byte(generateTOTPCode(key, uint64(counter+offset))), []byte(code)) {
			return true, nil
		}
	}
	return false, nil
}

func decodeTOTPSecret(secret string) ([]byte, error) {
	key, err := totpEncoding.DecodeString(strings.ToUpper(strings.TrimSpace(secret)))
	if err != nil || len(key) != 20 {
		return nil, errors.New("invalid TOTP secret")
	}
	return key, nil
}

func generateTOTPCode(key []byte, counter uint64) string {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], counter)

	mac := hmac.New(sha1.New, key)
	_, _ = mac.Write(buf[:])
	hash := mac.Sum(nil)
	offset := hash[len(hash)-1] & 0x0f
	value := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7fffffff
	return fmt.Sprintf("%06d", value%1_000_000)
}