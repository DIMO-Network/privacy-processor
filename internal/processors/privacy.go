package processors

import (
	"encoding/json"

	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/uber/h3-go"
)

type Privacy struct {
	Group        goka.Group
	StatusInput  goka.Stream
	FenceTable   goka.Table
	StatusOutput goka.Stream

	Logger *zerolog.Logger
}

type FenceData struct {
	H3Indexes []string `json:"h3Indexes"`
}

type FenceEvent struct {
	CloudEvent
	Data FenceData `json:"data"`
}

type StatusData struct {
	Latitude  *float64               `json:"latitude"`
	Longitude *float64               `json:"longitude"`
	Overflow  map[string]interface{} `json:"-"`
}

func (d *StatusData) MarshalJSON() ([]byte, error) {
	if d.Latitude != nil {
		d.Overflow["latitude"] = d.Latitude
	} else {
		delete(d.Overflow, "latitude")
	}

	if d.Longitude != nil {
		d.Overflow["longitude"] = d.Longitude
	} else {
		delete(d.Overflow, "longitude")
	}

	return json.Marshal(d.Overflow)
}

func (d *StatusData) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, d); err != nil {
		return err
	}

	return json.Unmarshal(data, &d.Overflow)
}

type StatusEvent struct {
	CloudEvent
	Data StatusData `json:"data"`
}

var StatusCodec = &JSONCodec{Factory: func() interface{} { return new(StatusEvent) }}
var FenceCodec = &JSONCodec{Factory: func() interface{} { return new(FenceEvent) }}

func (g *Privacy) Define() *goka.GroupGraph {
	return goka.DefineGroup(g.Group,
		goka.Input(g.StatusInput, StatusCodec, g.processStatusEvent),
		goka.Join(g.FenceTable, FenceCodec),
		goka.Output(g.StatusOutput, StatusCodec),
	)
}

func (g *Privacy) processStatusEvent(ctx goka.Context, msg interface{}) {
	fence := g.getFence(ctx)
	event := msg.(*StatusEvent)

	sanitizeEvent(event, fence)

	ctx.Emit(g.StatusOutput, ctx.Key(), event)
}

// sanitizeEvent modifies the given CloudEvent using fence.
func sanitizeEvent(event *StatusEvent, fence []h3.H3Index) {
	if event.Data.Latitude == nil || event.Data.Longitude == nil {
		return
	}

	geo := h3.GeoCoord{Latitude: *event.Data.Latitude, Longitude: *event.Data.Longitude}

	for _, fenceInd := range fence {
		// TODO: Should really validate res more.
		res := h3.Resolution(fenceInd)
		// TODO: Cache these.
		statusInd := h3.FromGeo(geo, res)
		if statusInd == fenceInd {
			outGeo := h3.ToGeo(h3.ToParent(statusInd, res-1))
			event.Data.Latitude, event.Data.Longitude = &outGeo.Latitude, &outGeo.Longitude
			break
		}
	}
}

func (g *Privacy) getFence(ctx goka.Context) []h3.H3Index {
	val := ctx.Join(g.FenceTable)
	if val == nil {
		return nil
	}

	sIndexes := val.(*FenceEvent).Data.H3Indexes
	out := make([]h3.H3Index, len(sIndexes))
	for i, s := range sIndexes {
		out[i] = h3.FromString(s)
	}

	return out
}
