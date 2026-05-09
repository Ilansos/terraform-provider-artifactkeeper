package provider

import (
	"os"
	"testing"
)

func TestAccPreCheck(t *testing.T) {
	if os.Getenv("ARTIFACTKEEPER_ACC") != "1" {
		t.Skip("set ARTIFACTKEEPER_ACC=1 to run acceptance tests")
	}
	if os.Getenv("ARTIFACTKEEPER_URL") == "" {
		t.Fatal("ARTIFACTKEEPER_URL must be set for acceptance tests")
	}
	if os.Getenv("ARTIFACTKEEPER_TOKEN") == "" && (os.Getenv("ARTIFACTKEEPER_USERNAME") == "" || os.Getenv("ARTIFACTKEEPER_PASSWORD") == "") {
		t.Fatal("set ARTIFACTKEEPER_TOKEN or ARTIFACTKEEPER_USERNAME/ARTIFACTKEEPER_PASSWORD for acceptance tests")
	}
}
