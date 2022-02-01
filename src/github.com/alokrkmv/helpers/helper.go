package helpers

import (
	b64 "encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Random seed
var seededRand *rand.Rand = rand.New(
	rand.NewSource(time.Now().UnixNano()))

// A helper function that removes duplicates from array
func RemoveDuplicateValues(intSlice []string) (list []string, duplicate_count int, duplicate_ids []string) {
	keys := make(map[string]bool)

	// If the key(values of the slice) is not equal
	// to the already present value in new slice (list)
	// then we append it. else we jump on another element.
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		} else {
			duplicate_ids = append(duplicate_ids, entry)
			duplicate_count++
		}
	}
	return list, duplicate_count, duplicate_ids
}

// Base64 encoder
func AuthHeaderGenerator(id string) string {
	return b64.StdEncoding.EncodeToString([]byte(id))
}

func GenerateIDs(number_of_id int) {
	var generated_ids []string

	for i := 0; i < number_of_id; i++ {
		max := 27
		min := 10
		generated_ids = append(generated_ids, string_with_charset(rand.Intn(max-min)+min))
	}
	// Fetch the location of the current directory
	mydir, _ := os.Getwd()
	// Generate the complete file path
	filename := fmt.Sprintf("%s/%s", mydir, "read_from_file/ids.txt")
	f, err := os.OpenFile(filename, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer f.Close()
	text := strings.Join(generated_ids[:], ",")
	text =","+text

	if _, err = f.WriteString(text); err != nil {
		panic(err)
	}

}

func string_with_charset(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}
