package config

// Settings contains the application config
type Settings struct {
	KafkaBrokers                  string `yaml:"KAFKA_BROKERS"`
	PrivacyProcessorConsumerGroup string `yaml:"PRIVACY_PROCESSOR_CONSUMER_GROUP"`
	DeviceStatusTopic             string `yaml:"DEVICE_STATUS_TOPIC"`
	PrivacyFenceTopic             string `yaml:"PRIVACY_FENCE_TOPIC"`
	DeviceStatusPrivateTopic      string `yaml:"DEVICE_STATUS_PRIVATE_TOPIC"`
}
