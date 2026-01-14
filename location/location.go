package location

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type IPLocation struct {
	Query       string  `json:"query"`
	Country     string  `json:"country"`
	RegionName  string  `json:"regionName"`
	City        string  `json:"city"`
	Zip         string  `json:"zip"`
	Lat         float64 `json:"lat"`
	Lon         float64 `json:"lon"`
	ISP         string  `json:"isp"`
	Org         string  `json:"org"`
	As          string  `json:"as"`
	Status      string  `json:"status"`
	Message     string  `json:"message"`
}

func GetCurrentLocation() (string, error) {
	location, err := tryLocationServices()
	if err != nil {
		return "", fmt.Errorf("automatic location detection failed: %v\n\n💡 Try: 'commute --from \"your current address\"' or set a default location with 'commute config set-current \"address\"'", err)
	}
	return location, nil
}

func tryLocationServices() (string, error) {
	services := []func() (string, error){
		tryPreciseLocation,
		tryIPAPI,
		tryIPInfo,
		tryDefault,
	}

	for _, service := range services {
		if location, err := service(); err == nil {
			return location, nil
		}
	}

	return "", fmt.Errorf("all location services failed")
}

func tryPreciseLocation() (string, error) {
	return GetPreciseLocation()
}

func tryIPAPI() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://ip-api.com/json/")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status: %d", resp.StatusCode)
	}

	var loc IPLocation
	if err := json.NewDecoder(resp.Body).Decode(&loc); err != nil {
		return "", err
	}

	if loc.Status != "success" {
		return "", fmt.Errorf(loc.Message)
	}

	return fmt.Sprintf("%f,%f", loc.Lat, loc.Lon), nil
}

func tryIPInfo() (string, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://ipinfo.io/json")
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("status: %d", resp.StatusCode)
	}

	var result struct {
		Loc string `json:"loc"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if result.Loc == "" {
		return "", fmt.Errorf("no location data")
	}

	return result.Loc, nil
}

func tryDefault() (string, error) {
	return "47.6062,-122.3321", nil
}