package controllers

import (
	"reflect"
	"testing"

	"jobbin/backend/app/models"
)

func TestAuthMethods(t *testing.T) {
	password := "hash"
	googleID := "google-user"

	tests := []struct {
		name string
		user models.User
		want []string
	}{
		{name: "password", user: models.User{Password: &password}, want: []string{"password"}},
		{name: "google", user: models.User{GoogleID: &googleID}, want: []string{"google"}},
		{name: "linked", user: models.User{Password: &password, GoogleID: &googleID}, want: []string{"password", "google"}},
		{name: "none", user: models.User{}, want: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := authMethods(tt.user); !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("authMethods() = %v, want %v", got, tt.want)
			}
		})
	}
}
