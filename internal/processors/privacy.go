package processors

import (
	"github.com/DIMO-Network/shared"
	"github.com/lovoo/goka"
	"github.com/rs/zerolog"
	"github.com/uber/h3-go/v4"
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

func (g *Privacy) Define() *goka.GroupGraph {
	return goka.DefineGroup(g.Group,
		goka.Input(g.StatusInput, new(shared.JSONCodec[shared.CloudEvent[StatusData]]), g.processStatusEvent),
		goka.Join(g.FenceTable, new(shared.JSONCodec[shared.CloudEvent[FenceData]])),
		goka.Output(g.StatusOutput, new(shared.JSONCodec[shared.CloudEvent[StatusData]])),
	)
}

func (g *Privacy) processStatusEvent(ctx goka.Context, msg interface{}) {
	fence := g.getFence(ctx)
	event := msg.(*shared.CloudEvent[StatusData])

	sanitizeEvent(event, fence)

	// Key should be the DIMO device id.
	ctx.Emit(g.StatusOutput, ctx.Key(), event)
}

// sanitizeEvent modifies the given CloudEvent using fence.
func sanitizeEvent(event *shared.CloudEvent[StatusData], fence []h3.Cell) {
	if event.Data.Latitude == nil || event.Data.Longitude == nil {
		return
	}

	geo := h3.NewLatLng(*event.Data.Latitude, *event.Data.Longitude)

	for _, fenceInd := range fence {
		// TODO: Should really validate res more.
		res := fenceInd.Resolution()
		// TODO: Cache these.
		statusInd := h3.LatLngToCell(geo, res)
		if statusInd == fenceInd {
			outGeo := statusInd.Parent(res - 1).LatLng()

			event.Data.Latitude, event.Data.Longitude = &outGeo.Lat, &outGeo.Lng
			event.Data.IsRedacted = ref(true)

			return
		}
	}

	event.Data.IsRedacted = ref(false)
}

func (g *Privacy) getFence(ctx goka.Context) []h3.Cell {
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

func ref[A any](a A) *A {
	return &a
}
