package main

// Lag is the parent object that will contain the sumary for the particular instance of the consumer
type Lag struct {
	Lag   int            `json:"totalLag"` //aggregates all the partition lag and displays the sum
	Topic string         `json:"topic"`    //topic assigned to the consumer
	Info  []ConsumerInfo `json:"info"`     //array of info per partition
}

// ConsumerInfo is the specific object that will wrap the information per partition
type ConsumerInfo struct {
	Lag       int    `json:"lag"`
	Topic     string `json:"topic"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
}
