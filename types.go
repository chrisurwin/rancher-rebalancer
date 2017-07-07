package main

type serviceDef struct {
	name        string
	id          string
	instanceIds []string
}

type host struct {
	id       string
	conCount int
	service  string
}
