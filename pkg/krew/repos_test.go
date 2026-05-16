package krew

import "testing"

func TestDefaultIndexRepoName(t *testing.T) {
	if DefaultIndexRepoName != "krew-index" {
		t.Fatalf("unexpected default repo name: %s", DefaultIndexRepoName)
	}
}

func TestDefaultIndexRepoOwner(t *testing.T) {
	if DefaultIndexRepoOwner != "kubernetes-sigs" {
		t.Fatalf("unexpected default repo owner: %s", DefaultIndexRepoOwner)
	}
}
