package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/reynld/carbtographer/pkg/database"

	"github.com/gorilla/mux"
	"github.com/machinebox/graphql"
	"github.com/reynld/carbtographer/pkg/models"
)

// searchBusiness gets ran in goroutine and returns the response to channel when done
func searchBusiness(cl *graphql.Client, rest models.Restaurants, ch chan<- models.YelpResponse, lat *float64, lon *float64) {
	key := os.Getenv("YELP_API_KEY")

	var res models.YelpResponse

	req := graphql.NewRequest(models.YelpQuery)
	req.Var("name", rest.Name)
	req.Var("lat", lat)
	req.Var("lon", lon)
	req.Header.Add("Authorization", "Bearer "+key)
	ctx := context.Background()

	err := cl.Run(ctx, req, &res)
	if err != nil {
		log.Fatal(err)
	}

	for i := range res.Search.Business {
		id := &res.Search.Business[i]
		id.RID = rest.ID
	}

	ch <- res
}

// GetLocations returns local businees that names match restuarants in our db
func GetLocations(w http.ResponseWriter, req *http.Request) {
	params := mux.Vars(req)

	lat, err := strconv.ParseFloat(params["lat"], 64)
	if err != nil {
		log.Fatal(err)
	}
	lon, err := strconv.ParseFloat(params["lon"], 64)
	if err != nil {
		log.Fatal(err)
	}

	if lon == float64(-74.0060) && lat == float64(40.7128) {
		jsonRes := make([]models.Business, 0)
		json.Unmarshal(DefaultLocation, &jsonRes)

		json.NewEncoder(w).Encode(jsonRes)
		return
	}

	var names []models.Restaurants
	database.GetNames(&names)

	client := graphql.NewClient("https://api.yelp.com/v3/graphql")
	var ab []models.Business // all businesses
	var uid []string         // unique id

	c := make(chan models.YelpResponse)
	var wg sync.WaitGroup

	for i, name := range names {
		wg.Add(1)
		go func(n models.Restaurants, m int) {
			time.Sleep(time.Millisecond * time.Duration(150*m))
			defer wg.Done()
			searchBusiness(client, n, c, &lat, &lon)
		}(name, i)
	}

	go func() {
		for yr := range c {
			for _, business := range yr.Search.Business {
				exist := false
				for _, id := range uid {
					if id == business.ID {
						exist = true
					}
				}

				vn := false //valid name
				for _, name := range names {
					if name.Name == business.Name {
						vn = true
					}
				}

				if exist == false && vn == true {
					ab = append(ab, business)
					uid = append(uid, business.ID)
				}
			}
		}
	}()

	wg.Wait()
	json.NewEncoder(w).Encode(&ab)
}
