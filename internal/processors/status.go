package processors

import (
	"encoding/json"
	"fmt"
)

type StatusData struct {
	Latitude  *float64               `json:"latitude"`
	Longitude *float64               `json:"longitude"`
	Overflow  map[string]interface{} `json:"-"`
}

func (d *StatusData) MarshalJSON() ([]byte, error) {
	if d.Latitude != nil {
		d.Overflow["latitude"] = d.Latitude
	}

	if d.Longitude != nil {
		d.Overflow["longitude"] = d.Longitude
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
			d.Latitude = &lngF
		}
		delete(d.Overflow, "longitude")
	}

	return nil
}

type StatusEvent struct {
	CloudEvent
	Data StatusData `json:"data"`
}
