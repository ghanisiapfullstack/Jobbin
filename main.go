package main

import (
	"jobbin/backend/bootstrap"
)

func main() {
	app := bootstrap.Boot()

	app.Start()
}
