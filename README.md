# GeoService

A service to quickly get the city, state, and community ID associated with a lat/lon pair. 

Usage note - it requires our TRP GeoJSON files to work.

Runs on port `8083`.

Expects a `GET` request with the following body `{"lat":float64, "lon":float64}`.
