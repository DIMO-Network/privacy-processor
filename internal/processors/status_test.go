package processors

import (
	"encoding/json"
	"testing"
)

func TestStatusUnmarshal(t *testing.T) {
	d := new(StatusData)
	if err := json.Unmarshal([]byte(`{"latitude": 44.5, "odometer": 22.1}`), d); err != nil {
		t.Errorf("Failed to unmarshal status data: %v", err)
	}
	if d.Latitude == nil || *d.Latitude != 44.5 {
		t.Errorf("Expected Latitude field to be a pointer to 44.5")
	}
	if d.Longitude != nil {
		t.Errorf("Expected Longitude field to be absent")
	}
	if d.Overflow["odometer"] != 22.1 {
		t.Errorf("Expected JSON odometer field to be be stored in overflow map")
	}
}

func TestStatusMarshal(t *testing.T) {
	lat, lng := 1.2, 3.4
	d := &StatusData{
		Latitude:   &lat,
		Longitude:  &lng,
		IsRedacted: ref(true),
		Overflow:   map[string]any{"charging": false},
	}

	bytes, err := json.Marshal(d)
	if err != nil {
		t.Errorf("Failed to marshal status data: %v", err)
	}

	m := make(map[string]any)
	if err := json.Unmarshal(bytes, &m); err != nil {
		t.Errorf("Failed to parse marshal output: %v", err)
	}

	if m["latitude"] != lat {
		t.Errorf("Expected latitude field to be 1.2")
	}

	if m["longitude"] != lng {
		t.Errorf("Expected longitude field to be 3.4")
	}

	if m["charging"] != false {
		t.Errorf("Expected charging field to be false, but was %v", m["charging"])
	}

	if m["isRedacted"] != true {
		t.Errorf("Expected isRedacted field to be true, but was %v", m["isRedacted"])
	}
}
