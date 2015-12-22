package main

type ChangeNameEvent struct {
	Id      string
	NewName string
}

type CreateEvent struct {
	Id   string
	Name string
}
