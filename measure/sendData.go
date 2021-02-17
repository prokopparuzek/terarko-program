package main

import (
	"fmt"
	"encoding/json"
	"time"

	"periph.io/x/conn/v3/physic"

	log "github.com/sirupsen/logrus"
)

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
	var toCsv [SENSORSCOUNT][]string
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
			toCsv[BME280] = []string{fmt.Sprint(msgB.Timestamp), fmt.Sprint(msgB.Temperature), fmt.Sprint(msgB.Humidity), fmt.Sprint(msgB.Pressure)}
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
			toCsv[DS18B20] = []string{fmt.Sprint(msgB.Timestamp), fmt.Sprint(msgB.Temperature)}
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
			toCsv[DHT11] = []string{fmt.Sprint(msgB.Timestamp), fmt.Sprint(msgB.Temperature), fmt.Sprint(msgB.Humidity)}
			go sendMsg(Jmsg, "DHT11", logger)
		} else {
			logger.Error(err)
		}
	}
	saveCsv(toCsv)
}
