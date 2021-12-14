package ahps2

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"
)

const (
	AHPS2_URL = "https://water.weather.gov/ahps2/hydrograph_to_xml.php?output=xml"
)

// Site is the object containing all of the information about this measuring site.
type Site struct {
	XMLName                   xml.Name `xml:"site"`
	Text                      string   `xml:",chardata"`
	Xsi                       string   `xml:"xsi,attr"`
	Timezone                  string   `xml:"timezone,attr"`
	Originator                string   `xml:"originator,attr"`
	Name                      string   `xml:"name,attr"`
	ID                        string   `xml:"id,attr"`
	NoNamespaceSchemaLocation string   `xml:"noNamespaceSchemaLocation,attr"`
	Generationtime            string   `xml:"generationtime,attr"`
	Disclaimers               struct {
		Text           string `xml:",chardata"`
		AHPSXMLversion string `xml:"AHPSXMLversion"`
		Status         string `xml:"status"`
		Quality        struct {
			Text  string `xml:",chardata"`
			Trace string `xml:"trace,attr"`
		} `xml:"quality"`
		Observed struct {
			Text     string `xml:",chardata"`
			Trace    string `xml:"trace,attr"`
			Audience string `xml:"audience,attr"`
		} `xml:"observed"`
		General struct {
			Text     string `xml:",chardata"`
			Audience string `xml:"audience,attr"`
		} `xml:"general"`
		Standing struct {
			Text     string `xml:",chardata"`
			Trace    string `xml:"trace,attr"`
			Audience string `xml:"audience,attr"`
			Dignity  string `xml:"dignity,attr"`
		} `xml:"standing"`
	} `xml:"disclaimers"`
	Sigstages struct {
		Text string `xml:",chardata"`
		Low  struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"low"`
		Action struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"action"`
		Bankfull struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"bankfull"`
		Flood struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"flood"`
		Moderate struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"moderate"`
		Major struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"major"`
		Record struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"record"`
	} `xml:"sigstages"`
	Sigflows struct {
		Text string `xml:",chardata"`
		Low  struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"low"`
		Action   string `xml:"action"`
		Bankfull struct {
			Text  string `xml:",chardata"`
			Units string `xml:"units,attr"`
		} `xml:"bankfull"`
		Flood    string `xml:"flood"`
		Moderate string `xml:"moderate"`
		Major    string `xml:"major"`
		Record   string `xml:"record"`
	} `xml:"sigflows"`
	Zerodatum struct {
		Text  string `xml:",chardata"`
		Units string `xml:"units,attr"`
	} `xml:"zerodatum"`
	Rating struct {
		Text    string `xml:",chardata"`
		Dignity string `xml:"dignity,attr"`
		Datum   []struct {
			Text       string `xml:",chardata"`
			Stage      string `xml:"stage,attr"`
			StageUnits string `xml:"stageUnits,attr"`
			Flow       string `xml:"flow,attr"`
			FlowUnits  string `xml:"flowUnits,attr"`
		} `xml:"datum"`
	} `xml:"rating"`
	AltRating struct {
		Text    string `xml:",chardata"`
		Dignity string `xml:"dignity,attr"`
		Datum   []struct {
			Text       string `xml:",chardata"`
			Stage      string `xml:"stage,attr"`
			StageUnits string `xml:"stageUnits,attr"`
			Flow       string `xml:"flow,attr"`
			FlowUnits  string `xml:"flowUnits,attr"`
		} `xml:"datum"`
	} `xml:"alt_rating"`
	Observed struct {
		Text  string `xml:",chardata"`
		Datum []struct {
			Text  string `xml:",chardata"`
			Valid struct {
				Text     string `xml:",chardata"`
				Timezone string `xml:"timezone,attr"`
			} `xml:"valid"`
			Primary struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"name,attr"`
				Units string `xml:"units,attr"`
			} `xml:"primary"`
			Secondary struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"name,attr"`
				Units string `xml:"units,attr"`
			} `xml:"secondary"`
			Pedts string `xml:"pedts"`
		} `xml:"datum"`
	} `xml:"observed"`
	Forecast struct {
		Text     string `xml:",chardata"`
		Timezone string `xml:"timezone,attr"`
		Issued   string `xml:"issued,attr"`
		Datum    []struct {
			Text  string `xml:",chardata"`
			Valid struct {
				Text     string `xml:",chardata"`
				Timezone string `xml:"timezone,attr"`
			} `xml:"valid"`
			Primary struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"name,attr"`
				Units string `xml:"units,attr"`
			} `xml:"primary"`
			Secondary struct {
				Text  string `xml:",chardata"`
				Name  string `xml:"name,attr"`
				Units string `xml:"units,attr"`
			} `xml:"secondary"`
			Pedts string `xml:"pedts"`
		} `xml:"datum"`
	} `xml:"forecast"`
}

// RiverPoint is a specific datapoint for a given time. This is usually returned by the Site object's public functions.
type RiverPoint struct {
	Value     float64
	Unit      string
	Timestamp time.Time
}

// GetStage returns the current flood stage level
// This function assumes Sigstages are always in order
func (s *Site) GetStage() (string, error) {
	resp := "unknown"
	mostRecent := s.Observed.Datum[0]
	cLevel, err := strconv.ParseFloat(mostRecent.Primary.Text, 32)
	if err != nil {
		return "", err
	}
	stages := s.Sigstages
	v := reflect.ValueOf(stages)
	typeOfS := v.Type()
	for i := 1; i < v.NumField(); i++ {
		FName := typeOfS.Field(i).Name
		FVal, err := strconv.ParseFloat(v.Field(i).FieldByName("Text").String(), 32)
		if err != nil {
			return "", err
		}
		if cLevel >= FVal {
			resp = FName
		}
	}
	return resp, nil
}

// GetLevel returns a RiverPoint containing the current river level and any error occurred in parsing the data.
func (s *Site) GetLevel() (*RiverPoint, error) {
	mostRecent := s.Observed.Datum[0]
	cLevel, err := strconv.ParseFloat(mostRecent.Primary.Text, 32)
	if err != nil {
		return nil, err
	}
	unit := mostRecent.Primary.Units
	timeStamp, err := time.Parse(time.RFC3339, mostRecent.Valid.Text)
	if err != nil {
		return nil, err
	}
	return &RiverPoint{Value: cLevel, Unit: unit, Timestamp: timeStamp}, nil
}

// GetCrest returns a RiverPoint containing the projected crest and any error occurred in parsing the data.
func (s *Site) GetCrest() (*RiverPoint, error) {
	var forecast = s.Forecast.Datum
	var crest = &RiverPoint{}
	var err error
	crest.Unit = forecast[0].Primary.Units
	crest.Value, err = strconv.ParseFloat(forecast[0].Primary.Text, 32)
	if err != nil {
		return nil, err
	}
	crest.Timestamp, err = time.Parse(time.RFC3339, forecast[0].Valid.Text)
	if err != nil {
		return nil, err
	}
	for i := range forecast {
		cV, err := strconv.ParseFloat(forecast[i].Primary.Text, 32)
		if err != nil {
			return nil, err
		}
		if cV > crest.Value {
			crest.Value = cV
			crest.Unit = forecast[i].Primary.Units
			crest.Timestamp, err = time.Parse(time.RFC3339, forecast[i].Valid.Text)
			if err != nil {
				return nil, err
			}
		}
	}
	return crest, nil
}

// GetSite retrieves the data for the specified gauge and unmarshals it into a Site object
// It returns a site object and any error occurred during retrieval and unmarshalling
func GetSite(gauge string) (*Site, error) {
	return getSite(AHPS2_URL, gauge)
}

func getSite(ahps2url, gauge string) (*Site, error) {
	u, _ := url.Parse(ahps2url)
	q := u.Query()
	q.Set("gage", gauge)
	u.RawQuery = q.Encode()

	Client := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	res, err := Client.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	return unMarshalSite(body)
}

func unMarshalSite(data []byte) (*Site, error) {
	site := &Site{}
	err := xml.Unmarshal(data, &site)
	if err != nil {
		return nil, err
	}

	return site, nil
}
