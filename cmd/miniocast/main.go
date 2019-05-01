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

	ps, ct, err := miniocast.Initialize()
	if err != nil {
		log.Printf("alert: 初期化に失敗しました：%s", err)
		exitCode = 1
		return
	}

	for _, p := range ps {
		p.Update(ct)
	}

	return
}
