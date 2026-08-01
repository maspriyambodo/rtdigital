package auth_test

import (
	"strings"
	"testing"

	"github.com/maspriyambodo/rtdigital/services/api/internal/auth"
)

func TestPasswordHash(t *testing.T) {
	t.Parallel()

	password := "correct horse battery staple"
	hash1, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword(): %v", err)
	}
	if !strings.HasPrefix(hash1, "$argon2id$v=19$m=65536,t=3,p=2$") {
		t.Errorf("unexpected hash format: %q", hash1)
	}

	hash2, err := auth.HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() second call: %v", err)
	}
	if hash1 == hash2 {
		t.Error("equal hashes; salt must be random")
	}

	ok, err := auth.VerifyPassword(password, hash1)
	if err != nil || !ok {
		t.Errorf("VerifyPassword(correct) = %v, %v; want true, nil", ok, err)
	}

	ok, err = auth.VerifyPassword("wrong password", hash1)
	if err != nil || ok {
		t.Errorf("VerifyPassword(wrong) = %v, %v; want false, nil", ok, err)
	}

	for _, invalid := range []string{
		"invalid",
		"$argon2id$v=18$m=65536,t=3,p=2$c2FsdDEyMzQ1Njc4OTAxMg$MTIzNDU2Nzg5MDEyMzQ1Ng",
		"$argon2id$v=19$m=0,t=3,p=2$c2FsdDEyMzQ1Njc4OTAxMg$MTIzNDU2Nzg5MDEyMzQ1Ng",
	} {
		if _, err := auth.VerifyPassword(password, invalid); err == nil {
			t.Errorf("VerifyPassword(%q): expected error", invalid)
		}
	}
}

func TestNormalizeEmail(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, want string
		wantErr     bool
	}{
		{"USER@Example.COM", "user@example.com", false},
		{"  user.name+tag@example.co.id  ", "user.name+tag@example.co.id", false},
		{"user", "", true},
		{"@example.com", "", true},
		{"user@example", "", true},
		{"", "", true},
	}

	for _, test := range tests {
		got, err := auth.NormalizeEmail(test.input)
		if (err != nil) != test.wantErr || got != test.want {
			t.Errorf("NormalizeEmail(%q) = %q, %v; want %q, error=%v", test.input, got, err, test.want, test.wantErr)
		}
	}
}

func TestNormalizePhone(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input, want string
		wantErr     bool
	}{
		{"+6281234567890", "+6281234567890", false},
		{"081234567890", "+6281234567890", false},
		{"6281234567890", "+6281234567890", false},
		{"+62 (812) 3456-7890", "+6281234567890", false},
		{"+1 (202) 555-0123", "+12025550123", false},
		{"invalid", "", true},
		{"0812", "", true},
		{"+6281234567890123456", "", true},
		{"", "", true},
	}

	for _, test := range tests {
		got, err := auth.NormalizePhone(test.input)
		if (err != nil) != test.wantErr || got != test.want {
			t.Errorf("NormalizePhone(%q) = %q, %v; want %q, error=%v", test.input, got, err, test.want, test.wantErr)
		}
	}
}