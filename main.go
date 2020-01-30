package main

import (
	"fmt"
	"log"
	"time"

	"fmradio/display"
	"fmradio/radio"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	stationName := "DlSnIpEr Inc."
	rdsMessage := "DlSnIpEr in the mix"

	adaptor := raspi.NewAdaptor()

	radioConfig := radio.Si4713Config{
		TransmitFrequency:  9550,
		TransmitPower:      115,
		HasRDS:             true,
		ProgramID:          0x3104,
		StationName:        stationName,
		RdsMessage:         rdsMessage,
		Log:                log.Printf,
	}
	rdio, err := radio.NewSi4713Driver(adaptor, radioConfig)
	if err != nil {
		log.Fatalln(err)
	}

	lcd, err := display.NewLCD1602Driver(adaptor)
	if err != nil {
		log.Fatalln(err)
	}

	work := func() {
		if err = lcd.DisplayMessage("Starting the FM station"); err != nil {
			log.Fatalln(err)
		}

		// wait a bit to get the new device status
		if err = rdio.SetRDSMessage(rdsMessage); err != nil {
			log.Fatalln(err)
		}

		stationFrequency := fmt.Sprintf(" - %.2fMHz", float32(radioConfig.TransmitFrequency)/100)
		if err = lcd.DisplayMessage(rdsMessage + stationFrequency); err != nil {
			log.Fatalln(err)
		}

		gobot.Every(1*time.Second, func() {
			if err = rdio.Loop(); err != nil {
				log.Fatalln(err)
			}

			timeNow := time.Now().Format("2006-01-02 15:04:05 -0700 MST")
			if err = lcd.DisplayMessage(timeNow); err != nil {
				log.Fatalln(err)
			}
		})
	}

	robot := gobot.NewRobot("FM Transmitter Station demo",
		[]gobot.Connection{adaptor},
		[]gobot.Device{rdio, lcd},
		work,
	)

	if err = robot.Start(); err != nil {
		log.Fatalln(err)
	}
}
