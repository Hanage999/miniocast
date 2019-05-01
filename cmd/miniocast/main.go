package main

import (
	"log"
	"os"

	"github.com/hanage999/miniocast"
)

func main() {
	os.Exit(run())
}

func run() (exitCode int) {
	exitCode = 0

	prefs, ct, err := miniocast.Initialize()
	if err != nil {
		log.Printf("alert: 初期化に失敗しました：%s", err)
		exitCode = 1
		return
	}

	for _, pref := range prefs {
		pref.Update(ct)
	}

	return
}
