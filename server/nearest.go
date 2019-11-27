package main

import (
	"encoding/json"
	"errors"
	"log"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
)

const (
	earthRadius    = float64(6371)
	defaultPage    = 1
	defaultPerPage = 10
)

type CarparkAvailability struct {
	Address       string  `json:"address"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	TotalLots     int64   `json:"total_lots"`
	AvailableLots int64   `json:"available_lots"`
}

type NearestRequest struct {
	Latitude  float64
	Longitude float64
	Page      int64
	PerPage   int64
}

func (s *Server) GetNearest(w http.ResponseWriter, r *http.Request) {
	// fetch data
	defer func(now time.Time) {
		log.Println("query nearest took ", time.Now().Sub(now).Seconds(), "sec")
	}(time.Now())

	// read lat, lng from request
	req, err := s.parseRequest(r)
	if err != nil {
		log.Fatalln("invalid request: ", err)
		w.WriteHeader(400)
		return
	}
	log.Println("query nearest with param: ", req)

	// fetch carpark list from cache; if not exists, fallback to DB
	carparkList, ok := s.getCache()
	if !ok {
		carparkList, err = s.queryDB()
		if err != nil {
			log.Fatalln("failed to query carpark info ", err)
			w.WriteHeader(500)
			return
		}

		s.updateCache(carparkList)
	}

	sortCarparksByDist(carparkList, req)

	start, end := getSliceStartAndEnd(len(carparkList), req)
	encoded, _ := json.Marshal(carparkList[start:end])
	w.Write(encoded)
	w.WriteHeader(200)
}

func getSliceStartAndEnd(size int, req NearestRequest) (int64, int64) {
	start := (req.Page - 1) * req.PerPage
	if start > int64(size) {
		return 0, 0
	}
	end := req.Page * req.PerPage
	if end > int64(size) {
		return start, int64(size)
	}
	return start, end
}

func sortCarparksByDist(carparkList []CarparkAvailability, req NearestRequest) {
	sort.SliceStable(carparkList, func(i, j int) bool {
		lat1, lng1 := carparkList[i].Latitude, carparkList[i].Longitude
		dist1 := Haversine(req.Longitude, req.Latitude, lng1, lat1)

		lat2, lng2 := carparkList[j].Latitude, carparkList[j].Longitude
		dist2 := Haversine(req.Longitude, req.Latitude, lng2, lat2)
		// the one shorter distance will be sorted in preceding order
		return dist1 < dist2
	})

}

func (s *Server) parseRequest(r *http.Request) (NearestRequest, error) {
	latStr := r.URL.Query()["latitude"]
	lngStr := r.URL.Query()["longitude"]
	pageStr := r.URL.Query()["page"]
	perPageStr := r.URL.Query()["per_page"]

	if len(latStr) == 0 || len(lngStr) == 0 {
		return NearestRequest{}, errors.New("no latitude or longitude was passed")
	}

	lat, err := strconv.ParseFloat(latStr[0], 64)
	if err != nil {
		return NearestRequest{}, err
	}
	lng, err := strconv.ParseFloat(lngStr[0], 64)
	if err != nil {
		return NearestRequest{}, err
	}

	// set default value for page and defa
	req := NearestRequest{
		Latitude:  lat,
		Longitude: lng,
		Page:      defaultPage,
		PerPage:   defaultPerPage,
	}

	if len(pageStr) > 0 {
		page, err := strconv.ParseInt(pageStr[0], 10, 64)
		if err != nil {
			return NearestRequest{}, err
		}
		if page < 1 {
			return NearestRequest{}, errors.New("invalid page number")
		}
		req.Page = page
	}

	if len(perPageStr) > 0 {
		perPage, err := strconv.ParseInt(perPageStr[0], 10, 64)
		if err != nil {
			return NearestRequest{}, err
		}
		if perPage < 1 {
			return NearestRequest{}, errors.New("invalid per_page number")
		}
		req.PerPage = perPage
	}

	return req, nil
}

/*
 * The haversine formula will calculate the spherical distance as the crow flies
 * between lat and lon for two given points in km
 */
func Haversine(lonFrom float64, latFrom float64, lonTo float64, latTo float64) (distance float64) {
	var deltaLat = (latTo - latFrom) * (math.Pi / 180)
	var deltaLon = (lonTo - lonFrom) * (math.Pi / 180)

	var a = math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(latFrom*(math.Pi/180))*math.Cos(latTo*(math.Pi/180))*
			math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	var c = 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	distance = earthRadius * c

	return
}

func (s *Server) queryDB() ([]CarparkAvailability, error) {
	statement := `SELECT address, latitude, longitude, total_lots, lots_available FROM carpark_info WHERE lots_available > 0`
	rows, err := s.db.Query(statement)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	res := []CarparkAvailability{}
	for rows.Next() {
		var address string
		var lat, lng float64
		var totalLots, availableLots int64

		if err := rows.Scan(&address, &lat, &lng, &totalLots, &availableLots); err != nil {
			return nil, err
		}

		record := CarparkAvailability{
			Address:       address,
			Latitude:      lat,
			Longitude:     lng,
			TotalLots:     totalLots,
			AvailableLots: availableLots,
		}
		res = append(res, record)
	}
	return res, nil
}
