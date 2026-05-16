package cicd

import "testing"

type testProvider struct{}

func (p *testProvider) GetTag() (string, error)                   { return "", nil }
func (p *testProvider) GetActor() (string, error)                 { return "", nil }
func (p *testProvider) GetOwnerAndRepo() (string, string, error)  { return "", "", nil }
func (p *testProvider) GetWorkDirectory() string                  { return "" }
func (p *testProvider) GetTemplateFile() string                   { return "" }
func (p *testProvider) IsPreRelease(_, _, _ string) (bool, error) { return false, nil }

func TestGetProviderUsesRegisteredProviders(t *testing.T) {
	originalProviders := providers
	providers = nil
	defer func() {
		providers = originalProviders
	}()

	RegisterProvider(Registration{
		Name: "never-match",
		Detect: func() bool {
			return false
		},
		New: func() Provider {
			t.Fatal("unexpected provider construction")
			return nil
		},
	})

	expected := &testProvider{}
	RegisterProvider(Registration{
		Name: "match",
		Detect: func() bool {
			return true
		},
		New: func() Provider {
			return expected
		},
	})

	got := GetProvider()
	if got != expected {
		t.Fatalf("expected registered provider to be returned")
	}
}
