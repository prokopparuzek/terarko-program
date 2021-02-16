package main

import (
	"fmt"
	"encoding/json"
	"os"
	"time"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"github.com/prokopparuzek/go-dht"
	"github.com/yryz/ds18b20"

	stan "github.com/nats-io/stan.go"
	cron "github.com/rk/go-cron"
	log "github.com/sirupsen/logrus"
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
	csvFile = "/home/pi/data/terarko.csv"
	subjectPrefix = "terarko"
	MAXTRY = 3
	MAX = 120
)

const (
	BME280 = iota
	DS18B20
	DHT11
)

var (
	scon stan.Conn
)

func getBME280() (e physic.Env, err error) {
	b, err := i2creg.Open("")
	if err != nil {
		log.Error(err)
		return
	}
	defer b.Close()
	dev, err := bmxx80.NewI2C(b, 0x77, &bmxx80.DefaultOpts)
	if err != nil {
		log.Error(err)
		return
	}
	defer dev.Halt()
	dev.Sense(&e)
	return e, nil
}

func getDS18B20() (e physic.Env, err error) {
	sensors, err := ds18b20.Sensors()
	if err != nil {
		log.Error(err)
		return
	}
	t, err := ds18b20.Temperature(sensors[0])
	if err != nil {
		log.Error(err)
		return
	}
	e.Temperature.Set(fmt.Sprintf("%fC", t))
	return
}

func getDHT11() (e physic.Env, err error) {
	dht, err := dht.NewDHT("GPIO17", dht.Celsius, "dht11")
	if err != nil {
		log.Error(err)
		return
	}
	h, t, err := dht.ReadRetry(10)
	if err != nil {
		log.Error(err)
		return
	}
	e.Temperature.Set(fmt.Sprintf("%fC", t))
	e.Humidity.Set(fmt.Sprintf("%f%%", h))
	return
}

func getMeasure() (out [3]measure){
	var e physic.Env
	var err error
	var logger *log.Entry

	// BME280
	logger = log.WithField("sensor", "BME280")
	for i := 0; i < MAXTRY; i++ {
		e, err = getBME280()
		if err != nil {
			logger.Error(err, " :will retry")
			continue
		} else {
			logger.Debug("Read succesfully")
			break
		}
	}
	if err != nil {
		logger.Error(err)
	}
	logger.Debug(e)
	out[BME280] = measure{e, err, time.Now()}
	// DS18B20
	logger = log.WithField("sensor", "DS18B20")
	for i := 0; i < MAXTRY; i++ {
		e, err = getDS18B20()
		if err != nil {
			logger.Error(err, " :will retry")
			continue
		} else {
			logger.Debug("Read succesfully")
			break
		}
	}
	if err != nil {
		logger.Error(err)
	}
	logger.Debug(e)
	out[DS18B20] = measure{e, err, time.Now()}
	// DHT11
	logger = log.WithField("sensor", "DHT11")
	for i := 0; i < MAXTRY; i++ {
		e, err = getDHT11()
		if err != nil {
			logger.Error(err, " :will retry")
			continue
		} else {
			logger.Debug("Read succesfully")
			break
		}
	}
	if err != nil {
		logger.Error(err)
	}
	logger.Debug(e)
	out[DHT11] = measure{e, err, time.Now()}
	return
}

func sendMsg(Jmsg []byte, subjectSuffix string, logger *log.Entry) {
	var tries int

	for tries = 0; tries < MAX; tries++ {
		err := scon.Publish(subjectPrefix+":"+subjectSuffix, Jmsg)
		if err != nil {
			logger.Debug("Error, will retry")
			time.Sleep(60 * time.Second)
		} else {
			logger.Debug("Deliver")
			break
		}
	}
	if tries >= MAX {
		logger.Error("Cannot deliver")
	}
}

func sendMeasures(_ time.Time) {
	var sense measure
	data := getMeasure()

	// BME280
	var msgB BME280Msg
	sense = data[BME280]
	if sense.err == nil {
		msgB.Timestamp = sense.timestamp.Unix()
		msgB.Humidity = float64(sense.sense.Humidity) / float64(physic.PercentRH)
		msgB.Temperature = sense.sense.Temperature.Celsius()
		msgB.Pressure = float64(sense.sense.Pressure) / float64(physic.Pascal)
		Jmsg, err := json.Marshal(msgB)
		logger := log.WithField("message", "BME280" + string(Jmsg))
		if err == nil {
			go sendMsg(Jmsg, "BME280", logger)
		} else {
			logger.Error(err)
		}
	}
	// DS18B20
	var msgD DS18B20Msg
	sense = data[DS18B20]
	if sense.err == nil {
		msgD.Timestamp = sense.timestamp.Unix()
		msgD.Temperature = sense.sense.Temperature.Celsius()
		Jmsg, err := json.Marshal(msgD)
		logger := log.WithField("message", "DS18B20" + string(Jmsg))
		if err == nil {
			go sendMsg(Jmsg, "DS18B20", logger)
		} else {
			logger.Error(err)
		}
	}
	//DHT11
	var msgT DHT11Msg
	sense = data[DHT11]
	if sense.err == nil {
		msgT.Timestamp = sense.timestamp.Unix()
		msgT.Temperature = sense.sense.Temperature.Celsius()
		msgT.Humidity = float64(sense.sense.Humidity) / float64(physic.PercentRH)
		Jmsg, err := json.Marshal(msgT)
		logger := log.WithField("message", "DHT11" + string(Jmsg))
		if err == nil {
			go sendMsg(Jmsg, "DHT11", logger)
		} else {
			logger.Error(err)
		}
	}
}

func main() {
	var err error
	// logrus
	log.SetOutput(os.Stderr)
	log.SetReportCaller(true)
	log.SetLevel(log.ErrorLevel)
	log.SetFormatter(&log.JSONFormatter{})
	f, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0664)
	if err != nil {
		log.WithField("file", logFile).Error(err)
	} else {
		log.SetOutput(f)
	}
	forever := make(chan bool)
	for {
		scon, err = stan.Connect("measures", "rpi2", stan.NatsURL("nats://rpi3:4222"), stan.Pings(60, 1440))
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second * 30)
			continue
		}
		break
	}
	defer scon.Close()
	log.Debug("Connected")
	// init periph host
	_, err = host.Init()
	if err != nil {
		log.Fatal(err)
	}
	// Cron
	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, 00, 10, sendMeasures)
	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, 15, 10, sendMeasures)
	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, 30, 10, sendMeasures)
	cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, 45, 10, sendMeasures)
	//cron.NewCronJob(cron.ANY, cron.ANY, cron.ANY, cron.ANY, cron.ANY, 10, sendMeasures)
	log.Debug("Set CRON")
	<-forever
}
