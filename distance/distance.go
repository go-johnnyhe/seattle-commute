package distance

import (
	"context"
	"fmt"
	"time"

	"googlemaps.github.io/maps"
)

type DistanceChecker struct {
	client *maps.Client
}

func NewDistanceChecker(apiKey string) (*DistanceChecker, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %v", err)
	}
	return &DistanceChecker{client: client}, nil
}

func (dc *DistanceChecker) GetWalkingDistance(origin, destination string) (time.Duration, string, error) {
	ctx := context.Background()

	req := &maps.DirectionsRequest{
		Origin:      origin,
		Destination: destination,
		Mode:        maps.TravelModeWalking,
		Units:       maps.UnitsImperial,
	}

	resp, _, err := dc.client.Directions(ctx, req)
	if err != nil {
		return 0, "", fmt.Errorf("failed to get walking directions: %v", err)
	}

	if len(resp) == 0 || len(resp[0].Legs) == 0 {
		return 0, "", fmt.Errorf("no walking route found")
	}

	leg := resp[0].Legs[0]
	return leg.Duration, leg.Distance.HumanReadable, nil
}

func (dc *DistanceChecker) IsWithinWalkingDistance(origin, destination string) (bool, time.Duration, string, error) {
	walkTime, walkDistance, err := dc.GetWalkingDistance(origin, destination)
	if err != nil {
		return false, 0, "", err
	}

	// Consider "walking distance" as 15 minutes or less
	isWalkable := walkTime <= 15*time.Minute

	return isWalkable, walkTime, walkDistance, nil
}