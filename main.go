package main

import (
	"log"
	"mycodemystyle_api/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
