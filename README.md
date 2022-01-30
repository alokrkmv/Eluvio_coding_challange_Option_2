[![forthebadge](https://forthebadge.com/images/badges/made-with-go.svg)](https://forthebadge.com)

[![forthebadge](https://forthebadge.com/images/badges/gluten-free.svg)](https://forthebadge.com)

[![forthebadge](https://forthebadge.com/images/badges/powered-by-coffee.svg)](https://forthebadge.com)

## Problem Statement


Imagine you have a program that needs to look up information about items using their item ID, often in large batches.

Unfortunately, the only API available for returning this data takes one item at a time, which means you will have to perform one query per item. Additionally, the API is limited to five simultaneous requests. Any additional requests will be served with HTTP 429 (too many requests).

Write a client utility for your program to use that will retrieve the information for all given IDs as quickly as possible without triggering the simultaneous requests limit, and without performing unnecessary queries for item IDs that have already been seen.

**API Usage:**

GET [https://challenges.qluv.io/items/:id](https://eluv.io/items/:id)

**Required headers:**

Authorization: Base64(:id)

**Example:**

curl [https://challenges.qluv.io/items/cRF2dvDZQsmu37WGgK6MTcL7XjH](https://challenges.qluv.io/items/cRF2dvDZQsmu37WGgK6MTcL7XjH) -H "Authorization: Y1JGMmR2RFpRc211MzdXR2dLNk1UY0w3WGpI"​

***

### My Approach

To solve this problem, I broke down the problem into three major chunks and then solved the chunks one by one. The major three chunks of this problem are:
1.   The most basic part of the problem statement requires performing a **GET** request and parsing the response. I wrote a wrapper function to perform GET requests using go's standard HTTP package. The function is capable of handling various kinds of errors and parsing the response body.
2. After creating the handler I wrote another function capable of performing simultaneous API requests to the given endpoint. To achieve this I used goroutines and managed them using wait groups. Also as stated in the problem statement the API is capable of handling *only 5 simultaneous requests at a time* so I spawned the goroutines in a group of five so that API doesn't throw 429 and also provides maximum throughput.
3. Last but not least as per the requirement of a problem statement, I added the logic to prevent API calls for duplicate ids. For this, I used **Redis** caching layer. The reason for choosing Redis caching over a local data structure like Map or array is that Redis is a high performant persistent data store that allows data to persist even after the execution of the program finishes. This helps in avoiding API calls for already fetched ids not in a single run but even in the subsequent run of the program.

#### Why Go??
I choose **Go** for this project as the soul of this project is handling concurrent calls and Go is a language build for concurrency. The goroutines, channels and wait groups provides an elegant and highly efficient way of handling concurrency compared to thread and processes which are used by most other languages like Python, Java etc. Handling race condition using Go's inbuilt mutex implementation is quite convenient. Because of all these aspects I felt that Go would be an ideal choice for this project.

*Brief code Snippet of the Concurrent data Fetcher function is provided below*

#### Talk is cheap show me the code

```go
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

```

***
#### Steps to run the project in local

*These steps have only been tested on linux operating system with Debian distribution. As the project is independent of Operating system, ideally it should run fine on any OS although setup process may vary for other OS.*

1. Clone the project to your local machine using ````git clone https://github.com/alokrkmv/Eluvio_coding_challange_Option_2.git ````
2. Setup go environment on your local machine. Checkout the official documentation for the same [Download and install go](https://go.dev/doc/install)
3. Once the Go is setup in your machine, add project GOPATH using the following steps
	* Go to the root folder of the project (Eluvio_coding_challange_Option_2)(https://github.com/alokrkmv/Eluvio_coding_challange_Option_2)** and use *pwd* command to get the current path of the root folder. 
	* Now export the path that you got in previous step as GOPATH using ````export GOPTAH=<path_obtained_from_pwd>````
	* Verify the GOPATH using ````echo $GOPATH```
4. Once GOPATH is properly setup install all the dependencies by running ````bash install_dependecies.sh```` from root folder.
5. This project uses REDIS as a caching layer so an instance of up and running redis server is required in the local machine. To setup redis server you can use either of the following two ways.
	* *Setting up the Redis server using docker (recommended)*
		*  Setup docker in your local machine by following the official documentation [setup docker for ubuntu](https://docs.docker.com/engine/install/ubuntu/)
		* Pull docker redis container using ````docker pull redis```. This will pull the official redis image to your local machine.
		* Make container up and running on port *6379* using ````docker run --name redis-test-instance -p 6379:6379 -d redisdocker run --name redis-test-instance -p 6379:6379 -d redis````
		* Get the container id using ````docker ps````
		* Exec inside the container using ````docker exec -it container_id bash```
		* Once inside the container start the redis client using ``redis-cli``
		* Now you can see any changes happening into the cache in real time.
	* Setting up the Redis server without using docker (not recommended)
		* Install redis in your local machine by following the official documentation [getting started with redis](https://redis.io/topics/quickstart)
		* Once redis is successfully installed in your machine start the redis server using ````redis-server```
		* Start the redis client in a separate terminal using ````redis-cli````. 
		* Now you can see any changes happening into the cache in real time.
6. Once the setup part for the project is complete cd into ````src/github.com/alokrkmv```` and then execute ````go run main.go````
7. If everything goes right you will see the message ````Program Executed Successfully```` in the console. In case of any error some error message will pop up.
8. If the execution goes right two files **output.json** and **meta_data.json** will be generated inside the ````src/github.com/alokrkmv/writer```` 
9. ````output.json```` will have all the response data for all the unique request ids. This is the final output of the program
10. ````meta_data.json```` will have various meta data like * total run time, number of duplicate calls prevented, list of failed ids, number of 429 requests occurred* etc. This data can help in providing further analysis about the program.







