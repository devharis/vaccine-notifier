package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type Location struct {
	name    string
	address string
	url     string
	link    string
}

type timeslot struct {
	Date  string `json:"date"`
	Slots []slot `json:"slots"`
}

type slot struct {
	When      string `json:"when"`
	Available bool   `json:"available"`
}

var (
	locations = []Location{
		{
			name:    "Kronans Apotek Ale Torg",
			address: "Ale Torg 7, Nödinge",
			url:     "https://booking-api.mittvaccin.se/clinique/2096/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2096",
		},
		{
			name:    "Kronans Apotek Alingsås",
			address: "Kungsgatan 34, Alingsås",
			url:     "https://booking-api.mittvaccin.se/clinique/2086/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2086",
		},
		{
			name:    "Kronans Apotek Borås Allégatan 43",
			address: "Allégatan 43, Borås",
			url:     "https://booking-api.mittvaccin.se/clinique/2081/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2081",
		},
		{
			name:    "Kronans Apotek Eriksbergs Köpcenter",
			address: "Kolhamnsgatan 1, Göteborg ",
			url:     "https://booking-api.mittvaccin.se/clinique/2092/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2092",
		},
		{
			name:    "Kronans Apotek Göteborg Wieselgrensplatsen 5",
			address: "Wieselgrensplatsen 5, Göteborg",
			url:     "https://booking-api.mittvaccin.se/clinique/2087/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2087",
		},
		{
			name:    "Kronans Apotek Hisings Backa",
			address: "Selma Lagerlöfs Torg 1, Hisings Backa",
			url:     "https://booking-api.mittvaccin.se/clinique/2091/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2091",
		},
		{
			name:    "Kronans Apotek Kongahälla Center",
			address: "Älvebacken 1, Kungälv",
			url:     "https://booking-api.mittvaccin.se/clinique/2082/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2082",
		},
		{
			name:    "Kronans Apotek Mariestad Vårdcentral",
			address: "Lockerudsvägen 10, Mariestad",
			url:     "https://booking-api.mittvaccin.se/clinique/2079/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2079",
		},
		{
			name:    "Kronans Apotek Mölnlycke Centrum",
			address: "Biblioteksgatan 4A, Mölnlycke",
			url:     "https://booking-api.mittvaccin.se/clinique/2095/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2095",
		},
		{
			name:    "Kronans Apotek Sjukhuset Falköping",
			address: "Danska Vägen 62, Falköping",
			url:     "https://booking-api.mittvaccin.se/clinique/2085/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2085",
		},
		{
			name:    "Kronans Apotek Skene",
			address: "Varbergsvägen 71, Skene",
			url:     "https://booking-api.mittvaccin.se/clinique/2084/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2084",
		},
		{
			name:    "Kronans Apotek Strömstad",
			address: "Södra Hamngatan 4, Strömstad",
			url:     "https://booking-api.mittvaccin.se/clinique/2083/appointments/16573/slots/210614-211201",
			link:    "https://bokning.mittvaccin.se/klinik/2083",
		},
	}
)

func main() {
	wait := make(chan struct{})
	for {
		fmt.Println("Wake up, time to search...")
		go search(wait)
		<-wait
	}
}

func search(ch chan struct{}) {
	for _, location := range locations {
		fmt.Println(location.name)
		resp, err := http.Get(location.url)
		if err != nil {
			log.Fatalln(err)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}
		//timeslots = make([]*TimeSlot, 0)
		timeslots := []timeslot{}
		if json.Unmarshal(body, &timeslots); err != nil {
			panic(err)
		}

		for _, timeslot := range timeslots {
			for _, slot := range timeslot.Slots {
				if slot.Available {
					message := fmt.Sprintf("Ledig tid %s, %s finns vid %s som ligger på adressen %s\n\n %s", timeslot.Date, slot.When, location.name, location.address, location.link)
					notify(message)
				}
			}
		}
	}
	fmt.Println("Search complete, go back to sleep...")
	time.Sleep(1 * time.Minute)
	ch <- struct{}{}
}

func notify(message string) {
	data := url.Values{
		"from":    {"Vaccintid"},
		"to":      {"+46000000000"},
		"message": {message}}

	req, _ := http.NewRequest("POST", "https://api.46elks.com/a1/SMS", bytes.NewBufferString(data.Encode()))

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	req.SetBasicAuth("clientid", "clientsecret")

	client := &http.Client{}
	resp, _ := client.Do(req)

	defer resp.Body.Close()
	ioutil.ReadAll(resp.Body)
}
