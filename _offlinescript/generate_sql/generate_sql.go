package main

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

const (
	initSql = `
CREATE DATABASE IF NOT EXISTS carpark ; 

USE carpark;

DROP TABLE IF EXISTS carpark_info;

CREATE TABLE carpark_info (
	car_park_no varchar(50) NOT NULL,
  	address varchar(255) NOT NULL,
  	x_coord double,
  	y_coord double,
  	car_park_type varchar(50) NOT NULL,
	type_of_parking_system varchar(50) NOT NULL,
	short_term_parking varchar(50) NOT NULL,
	free_parking varchar(50) NOT NULL,
	night_parking varchar(50) NOT NULL,
	car_park_decks int NOT NULL,
	gantry_height double NOT NULL,
	car_park_basement varchar(50) NOT NULL,
	latitude double NOT NULL,
	longitude double NOT NULL,
	total_lots int NOT NULL,
	lot_type varchar(25) NOT NULL,
	lots_available int NOT NULL,
  	updated_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
  	created_at datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY (car_park_no)
)

;
`
)

type Response struct {
	Help    string `json:"help"`
	Success bool   `json:"success"`
	Result  Result `json:"result"`
}

type Result struct {
	ResourceID string   `json:"resource_id"`
	Fields     []Field  `json:"fields"`
	Records    []Record `json:"records"`
	Limit      int      `json:"limit"`
	Total      int      `json:"total"`
}

type Field struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type Record struct {
	ID               string `json:"_id"`
	ShortTermParking string `json:"short_term_parking"`
	CarParkType      string `json:"car_park_type"`
	YCoord           string `json:"y_coord"`
	XCoord           string `json:"x_coord"`
	FreeParking      string `json:"free_parking"`
	GantryHeight     string `json:"gantry_height"`
}

type Coord struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type Metadata struct {
	Resources []Resource `yaml:"Resources"`
}

type Resource struct {
	Schema []Schema `yaml:"Schema"`
}

type Schema struct {
	Name string `yaml:"Name"`
	Type string `yaml:"Type"`
}

var schema []Schema

func main() {
	yamlFile, err := ioutil.ReadFile("data/metadata-hdb-carpark-information.txt")
	if err != nil {
		log.Fatalf("failed to read yaml ", err)
	}

	metadata := Metadata{}
	err = yaml.Unmarshal(yamlFile, &metadata)
	if err != nil {
		log.Fatalf("failed to unmarshal: %s", err)
	}
	fmt.Println(metadata)
	schema = append(metadata.Resources[0].Schema,
		Schema{Name: "latitude", Type: "numeric"},
		Schema{Name: "longitude", Type: "numeric"},
		Schema{Name: "total_lots", Type: "numeric"},
		Schema{Name: "lot_type", Type: "text"},
		Schema{Name: "lots_available", Type: "numeric"},
	)

	csvfile, err := os.Open("./data/hdb-carpark-info-coverted.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	newfile, err := os.Create("./data/db-init.sql")
	if err != nil {
		log.Fatalln("Couldn't create new file", err)
	}

	r := csv.NewReader(csvfile)
	w := bufio.NewWriter(newfile)
	defer w.Flush()

	//read header
	w.WriteString(initSql)

	// skip the first header row
	_, _ = r.Read()
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

		sqlCommand := generateSql(append(record, "0", "", "0"))
		w.WriteString(sqlCommand)
	}
}

func generateSql(record []string) string {
	for i, tuple := range record {
		switch schema[i].Type {
		case "text":
			record[i] = fmt.Sprintf("\"%s\"", tuple)
		case "geo_coordinate", "numeric":
			record[i] = fmt.Sprintf("%s", tuple)
		}
	}

	return fmt.Sprintf(`INSERT INTO carpark_info (car_park_no, address, x_coord, y_coord, car_park_type,
	type_of_parking_system, short_term_parking, free_parking, night_parking, car_park_decks, gantry_height,
	car_park_basement, latitude, longitude, total_lots, lot_type, lots_available ) VALUES (%s);\n`, strings.Join(record, ","))
}
