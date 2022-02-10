package processors

import (
	"context"
	"testing"

	"github.com/lovoo/goka"
	"github.com/lovoo/goka/tester"
	"github.com/rs/zerolog"
)

func ref(x float64) *float64 { return &x }

func TestMain(t *testing.T) {
	gt := tester.New(t)
	log := zerolog.Nop()

	fg := Privacy{
		Group:        "privacy-processor",
		StatusInput:  "topic.device.status",
		FenceTable:   "table.device.privacyfence",
		StatusOutput: "topic.device.status.private",
		Logger:       &log,
	}

	fgg := fg.Define()

	p, _ := goka.NewProcessor([]string{}, fgg, goka.WithTester(gt))

	go p.Run(context.TODO())

	out := gt.NewQueueTracker(string(fg.StatusOutput))

	deviceID := "24c14Q2GGmXRT4JL0Gazu0MJ9XI"

	gt.Consume(string(fg.FenceTable), deviceID, &FenceEvent{Data: FenceData{
		H3Indexes: []string{"872ab259affffff", "872ab259effffff"},
	}})

	gt.Consume(string(fg.StatusInput), deviceID, &StatusEvent{Data: StatusData{
		Latitude:  ref(42.26172693660968),
		Longitude: ref(-83.71029708818693),
		Overflow:  map[string]interface{}{},
	}})

	key, value, valid := out.Next()
	if !valid {
		t.Error("No output")
	}
	if key != deviceID {
		t.Errorf("Expected output to maintain the device ID %s as the key, but got %s", deviceID, key)
	}

	event := value.(*StatusEvent)
	if *event.Data.Latitude != 42.25362819577089 || *event.Data.Longitude != -83.68562802176137 {
		t.Errorf("Expected %f, %f in the output but got %f, %f",
			42.25362819577089, -83.68562802176137,
			*event.Data.Latitude, *event.Data.Longitude,
		)
	}
}
