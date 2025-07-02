package metricscollector

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_New(t *testing.T) {
	tests := []struct {
		name   string
		mc     MetricsCollector
		topics []string
	}{
		{
			name:   "single topic collector",
			mc:     MetricsCollector{},
			topics: []string{"test-topic"},
		},
		{
			name:   "multi topic collector",
			mc:     MetricsCollector{},
			topics: []string{"test-topic", "another-topic"},
		},
	}
	for _, test := range tests {
		err := test.mc.New(test.topics)
		assert.Nil(t, err)

		structValues := reflect.ValueOf(test.mc)
		numField := structValues.NumField()

		// ensures all fields in struct are properly instantiated
		for i := 0; i < numField; i++ {
			field := structValues.Field(i)
			assert.True(t, field.IsValid())
			assert.True(t, !field.IsZero())
		}
		// ensures the number of fields in the type and instantiated version match
		assert.Equal(t, reflect.TypeOf(MetricsCollector{}).NumField(), reflect.TypeOf(test.mc).NumField())
	}
}
