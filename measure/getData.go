package main

import (
	"fmt"
	"time"

	"periph.io/x/conn/v3/physic"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"github.com/prokopparuzek/go-dht"
	"github.com/yryz/ds18b20"

	log "github.com/sirupsen/logrus"
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

func getMeasure() (out [SENSORSCOUNT]measure){
	var e physic.Env
	var err error
	var logger *log.Entry
	var functions [SENSORSCOUNT]func()(physic.Env, error)

	functions[BME280] = getBME280
	functions[DS18B20] = getDS18B20
	functions[DHT11] = getDHT11

	for i := 0; i < SENSORSCOUNT; i++ {
		logger = log.WithField("sensor", sensors[i])
		for j := 0; j < MAXTRY; j++ {
			e, err = functions[i]()
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
		out[i] = measure{e, err, time.Now()}
	}
	return
}
