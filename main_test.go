package main

import (
	"runtime"
	"testing"
)

func BenchmarkCreateNetworkBridge(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("bridge", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

func BenchmarkCreateNetworkIPvlan(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("ipvlan", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

func BenchmarkCreateNetworkMacvlan(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("macvlan", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

func BenchmarkCreateNetworkPTP(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("ptp", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}
