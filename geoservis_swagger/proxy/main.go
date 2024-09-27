//   Product Api:
//    version: 0.1
//    title: Product Api
//   Schemes: http, https
//   Host:
//   BasePath: /api/v1
//      Consumes:
//      - application/json
//   Produces:
//   - application/json
//   SecurityDefinitions:
//    Bearer:
//     type: apiKey
//     name: Authorization
//     in: header
//   swagger:meta

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"test/swagger"

	"github.com/go-chi/chi"
)

type application struct {
	geo    GeoProvider
	logger *slog.Logger
}

// swagger:parameters GetAddress
type SearchRequest struct {
	//A search request in JSON format
	//example: Москва Обуховская 11
	Query string `json:"query"`
}

//
//swagger:model
type SearchResponse struct {
	// An array of addresses
	Addresses []*Address `json:"addresses"`
}

// swagger:parameters GetAddressByGeocode
type GeocodeRequest struct {
	//latitude
	Lat string `json:"lat"`
	//longitude
	Lng string `json:"lng"`
}

//swagger:model
type GeocodeResponse struct {
	//An array of addresses
	Addresses []*Address `json:"addresses"`
}

// возвращать информацию о городе, в котором находится данный адрес.

func (app *application) SearchHandler(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/address/search GetAddress
	// swagger:operation POST /api/address/search GetAddress
	//
	// gets addresses either from URL query param or request body
	//
	//
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: addr_query
	//   in: query
	//   type: string
	// - name: addr_query
	//   in: body
	//   type: string
	// responses:
	//   '200':
	//     description: an array of addresses
	//     schema:
	//         items:
	//         "$ref": "#/definitions/SearchResponse"
	//   '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//   '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	var req SearchRequest
	req.Query = r.URL.Query().Get("query")
	if req.Query == "" {

		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	fmt.Println(req.Query)
	addresses, err := app.geo.AddressSearch(req.Query)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := SearchResponse{Addresses: addresses}
	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	w.Write(responseJSON)

}

func (app *application) GeocodeHandler(w http.ResponseWriter, r *http.Request) {
	//swagger:route POST /api/address/geocode GetAddressByGeocode
	// swagger:operation POST /api/address/geocode GetAddressByGeocode
	//
	// gets addresses based on geographic coordinates submitted in URL query param or request body
	//
	//
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: lat
	//   in: query
	//   type: string
	// - name: lng
	//   in: query
	//   type: string
	// - name: lat_lng
	//   in: body
	//   type: string
	// responses:
	//  '200':
	//     description: an array of addresses
	//     schema:
	//         items:
	//         "$ref": "#/definitions/GeocodeResponse"
	//  '400':
	//      description: invalid request body
	//      schema:
	//	        type: string
	//  '500':
	//        description: internal server error
	//        schema:
	//	        type: string

	//

	var req GeocodeRequest
	req.Lat = r.URL.Query().Get("lat")
	req.Lng = r.URL.Query().Get("lng")
	if req.Lat == "" || req.Lng == "" {
		err := json.NewDecoder(r.Body).Decode(&req)
		if err != nil {
			app.logger.Error(err.Error())
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}
	addresses, err := app.geo.GeoCode(req.Lat, req.Lng)
	if err != nil {
		app.logger.Error(err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	response := GeocodeResponse{Addresses: addresses}

	responseJSON, _ := json.Marshal(response)

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Accept", "application/json")
	w.Write(responseJSON)
}

func (app *application) setupRouter() *chi.Mux {
	r := chi.NewRouter()

	proxy := ReverseProxy{
		host: "hugo_task",
		port: "1313",
	}
	r.Use(proxy.ReverseProxy)
	r.Post("/api/address/search", app.SearchHandler)
	r.Post("/api/address/geocode", app.GeocodeHandler)

	// r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
	// 	http.ServeFile(w, r, "./swagger/swagger.yaml")
	// })

	fileServer := http.FileServerFS(swagger.Swaggerfile)
	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/swagger", fileServer)
		fs.ServeHTTP(w, r)
	})

	return r
}

// func (app *application) SwaggerHandler(w http.ResponseWriter, r *http.Request) {
// 	config, err := swagger.Swaggerfile.ReadFile("swagger.yaml")
// 	if err != nil {
// 		app.logger.Error(err.Error())
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 		return
// 	}
// 	w.Header().Set("Content-Type", "application/json")
// 	w.Header().Set("Accept", "application/json")
// 	w.Write(config)

// }

func newServer(r *chi.Mux) *http.Server {
	return &http.Server{
		Addr:    ":8080",
		Handler: r,
	}
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	app := &application{
		geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
		logger: logger,
	}

	httpServer := newServer(app.setupRouter())
	log.Fatal(httpServer.ListenAndServe())

}
