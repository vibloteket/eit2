package version

import "testing"

func TestDevelopmentVersionIsNotEmpty(t *testing.T) {
	if Value == "" {
		t.Fatal("version must not be empty")
	}
}
