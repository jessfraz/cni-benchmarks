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

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		if err := a.createNetwork("bridge"); err != nil {
			b.Fatalf("[%d] %v", n, err)
		}
	}
}
