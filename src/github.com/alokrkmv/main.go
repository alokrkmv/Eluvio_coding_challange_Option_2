package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/alokrkmv/fetch_data"
	"github.com/alokrkmv/read_from_file"
	"github.com/alokrkmv/writer"
)

func main() {
	// Uncomment this function to generate new ids
	// helpers.GenerateIDs(10)
	ids := read_from_file.ReadFromTextFile()
	start_time := time.Now()
	final_res, meta_map := fetch_data.GetConcurrentData(ids)
	end_time := time.Now()
	total_response_time := end_time.Sub(start_time).Seconds()
	converted_string_time := fmt.Sprintf("%f", total_response_time)
	meta_map["total_response_time"] = converted_string_time + " sec"

	// Writing meta data to file is not a primary task so we will spawn a go routine to
	// write meta data hence it won't hamper the performance of the actual program
	var wg sync.WaitGroup
	wg.Add(1)
	go writer.MetaDataWriter(meta_map, &wg)
	err := writer.WriteOutput(final_res)

	if err != nil {
		fmt.Println("Something went wrong in writing response to the output file")
	}
	wg.Wait()
	fmt.Println("Program Executed Successfully")

}
