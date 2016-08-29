// +build windows

package app

import "testing"

var interpolateEnvironmentVariablesTests = []struct {
	id       int
	data     string
	expected string
}{
	{1, "%", ""},
	{2, "%SYSTEMROOT%\\test", "C:\\Windows\\test"},
	{3, "test\\%SYSTEMROOT%\\test", "test\\C:\\Windows\\test"},
	{4, "test\\%SYSTEMROOT%", "test\\C:\\Windows"},
}

func TestInterpolateEnvironmentVariables(t *testing.T) {
	for _, tt := range interpolateEnvironmentVariablesTests {
		actual, err := interpolateEnvironmentVariables(tt.data)
		if err != nil {
			t.Errorf("[id: %d] interpolateEnvironmentVariables: failed with %s", tt.id, err)
			continue
		}
		if tt.expected != actual {
			t.Errorf("[id: %d] interpolateEnvironmentVariables: expected %s, actual %s", tt.id, tt.expected, actual)
		}
	}
}
