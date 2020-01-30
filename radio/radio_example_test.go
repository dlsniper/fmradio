package radio_test

import (
	"log"
	"time"

	"fmradio/radio"

	"gobot.io/x/gobot"
	"gobot.io/x/gobot/platforms/raspi"
)

func ExampleSi4713Driver() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	stationName := "DlSnIpEr Inc."
	rdsMessage := "DlSnIpEr in the mix"

	adaptor := raspi.NewAdaptor()

	radioConfig := radio.Si4713Config{
		TransmitFrequency: 8850,
		TransmitPower:     115,
		ResetPin:          "29",
		DebugMode:         false,
		HasRDS:            true,
		ProgramID:         0x3104,
		StationName:       stationName,
		RdsMessage:        rdsMessage,
		Log:               log.Printf,
		DebugLog:          nil,
	}
	rdio, err := radio.NewSi4713Driver(adaptor, radioConfig)
	if err != nil {
		log.Fatalln(err)
	}

	work := func() {
		if err = rdio.SetRDSMessage(rdsMessage); err != nil {
			log.Fatalln(err)
		}

		gobot.Every(1*time.Second, func() {
			if err = rdio.Loop(); err != nil {
				log.Fatalln(err)
			}
		})
	}

	robot := gobot.NewRobot("FM Transmitter Station demo",
		[]gobot.Connection{adaptor},
		[]gobot.Device{rdio},
		work,
	)

	if err = robot.Start(); err != nil {
		log.Fatalln(err)
	}
}
