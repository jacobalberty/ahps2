package ahps2

import (
	_ "embed"
	"math"
	"testing"
	"time"
)

const (
	validGauge   = "btrl1"
	invalidGauge = "invalidgauge"
)

//go:embed testdata/btrl1.xml
var btrl1 []byte

func returnDummySite() *Site {
	site, err := unMarshalSite(btrl1)
	if err != nil {
		panic(err)
	}
	return site
}

func TestValidSite(t *testing.T) {
	site, err := GetSite(validGauge)
	if err != nil {
		t.Errorf("Error in valid gauge: %s", err.Error())
	}
	if site == nil {
		t.Errorf("Should not recieve nil for a valid gauge")
	}
}

func TestInvalidSite(t *testing.T) {
	site, err := GetSite(invalidGauge)
	if err == nil {
		t.Errorf("Expected error for invalid gauge")
	}
	if site != nil {
		t.Errorf("Something other than nil returned for invalid gauge")
	}
}

func TestGetStage(t *testing.T) {
	validStages := map[string]bool{"low": true, "action": true, "bankfull": true, "flood": true, "moderate": true, "major": true, "record": true}
	site := returnDummySite()
	stage, err := site.GetStage()
	if err != nil {
		t.Errorf("Got error from GetStage: %s", err.Error())
	}
	if _, ok := validStages[stage]; ok {
		t.Errorf("Got invalid stage '%s'", stage)
	}
}

func TestGetLevel(t *testing.T) {
	cst := time.FixedZone("CST", int((-6 * time.Hour).Seconds()))
	site := returnDummySite()
	level, err := site.GetLevel()
	if err != nil {
		t.Errorf("Got error from GetStage: %s", err.Error())
	}
	eTime := time.Date(2021, 12, 14, 10, 00, 00, 00, cst)
	eValue := 7.67

	if !level.Timestamp.Equal(eTime) {
		t.Errorf("Time for most recent level was not correct.\nExpected: %v, got %v", eTime, level.Timestamp)
	}

	if diff := math.Abs(level.Value - eValue); diff > 0.001 {
		t.Errorf("Value for most recent level was not correct.\nExpected: %.2f, got %.2f", eValue, level.Value)
	}
}
