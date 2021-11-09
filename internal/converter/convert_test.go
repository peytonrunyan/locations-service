package converter

import (
	"log"
	"testing"
)

const (
	STATES_GEOJSON   string = "/Users/peytonrunyan/projects/geoapp/data/us_states_500k.geojson"
	COUNTIES_GEOJSON string = "/Users/peytonrunyan/projects/geoapp/data/Merged_Counties_Subcounties_Communities.geojson"
)

var (
	StatesFC, _ = GetFeatureCollection(STATES_GEOJSON)
	sortedMap   = setupMap()
	Durham      = MakePoint(35.9, -78.9)
	CostaMesa   = MakePoint(33.64, -117.92)
)

// The map is expensive to create, so create it once
func setupMap() map[string]*[]Location {
	CountiesFC, err := GetFeatureCollection(COUNTIES_GEOJSON)
	if err != nil {
		log.Fatal(err)
	}
	locationsMap, err := MapLocations(CountiesFC)
	if err != nil {
		log.Fatal(err)
	}
	locationsMap, err = SortBySize(locationsMap)
	if err != nil {
		log.Fatal(err)
	}

	return locationsMap
}

func BenchmarkDurham(b *testing.B) {
	for i := 0; i < b.N; i++ {
		state, _ := FindState(StatesFC, Durham)
		countiesToSearch := sortedMap[state]
		FindCityCounty(countiesToSearch, Durham)
	}
}
func BenchmarkCostaMesa(b *testing.B) {
	for i := 0; i < b.N; i++ {
		state, _ := FindState(StatesFC, CostaMesa)
		countiesToSearch := sortedMap[state]
		FindCityCounty(countiesToSearch, CostaMesa)
	}
}
