package cicd

import (
	"strings"
	"testing"
)

type testProvider struct{}

func (p *testProvider) GetTag() (string, error)                   { return "", nil }
func (p *testProvider) GetActor() (string, error)                 { return "", nil }
func (p *testProvider) GetOwnerAndRepo() (string, string, error)  { return "", "", nil }
func (p *testProvider) GetWorkDirectory() string                  { return "" }
func (p *testProvider) GetTemplateFile() string                   { return "" }
func (p *testProvider) IsPreRelease(_, _, _ string) (bool, error) { return false, nil }

func TestRegistryGetProviderUsesRegisteredProviders(t *testing.T) {
	registry, err := NewRegistry(
		Registration{
		Name: "never-match",
		Detect: func() bool {
			return false
		},
		New: func() Provider {
			t.Fatal("unexpected provider construction")
			return nil
		},
		},
	)
	if err != nil {
		t.Fatalf("NewRegistry() error = %v", err)
	}

	expected := &testProvider{}
	err = registry.RegisterProvider(Registration{
		Name: "match",
		Detect: func() bool {
			return true
		},
		New: func() Provider {
			return expected
		},
	})
	if err != nil {
		t.Fatalf("RegisterProvider() error = %v", err)
	}

	got := registry.GetProvider()
	if got != expected {
		t.Fatalf("expected registered provider to be returned")
	}
}

func TestRegisterProviderRejectsInvalidRegistrations(t *testing.T) {
	testcases := []struct {
		name        string
		reg         Registration
		expectedErr string
	}{
		{
			name:        "missing name",
			reg:         Registration{},
			expectedErr: "provider registration name is required",
		},
		{
			name: "missing detect",
			reg: Registration{
				Name: "github-actions",
			},
			expectedErr: `provider registration "github-actions" is missing Detect`,
		},
		{
			name: "missing new",
			reg: Registration{
				Name: "github-actions",
				Detect: func() bool {
					return true
				},
			},
			expectedErr: `provider registration "github-actions" is missing New`,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			registry := &Registry{}
			err := registry.RegisterProvider(tc.reg)
			if err == nil {
				t.Fatalf("expected error %q, got nil", tc.expectedErr)
			}
			if !strings.Contains(err.Error(), tc.expectedErr) {
				t.Fatalf("expected error %q, got %q", tc.expectedErr, err.Error())
			}
		})
	}
}
