package main

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHaversine(t *testing.T) {
	// Distance from York to Bristol
	dist := Haversine(1.0803, 53.9583, 2.5833, 51.4500)
	assert.Equal(t, "296.71", fmt.Sprintf("%.2f", dist))
}

func TestSortCarparksByDist(t *testing.T) {
	carparkList := []CarparkAvailability{
		{
			Address:       "BLK 364 / 365 UPP SERANGOON RD",
			Latitude:      1.37011,
			Longitude:     103.897,
			TotalLots:     471,
			AvailableLots: 324,
		},
		{
			Address:       "BLK 401-413, 460-463 HOUGANG AVENUE 10",
			Latitude:      1.37429,
			Longitude:     103.896,
			TotalLots:     693,
			AvailableLots: 182,
		},
		{
			Address:       "BLK 351-357 HOUGANG AVENUE 7",
			Latitude:      1.37234,
			Longitude:     103.899,
			TotalLots:     249,
			AvailableLots: 143,
		},
	}

	expected := []CarparkAvailability{
		{
			Address:       "BLK 401-413, 460-463 HOUGANG AVENUE 10",
			Latitude:      1.37429,
			Longitude:     103.896,
			TotalLots:     693,
			AvailableLots: 182,
		},
		{
			Address:       "BLK 351-357 HOUGANG AVENUE 7",
			Latitude:      1.37234,
			Longitude:     103.899,
			TotalLots:     249,
			AvailableLots: 143,
		},
		{
			Address:       "BLK 364 / 365 UPP SERANGOON RD",
			Latitude:      1.37011,
			Longitude:     103.897,
			TotalLots:     471,
			AvailableLots: 324,
		},
	}

	req := NearestRequest{
		Latitude:  1.37326,
		Longitude: 103.897,
	}

	sortCarparksByDist(carparkList, req)
	assert.Equal(t, expected, carparkList)
}

func TestParseRequest(t *testing.T) {
	testData := []struct {
		desc           string
		lat            string
		lng            string
		expectedResult NearestRequest
		expectedErr    bool
	}{
		{
			desc: "happy path",
			lat:  "12.31",
			lng:  "3.1",
			expectedResult: NearestRequest{
				Latitude:  12.31,
				Longitude: 3.1,
				Page:      1,
				PerPage:   10,
			},
		},
		{
			desc:        "invalid coordinate",
			lat:         "12.31",
			lng:         "test",
			expectedErr: true,
		},
	}

	for _, td := range testData {
		s := &Server{}
		baseUrl, _ := url.Parse("http://test")
		req := &http.Request{URL: baseUrl}
		query := url.Values{}
		query.Set("latitude", td.lat)
		query.Set("longitude", td.lng)
		req.URL.RawQuery = query.Encode()
		_, err := s.parseRequest(req)
		assert.Equal(t, td.expectedErr, err != nil, td.desc)
	}
}

func TestGetSliceStartAndEnd(t *testing.T) {
	testData := []struct {
		desc          string
		size          int
		page          int64
		perPage       int64
		expectedStart int64
		expectedEnd   int64
	}{
		{
			desc:          "end within range",
			size:          30,
			page:          3,
			perPage:       3,
			expectedStart: 6,
			expectedEnd:   9,
		},
		{
			desc:          "end equal to last one",
			size:          30,
			page:          3,
			perPage:       10,
			expectedStart: 20,
			expectedEnd:   30,
		},
		{
			desc:          "end over index",
			size:          35,
			page:          4,
			perPage:       10,
			expectedStart: 30,
			expectedEnd:   35,
		},
		{
			desc:          "start over index",
			size:          35,
			page:          5,
			perPage:       10,
			expectedStart: 0,
			expectedEnd:   0,
		},
	}

	for _, td := range testData {
		req := NearestRequest{
			PerPage: td.perPage,
			Page:    td.page,
		}
		start, end := getSliceStartAndEnd(td.size, req)
		assert.Equal(t, td.expectedStart, start, td.desc)
		assert.Equal(t, td.expectedEnd, end, td.desc)
	}
}
