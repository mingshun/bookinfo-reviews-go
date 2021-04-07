package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

const (
	ratings_enabled = true
)

func star_color() string {
	if sc := os.Getenv("STAR_COLOR"); sc!="" {
		return sc
	} else {
		return "black"
	}
}

func services_domain() string {
	if sd := os.Getenv("SERVICES_DOMAIN"); sd != "" {
		return "." + sd
	} else {
		return ""
	}
}

func ratings_hostname() string {
	if rh := os.Getenv("RATINGS_HOSTNAME"); rh != "" {
		return rh
	} else {
		return "ratings"
	}
}

func ratings_service() string {
	return "http://" + ratings_hostname() + services_domain() + ":9080/ratings"
}

type RatingResp struct {
	Id int `json:"id"`
	Ratings map[string]int `json:"ratings"`
}

 type ReviewResp struct {
	 Id string `json:"id"`
	 Reviews []*Review `json:"reviews"`
 }

 type Review struct {
	Reviewer string `json:"reviewer"`
	Text string `json:"text"`
	Rating *Rating `json:"rating"`
 }

 type Rating struct {
	Stars int `json:"stars,omitempty"`
	Color string `json:"color,omitempty"`
	Error string `json:"error,omitempty"`
 }

func main() {
	r := mux.NewRouter()
    r.PathPrefix("/health").Subrouter().Methods(http.MethodGet).Subrouter().HandleFunc("", health)
    r.PathPrefix("/reviews/{productId}").Subrouter().Methods(http.MethodGet).Subrouter().HandleFunc("", bookReviewsById)
    http.Handle("/", r)

	log.Println("Start listening http port 9080 ...")
	if err := http.ListenAndServe(":9080", nil); err != nil {
		panic(err)
	}
}

func getJsonResponse(productId string, starsReviewer1, starsReviewer2 int) ([]byte, error) {
	rating1 := &Rating{}
	if starsReviewer1 != -1 {
		rating1.Stars = starsReviewer1
		rating1.Color = star_color()
	} else {
		rating1.Error = "Ratings service is currently unavailable"
	}
	
	rating2 := &Rating{}
	if starsReviewer2 != -1 {
		rating2.Stars = starsReviewer2
		rating2.Color = star_color()
	} else {
		rating2.Error = "Ratings service is currently unavailable"
	}

	reviewResp := &ReviewResp{
		Id: productId,
		Reviews: []*Review{
			&Review{
				Reviewer: "Reviewer1",
				Text: "An extremely entertaining play by Shakespeare. The slapstick humour is refreshing!",
				Rating: rating1,
			},
			&Review{
				Reviewer: "Reviewer2",
				Text: "Absolutely fun and entertaining. The play lacks thematic depth when compared to other plays by Shakespeare.",
				Rating: rating2,
			},
		},
	}

	return json.Marshal(reviewResp)
}

func health(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	resp, err := json.Marshal(map[string]string{
		"status": "Reviews is healthy",
	})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(err)
		return
	}
	w.Write(resp)
}

func getRatings(productId string) ([]byte, error) {
	resp, err := http.Get(ratings_service() + "/" + productId)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func bookReviewsById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	productId := vars["productId"]

	starsReviewer1 := -1
	starsReviewer2 := -1

	if (ratings_enabled) {
		ratings_string, err := getRatings(productId)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		log.Println(string(ratings_string))
		resp := &RatingResp{}
		if err := json.Unmarshal(ratings_string, resp); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
		}
		
		starsReviewer1 = resp.Ratings["Reviewer1"]
		starsReviewer2 = resp.Ratings["Reviewer2"]
	}
	jsonResp, err := getJsonResponse(productId, starsReviewer1, starsReviewer2)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
			log.Println(err)
			return
	}
	w.Write(jsonResp)
}