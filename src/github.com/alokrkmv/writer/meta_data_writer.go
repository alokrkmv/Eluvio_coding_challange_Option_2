package writer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
)



func MetaDataWriter(meta_data map[string]interface{},wg *sync.WaitGroup) {
	// Fetch the location of the current directory
	defer wg.Done()
	mydir, _ := os.Getwd()
	// Generate the complete file path to write the data
	filePath := fmt.Sprintf("%s/%s", mydir, "writer/meta_data.json")
	file, _ := json.MarshalIndent(meta_data, "", " ")
	err := ioutil.WriteFile(filePath, file, 0644)
	if err != nil {
		log.Fatal(err)
	}

}
