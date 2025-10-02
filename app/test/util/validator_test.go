package util_test

import (
	"testing"

	"backend/service-platform/app/pkg/util/validator"
)

func TestIsDomainAllowed_Exact(t *testing.T) {
	allowed := []string{"app.savecoin.xyz"}
	if !validator.IsDomainAllowed("app.savecoin.xyz", allowed) {
		t.Fatalf("expected exact match to be allowed")
	}
	if validator.IsDomainAllowed("api.savecoin.xyz", allowed) {
		t.Fatalf("did not expect different subdomain to be allowed")
	}
}

func TestIsDomainAllowed_Wildcard(t *testing.T) {
	allowed := []string{"*.savecoin.xyz"}
	if !validator.IsDomainAllowed("app.savecoin.xyz", allowed) {
		t.Fatalf("expected wildcard subdomain to be allowed")
	}
	if !validator.IsDomainAllowed("savecoin.xyz", allowed) {
		t.Fatalf("expected apex domain to be allowed by wildcard rule")
	}
}

func TestIsDomainAllowed_LeadingDot(t *testing.T) {
	allowed := []string{".savecoin.xyz"}
	if !validator.IsDomainAllowed("app.savecoin.xyz", allowed) {
		t.Fatalf("expected subdomain to be allowed by leading dot rule")
	}
	if !validator.IsDomainAllowed("savecoin.xyz", allowed) {
		t.Fatalf("expected apex domain to be allowed by leading dot rule")
	}
}

func TestIsDomainAllowed_URLForms(t *testing.T) {
	allowed := []string{"https://app.savecoin.xyz", "http://www.savecoin.xyz"}
	if !validator.IsDomainAllowed("app.savecoin.xyz", allowed) {
		t.Fatalf("expected https origin host to be allowed")
	}
	if !validator.IsDomainAllowed("www.savecoin.xyz", allowed) {
		t.Fatalf("expected http origin host to be allowed")
	}
}

func TestIsDomainAllowed_EmptyInputs(t *testing.T) {
	if validator.IsDomainAllowed("", []string{"app.savecoin.xyz"}) {
		t.Fatalf("expected empty domain to be rejected")
	}
	if validator.IsDomainAllowed("app.savecoin.xyz", nil) {
		t.Fatalf("expected empty allowed list to reject")
	}
}
