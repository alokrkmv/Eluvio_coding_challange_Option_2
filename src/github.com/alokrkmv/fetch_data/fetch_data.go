package fetch_data

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/go-redis/redis"
)

// Adding the redis client which will act as persistant caching layer
// to cache the data for the ids which are already being fetched 

var cache = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

const BASE_URL = "https://challenges.qluv.io/items/"

// A basic get request handler
func get_request_handler(url string) (res interface{}, err error) {

	client := &http.Client{}
	final_url := BASE_URL + url
	req, _ := http.NewRequest("GET", final_url, nil)
	req.Header.Set("Authorization", "Y1JGMmR2RFpRc211MzdXR2dLNk1UY0w3WGpI")

	resp, err := client.Do(req)

	// Log and return the error in case the request fails
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	// Close the request body to avoid any kind of leaks
	defer resp.Body.Close()
	// Check for the API response type .
	// If the API response is something else than 200 then exit
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		err = errors.New("API response status is not 200")
		log.Fatal(err)
		return nil, err
	}
	// Parse the response body and handle any error that occurs while parsing
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.New("somthing went wrong in reading the response body")
		log.Fatal(err)
		return nil, err
	}
	return responseData, err
}

func GetData(urls []string) (final_res map[string]string, failed_ids []string) {

	final_res = make(map[string]string)

	for _, url := range urls {
		// Try to fetch the response for a particular id from redis cache
		// to avoid any API call in case of cache hit
		val, err := cache.Get(context.TODO(), url).Result()
		if err == nil {
			final_res[url] = val
			continue
		}

		res, err := get_request_handler(url)
		if err != nil {
			log.Fatal(err)
			failed_ids = append(failed_ids, url)
		}
		// Once the data is fetched caching it for one hour in redis cache so that 
		// repeated API calls for duplicate ids can be avoided.
		// The TTL is for one hour so that any update in the response for the same
		// id can be reflected after an hour. This value can be changed based on the
		// the actual scenario
		
		err = cache.Set(context.TODO(), url, string(res.([]uint8)), 1440*time.Second).Err()
		if err != nil {
			log.Fatal(err)
			failed_ids = append(failed_ids, url)
		}
		// Convert the response from []uint8 to string
		final_res[url] = string(res.([]uint8))
	}

	return final_res, failed_ids

}
