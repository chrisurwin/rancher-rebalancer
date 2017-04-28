package main

import "regexp"

func findIP(input string) string {
	numBlock := "(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])"
	regexPattern := numBlock + "\\." + numBlock + "\\." + numBlock + "\\." + numBlock

	regEx := regexp.MustCompile(regexPattern)
	return regEx.FindString(input)
}

func roundCount(serv int, cont int) int {
	average := cont / serv
	if average*serv < cont {
		average = average + 1
	}
	return average
}
