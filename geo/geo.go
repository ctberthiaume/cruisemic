// Package geo provides functions to convert GGA coordinates to decimal degree.
package geo

import (
	"fmt"
	"strconv"
	"strings"
)

// GGALat2DD converts a GGA latitude to decimal degrees.
// Latitude and north/south designation should be provided separately, with
// north/south presented as one of "NnSs". Decimal degree coordinates are
// returned with precision to 4 decimal places (11.132 m).
// e.g. "2116.6922" -> "21.2782"
func GGALat2DD(lat string, ns string) (string, error) {
	if len(lat) > 0 && (string(lat[0]) == "-" || string(lat[0]) == "+") {
		return "", fmt.Errorf("+/- should be passed as N/S")
	}
	if len(lat) < 4 {
		return "", fmt.Errorf("bad GGA latitude, len < 4: %v", lat)
	}
	deg, err := strconv.ParseFloat(lat[:2], 64)
	if err != nil {
		return "", fmt.Errorf("bad GGA latitude, deg not numeric: %v", lat)
	}
	min, err := strconv.ParseFloat(lat[2:], 64)
	if err != nil {
		return "", fmt.Errorf("bad GGA latitude, min not numeric: %v", lat)
	}
	if deg > 90 {
		return "", fmt.Errorf("bad GGA latitude, gga=%v,%v deg=%v", lat, ns, deg)
	}
	if min > 60 {
		return "", fmt.Errorf("bad GGA latitude, gga=%v,%v min=%v", lat, ns, min)
	}

	switch strings.ToUpper(ns) {
	case "N":
		return fmt.Sprintf("%.4f", deg+(min/60.0)), nil
	case "S":
		return fmt.Sprintf("%.4f", -1*(deg+(min/60.0))), nil
	default:
		return "", fmt.Errorf("bad GGA latitude, bad north/south char: %v", lat)
	}
}

// GGALon2DD converts a GGA longitude to decimal degrees
// Longitude and east/west designation should be provided separately, with
// east/west presented as one of "EeWw". Decimal degree coordinates are
// returned with precision to 4 decimal places (11.132 m).
// e.g. "15752.6526" -> "157.8775"
func GGALon2DD(lon string, ew string) (string, error) {
	if len(lon) > 0 && (string(lon[0]) == "-" || string(lon[0]) == "+") {
		return "", fmt.Errorf("+/- should be passed as E/W")
	}
	if len(lon) < 5 {
		return "", fmt.Errorf("bad GGA longitude, len < 5: %v", lon)
	}
	deg, err := strconv.ParseFloat(lon[:3], 64)
	if err != nil {
		return "", fmt.Errorf("bad GGA longitude, deg not numeric: %v", lon)
	}
	min, err := strconv.ParseFloat(lon[3:], 64)
	if err != nil {
		return "", fmt.Errorf("bad GGA longitude, min not numeric: %v", lon)
	}
	if deg > 180 {
		return "", fmt.Errorf("bad GGA longitude, deg > 180: gga=%v,%v deg=%v", lon, ew, deg)
	}
	if min > 60 {
		return "", fmt.Errorf("bad GGA longitude, min > 60: gga=%v,%v min=%v", lon, ew, min)
	}

	switch strings.ToUpper(ew) {
	case "E":
		return fmt.Sprintf("%.4f", deg+(min/60.0)), nil
	case "W":
		return fmt.Sprintf("%.4f", -1*(deg+(min/60.0))), nil
	default:
		return "", fmt.Errorf("bad GGA longitude, bad east/west char: %v", lon)
	}
}

// CheckLat checks if the decimal degree longitude string is valid.
func CheckLat(lat string) error {
	val, err := strconv.ParseFloat(lat, 64)
	if err != nil {
		return fmt.Errorf("bad latitude: %v", lat)
	}
	if val < -90 || val > 90 {
		return fmt.Errorf("latitude out of range: %v", lat)
	}
	return nil
}

// CheckLon checks if the decimal degree latitude string is valid.
func CheckLon(lon string) error {
	val, err := strconv.ParseFloat(lon, 64)
	if err != nil {
		return fmt.Errorf("bad longitude: %v", lon)
	}
	if val < -180 || val > 180 {
		return fmt.Errorf("longitude out of range: %v", lon)
	}
	return nil

}
