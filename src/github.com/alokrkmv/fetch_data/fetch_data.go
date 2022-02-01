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

const BASE_id = "https://challenges.qluv.io/items/"

// Intializing the mutex variable to prevent race condition
var mutex = &sync.RWMutex{}

// A basic get request handler
func get_request_handler(id string) (res interface{}, is_429 bool, err error) {
	client := &http.Client{}
	final_id := BASE_id + id
	req, _ := http.NewRequest("GET", final_id, nil)
	// req.Header.Set("Authorization", "Y1JGMmR2RFpRc211MzdXR2dLNk1UY0w3WGpI")
	req.Header.Set("Authorization", helpers.AuthHeaderGenerator(id))
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
		fmt.Println(err)

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
func GetConcurrentData(ids []string) (final_res map[string]string, meta_map map[string]interface{}) {
	// This map contains the final result
	final_res = make(map[string]string)
	// This map contains various meta data
	var failed_ids []string
	var successful_ids []string
	var duplicate_ids_count int
	var duplicate_id []string
	var count_429 int
	wg := sync.WaitGroup{}
	i := 0
	for i < len(ids) {
		var spliced_id []string
		// Splicing the array of ids into a set of 5 as the API can't handle more than
		// five concurrent requests at a time.
		// This will prevent API from throwing 429 error and also ensure to get
		// maximum throughput from the API
		if i+5 < len(ids) {
			spliced_id = ids[i : i+5]
		} else {
			spliced_id = ids[i:]
		}

		// Filtering any duplicate id to avoid unwanted goroutine spwans
		spliced_id, duplicate_count, duplicate_ids := helpers.RemoveDuplicateValues(spliced_id)
		duplicate_ids_count = duplicate_ids_count + duplicate_count
		duplicate_id = append(duplicate_id, duplicate_ids...)
		for _, id := range spliced_id {
			// If the result is already present in redis cache then serve from cache
			// This will prevent API calls for duplicate ids
			val, err := cache.Get(context.TODO(), id).Result()
			if err == nil {
				final_res[id] = val
				duplicate_ids_count++
				duplicate_id = append(duplicate_id, id)
				continue
			}
			wg.Add(1)
			// Initialize go routines for concurrent calls.
			go func(id string) {

				res, is_429, err := get_request_handler(id)
				if is_429 {
					count_429++
				}
				if err != nil {
					log.Fatal(err)
					// Adding mutex lock to prevent race condition
					mutex.Lock()
					failed_ids = append(failed_ids, id)
					mutex.Unlock()
				}
				// Adding mutex lock to prevent race condition while writing result to map
				mutex.Lock()
				final_res[id] = string(res.([]uint8))
				mutex.Unlock()

				// Once the data is fetched caching it for one hour in redis cache so that
				// repeated API calls for duplicate ids can be avoided.
				// The TTL is for one hour so that any update in the response for the same
				// id can be reflected after an hour. This value can be changed based on the
				// the actual scenario

				// We don't need to add any mutex lock while writing to redis cache
				// as redis is capable of handling concurrent read and writes

				err = cache.Set(context.TODO(), id, string(res.([]uint8)), 1440*time.Second).Err()
				mutex.Lock()
				successful_ids = append(successful_ids, id)
				mutex.Unlock()
				if err != nil {
					log.Fatal(err)
				}
				wg.Done()
			}(id)

		}
		wg.Wait()
		i = i + 5
	}
	// Storing unique duplicate ids
	duplicate_id, _, _ = helpers.RemoveDuplicateValues(duplicate_id)
	meta_map = map[string]interface{}{
		"total_ids":                            ids,
		"total_ids_count":                      len(ids),
		"failed_ids":                           failed_ids,
		"successful_ids":                       successful_ids,
		"duplicate_ids_count":                  duplicate_ids_count,
		"unique_duplicate_ids":                 duplicate_id,
		"count_429":                            count_429,
		"number_of_failed_ids":                 len(failed_ids),
		"number_of_api_calls":                  len(successful_ids),
		"number_of_repeated_api_calls_avoided": len(ids) - len(successful_ids),
	}
	return final_res, meta_map
}
