package provider

import "testing"

func TestTerminalBaremetalAllocationError(t *testing.T) {
	t.Parallel()

	if err := terminalBaremetalAllocationError("bm-1", "NoResource", ""); err != nil {
		t.Fatalf("empty lastErrMsg should not be terminal, got %v", err)
	}
	if err := terminalBaremetalAllocationError("bm-1", "Allocating", "No bms instance found in the general pool"); err != nil {
		t.Fatalf("Allocating with lastErrMsg should not be terminal yet, got %v", err)
	}

	err := terminalBaremetalAllocationError("bm-1", "NoResource", "No bms instance found in the general pool; provider: bmaas-provider-s1 | flavor: metal-c56-m1024")
	if err == nil {
		t.Fatal("expected terminal allocation error")
	}
	got := err.Error()
	want := `baremetal server "bm-1" allocation failed (state="NoResource"): No bms instance found in the general pool; provider: bmaas-provider-s1 | flavor: metal-c56-m1024`
	if got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}
}

func TestLooksLikeUUID(t *testing.T) {
	t.Parallel()
	if !looksLikeUUID("33c9f6c7-ba79-41e0-9009-0ec94ab1cdf4") {
		t.Fatal("expected network UUID to match")
	}
	if looksLikeUUID("copper-vpc1") {
		t.Fatal("expected VPC name not to match UUID")
	}
}

func TestIsProvisionedBaremetalState(t *testing.T) {
	t.Parallel()

	if isProvisionedBaremetalState("") || isProvisionedBaremetalState("NoResource") {
		t.Fatal("empty and NoResource must not count as provisioned")
	}
	if !isProvisionedBaremetalState("Allocating") || !isProvisionedBaremetalState("Ready") {
		t.Fatal("Allocating and Ready must count as provisioned")
	}
}
