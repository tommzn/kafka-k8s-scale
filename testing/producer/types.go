package main

import "time"

type Message struct {
	Text      string    `json:"text"`
	TimeStamp time.Time `json:"timestamp"`
}
