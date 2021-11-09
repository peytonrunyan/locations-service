package server

import (
	"encoding/json"
	"fmt"
	"geodb/internal/converter"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/paulmach/orb/geojson"
)

var (
	STATES, CITIES_COUNTIES = getGeoJSONLocations()
)

func NewHTTPServer(addr string) *http.Server {
	httpServer := newHTTPServer()
	r := mux.NewRouter()
	r.HandleFunc("/", httpServer.handleLocationRequest).Methods("Get")

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}
}

// Struct to hold our features to search against
type httpServer struct {
	statesFC     *geojson.FeatureCollection
	locationsMap map[string]*[]converter.Location
}

// Creates a new httpServer with initialized an initialized state Feature Collection and
// and initialized sorted map of states to city Features.
func newHTTPServer() *httpServer {
	CountiesFC, _ := converter.GetFeatureCollection(CITIES_COUNTIES)
	StatesFC, _ := converter.GetFeatureCollection(STATES)
	locationsMap, _ := converter.MapLocations(CountiesFC)
	locationsMap, _ = converter.SortBySize(locationsMap)

	return &httpServer{StatesFC, locationsMap}
}

type LocationRequest struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type LocationReponse struct {
	ErrorMsg    string `json:"errorMessage"`
	City        string `json:"city"`
	State       string `json:"state"`
	CommunityID string `json:"communityID"`
}

func (s *httpServer) handleLocationRequest(w http.ResponseWriter, r *http.Request) {
	var req LocationRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Printf("req: %v\n", req)
	loc := converter.MakePoint(req.Lat, req.Lon)
	state, err := converter.FindState(s.statesFC, loc)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	// state will be "" if the point isn't in the USA
	if state == "" {
		http.Error(w, noPointFound(&req), http.StatusNotFound)
		return
	}
	res := converter.FindCityCounty(s.locationsMap[state], loc)
	err = json.NewEncoder(w).Encode(LocationReponse{"", res.City, res.State, res.Community_ID})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func noPointFound(r *LocationRequest) string {
	return fmt.Sprintf(
		"Could not find a location within the United States at Lat: %v, Long: %v",
		r.Lat, r.Lon,
	)
}

// Loads the absolute file path for the GeoJSON files. It expects these to be
// located in a `.env` file in the root directory with the variable names `STATES`
// and `CITIES_COUNTIES` respectively.
func getGeoJSONLocations() (states string, citiesCounties string) {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Could not load GeoJSON files\n%s", err.Error())
	}
	states = os.Getenv("STATES")
	citiesCounties = os.Getenv("CITIES_COUNTIES")
	return states, citiesCounties

}
