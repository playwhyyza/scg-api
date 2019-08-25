package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	cache "github.com/victorspringer/http-cache"
	"github.com/victorspringer/http-cache/adapter/memory"
)

type xyz struct {
	X int
	Y int
	Z int
}

func helloSCG(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome to SCG API")
}

func findXYZ(w http.ResponseWriter, r *http.Request) {
	// X, 5, 9, 15, 23, Y, Z
	// curl localhost:8080/scg/xyz -d "[5,9,15,23]"

	switch r.Method {
	case http.MethodPost:
		var sequence []int

		reqBody, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		json.Unmarshal([]byte(reqBody), &sequence)

		numberList := sequence
		term := (numberList[2] - numberList[1]) - (numberList[1] - numberList[0])
		x := numberList[0] - term
		y := (term + len(numberList)*term) + numberList[len(numberList)-1]
		z := y + (y - numberList[len(numberList)-1]) + term
		mySequence := xyz{
			X: x,
			Y: y,
			Z: z,
		}

		json.Marshal(mySequence)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(mySequence)
	}

}

func findRestaurants(w http.ResponseWriter, r *http.Request) {
	// curl localhost:8080/scg/restaurants

	switch r.Method {
	case http.MethodGet:
		apiKey := "Insert-API-Key-Here"
		apiURL := "https://maps.googleapis.com/maps/api/place/nearbysearch/json?"
		location := "13.8234866,100.5081204" // Bangsue
		radius := "1000"
		typePlace := "restaurant"
		language := "th"

		fullURL := fmt.Sprintf("%slocation=%s&radius=%s&type=%s&lang=%s&key=%s", apiURL, location, radius, typePlace, language, apiKey)

		resp, err := http.Get(fullURL)

		if err != nil {
			panic(err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		var mapData map[string]interface{}

		json.Unmarshal(body, &mapData)

		json.Marshal(mapData)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(mapData)

	default:
		fmt.Println("allow only POST Method")
	}
}

func main() {
	memcached, err := memory.NewAdapter(
		memory.AdapterWithAlgorithm(memory.LRU),
		memory.AdapterWithCapacity(10000000),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	cacheClient, err := cache.NewClient(
		cache.ClientWithAdapter(memcached),
		cache.ClientWithTTL(5*time.Minute),
		cache.ClientWithRefreshKey("opn"),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	http.Handle("/scg", cacheClient.Middleware(http.HandlerFunc(helloSCG)))
	http.Handle("/scg/xyz", cacheClient.Middleware(http.HandlerFunc(findXYZ)))
	http.Handle("/scg/restaurants", cacheClient.Middleware(http.HandlerFunc(findRestaurants)))
	http.ListenAndServe(":8080", nil)
}
