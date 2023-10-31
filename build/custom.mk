# Include custom targets and environment variables here

# Generates mock golang interfaces for testing
.PHONY: mock
mock:
ifneq ($(HAS_SERVER),)
	go install github.com/golang/mock/mockgen@v1.6.0
	mockgen --build_flags=--mod=mod -destination server/prometheus/mock_client/mock_client.go github.com/mattermost/mattermost-plugin-metrics-comparison/server/prometheus PrometheusClient
endif
