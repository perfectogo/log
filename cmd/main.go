package main

import (
	"errors"

	"github.com/perfectogo/log"
)

func main() {
	err := errors.New("Xato")
	log.Println("Salom hammaga")
	log.Error("shu joyda", err)
	log.Error("Color", nil)
	log.Info("Salom Mani Zo'r narsam bor")
	log.Warning("Shu yerni qarab ketish yodizdan chiqamsin")
}
