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

	err = terminalBaremetalAllocationError("bm-1", "Failed", "volume fetch not found")
	if err == nil {
		t.Fatal("expected Failed UI state to stop waiting")
	}
	got = err.Error()
	want = `baremetal server "bm-1" provisioning failed (state="Failed"): volume fetch not found`
	if got != want {
		t.Fatalf("error = %q, want %q", got, want)
	}

	err = terminalBaremetalAllocationError("bm-1", "InitializingDependencies", "volume fetch not found")
	if err == nil {
		t.Fatal("volume fetch not found must stop waiting even before state flips to Failed")
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

func TestIsBaremetalReadyAndPoweredOn(t *testing.T) {
	t.Parallel()

	if isBaremetalReadyAndPoweredOn("InitializingDependencies", "Unknown") {
		t.Fatal("intermediate state must keep polling")
	}
	if isBaremetalReadyAndPoweredOn("Ready", "Unknown") {
		t.Fatal("Ready with power Unknown must keep polling")
	}
	if !isBaremetalReadyAndPoweredOn("Ready", "On") {
		t.Fatal("Ready and On must complete wait")
	}
	if !isBaremetalReadyAndPoweredOn("ready", "poweredon") {
		t.Fatal("Ready and PoweredOn must complete wait")
	}
}
