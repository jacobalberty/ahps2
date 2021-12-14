package ahps2

import (
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"time"
)

const (
	AHPS2_URL  = "https://water.weather.gov/ahps2/hydrograph_to_xml.php?output=xml"
	TIMEFORMAT = "2006-01-02T15:04:05-07:00"
)

type rateDatum struct {
	Text       string `xml:",chardata"`
	Stage      string `xml:"stage,attr"`
	StageUnits string `xml:"stageUnits,attr"`
	Flow       string `xml:"flow,attr"`
	FlowUnits  string `xml:"flowUnits,attr"`
}

type pointDatum struct {
	Text      string  `xml:",chardata"`
	Timestamp pdValid `xml:"valid"`
	Primary   struct {
		Value float64 `xml:",chardata"`
		Name  string  `xml:"name,attr"`
		Units string  `xml:"units,attr"`
	} `xml:"primary"`
	Secondary struct {
		Text  string `xml:",chardata"`
		Name  string `xml:"name,attr"`
		Units string `xml:"units,attr"`
	} `xml:"secondary"`
	Pedts string `xml:"pedts"`
}
type pdValid time.Time

func (pd *pdValid) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	var tV time.Time
	err := d.DecodeElement(&s, &start)
	if err != nil {
		return err
	}
	tV, err = time.Parse(TIMEFORMAT, s)
	if err != nil {
		return err
	}
	*pd = pdValid(tV)
	return nil
}

type Sigstage struct {
	Stage string
	Value float64
	Units string
}

// Site is the object containing all of the information about this measuring site.
type Site struct {
	site
	// This is a list of the different flood stages for the site you are examining.
	Sigstages map[string]Sigstage
}

type site struct {
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
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"low"`
		Action struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"action"`
		Bankfull struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"bankfull"`
		Flood struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"flood"`
		Moderate struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"moderate"`
		Major struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
		} `xml:"major"`
		Record struct {
			Value float64 `xml:",chardata"`
			Units string  `xml:"units,attr"`
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
		Text    string      `xml:",chardata"`
		Dignity string      `xml:"dignity,attr"`
		Datum   []rateDatum `xml:"datum"`
	} `xml:"rating"`
	AltRating struct {
		Text    string      `xml:",chardata"`
		Dignity string      `xml:"dignity,attr"`
		Datum   []rateDatum `xml:"datum"`
	} `xml:"alt_rating"`
	Observed struct {
		Text  string       `xml:",chardata"`
		Datum []pointDatum `xml:"datum"`
	} `xml:"observed"`
	Forecast struct {
		Text     string       `xml:",chardata"`
		Timezone string       `xml:"timezone,attr"`
		Issued   string       `xml:"issued,attr"`
		Datum    []pointDatum `xml:"datum"`
	} `xml:"forecast"`
}

// RiverPoint is a specific datapoint for a given time. This is usually returned by the Site object's public functions.
type RiverPoint struct {
	Value     float64
	Unit      string
	Timestamp time.Time
}

// GetStage returns the current flood stage level
func (s *Site) GetStage() (string, error) {
	resp := "unknown"
	mostRecent := s.site.Observed.Datum[0]
	cLevel := mostRecent.Primary.Value
	for _, v := range s.Sigstages {
		if cLevel >= v.Value {
			resp = v.Stage
		}
	}
	return resp, nil
}

// GetLevel returns a RiverPoint containing the current river level and any error occurred in parsing the data.
func (s *Site) GetLevel() (*RiverPoint, error) {
	mostRecent := s.Observed.Datum[0]

	timeStamp := time.Time(mostRecent.Timestamp)

	return &RiverPoint{Value: mostRecent.Primary.Value, Unit: mostRecent.Primary.Units, Timestamp: timeStamp}, nil
}

// GetCrest returns a RiverPoint containing the projected crest and any error occurred in parsing the data.
func (s *Site) GetCrest() (*RiverPoint, error) {
	var forecast = s.Forecast.Datum
	var crest = &RiverPoint{}
	var err error
	crest.Unit = forecast[0].Primary.Units
	crest.Value = forecast[0].Primary.Value
	if err != nil {
		return nil, err
	}
	crest.Timestamp = time.Time(forecast[0].Timestamp)

	for i := range forecast {
		cV := forecast[i].Primary.Value

		if cV > crest.Value {
			crest.Value = cV
			crest.Unit = forecast[i].Primary.Units
			crest.Timestamp = interface{}(forecast[i].Timestamp).(time.Time)
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
		Timeout: time.Second * 5, // Maximum of 5 seconds
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
	site := site{}
	err := xml.Unmarshal(data, &site)
	if err != nil {
		return nil, err
	}
	s := &Site{
		site:      site,
		Sigstages: make(map[string]Sigstage),
	}
	stages := site.Sigstages
	v := reflect.ValueOf(stages)
	typeOfS := v.Type()
	for i := 1; i < v.NumField(); i++ {
		FName := strings.ToLower(typeOfS.Field(i).Name)
		s.Sigstages[FName] = Sigstage{
			Stage: FName,
			Value: v.Field(i).FieldByName("Value").Float(),
			Units: v.Field(i).FieldByName("Units").String(),
		}
	}

	return s, nil
}
