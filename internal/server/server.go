package server

import (
	"encoding/json"
	"fmt"
	"geodb/internal/converter"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/paulmach/orb/geojson"
)

const (
	GEO_FILE1 = "/Users/peytonrunyan/projects/geoapp/data/us_states_500k.geojson"
	GEO_FILE2 = "/Users/peytonrunyan/projects/geoapp/data/Merged_Counties_Subcounties_Communities.geojson"
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
	CountiesFC, _ := converter.GetFeatureCollection(GEO_FILE2)
	StatesFC, _ := converter.GetFeatureCollection(GEO_FILE1)
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
