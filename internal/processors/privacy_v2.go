package processors

import (
	"github.com/DIMO-Network/shared"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/uber/h3-go/v4"
)

type PrivacyV2 struct {
	Group        goka.Group
	StatusInput  goka.Stream
	FenceTable   goka.Table
	StatusOutput goka.Stream

	Logger *zerolog.Logger
}

func (g *PrivacyV2) DefineV2() *goka.GroupGraph {
	return goka.DefineGroup(g.Group,
		goka.Input(g.StatusInput, new(shared.JSONCodec[StatusEventV2[StatusV2Data]]), g.processStatusEventV2),
		goka.Join(g.FenceTable, new(shared.JSONCodec[shared.CloudEvent[FenceData]])),
		goka.Output(g.StatusOutput, new(shared.JSONCodec[StatusEventV2[StatusV2Data]])),
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
func sanitizeEventV2(event *StatusEventV2[StatusV2Data], fence []h3.Cell) {
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

		geo := h3.NewLatLng(latVal, lngVal)

		for _, fenceInd := range fence {
			res := fenceInd.Resolution()
			statusInd := h3.LatLngToCell(geo, res)
			if statusInd == fenceInd {
				outGeo := statusInd.Parent(res - 1).LatLng()

				event.Data.Vehicle.Signals[latitudeIndx].Value = &outGeo.Lat
				event.Data.Vehicle.Signals[longitudeIndx].Value = &outGeo.Lng
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

func (g *PrivacyV2) getFenceV2(ctx goka.Context) []h3.Cell {
	val := ctx.Join(g.FenceTable)
	if val == nil {
		return nil
	}

	sIndexes := val.(*shared.CloudEvent[FenceData]).Data.H3Indexes
	out := make([]h3.Cell, len(sIndexes))
	for i, s := range sIndexes {
		out[i] = h3.Cell(h3.IndexFromString(s))
	}

	return out
}
