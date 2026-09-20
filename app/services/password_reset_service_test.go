package services

import "testing"

func TestPasswordResetTokensAreRandomAndHashed(t *testing.T) {
	rawA, hashA, err := newPasswordResetToken()
	if err != nil {
		t.Fatalf("newPasswordResetToken() error = %v", err)
	}
	rawB, hashB, err := newPasswordResetToken()
	if err != nil {
		t.Fatalf("newPasswordResetToken() error = %v", err)
	}

	if rawA == rawB || hashA == hashB {
		t.Fatal("password reset tokens must be unique")
	}
	if rawA == hashA {
		t.Fatal("raw reset token must not be stored as its hash")
	}
	if got := hashPasswordResetToken(rawA); got != hashA {
		t.Fatalf("hashPasswordResetToken() = %q, want %q", got, hashA)
	}
	if len(hashA) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hashA))
	}
}
