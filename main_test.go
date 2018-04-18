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

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("bridge", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

// You should run `make run-calico` before running this benchmark.
func BenchmarkCreateNetworkCalico(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("calico", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

// You should run `make run-cilium` before running this benchmark.
func BenchmarkCreateNetworkCilium(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("cilium", false); err != nil {
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

	b.ResetTimer()
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

	b.ResetTimer()
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

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("ptp", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}

// You should run `make run-weave` before running this benchmark.
func BenchmarkCreateNetworkWeave(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark()
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("weave", false); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}
