package main

import (
	"fmt"
	conv "geodb/converter"
)

const (
	GEO_FILE1 = "/Users/peytonrunyan/projects/geoapp/data/us_states_500k.geojson"
	GEO_FILE2 = "/Users/peytonrunyan/projects/geoapp/data/Merged_Counties_Subcounties_Communities.geojson"
)

func main() {
	Durham := conv.MakePoint(35.78, -78.6)
	CostaMesa := conv.MakePoint(33.64, -117.92)

	CountiesFC, _ := conv.GetFeatureCollection(GEO_FILE2)
	StatesFC, _ := conv.GetFeatureCollection(GEO_FILE1)

	locationsMap, _ := conv.MapLocations(CountiesFC)
	locationsMap, _ = conv.SortBySize(locationsMap)

	durhamState, _ := conv.FindState(StatesFC, Durham)
	cmState, _ := conv.FindState(StatesFC, CostaMesa)

	res1 := conv.FindCityCounty(locationsMap[durhamState], Durham)
	res2 := conv.FindCityCounty(locationsMap[cmState], CostaMesa)
	fmt.Println(res1)
	fmt.Println(res2)
}
