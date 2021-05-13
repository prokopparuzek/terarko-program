package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	stan "github.com/nats-io/stan.go"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

//const logFile = "/var/log/fire.log"
const logFile = "log"

var client *firestore.Client
var ctx context.Context

var subjects []string = []string{"terarko:BME280", "terarko:DHT11", "terarko:DS18B20"}

func handleMsg(msg *stan.Msg) {
	var payload map[string]interface{}
	payload = make(map[string]interface{})
	ref := client.Collection(msg.Subject)
	json.Unmarshal(msg.Data, &payload)
	log.WithField("payload", payload).Debug()
	_, _, err := ref.Add(ctx, payload)
	if err != nil {
		log.Println(err)
		return
	}
	log.Debug("Fired")
	msg.Ack()
}

func main() {
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
	var sc stan.Conn
	for {
		sc, err = stan.Connect("measures", "terarko-gun", stan.NatsURL("nats://rpi3:4222"), stan.Pings(60, 1440))
		if err != nil {
			log.Error(err)
			time.Sleep(time.Second * 30)
			continue
		}
		break
	}
	defer sc.Close()
	log.Debug("Connected")
	// firebase
	ctx = context.Background()
	opt := option.WithCredentialsFile("service.json")
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v\n", err)
	}
	log.Debug("Connected to firebase")
	client, err = app.Firestore(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	defer client.Close()
	log.Debug("Connected to firestore")
	for _, s := range subjects {
		_, err = sc.Subscribe(s, handleMsg, stan.DurableName("2"), stan.DeliverAllAvailable(), stan.MaxInflight(3), stan.SetManualAckMode())
		if err != nil {
			log.Error(err)
		}
	}
	log.Debug("Subscribed")
	<-forever
}
