package processors

import (
	"encoding/json"
	"fmt"
	"github.com/DIMO-Network/shared"
)

type StatusData struct {
	Latitude   *float64       `json:"latitude"`
	Longitude  *float64       `json:"longitude"`
	IsRedacted *bool          `json:"isRedacted"`
	Overflow   map[string]any `json:"-"`
}

func (d *StatusData) MarshalJSON() ([]byte, error) {
	if d.Latitude != nil {
		d.Overflow["latitude"] = d.Latitude
	}

	if d.Longitude != nil {
		d.Overflow["longitude"] = d.Longitude
	}

	if d.IsRedacted != nil {
		d.Overflow["isRedacted"] = *d.IsRedacted
	}

	return json.Marshal(d.Overflow)
}

func (d *StatusData) UnmarshalJSON(data []byte) error {
	if err := json.Unmarshal(data, &d.Overflow); err != nil {
		return err
	}

	if lat, ok := d.Overflow["latitude"]; ok {
		if lat != nil {
			latF, ok := lat.(float64)
			if !ok {
				return fmt.Errorf("latitude field was not a JSON number")
			}
			d.Latitude = &latF
		}
		delete(d.Overflow, "latitude")
	}

	if lng, ok := d.Overflow["longitude"]; ok {
		if lng != nil {
			lngF, ok := lng.(float64)
			if !ok {
				return fmt.Errorf("longitude field was not a JSON number")
			}
			d.Longitude = &lngF
		}
		delete(d.Overflow, "longitude")
	}

	if ir, ok := d.Overflow["isRedacted"]; ok {
		if ir != nil {
			irB, ok := ir.(bool)
			if !ok {
				return fmt.Errorf("isRedacted field was not a JSON boolean")
			}
			d.IsRedacted = &irB
		}
		delete(d.Overflow, "isRedacted")
	}

	return nil
}

type StatusEventV2[A any] struct {
	shared.CloudEvent[A]
	TokenID   uint64 `json:"vehicleTokenId"`
	Signature string `json:"signature"`
}

type Vehicle struct {
	Make    string       `json:"make,omitempty"`
	Model   string       `json:"model,omitempty"`
	Year    int          `json:"year,omitempty"`
	Signals []SignalData `json:"signals,omitempty"`
}

type SignalData struct {
	// Timestamp is in unix millis, when signal was queried
	Timestamp int64  `json:"timestamp,omitempty"`
	Name      string `json:"name,omitempty"`
	Value     any    `json:"value,omitempty"`
}
type StatusV2Data struct {
	// Timestamp is in unix millis, when payload was sent
	Timestamp int64                  `json:"timestamp,omitempty"`
	Device    map[string]interface{} `json:"device,omitempty"`
	Vehicle   Vehicle                `json:"vehicle,omitempty"`
}
