package processors

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/shared"
	"os"
	"testing"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/tester"
	"github.com/rs/zerolog"
)

func TestPrivacyV2(t *testing.T) {
	gt := tester.New(t)
	log := zerolog.Nop()

	fg := PrivacyV2{
		Group:        "privacy-processor",
		StatusInput:  "topic.device.status",
		FenceTable:   "table.device.privacyfence",
		StatusOutput: "topic.device.status.private",
		Logger:       &log,
	}

	fgg := fg.DefineV2()

	p, _ := goka.NewProcessor([]string{}, fgg, goka.WithTester(gt))

	go p.Run(context.TODO()) //nolint

	out := gt.NewQueueTracker(string(fg.StatusOutput))

	deviceID := "24c14Q2GGmXRT4JL0Gazu0MJ9XI"

	gt.SetTableValue(fg.FenceTable, deviceID, &shared.CloudEvent[FenceData]{Data: FenceData{
		H3Indexes: []string{"872ab259affffff", "872ab259effffff"},
	}})

	t.Run("WithinFence", func(t *testing.T) {
		statusV2 := StatusEventV2[StatusV2Data]{
			CloudEvent: shared.CloudEvent[StatusV2Data]{
				Data: StatusV2Data{
					Timestamp: 1713818407248,
					Vehicle: Vehicle{
						Signals: []SignalData{
							{
								Timestamp: 1713818407248,
								Name:      "latitude",
								Value:     42.26172693660968,
							},
							{
								Timestamp: 1713818407248,
								Name:      "longitude",
								Value:     -83.71029708818693,
							},
						},
					},
				},
			},
			TokenID:      12345,
			UserDeviceID: "2fbaXmHpdQiKyAH6o5hHTCYwU0U",
			Make:         "VW",
			Model:        "passat",
			Year:         2016,
		}

		gt.Consume(string(fg.StatusInput), deviceID, &statusV2)

		key, value, valid := out.Next()
		if !valid {
			t.Error("No output")
		}
		if key != deviceID {
			t.Errorf("Expected output to maintain the device ID %s as the key, but got %s", deviceID, key)
		}

		event := value.(*StatusEventV2[StatusV2Data])
		lat := event.Data.Vehicle.Signals[0].Value.(float64)
		lon := event.Data.Vehicle.Signals[1].Value.(float64)
		if lat != 42.25362819577089 || lon != -83.68562802176137 {
			t.Errorf("Expected %f, %f in the output but got %f, %f",
				42.25362819577089, -83.68562802176137,
				lat, lon,
			)
		}

		if *event.Data.IsRedacted != true {
			t.Errorf("Expected isRedacted to be true")
		}
	})

	t.Run("WithinFenceWithMultipleLocations", func(t *testing.T) {
		statusV2 := StatusEventV2[StatusV2Data]{
			CloudEvent: shared.CloudEvent[StatusV2Data]{
				Data: StatusV2Data{
					Timestamp: 1713818407248,
					Vehicle: Vehicle{
						Signals: []SignalData{
							{
								Timestamp: 1713818407248,
								Name:      "latitude",
								Value:     42.26172693660968,
							},
							{
								Timestamp: 1713818407248,
								Name:      "longitude",
								Value:     -83.71029708818693,
							},
							{
								Timestamp: 1713818407248,
								Name:      "hdop",
								Value:     0.8,
							},
							{
								Timestamp: 1713818407248,
								Name:      "nsat",
								Value:     0,
							},
							{
								Timestamp: 1713818400177,
								Name:      "latitude",
								Value:     42.261123478313145,
							},
							{
								Timestamp: 1713818400177,
								Name:      "longitude",
								Value:     -83.68613574673722,
							},
						},
					},
				},
			},
			TokenID:      12345,
			UserDeviceID: "2fbaXmHpdQiKyAH6o5hHTCYwU0U",
			Make:         "VW",
			Model:        "passat",
			Year:         2016,
		}

		gt.Consume(string(fg.StatusInput), deviceID, &statusV2)

		key, value, valid := out.Next()
		if !valid {
			t.Error("No output")
		}
		if key != deviceID {
			t.Errorf("Expected output to maintain the device ID %s as the key, but got %s", deviceID, key)
		}

		event := value.(*StatusEventV2[StatusV2Data])
		// fenced location
		lat := event.Data.Vehicle.Signals[0].Value.(float64)
		lon := event.Data.Vehicle.Signals[1].Value.(float64)
		if lat != 42.25362819577089 || lon != -83.68562802176137 {
			t.Errorf("Expected %f, %f in the output but got %f, %f",
				42.25362819577089, -83.68562802176137,
				lat, lon,
			)
		}

		// unfenced location
		lat = event.Data.Vehicle.Signals[4].Value.(float64)
		lon = event.Data.Vehicle.Signals[5].Value.(float64)
		if lat != 42.261123478313145 || lon != -83.68613574673722 {
			t.Errorf("Expected %f, %f in the output but got %f, %f",
				42.261123478313145, -83.68613574673722,
				lat, lon,
			)
		}

		if *event.Data.IsRedacted != true {
			t.Errorf("Expected isRedacted to be true")
		}
	})

	t.Run("WithinFenceWithFullPayload", func(t *testing.T) {
		file, err := os.Open("testdata/statusV2.json")
		if err != nil {
			fmt.Println("Error opening file:", err)
			return
		}
		defer file.Close()

		var statusV2 StatusEventV2[StatusV2Data]
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&statusV2)
		if err != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}

		gt.Consume(string(fg.StatusInput), deviceID, &statusV2)

		key, value, valid := out.Next()
		if !valid {
			t.Error("No output")
		}
		if key != deviceID {
			t.Errorf("Expected output to maintain the device ID %s as the key, but got %s", deviceID, key)
		}

		event := value.(*StatusEventV2[StatusV2Data])
		// fenced location
		lat := event.Data.Vehicle.Signals[23].Value.(float64)
		lon := event.Data.Vehicle.Signals[22].Value.(float64)
		if lat != 42.25362819577089 || lon != -83.68562802176137 {
			t.Errorf("Expected %f, %f in the output but got %f, %f",
				42.25362819577089, -83.68562802176137,
				lat, lon,
			)
		}

		if *event.Data.IsRedacted != true {
			t.Errorf("Expected isRedacted to be true")
		}
	})

	t.Run("OutsideFence", func(t *testing.T) {
		statusV2 := StatusEventV2[StatusV2Data]{
			CloudEvent: shared.CloudEvent[StatusV2Data]{
				Data: StatusV2Data{
					Timestamp: 1713818407248,
					Vehicle: Vehicle{
						Signals: []SignalData{
							{
								Timestamp: 1713818407248,
								Name:      "latitude",
								Value:     42.261123478313145,
							},
							{
								Timestamp: 1713818407248,
								Name:      "longitude",
								Value:     -83.68613574673722,
							},
						},
					},
				},
			},
			TokenID:      12345,
			UserDeviceID: "2fbaXmHpdQiKyAH6o5hHTCYwU0U",
			Make:         "VW",
			Model:        "passat",
			Year:         2016,
		}

		gt.Consume(string(fg.StatusInput), deviceID, &statusV2)

		key, value, valid := out.Next()
		if !valid {
			t.Error("No output")
		}
		if key != deviceID {
			t.Errorf("Expected output to maintain the device ID %s as the key, but got %s", deviceID, key)
		}

		event := value.(*StatusEventV2[StatusV2Data])
		lat := event.Data.Vehicle.Signals[0].Value.(float64)
		lon := event.Data.Vehicle.Signals[1].Value.(float64)
		if lat != 42.261123478313145 || lon != -83.68613574673722 {
			t.Errorf("Expected %f, %f in the output but got %f, %f",
				42.261123478313145, -83.68613574673722,
				lat, lon,
			)
		}

		if 12345 != event.TokenID {
			t.Errorf("Expected TokenID to be %d", 12345)
		}

		if *event.Data.IsRedacted != false {
			t.Errorf("Expected isRedacted to be false")
		}
	})
}
