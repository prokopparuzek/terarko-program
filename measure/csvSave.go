package main

import (
	"encoding/csv"
	"os"

	log "github.com/sirupsen/logrus"
)

func saveData(what string, data []string) {
	f, err := os.OpenFile(csvFilePrefix+what+".csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Panicln(err)
	}
	defer f.Close()
	log.WithField("file", what).Debug("Open file")
	writer := csv.NewWriter(f)
	writer.Write(data)
	writer.Flush()
	err = writer.Error()
	if err != nil {
		log.Panicln(err)
	}
	log.Debug("Stored csv")
}

func saveCsv(toCsv [SENSORSCOUNT][]string) {
	for i, data := range toCsv {
		saveData(sensors[i], data)
	}
}
