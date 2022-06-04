package main

import (
	"errors"

	"github.com/perfectogo/log"
)

func main() {
	err := errors.New("Xato")
	log.Println("Salom hammaga")
	log.Errorln("shu joyda", err)
	log.Errorln("Color", nil)
	log.Infoln("Salom Mani Zo'r narsam bor")
	log.Warning("Shu yerni qarab ketish yodizdan chiqamsin")
	Add()
}
func Add() {
	log.Println("ckddkdkdckmdc")
}
