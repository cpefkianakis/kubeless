package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type Page struct {
	Title      string
	Body       []byte
	Longitude1 string
	Latitude1  string
	Longitude2 string
	Latitude2  string
	Distance   string
	City       string
	regionName string
	Country    string
}

type IPGeolocation struct {
	Status     string  `json:"status"`
	Lat        float32 `json:"lat"`
	Lon        float32 `json:"lon"`
	City       string  `json:"city"`
	RegionName string  `json:"regionName"`
	Country    string  `json:"country"`
}

type ISSResponse struct {
	ISSPosition struct {
		Longitude string `json:"longitude"`
		Latitude  string `json:"latitude"`
	} `json:"iss_position"`
	Message   string `json:"message"`
	Timestamp int    `json:"timestamp"`
}

func loadPage(title string) (*Page, error) {

	filename := title + ".txt"
	body, err := os.ReadFile(filename)
	var longitude1 string
	var latitude1 string
	var longitude2 string
	var latitude2 string
	var distance string

	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body, Longitude1: longitude1, Latitude1: latitude1, Longitude2: longitude2, Latitude2: latitude2, Distance: distance}, nil
}

func renderTemplate(w http.ResponseWriter, tmpl string, p *Page) {

	t, _ := template.ParseFiles(tmpl + ".html")
	t.Execute(w, p)
}

func findMyIP() string {

	url := "https://api.ipify.org?format=text"
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(bodyText)

}

func getIPGeolocation() (string, string, string, string, string) {

	myIP := findMyIP()
	var url string = "http://ip-api.com/json/" + myIP
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	var result IPGeolocation
	if err := json.Unmarshal(bodyText, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON IP Geolocation")
	}
	var lon string = strconv.FormatFloat(float64(result.Lon), 'f', 4, 32)
	var lat string = strconv.FormatFloat(float64(result.Lat), 'f', 4, 32)
	return string(lon), lat, result.City, result.RegionName, result.Country
}

func getISSPosition() (string, string) {
	client := &http.Client{}
	var data = strings.NewReader(`{"url":"http://api.open-notify.org/iss-now.json"}`)
	req, err := http.NewRequest("POST", "http://localhost:8081/api/v1/namespaces/default/services/python-http-request:8080", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var result ISSResponse
	if err := json.Unmarshal(bodyText, &result); err != nil { // Parse []byte to the go struct pointer
		fmt.Println("Can not unmarshal JSON ISS Location")
	}
	fmt.Println(result.ISSPosition.Longitude, result.ISSPosition.Latitude)

	return result.ISSPosition.Longitude, result.ISSPosition.Latitude
}

//HTML page handlers

func indexHandler(w http.ResponseWriter, r *http.Request) {

	title := r.URL.Path[len("/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	renderTemplate(w, "templates/index", p)
}

func locationHandler(w http.ResponseWriter, r *http.Request) {

	lon, lat := getISSPosition()
	title := r.URL.Path[len("/location/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title, Longitude1: lon, Latitude1: lat}
	}
	renderTemplate(w, "templates/location", p)

}

func distanceHandler(w http.ResponseWriter, r *http.Request) {

	lonISS, latISS := getISSPosition()

	lonIP, latIP, city, region, country := getIPGeolocation()

	client := &http.Client{}
	u, err_marshal := json.Marshal(map[string]string{"longitude1": lonISS, "latitude1": latISS, "longitude2": lonIP, "latitude2": latIP})
	fmt.Println(string(u))
	if err_marshal != nil {
		log.Fatal(err_marshal)
	}

	var data = strings.NewReader(string(u))
	req, err := http.NewRequest("POST", "http://localhost:8082/api/v1/namespaces/default/services/calculate-distance:8080", data)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	bodyText, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	title := r.URL.Path[len("/location/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title, Longitude2: lonIP, Latitude2: latIP, City: city, regionName: region, Country: country, Distance: string(bodyText)}
	}
	renderTemplate(w, "templates/distance", p)

}

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/location/", locationHandler)
	http.HandleFunc("/distance/", distanceHandler)
	log.Fatal(http.ListenAndServe(":8090", nil))
}
