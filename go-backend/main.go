package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/rs/cors"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)
// Define structs to represent weather and city data in the application

type Weather struct {
	Temperature string `json:"temperature"`
	Humidity    string `json:"humidity"`
	WindSpeed   string `json:"windSpeed"`
	Conditions  string `json:"conditions"`
}
// City struct matches the structure of each json object from cities.json
type City struct {
	Name      string  `json:"name"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

var rdb *redis.Client
var ctx = context.Background()
var cities []City

// Load city data from cities.json file into the cities slice.
func loadCities() {
	file, err := os.Open("cities.json")
	//if err value is not null, then there must be an error, so return a log with the error
	if err != nil {
		log.Fatalf("Failed to open cities.json: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&cities)
	//if err value is not null, then there must be an error, so return a log with the error
	if err != nil {
		log.Fatalf("Failed to decode cities.json: %v", err)
	}
}


// getWeather handles the HTTP request, retrieves weather data from cache or API, and returns it to the webapp.
// http Response writer is named w, and http request value is named r
func getWeather(w http.ResponseWriter, r *http.Request) {
	//start a timer when process starts
	start := time.Now()

	// Retrieve location and coordinates from query parameters.
	location := r.URL.Query().Get("location")
	latStr := r.URL.Query().Get("lat")
	lonStr := r.URL.Query().Get("lon")

	var lat, lon float64
	var err error

	// Determine coordinates based on input type (location or lat/lon).
	if location != "" {
		//If location is provided, then it uses that to get the coordinates
		lat, lon, err = getCoordinates(location)
		if err != nil {
			// if there is an error retrieving the coordinates, return internal server error
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else if latStr != "" && lonStr != "" {
		// if latitude and longitude are provided, parse them and use them to get coordinates
		lat, err = strconv.ParseFloat(latStr, 64)
		if err != nil {
			// if there is an error parsing latitude, return a bad request
			http.Error(w, "Invalid latitude", http.StatusBadRequest)
			return
		}
		lon, err = strconv.ParseFloat(lonStr, 64)
		if err != nil {
			// if there is an error parsing longitude, return a bad request
			http.Error(w, "Invalid longitude", http.StatusBadRequest)
			return
		}
	} else {
		// Neither location nor Lat/Lon was provided, return bad request
		http.Error(w, "Location or coordinates required", http.StatusBadRequest)
		return
	}

	// Attempt to get weather data from the cache using a key composed of latitude and longitude, it converts the city name(if used) to a corresponding
	//lat and lon
	weatherData, err := rdb.Get(ctx, fmt.Sprintf("%f,%f", lat, lon)).Result()
	if err == redis.Nil {
		// If cache miss, fetch weather data from the API.
		weather, err := fetchWeatherFromAPI(lat, lon)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		weatherJSON, _ := json.Marshal(weather)
		rdb.Set(ctx, fmt.Sprintf("%f,%f", lat, lon), weatherJSON, 0)
		w.Write(weatherJSON)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	} else {
		// If cache hit, return the cached data.
		w.Write([]byte(weatherData))
	}
	//log time process ends
	totalDuration := time.Since(start)
	log.Printf("Total request processing time: %s", totalDuration)
}

// getCoordinates returns the latitude and longitude for a given city name by looking it up in the cities slice.
func getCoordinates(cityName string) (float64, float64, error) {
	for _, city := range cities {
		if city.Name == cityName {
			return city.Latitude, city.Longitude, nil
		}
	}
	log.Printf("City not found %s", cityName)
	return 0, 0, fmt.Errorf("failed to fetch weather data")
}


// fetchWeatherFromAPI fetches weather data from the weather.gov API based on latitude and longitude.
// It uses the fetchForecast function to get detailed forecast data.
func fetchWeatherFromAPI(lat, lon float64) (*Weather, error) {
	url := fmt.Sprintf("https://api.weather.gov/points/%f,%f", lat, lon)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API request failed with status %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch weather data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//Struct to parse JSON response from API
	var pointData struct {
		Properties struct {
			Forecast string `json:"forecast"`
		} `json:"properties"`
	}

	err = json.Unmarshal(body, &pointData)
	if err != nil {
		return nil, err
	}

	if pointData.Properties.Forecast == "" {
		log.Printf("No forecast URL available")
		return nil, fmt.Errorf("Failed to fetch city data")
	}
    //Extract the forecast URL from the response and call fetchForecast for a detailed weather forecast
	forecastURL := pointData.Properties.Forecast
	return fetchForecast(forecastURL)
}

// fetchForecast fetches detailed weather forecast data from the given URL.
func fetchForecast(url string) (*Weather, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("API request failed with status %d", resp.StatusCode)
		return nil, fmt.Errorf("failed to fetch weather data")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	//struct to parse the response from the API
	var forecastData struct {
		Properties struct {
			Periods []struct {
				Temperature      int    `json:"temperature"`
				WindSpeed        string `json:"windSpeed"`
				ShortForecast    string `json:"shortForecast"`
				RelativeHumidity struct {
					Value int `json:"value"`
				} `json:"relativeHumidity"`
			} `json:"periods"`
		} `json:"properties"`
	}

	err = json.Unmarshal(body, &forecastData)
	if err != nil {
		return nil, err
	}
	//Periods array is part of the JSON response from weather.gov
	//if the parsed response is empty, then no weather data available
	if len(forecastData.Properties.Periods) == 0 {
		return nil, fmt.Errorf("No weather data available")
	}
	//Extract first periods weather data and then return it as a weather struct
	period := forecastData.Properties.Periods[0]
	return &Weather{
		Temperature: fmt.Sprintf("%d°C", period.Temperature),
		Humidity:    fmt.Sprintf("%d%%", period.RelativeHumidity.Value),
		WindSpeed:   period.WindSpeed,
		Conditions:  period.ShortForecast,
	}, nil
}

// main initializes the application, sets up the Redis client, and starts the HTTP server.
func main() {
	loadCities()

	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	mux := http.NewServeMux()
	mux.HandleFunc("/weather", getWeather)

	handler := cors.Default().Handler(mux)

	fmt.Println("Server is running on port 8080")
	log.Fatal(http.ListenAndServe(":8080", handler))
}
