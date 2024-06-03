package processors

import (
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/uber/h3-go/v3"
)

type PrivacyV2 struct {
	Group        goka.Group
	StatusInput  goka.Stream
	FenceTable   goka.Table
	StatusOutput goka.Stream

	Logger *zerolog.Logger
}

var StatusV2Codec = &JSONCodec{Factory: func() interface{} { return new(StatusEventV2[StatusV2Data]) }}

func (g *PrivacyV2) DefineV2() *goka.GroupGraph {
	return goka.DefineGroup(g.Group,
		goka.Input(g.StatusInput, StatusV2Codec, g.processStatusEventV2),
		goka.Join(g.FenceTable, FenceCodec),
		goka.Output(g.StatusOutput, StatusV2Codec),
	)
}

func (g *PrivacyV2) processStatusEventV2(ctx goka.Context, msg interface{}) {
	fence := g.getFenceV2(ctx)
	event := msg.(*StatusEventV2[StatusV2Data])

	sanitizeEventV2(event, fence)

	// Key should be the DIMO device id.
	ctx.Emit(g.StatusOutput, ctx.Key(), event)
}

// sanitizeEventV2 modifies the given CloudEvent using fence.
func sanitizeEventV2(event *StatusEventV2[StatusV2Data], fence []h3.H3Index) {
	locationIndexesByTimestamp := findIndexPairsWithSameTimestamp(event.Data.Vehicle.Signals)

	if len(locationIndexesByTimestamp) == 0 {
		return
	}

	for _, signals := range locationIndexesByTimestamp {
		latitudeIndx, ok := signals["latitude"]
		if !ok {
			continue
		}

		longitudeIndx, ok := signals["longitude"]
		if !ok {
			continue
		}

		latVal, ok := event.Data.Vehicle.Signals[latitudeIndx].Value.(float64)
		if !ok {
			continue
		}

		lngVal, ok := event.Data.Vehicle.Signals[longitudeIndx].Value.(float64)
		if !ok {
			continue
		}

		geo := h3.GeoCoord{Latitude: latVal, Longitude: lngVal}

		for _, fenceInd := range fence {
			res := h3.Resolution(fenceInd)
			statusInd := h3.FromGeo(geo, res)
			if statusInd == fenceInd {
				outGeo := h3.ToGeo(h3.ToParent(statusInd, res-1))

				event.Data.Vehicle.Signals[latitudeIndx].Value = &outGeo.Latitude
				event.Data.Vehicle.Signals[longitudeIndx].Value = &outGeo.Longitude
				event.Data.IsRedacted = ref(true)

				return
			}
		}
		event.Data.IsRedacted = ref(false)
	}
}

// findIndexPairsWithSameTimestamp returns a map of timestamps to a map of signal names(long and lat ) to their index in the slice
func findIndexPairsWithSameTimestamp(locationSignals []SignalData) map[int64]map[string]int {
	result := make(map[int64]map[string]int)

	for i, signal := range locationSignals {
		if signal.Name == "longitude" || signal.Name == "latitude" {
			if _, ok := result[signal.Timestamp]; !ok {
				result[signal.Timestamp] = make(map[string]int)
			}
			result[signal.Timestamp][signal.Name] = i
		}
	}

	return result
}

func (g *PrivacyV2) getFenceV2(ctx goka.Context) []h3.H3Index {
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
