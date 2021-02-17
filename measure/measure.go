package main

import (
	"os"
	"time"

	"periph.io/x/host/v3"

	stan "github.com/nats-io/stan.go"
	cron "github.com/rk/go-cron"
	log "github.com/sirupsen/logrus"
)

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
