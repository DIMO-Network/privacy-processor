package config

// Settings contains the application config
type Settings struct {
	Environment                   string `yaml:"ENVIRONMENT"`
	Port                          string `yaml:"PORT"`
	LogLevel                      string `yaml:"LOG_LEVEL"`
	KafkaBrokers                  string `yaml:"KAFKA_BROKERS"`
	PrivacyProcessorConsumerGroup string `yaml:"PRIVACY_PROCESSOR_CONSUMER_GROUP"`
	DeviceStatusTopic             string `yaml:"DEVICE_STATUS_TOPIC"`
	PrivacyFenceTopic             string `yaml:"PRIVACY_FENCE_TOPIC"`
	DeviceStatusPrivateTopic      string `yaml:"DEVICE_STATUS_PRIVATE_TOPIC"`
	// V2
	PrivacyProcessorConsumerGroupV2 string `yaml:"PRIVACY_PROCESSOR_CONSUMER_GROUP_V2"`
	DeviceStatusTopicV2             string `yaml:"DEVICE_STATUS_TOPIC_V2"`
	DeviceStatusPrivateTopicV2      string `yaml:"DEVICE_STATUS_PRIVATE_TOPIC_V2"`
	PrivacyFenceTopicV2             string `yaml:"PRIVACY_FENCE_TOPIC_V2"`
}
