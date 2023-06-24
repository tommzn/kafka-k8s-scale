package main

type KafKaMetrics interface {
	ConsumerLag(topic string, groupId string) (error, int)
}
