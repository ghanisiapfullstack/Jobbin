package tests

import (
	"github.com/goravel/framework/testing"

	"jobbin/backend/bootstrap"
)

func init() {
	bootstrap.Boot()
}

type TestCase struct {
	testing.TestCase
}
