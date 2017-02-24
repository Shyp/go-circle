package circle

import (
	"strings"
	"testing"
)

func TestCaseInsensitive(t *testing.T) {
	cfg := CircleConfig{
		Organizations: map[string]organization{
			"sHYp": organization{Token: "foo"},
		},
	}
	o, err := getCaseInsensitiveOrg("ShyP", cfg.Organizations)
	if err != nil {
		t.Fatalf(err.Error())
	}
	if o.Token != "foo" {
		t.Fatalf("expected o.Token to be foo, was %v", o.Token)
	}

	_, err = getCaseInsensitiveOrg("ShyPmorelongname", cfg.Organizations)
	if err == nil {
		t.Fatalf("should not have found Shypmorelongname in the config, but did")
	}
	if !strings.Contains(err.Error(), "Couldn't find organization ShyPmorelongname in the config") {
		t.Fatalf("expected Couldn't find error message, got %v", err)
	}
}
