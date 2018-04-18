package main

import (
	"runtime"
	"testing"
)

func BenchmarkCreateNetworkBridge(b *testing.B) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	b, err := newCNIBenchmark()
	if err != nil {
		t.Fatal(err)
	}
	defer b.originalNS.Close()

	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		if err := b.createNetwork("bridge"); err != nil {
			t.Fatalf("[%d] %v", n, err)
		}
	}
}
