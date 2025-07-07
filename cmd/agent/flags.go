package main

import (
	"flag"
)

var addrServer string
var pollInterval int
var reportInterval int

// Парсим командную строку, получаем адрес сервера, интервалы сбора и отправки метрик
func parseFlags() {
	flag.StringVar(&addrServer, "a", ":8080", "server address")
	flag.IntVar(&pollInterval, "p", 2, "poll interval")
	flag.IntVar(&reportInterval, "r", 10, "report interval")
	flag.Parse()
}
