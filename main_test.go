package main

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/vishvananda/netns"
)

func BenchmarkAzure(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "azure")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "azure")
	})
}

func BenchmarkBridge(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "bridge")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "bridge")
	})
}

// You should run `make run-calico` before running these benchmarks.
func BenchmarkCalico(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "calico")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "calico")
	})
}

// You should run `make run-cilium` before running these benchmarks.
func BenchmarkCilium(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "cilium")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "cilium")
	})
}

// You should run `make run-flannel` before running these benchmarks.
func BenchmarkFlannelIPvlan(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "flannel-ipvlan")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "flannel-ipvlan")
	})
}

// You should run `make run-flannel` before running these benchmarks.
func BenchmarkFlannelBridge(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "flannel-bridge")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "flannel-bridge")
	})
}

func BenchmarkIPvlan(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "ipvlan")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "ipvlan")
	})
}

func BenchmarkMacvlan(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "macvlan")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "macvlan")
	})
}

func BenchmarkPTP(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "ptp")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "ptp")
	})
}

// You should run `make run-weave` before running these benchmarks.
func BenchmarkWeave(b *testing.B) {
	b.Run("setup network in netns", func(b *testing.B) {
		runBenchmarkSetupNetNS(b, "weave")
	})
	b.Run("delete network from netns", func(b *testing.B) {
		runBenchmarkDeleteNetwork(b, "weave")
	})
}

func runBenchmarkSetupNetNS(b *testing.B, plugin string) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark(false)
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	b.ResetTimer()
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		if err := a.createProcess(plugin); err != nil {
			b.Fatal(err)
		}

		if err := a.loadCNIConfig(plugin); err != nil {
			b.Fatal(err)
		}

		b.StartTimer()
		if _, err := a.setupNetNS(); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()
		defer a.nsHandle.Close()

		if err := a.setNS(); err != nil {
			b.Fatal(err)
		}

		if err := netns.Set(a.originalNS); err != nil {
			b.Fatalf("returning to original namespace failed: %v", err)
		}

		if err := a.libcni.Remove(fmt.Sprintf("%d", a.process.Pid), a.netnsFD); err != nil {
			b.Fatal(err)
		}

		if err := a.process.Kill(); err != nil {
			b.Fatal(err)
		}
	}
}

func runBenchmarkDeleteNetwork(b *testing.B, plugin string) {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	a, err := newCNIBenchmark(false)
	if err != nil {
		b.Fatal(err)
	}
	defer a.originalNS.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		if err := a.createProcess(plugin); err != nil {
			b.Fatal(err)
		}

		if err := a.loadCNIConfig(plugin); err != nil {
			b.Fatal(err)
		}

		if _, err := a.setupNetNS(); err != nil {
			b.Fatal(err)
		}
		defer a.nsHandle.Close()

		if err := a.setNS(); err != nil {
			b.Fatal(err)
		}

		if err := netns.Set(a.originalNS); err != nil {
			b.Fatalf("returning to original namespace failed: %v", err)
		}

		b.StartTimer()
		if err := a.libcni.Remove(fmt.Sprintf("%d", a.process.Pid), a.netnsFD); err != nil {
			b.Fatal(err)
		}
		b.StopTimer()

		if err := a.process.Kill(); err != nil {
			b.Fatal(err)
		}
	}
}
