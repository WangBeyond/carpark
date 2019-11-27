package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	availabilityAPI = "https://api.data.gov.sg/v1/transport/carpark-availability"
)

type Availability struct {
	Items []Item `json"items"`
}

type Item struct {
	Timestamp   string        `json:"timestamp"`
	CarparkData []CarparkData `json:"carpark_data"`
}

type CarparkData struct {
	CarparkInfo    []CarparkInfo `json:"carpark_info"`
	CarparkNumber  string        `json:"carpark_number"`
	UpdateDatetime string        `json:"update_datetime"`
}

type CarparkInfo struct {
	TotalLots     string `json:"total_lots"`
	LotType       string `json:"lot_type"`
	LotsAvailable string `json:"lots_available"`
}

func (s *Server) UpdateAvailability(w http.ResponseWriter, r *http.Request) {
	// fetch data
	defer func(now time.Time) {
		log.Println("update availability took ", time.Now().Sub(now).Seconds(), "sec")
	}(time.Now())

	url := getUrl()
	log.Println("trying to query with URL: ", url)
	res, err := getAvailability(url)
	if err != nil {
		log.Fatalln("query url got error ", err)
		w.WriteHeader(500)
		return
	}

	// update db
	err = s.updateDB(*res)
	if err != nil {
		log.Fatalln("update DB failed ", err)
		w.WriteHeader(500)
		return
	}
	s.invalidateCache()

	w.WriteHeader(200)
}

func (s *Server) updateDB(data Availability) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	statement, err := tx.Prepare("UPDATE carpark_info SET total_lots=?, lot_type=?, lots_available=? WHERE car_park_no=?")
	if err != nil {
		tx.Rollback()
		log.Fatalln("prepare statement failed", err)
		return err
	}
	defer statement.Close()

	for _, item := range data.Items {
		for _, data := range item.CarparkData {
			if len(data.CarparkInfo) == 0 {
				continue
			}
			totalLots, err := strconv.ParseInt(data.CarparkInfo[0].TotalLots, 10, 64)
			if err != nil {
				return err
			}

			lotsAvailable, err := strconv.ParseInt(data.CarparkInfo[0].LotsAvailable, 10, 64)
			if err != nil {
				return err
			}

			lotType := data.CarparkInfo[0].LotType
			carparkNo := data.CarparkNumber

			if _, err := statement.Exec(totalLots, lotType, lotsAvailable, carparkNo); err != nil {
				tx.Rollback()
				log.Fatalln("execute statement failed", err)
				return err
			}
		}

	}
	err = tx.Commit()
	if err != nil {
		log.Fatalln("committing transaction failed", err)
		return err
	}
	log.Println("updating DB succeeded")
	return nil
}

func getUrl() string {
	baseUrl, _ := url.Parse(availabilityAPI)
	params := url.Values{}
	now := time.Now().Format("2006-01-02T15:04:05")
	params.Set("date_time", now)
	baseUrl.RawQuery = params.Encode()
	return baseUrl.String()
}

func getAvailability(url string) (*Availability, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	res := Availability{}
	if err = json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	return &res, nil
}
