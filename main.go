package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
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
	var wg sync.WaitGroup

	loc := "mumbai"
	if len(os.Args) >= 2 {
		loc = os.Args[1]
	}
	url := "http://api.weatherapi.com/v1/forecast.json?key=588129751b7940ca824160458242001&q=" + loc + "&days=1&aqi=no&alerts=no"

	ch := make(chan Weather, 1)

	wg.Add(1)
	go func() { // a goroutine for the HTTP request
		defer wg.Done()
		weather, err := fetchWeather(url)
		if err != nil {
			panic(err)
		}
		ch <- weather
	}()

	wg.Add(1)
	go func() { // a goroutine for unmarshalling JSON
		defer wg.Done()
		weather := <-ch
		printweather(weather)
	}()

	wg.Wait()
}

func fetchWeather(url string) (Weather, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	res, err := client.Get(url)
	if err != nil {
		return Weather{}, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return Weather{}, fmt.Errorf("API error: %d", res.StatusCode)
	}
	var weather Weather
	decoder := json.NewDecoder(res.Body)
	if err := decoder.Decode(&weather); err != nil {
		return Weather{}, err
	}
	return weather, nil
}

func printweather(weather Weather) {
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
