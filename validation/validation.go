package validation

import (
	"context"
	"fmt"
	"strings"

	"googlemaps.github.io/maps"
)

type AddressValidator struct {
	client *maps.Client
}

func NewAddressValidator(apiKey string) (*AddressValidator, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %v", err)
	}
	return &AddressValidator{client: client}, nil
}

func (av *AddressValidator) ValidateSeattleAddress(address string) (string, error) {
	ctx := context.Background()

	req := &maps.GeocodingRequest{
		Address: address,
	}

	resp, err := av.client.Geocode(ctx, req)
	if err != nil {
		return "", fmt.Errorf("failed to validate address: %v", err)
	}

	if len(resp) == 0 {
		return "", fmt.Errorf("address not found. Try being more specific (e.g., '123 Main St, Seattle, WA')")
	}

	result := resp[0]

	isInSeattleArea := false
	var city, state, formattedAddress string

	for _, component := range result.AddressComponents {
		for _, typ := range component.Types {
			switch typ {
			case "locality":
				city = component.LongName
			case "administrative_area_level_1":
				state = component.ShortName
			}
		}
	}

	formattedAddress = result.FormattedAddress

	seattleAreaCities := []string{
		"Seattle", "Bellevue", "Redmond", "Kirkland", "Bothell", "Lynnwood",
		"Everett", "Renton", "Kent", "Federal Way", "Tacoma", "Burien",
		"Shoreline", "Edmonds", "Mukilteo", "Mill Creek", "Woodinville",
	}

	for _, seattleCity := range seattleAreaCities {
		if strings.Contains(strings.ToLower(city), strings.ToLower(seattleCity)) ||
		   strings.Contains(strings.ToLower(formattedAddress), strings.ToLower(seattleCity)) {
			isInSeattleArea = true
			break
		}
	}

	if state == "WA" && strings.Contains(strings.ToLower(formattedAddress), "king county") {
		isInSeattleArea = true
	}

	if !isInSeattleArea {
		return "", fmt.Errorf("⚠️  '%s' appears to be outside the Seattle area (%s, %s). This tool works best with Seattle-area addresses", address, city, state)
	}

	return formattedAddress, nil
}