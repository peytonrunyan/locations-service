// Functions to convert a long/lat pair into a city/state result
package converter

import (
	"errors"
	"fmt"
	"io/ioutil"
	"sort"

	"github.com/paulmach/orb"
	"github.com/paulmach/orb/geojson"
	"github.com/paulmach/orb/planar"
)

// create a struct that contains the geometry type, Features, population
// and State and city name for a given location (note - could be a County)
type Location struct {
	IsMultiPolygon bool
	Features       *geojson.Feature
	HouseUnits     int32
	State          string
	City           string
}

// used to return the results of `FindCounty`
type SearchResult struct {
	City         string
	State        string
	Community_ID string
}

// Returns a feature collection of states for high-granularity search before moving to search
// for the individual county that a coordinate pair belongs to
func GetFeatureCollection(filename string) (*geojson.FeatureCollection, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, errors.New("Problem reading geojson file: " + err.Error())
	}
	FeatureCollection, err := geojson.UnmarshalFeatureCollection(b)
	if err != nil {
		return nil, errors.New("Could not unmarshall geojson into feature collection: " + err.Error())
	}
	return FeatureCollection, nil
}

// Iterates through state geometries to see if a coordinate is inside of a given state
func FindState(fc *geojson.FeatureCollection, p orb.Point) (string, error) {
	for _, feature := range fc.Features {
		multiPoly, isMulti := feature.Geometry.(orb.MultiPolygon)
		if isMulti {
			if planar.MultiPolygonContains(multiPoly, p) {
				state := feature.Properties["NAME"]
				return state.(string), nil
			}
		} else {
			polygon, isPoly := feature.Geometry.(orb.Polygon)
			if isPoly {
				if planar.PolygonContains(polygon, p) {
					state := feature.Properties["NAME"]
					return state.(string), nil
				}
			}
		}
	}
	fmt.Println("Point not found in any states")
	return "", nil // uncertain whether or not I should return an error here
}

// Create a map that has an array of locations for each state subunit (e.g. city, county, etc).
// This is used to cut down the search space to a single state, since each individual check is
// expensive.
func MapLocations(fc *geojson.FeatureCollection) (map[string]*[]Location, error) {
	stateLocations := make(map[string]*[]Location)

	for _, feature := range fc.Features {
		var isMultiPolygon bool
		if _, isMulti := feature.Geometry.(orb.MultiPolygon); isMulti {
			isMultiPolygon = true
		} else if _, isPoly := feature.Geometry.(orb.Polygon); isPoly {
			isMultiPolygon = false
		} else {
			return nil, fmt.Errorf("Featured was passed that is neither a polygon nor multi-polygon")
		}
		loc := Location{
			IsMultiPolygon: isMultiPolygon,
			Features:       feature,
			City:           feature.Properties["NAME"].(string),
		}

		// Cities and counties has different datatypes and keys, so we need separate
		// handling. Cities use "ST" for state and Counties use either "STATE" or "STATE_NAME".
		// "HOUSEUNITS" is a float64 for cities and an int32 for counties. We use "ST" to
		// distinguish between cities and counties.
		if feature.Properties["ST"] != nil { // Cities
			loc.State = StateFromAbbrev[feature.Properties["ST"].(string)]
			if feature.Properties["HOUSEUNITS"] != nil {
				loc.HouseUnits = int32(feature.Properties["HOUSEUNITS"].(float64))
			} else {
				loc.HouseUnits = 10
			}
			// Counties logic
		} else {
			if feature.Properties["STATE"] != nil {
				loc.State = StateFromNum[feature.Properties["STATE"].(string)]
			} else if feature.Properties["STATE_NAME"] != nil {
				loc.State = feature.Properties["STATE_NAME"].(string)
			} else {
				fmt.Printf("CityName: %v", feature.Properties["NAME"])
				continue
			}
		}
		if feature.Properties["HOUSEUNITS"] != nil {
			loc.HouseUnits = feature.Properties["HOUSEUNITS"].(int32)
		} else {
			loc.HouseUnits = 10
		}
		stateSlice, ok := stateLocations[loc.State]
		if ok {
			*stateSlice = append(*stateSlice, loc)
		} else {
			stateLocations[loc.State] = &[]Location{loc}
		}
	}
	return stateLocations, nil
}

// Checks to see if a point is within the boundaries of a specific location within a
// state's slice of locations
func FindCityCounty(locs *[]Location, p orb.Point) *SearchResult {

	makeRes := func(l Location) *SearchResult {
		res := SearchResult{
			City:         l.City,
			State:        l.State,
			Community_ID: l.Features.Properties["Community_ID"].(string),
		}
		return &res
	}

	for _, loc := range *locs {
		fmt.Printf("City: %v, Units: %v\n", loc.City, loc.HouseUnits)
		if loc.IsMultiPolygon {
			multiPoly, _ := loc.Features.Geometry.(orb.MultiPolygon)
			if planar.MultiPolygonContains(multiPoly, p) {
				fmt.Println("Hit A")
				return makeRes(loc)
			}
		} else {
			polygon, _ := loc.Features.Geometry.(orb.Polygon)
			if planar.PolygonContains(polygon, p) {
				fmt.Println("Hit B")
				return makeRes(loc)
			}
		}
	}
	return nil
}

// Order our arrays of locations by population to make lookup faster - we'll have more
// requests from more populous locations, so they should be the first thing searched.
func SortBySize(LocMap map[string]*[]Location) (map[string]*[]Location, error) {
	for k := range LocMap {
		stateArray := *LocMap[k]
		sort.Slice(stateArray, func(i, j int) bool {
			return stateArray[i].HouseUnits > stateArray[j].HouseUnits
		})
		LocMap[k] = &stateArray
	}
	return LocMap, nil
}

// Makes a point from a lat/lon pair. Used to spare the `orb` import for anything
// using this package
func MakePoint(lat, lon float64) orb.Point {
	return orb.Point{lon, lat}
}

//TODO split states into group of N on startup for goroutines
// Take ordered items and pass them into arrays, round robin style

// Converts to state strings from state numbers
// Source: https://www.census.gov/geographies/reference-files/time-series/geo/tallies.html
