// +build darwin

package location

/*
#cgo CFLAGS: -x objective-c -mmacosx-version-min=10.14
#cgo LDFLAGS: -framework CoreLocation -framework Foundation
#include "corelocation.h"
*/
import "C"
import (
	"fmt"
	"runtime"
)

func init() {
	if runtime.GOOS == "darwin" {
		C.initLocationManager()
	}
}

func GetPreciseLocation() (string, error) {
	if runtime.GOOS != "darwin" {
		return "", fmt.Errorf("Core Location only available on macOS")
	}

	result := C.requestLocation()

	switch result {
	case 0:
		// Success
		lat := float64(C.getLatitude())
		lon := float64(C.getLongitude())
		return fmt.Sprintf("%f,%f", lat, lon), nil
	case -1:
		return "", fmt.Errorf("location manager not initialized")
	case -2:
		return "", fmt.Errorf("location permission denied. Please allow location access for this app in System Preferences > Security & Privacy > Privacy > Location Services")
	case -3:
		return "", fmt.Errorf("location services disabled. Please enable Location Services in System Preferences > Security & Privacy > Privacy > Location Services")
	case -4:
		return "", fmt.Errorf("failed to get location (GPS/WiFi issue)")
	case -5:
		return "", fmt.Errorf("location request timed out")
	default:
		return "", fmt.Errorf("unknown location error (%d)", result)
	}
}