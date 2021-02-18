package main

// Checks if val is contained in the slice
func containsByte(slice []byte, val byte) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func containsString(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
