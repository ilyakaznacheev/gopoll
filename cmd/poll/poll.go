package main

import (
	"flag"
	"fmt"

	"github.com/ilyakaznacheev/gopoll/internal/poll"
)

func main() {
	var (
		port     string
		confPath string
		static   string
	)

	flag.StringVar(&port, "port", "8000", "port to listen HTTP requests")
	flag.StringVar(&confPath, "config", "config/config.yml", "path to config file")
	flag.StringVar(&static, "static", "static", "path to static file directory")

	flag.Parse()

	err := poll.Run(port, confPath, static)
	if err != nil {
		fmt.Println(err)
	}
}
