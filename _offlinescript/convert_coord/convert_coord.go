package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

const (
	converterAPI = "https://developers.onemap.sg/commonapi/convert/3414to4326"
)

type Coord struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

func main() {
	csvfile, err := os.Open("./data/hdb-carpark-information.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	newfile, err := os.Create("./data/hdb-carpark-info-coverted.csv")
	if err != nil {
		log.Fatalln("Couldn't create new file", err)
	}

	r := csv.NewReader(csvfile)
	w := csv.NewWriter(newfile)
	defer w.Flush()

	//read header
	header, _ := r.Read()
	idxX, idxY := getIdxForCoords(header)

	w.Write(append(header, "latitude", "longitude"))

	// Iterate through the records and enrich them
	for {
		// Read each record from csv
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		url := getUrl(record[idxX], record[idxY])
		res, err := getData(url)
		if err != nil {
			log.Fatalln("cannot get coordinate", err)
		}

		w.Write(append(record, fmt.Sprintf("%f", res.Latitude), fmt.Sprintf("%f", res.Longitude)))
	}
}

func getIdxForCoords(header []string) (int, int) {
	idxX := 0
	idxY := 0
	for i, col := range header {
		if col == "x_coord" {
			idxX = i
		} else if col == "y_coord" {
			idxY = i
		}
	}
	return idxX, idxY
}

func getUrl(x string, y string) string {
	baseUrl, _ := url.Parse(converterAPI)
	params := url.Values{}
	params.Add("X", x)
	params.Add("Y", y)
	baseUrl.RawQuery = params.Encode()
	return baseUrl.String()
}

func getData(url string) (*Coord, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := Coord{}
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
