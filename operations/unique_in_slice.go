package operations

func Unique(slice []string) []string {
	encountered := map[string]int{}
	var diff []string

	for _, v := range slice {
		encountered[v] = encountered[v]+1
	}

	for _, v := range slice {
		if encountered[v] == 1 {
			diff = append(diff, v)
		}
	}
	return diff
}

func removeDuplicates(elements []string) (result []string) {
	encountered := map[string]bool{}
	for _, v := range elements {
		if !encountered[v] {
			encountered[v] = true
			result = append(result, v)
		}
	}

	return
}