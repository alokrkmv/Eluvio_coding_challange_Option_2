package fetch_data

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/alokrkmv/helpers"
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

// Intializing the mutex variable to prevent race condition
var mutex = &sync.RWMutex{}

// A basic get request handler
func get_request_handler(url string) (res interface{}, is_429 bool, err error) {
	client := &http.Client{}
	final_url := BASE_URL + url
	req, _ := http.NewRequest("GET", final_url, nil)
	req.Header.Set("Authorization", "Y1JGMmR2RFpRc211MzdXR2dLNk1UY0w3WGpI")

	resp, err := client.Do(req)

	// Log and return the error in case the request fails
	if err != nil {
		log.Fatal(err)
		return nil, false, err
	}
	// Close the request body to avoid any kind of leaks
	defer resp.Body.Close()
	// Check for the API response type .
	// If the API response is something else than 200 then exit
	if !(resp.StatusCode >= 200 && resp.StatusCode <= 299) {
		err = errors.New("API response status is not 200")
		fmt.Println(resp.StatusCode)
		log.Fatal(err)

		if resp.StatusCode == 429 {
			is_429 = true

		}
		return nil, is_429, err
	}
	// Parse the response body and handle any error that occurs while parsing
	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		err = errors.New("somthing went wrong in reading the response body")
		log.Fatal(err)
		return nil, false, err
	}
	return responseData, false, err
}

// Function to fetch data from the API concurrently
func GetConcurrentData(urls []string) (final_res map[string]string, meta_map map[string]interface{}) {
	// This map conatins the final result
	final_res = make(map[string]string)
	// This map contains various meta data
	var failed_ids []string
	var successful_ids []string
	var duplicate_ids_count int
	var duplicate_id []string
	var count_429 int
	wg := sync.WaitGroup{}
	i := 0
	for i < len(urls) {
		var spliced_url []string
		// Splicing the array of ids into a set of 5 as the API can't handle more than
		// five concurrent requests at a time.
		// This will prevent API from throwing 429 error and also ensure to get
		// maximum throughput from the API
		if i+5 < len(urls) {
			spliced_url = urls[i : i+5]
		} else {
			spliced_url = urls[i:]
		}

		// Filtering any duplicate id to avoid unwanted goroutine spwans
		spliced_url, duplicate_count, duplicate_ids := helpers.RemoveDuplicateValues(spliced_url)
		duplicate_ids_count = duplicate_ids_count + duplicate_count
		duplicate_id = append(duplicate_id, duplicate_ids...)
		for _, url := range spliced_url {
			// If the result is already present in redis cache then serve from cache
			// This will prevent API calls for duplicate ids
			val, err := cache.Get(context.TODO(), url).Result()
			if err == nil {
				final_res[url] = val
				duplicate_ids_count++
				continue
			}
			wg.Add(1)
			// Intialize go routines for concurrent calls.
			go func(url string) {

				res, is_429, err := get_request_handler(url)
				if is_429 {
					count_429++
				}
				if err != nil {
					log.Fatal(err)
					// Adding mutex lock to prevent race condition
					mutex.Lock()
					failed_ids = append(failed_ids, url)
					mutex.Unlock()
				}
				// Adding mutex lock to prevent race condition while writing result to map
				mutex.Lock()
				final_res[url] = string(res.([]uint8))
				mutex.Unlock()

				// Once the data is fetched caching it for one hour in redis cache so that
				// repeated API calls for duplicate ids can be avoided.
				// The TTL is for one hour so that any update in the response for the same
				// id can be reflected after an hour. This value can be changed based on the
				// the actual scenario

				// We don't need to add any mutex lock while writing to redis cache
				// as redis is capable of handling concurrent read and writes

				err = cache.Set(context.TODO(), url, string(res.([]uint8)), 1440*time.Second).Err()
				mutex.Lock()
				successful_ids = append(successful_ids, url)
				mutex.Unlock()
				if err != nil {
					log.Fatal(err)
				}
				wg.Done()
			}(url)

		}
		wg.Wait()
		i = i + 5
	}
	// Storing unique duplicate ids
	duplicate_id, _, _ = helpers.RemoveDuplicateValues(duplicate_id)

	meta_map = map[string]interface{}{
		"total_ids":            urls,
		"total_ids_count":      len(urls),
		"failed_ids":           failed_ids,
		"successful_ids":       successful_ids,
		"duplicate_ids_count":  duplicate_ids_count,
		"unique_duplicate_ids": duplicate_id,
		"count_429":            count_429,
		"number_of_failed_ids": len(failed_ids),
		"number_of_api_calls":  len(successful_ids),
	}
	return final_res, meta_map
}
