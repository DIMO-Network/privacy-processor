package processors

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/lovoo/goka"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
	"github.com/uber/h3-go"
)

type CloudEvent struct {
	ID          string          `json:"id"`
	Source      string          `json:"source"`
	SpecVersion string          `json:"specversion"`
	Subject     string          `json:"subject"`
	Time        time.Time       `json:"time"`
	Type        string          `json:"type"`
	Data        json.RawMessage `json:"data"`
}

type CloudEventCodec struct{}

func (c *CloudEventCodec) Encode(value interface{}) ([]byte, error) {
	return json.Marshal(value)
}

func (c *CloudEventCodec) Decode(data []byte) (interface{}, error) {
	ce := new(CloudEvent)
	err := json.Unmarshal(data, ce)
	return ce, err
}

type Privacy struct {
	Group        goka.Group
	StatusInput  goka.Stream
	FenceTable   goka.Table
	StatusOutput goka.Stream
}

var fastOptions = &sjson.Options{
	Optimistic:     true,
	ReplaceInPlace: true,
}

func (g *Privacy) Define() *goka.GroupGraph {
	return goka.DefineGroup(g.Group,
		goka.Input(g.StatusInput, new(CloudEventCodec), g.processStatusEvent),
		goka.Join(g.FenceTable, new(CloudEventCodec)),
		goka.Output(g.StatusOutput, new(CloudEventCodec)),
	)
}

func (g *Privacy) processStatusEvent(ctx goka.Context, msg interface{}) {
	fence := g.getFence(ctx)
	event := msg.(*CloudEvent)

	sanitizeEvent(event, fence)

	ctx.Emit(g.StatusOutput, ctx.Key(), event)
}

// sanitizeEvent modifies the given CloudEvent using fence.
func sanitizeEvent(event *CloudEvent, fence []h3.H3Index) {
	lat, err := getNumber(event.Data, "latitude")
	if err != nil {
		return
	}

	lng, err := getNumber(event.Data, "longitude")
	if err != nil {
		return
	}

	geo := h3.GeoCoord{Latitude: lat, Longitude: lng}

	for _, fenceInd := range fence {
		// TODO: Should really validate res more
		res := h3.Resolution(fenceInd)
		statusInd := h3.FromGeo(geo, res)
		if statusInd == fenceInd {
			outGeo := h3.ToGeo(h3.ToParent(statusInd, res-1))
			event.Data, _ = sjson.SetBytesOptions(event.Data, "latitude", outGeo.Latitude, fastOptions)
			event.Data, _ = sjson.SetBytesOptions(event.Data, "longitude", outGeo.Longitude, fastOptions)
			break
		}
	}
}

// getNumber takes the marshaled JSON in data and tries to retrieve the numeric value associated
// with the given key.
func getNumber(data []byte, key string) (float64, error) {
	val := gjson.GetBytes(data, key)
	if !val.Exists() {
		return 0, fmt.Errorf("no field %s in document", key)
	}
	if val.Type != gjson.Number {
		return 0, fmt.Errorf("field %s had non-numeric type %s", key, val.Type)
	}
	return val.Num, nil
}

type FenceData struct {
	H3Indexes []string `json:"h3Indexes"`
}

func (g *Privacy) getFence(ctx goka.Context) []h3.H3Index {
	val := ctx.Join(g.FenceTable)
	if val == nil {
		return nil
	}

	event := val.(*CloudEvent)
	fence := new(FenceData)

	err := json.Unmarshal(event.Data, fence)
	if err != nil {
		return nil
	}

	out := make([]h3.H3Index, len(fence.H3Indexes))
	for i, s := range fence.H3Indexes {
		out[i] = h3.FromString(s)
	}

	return out
}
