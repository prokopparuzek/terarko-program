package main

import (
	"time"

	"periph.io/x/conn/v3/physic"

	stan "github.com/nats-io/stan.go"
)

type measure struct {
	sense physic.Env
	err error
	timestamp time.Time
}

type BME280Msg struct {
	Timestamp   int64   `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Humidity float64 `json:"humidity"`
	Pressure float64 `json:"pressure"`
}

type DS18B20Msg struct {
	Timestamp   int64   `json:"timestamp"`
	Temperature float64 `json:"temperature"`
}

type DHT11Msg struct {
	Timestamp   int64   `json:"timestamp"`
	Temperature float64 `json:"temperature"`
	Humidity float64 `json:"humidity"`
}

const (
	//logFile = "/var/log/measures-terarko.log"
	logFile = "log"
	csvFilePrefix = "/home/pi/data/terarko-"
	subjectPrefix = "terarko"
	MAXTRY = 3
	MAX = 120
	SENSORSCOUNT = 3
)

const (
	BME280 = iota
	DS18B20
	DHT11
)

var sensors []string = []string{"BME280", "DS18B20", "DHT11"}

var (
	scon stan.Conn
)
