package routes

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/machinebox/graphql"
	"github.com/reynld/carbtographer/pkg/models"
)

func searchBusiness(cl *graphql.Client, s string, ch chan models.YelpResponse) {
	key := os.Getenv("YELP_API_KEY")
	var res models.YelpResponse

	req := graphql.NewRequest(yelpQuery)
	req.Var("name", s)
	req.Header.Add("Authorization", "Bearer "+key)

	ctx := context.Background()

	err := cl.Run(ctx, req, &res)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("\nYR: %s\n%v\n", s, res)
	ch <- res
}

func getLocations(w http.ResponseWriter, req *http.Request) {
	url := os.Getenv("YELP_URL")
	var names []models.Restaurants
	db.Find(&names)

	client := graphql.NewClient(url)
	// all businesses
	var ab []models.Business
	// unique id
	var uid []string

	c := make(chan models.YelpResponse)

	for _, name := range names {
		go searchBusiness(client, name.Name, c)
	}

	count := 0
	for yr := range c {
		for _, business := range yr.Search.Business {
			exist := false
			for _, id := range uid {
				if id == business.ID {
					exist = true
				}
			}

			//valid name
			vn := false
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
		count++
		if count >= len(names) {
			close(c)
		}
	}

	json.NewEncoder(w).Encode(&ab)

}
