package merki

import (
	"bytes"
	"strings"
	"testing"
)

func TestFilterMultipleMeasurements(t *testing.T) {
	var output bytes.Buffer
	m := New(',', &output)
	input := strings.NewReader("2024-01-01 10:00:00,pulse,2\n2024-01-01 10:01:00,weight,70\n2024-01-01 10:02:00,steps,100\n")

	if err := m.Filter(input, []string{"pulse", "weight", "pulse"}, IntervalNone, TypeAverage, nil); err != nil {
		t.Fatalf("Filter returned an error: %v", err)
	}

	want := "2024-01-01 10:00:00,pulse,2.000\n2024-01-01 10:01:00,weight,70.000\n"
	if got := output.String(); got != want {
		t.Errorf("Filter output = %q, want %q", got, want)
	}
}

func TestFilterGroupsMeasurementsIndependently(t *testing.T) {
	var output bytes.Buffer
	m := New(',', &output)
	input := strings.NewReader("2024-01-01 10:00:00,pulse,2\n2024-01-01 10:01:00,weight,70\n2024-01-01 10:02:00,pulse,3\n2024-01-01 10:03:00,steps,100\n")

	if err := m.Filter(input, []string{"pulse", "weight", "pulse"}, IntervalDaily, TypeSum, nil); err != nil {
		t.Fatalf("Filter returned an error: %v", err)
	}

	want := "2024-01-01,pulse,5.000\n2024-01-01,weight,70.000\n"
	if got := output.String(); got != want {
		t.Errorf("Filter output = %q, want %q", got, want)
	}
}
