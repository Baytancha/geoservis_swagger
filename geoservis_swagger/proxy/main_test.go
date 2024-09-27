package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
)

func TestMainfunc(t *testing.T) {
	go func() {
		main()
	}()
	time.Sleep(2 * time.Second)
	t.Log("main finished")
}

func TestReverseProxy_proxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/api", nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	proxy := NewReverseProxy("hugo_task", "1313")
	r.Use(proxy.ReverseProxy)
	r.Get("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from API"))
	}))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if w.Body.String() != "Hello from API" {
		t.Errorf("Expected 'Hello from API', got %s", w.Body.String())
	}
}

func TestReverseProxy_target(t *testing.T) {
	mux := http.NewServeMux()
	server := &http.Server{
		Addr:    ":1313",
		Handler: mux,
	}
	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("TASK LIST"))
	})
	go func() {
		log.Fatal(server.ListenAndServe())
	}()
	req := httptest.NewRequest(http.MethodGet, "http://localhost:8080/tasks", nil)
	w := httptest.NewRecorder()
	r := chi.NewRouter()
	proxy := NewReverseProxy("localhost", "1313")
	r.Use(proxy.ReverseProxy)
	r.Get("/api", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello from API"))
	}))
	r.ServeHTTP(w, req)
	if w.Code != 200 {
		t.Errorf("Expected 200, got %d", w.Code)
	}

	if w.Body.String() != "TASK LIST" {
		t.Errorf("Expected 'Hello from API', got %s", w.Body.String())
	}
}

func TestAddressSearch(t *testing.T) {

	geo := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	addresses, err := geo.AddressSearch("Москва, ул Сухонская")
	if err != nil {
		t.Error(err)
	}
	if len(addresses) == 0 {
		t.Error("no addresses")
	}

	empty, err := geo.AddressSearch("Босква, ул Бухонская")
	if err != nil {
		t.Error(err)
	}
	if len(empty) != 0 {
		t.Error("should be empty addresses")
	}

}

func TestGeoCode(t *testing.T) {
	geo := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	geoCode, err := geo.GeoCode("55.878", "37.653")
	if err != nil {
		t.Error(err)
	}
	if len(geoCode) == 0 {
		t.Error("no addresses")
	}

	empty, err := geo.GeoCode("-7575", "-867868")
	if err != nil {
		t.Error(err)
	}
	if len(empty) != 0 {
		t.Error("should be empty")
	}

	empty2, err := geo.GeoCode("sdfsfsfsf", "fsfsf")
	if err != nil {
		t.Error(err)
	}
	if len(empty2) != 0 {
		t.Error("should be empty")
	}
}

func TestMarshalUnMarshalGeoCode(t *testing.T) {
	client := NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9")
	lat, lng := "55.878", "37.653"
	httpClient := &http.Client{}
	var data = strings.NewReader(fmt.Sprintf(`{"lat": %s, "lon": %s}`, lat, lng))
	req, err := http.NewRequest("POST", "https://suggestions.dadata.ru/suggestions/api/4_1/rs/geolocate/address", data)
	if err != nil {
		t.Error(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Token %s", client.apiKey))
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	buf, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
	}
	fmt.Println("buffer", string(buf))

	geocode, err := UnmarshalGeoCode(buf)
	if err != nil {
		t.Error(err)
	}

	if geocode.Suggestions[0].Data.City != "Москва" {
		t.Error("wrong city")
	}

	_, err = geocode.Marshal()
	if err != nil {
		t.Error(err)
	}

	fmt.Println("end")
}

func TestSearchHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"SearchHandler1_valid", "/api/address/search", http.StatusOK, []byte(`{"query":"Москва, ул Сухонская"}`)},
		{"SearchHandler2_invalid", "/api/address/search", http.StatusBadRequest, []byte(`"name":"John","age":30}`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("POST", tt.path, bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestGeoCodeHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"GeocodeHandler1_valid", "/api/address/geocode", http.StatusOK, []byte(`{"lat": "55.878", "lng": "37.653"}`)},
		{"GeocodeHandler2_invalid", "/api/address/geocode", http.StatusBadRequest, []byte(` "lat": "55.878", "lng": "37.653"`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("POST", tt.path, bytes.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}

func TestFileHandler(t *testing.T) {
	tests := []struct {
		name       string
		path       string
		statusCode int
		body       []byte
	}{
		{"Fileserver_valid", "/swagger/", http.StatusOK, nil},
		{"Fileserver_invalid", "/swagger/smth", http.StatusNotFound, nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
			app := &application{
				geo:    NewGeoService("fc47d9338dbcf9a2199f193ec2e5e57857e37378", "954baf5559aa44c49bde9a4dc572801bf48b69e9"),
				logger: logger,
			}
			r := app.setupRouter()

			r.ServeHTTP(w, req)
			if w.Code != tt.statusCode {
				t.Errorf("expected status code %d but got %d", tt.statusCode, w.Code)
			}
		})
	}
}
