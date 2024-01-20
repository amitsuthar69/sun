package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/fatih/color"
)

type Weather struct {
	Location struct {
		Name    string `json:"name"`
		Country string `json:"country"`
	} `json:"location"`
	Current struct {
		TempC     float64 `json:"temp_c"`
		Condition struct {
			Text string `json:"text"`
		} `json:"condition"`
	} `json:"current"`
	Forecast struct {
		Forecastday []struct {
			Hour []struct {
				TimeEpoch int64   `json:"time_epoch"`
				TempC     float64 `json:"temp_c"`
				Condition struct {
					Text string `json:"text"`
				} `json:"condition"`
				ChanceOfRain float64 `json:"chance_of_rain"`
			} `json:"hour"`
		} `json:"forecastday"`
	} `json:"forecast"`
}

func main() {
	loc := "mumbai"
	if len(os.Args) >= 2 {
		loc = os.Args[1]
	}
	url := "http://api.weatherapi.com/v1/forecast.json?key=588129751b7940ca824160458242001&q=" + loc + "&days=1&aqi=no&alerts=no"
	res, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		panic("api error")
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic("error reading response body")
	}
	var weather Weather
	err = json.Unmarshal(body, &weather)
	if err != nil {
		panic(err)
	}
	// fmt.Println(weather)
	location, current, hours := weather.Location, weather.Current, weather.Forecast.Forecastday[0].Hour
	fmt.Printf(
		"%s, %s: %.0f°C, %s, %s\n",
		location.Country,
		location.Name,
		current.TempC,
		time.Now().Format("15:04"),
		current.Condition.Text,
	)
	for _, hour := range hours {
		date := time.Unix(hour.TimeEpoch, 0)
		if date.Before(time.Now().Truncate(time.Hour)) {
			continue
		}
		message := fmt.Sprintf(
			"%s - %.0f°C, Sky: %s, Chances of Rain: %.0f%%\n",
			date.Format("15:04"),
			hour.TempC,
			hour.Condition.Text,
			hour.ChanceOfRain,
		)
		if hour.ChanceOfRain < 40 {
			color.Cyan(message)
		} else {
			color.Red(message)
		}
	}
}
