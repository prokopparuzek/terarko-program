package main

import (
	"fmt"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/host/v3"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/devices/v3/bmxx80"
	"github.com/prokopparuzek/go-dht"
	"github.com/yryz/ds18b20"
)
func main() {
	// BME280
	var e physic.Env
	_, err := host.Init()
	if err != nil {
		fmt.Println(err)
	}
	b, err := i2creg.Open("")
	if err != nil {
		fmt.Println(err)
	}
	defer b.Close()
	dev, err := bmxx80.NewI2C(b, 0x77, &bmxx80.DefaultOpts)
	if err != nil {
		fmt.Println(err)
	}
	defer dev.Halt()
	dev.Sense(&e)
	fmt.Println(e)
	// DS18B20
	sensors, err := ds18b20.Sensors()
	if err != nil {
		fmt.Println(err)
	}
	t, err := ds18b20.Temperature(sensors[0])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t)
	// DHT11
	dht, err := dht.NewDHT("GPIO17", dht.Celsius, "dht11")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dht.ReadRetry(15))
}
