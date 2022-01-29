package helpers

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
