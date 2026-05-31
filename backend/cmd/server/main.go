package main

import "github.com/devathh/hive/backend/internal/app"

func main() {
	var di app.DIContainer
	di.Log().Info("di container was loaded")
}
