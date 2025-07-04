package main

import "testing"

func TestSendMetrics(t *testing.T) {
	tests := []struct {
		name        string
		metricType  string
		metricName  string
		metricValue float64
	}{
		{
			name:        "1",
			metricType:  "gauge",
			metricName:  "TestGauge",
			metricValue: 10,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

		})
	}

}
