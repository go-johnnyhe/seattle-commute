package transit

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"googlemaps.github.io/maps"
)

type Route struct {
	Summary      string
	Duration     time.Duration
	DepartureTime time.Time
	ArrivalTime   time.Time
	Steps        []Step
	Distance     string
}

type Step struct {
	Instructions string
	Duration     time.Duration
	Mode         string
	LineInfo     string
	DepartTime   time.Time
	ArrivalTime  time.Time
}

type TransitService struct {
	client *maps.Client
}

func NewTransitService(apiKey string) (*TransitService, error) {
	client, err := maps.NewClient(maps.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Google Maps client: %v", err)
	}

	return &TransitService{client: client}, nil
}

func (ts *TransitService) GetRoutes(origin, destination string) ([]Route, error) {
	ctx := context.Background()
	now := time.Now()

	req := &maps.DirectionsRequest{
		Origin:        origin,
		Destination:   destination,
		Mode:          maps.TravelModeTransit,
		DepartureTime: fmt.Sprintf("%d", now.Unix()),
		Alternatives:  true,
		Units:         maps.UnitsImperial,
	}

	resp, _, err := ts.client.Directions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get directions: %v", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	var routes []Route
	for _, route := range resp {
		if len(route.Legs) == 0 {
			continue
		}

		leg := route.Legs[0]
		r := Route{
			Summary:      route.Summary,
			Duration:     leg.Duration,
			DepartureTime: leg.DepartureTime,
			ArrivalTime:   leg.ArrivalTime,
			Distance:     leg.Distance.HumanReadable,
		}

		for _, step := range leg.Steps {
			s := Step{
				Instructions: cleanHTML(step.HTMLInstructions),
				Duration:     step.Duration,
				Mode:         string(step.TravelMode),
			}

			if step.TransitDetails != nil {
				s.DepartTime = step.TransitDetails.DepartureTime
				s.ArrivalTime = step.TransitDetails.ArrivalTime

				if step.TransitDetails.Line.ShortName != "" {
					s.LineInfo = fmt.Sprintf("%s %s",
						step.TransitDetails.Line.Vehicle.Name,
						step.TransitDetails.Line.ShortName)
				} else {
					s.LineInfo = step.TransitDetails.Line.Name
				}
			}

			r.Steps = append(r.Steps, s)
		}

		routes = append(routes, r)
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].DepartureTime.Before(routes[j].DepartureTime)
	})

	return routes, nil
}

func (ts *TransitService) GetNextRoutes(origin, destination string, hours int) ([]Route, error) {
	ctx := context.Background()

	req := &maps.DirectionsRequest{
		Origin:        origin,
		Destination:   destination,
		Mode:          maps.TravelModeTransit,
		DepartureTime: "now",
		Alternatives:  true,
		Units:         maps.UnitsImperial,
	}

	resp, _, err := ts.client.Directions(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get directions: %v", err)
	}

	if len(resp) == 0 {
		return nil, fmt.Errorf("no routes found")
	}

	var routes []Route
	for _, route := range resp {
		if len(route.Legs) == 0 {
			continue
		}

		leg := route.Legs[0]

		if leg.DepartureTime.Before(time.Now().Add(-5 * time.Minute)) {
			continue
		}

		r := Route{
			Summary:      route.Summary,
			Duration:     leg.Duration,
			DepartureTime: leg.DepartureTime,
			ArrivalTime:   leg.ArrivalTime,
			Distance:     leg.Distance.HumanReadable,
		}

		for _, step := range leg.Steps {
			s := Step{
				Instructions: cleanHTML(step.HTMLInstructions),
				Duration:     step.Duration,
				Mode:         string(step.TravelMode),
			}

			if step.TransitDetails != nil {
				s.DepartTime = step.TransitDetails.DepartureTime
				s.ArrivalTime = step.TransitDetails.ArrivalTime

				if step.TransitDetails.Line.ShortName != "" {
					s.LineInfo = fmt.Sprintf("%s %s",
						step.TransitDetails.Line.Vehicle.Name,
						step.TransitDetails.Line.ShortName)
				} else {
					s.LineInfo = step.TransitDetails.Line.Name
				}
			}

			r.Steps = append(r.Steps, s)
		}

		routes = append(routes, r)
	}

	if len(routes) == 0 {
		return ts.getFallbackRoutes(origin, destination, hours)
	}

	sort.Slice(routes, func(i, j int) bool {
		return routes[i].DepartureTime.Before(routes[j].DepartureTime)
	})

	return routes, nil
}

func (ts *TransitService) getFallbackRoutes(origin, destination string, hours int) ([]Route, error) {
	var allRoutes []Route

	for i := 0; i < 3; i++ {
		departTime := time.Now().Add(time.Duration(i*20) * time.Minute)

		ctx := context.Background()
		req := &maps.DirectionsRequest{
			Origin:        origin,
			Destination:   destination,
			Mode:          maps.TravelModeTransit,
			DepartureTime: fmt.Sprintf("%d", departTime.Unix()),
			Units:         maps.UnitsImperial,
		}

		resp, _, err := ts.client.Directions(ctx, req)
		if err != nil {
			continue
		}

		if len(resp) == 0 || len(resp[0].Legs) == 0 {
			continue
		}

		route := resp[0]
		leg := route.Legs[0]

		r := Route{
			Summary:      route.Summary,
			Duration:     leg.Duration,
			DepartureTime: leg.DepartureTime,
			ArrivalTime:   leg.ArrivalTime,
			Distance:     leg.Distance.HumanReadable,
		}

		for _, step := range leg.Steps {
			s := Step{
				Instructions: cleanHTML(step.HTMLInstructions),
				Duration:     step.Duration,
				Mode:         string(step.TravelMode),
			}

			if step.TransitDetails != nil {
				s.DepartTime = step.TransitDetails.DepartureTime
				s.ArrivalTime = step.TransitDetails.ArrivalTime

				if step.TransitDetails.Line.ShortName != "" {
					s.LineInfo = fmt.Sprintf("%s %s",
						step.TransitDetails.Line.Vehicle.Name,
						step.TransitDetails.Line.ShortName)
				} else {
					s.LineInfo = step.TransitDetails.Line.Name
				}
			}

			r.Steps = append(r.Steps, s)
		}

		allRoutes = append(allRoutes, r)
	}

	uniqueRoutes := removeDuplicateRoutes(allRoutes)

	sort.Slice(uniqueRoutes, func(i, j int) bool {
		return uniqueRoutes[i].DepartureTime.Before(uniqueRoutes[j].DepartureTime)
	})

	return uniqueRoutes, nil
}

func cleanHTML(html string) string {
	html = strings.ReplaceAll(html, "<b>", "")
	html = strings.ReplaceAll(html, "</b>", "")
	html = strings.ReplaceAll(html, "<div>", "")
	html = strings.ReplaceAll(html, "</div>", "")
	html = strings.ReplaceAll(html, "<div style=\"font-size:0.9em\">", " - ")
	return html
}

func removeDuplicateRoutes(routes []Route) []Route {
	seen := make(map[string]bool)
	var unique []Route

	for _, route := range routes {
		key := fmt.Sprintf("%s_%s_%s",
			route.Summary,
			route.DepartureTime.Format("15:04"),
			route.ArrivalTime.Format("15:04"))

		if !seen[key] {
			seen[key] = true
			unique = append(unique, route)
		}
	}

	return unique
}