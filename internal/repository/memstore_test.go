package repository

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestNew(t *testing.T) {
	memStore := New()
	assert.NotNil(t, memStore)
}

func TestSetGauge(t *testing.T) {
	memStore := New()
	testCases := []struct {
		name      string
		mName     string
		wantValue float64
		wantOk    bool
	}{
		{
			name:      "Test new gauge",
			mName:     "Test name",
			wantValue: 5.4,
			wantOk:    true,
		},
		{
			name:      "Update old gauge",
			mName:     "Test name",
			wantValue: 7.8,
			wantOk:    true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStore.SetGauge(tc.mName, tc.wantValue)
			v, ok := memStore.GetGauge(tc.mName)
			require.Equal(t, tc.wantValue, v)
			require.Equal(t, tc.wantOk, ok)
		})

	}
}

func TestSetCounter(t *testing.T) {
	memStore := New()
	testCases := []struct {
		name      string
		mName     string
		addValue  int64
		wantValue int64
	}{
		{
			name:      "Test new counter",
			mName:     "Test name",
			addValue:  5,
			wantValue: 5,
		},
		{
			name:      "Update counter",
			mName:     "Test name",
			addValue:  7,
			wantValue: 12,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			memStore.AddCounter(tc.mName, tc.addValue)
			v, _ := memStore.GetCounter(tc.mName)
			assert.Equal(t, tc.wantValue, v)
		})
	}
}
