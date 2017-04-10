package main

func roundCount(serv int, cont int) int {
	average := cont / serv
	if average*serv < cont {
		average = average + 1
	}
	return average
}
