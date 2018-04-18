package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"syscall"

	cni "github.com/containerd/go-cni"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netns"
)

var (
	cniBinDir = filepath.Join(os.Getenv("GOPATH"), "src/github.com/containernetworking/plugins/bin")
)

func main() {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// Save the current network namespace.
	originalNS, err := netns.Get()
	if err != nil {
		logrus.Fatalf("getting current netns failed: %v", err)
	}
	defer originalNS.Close()

	// Create a new network namespace.
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNET}
	if err := cmd.Start(); err != nil {
		logrus.Fatalf("unsharing command failed: %v", err)
	}
	pid := cmd.Process.Pid

	/*newNS, err := netns.New()
	if err != nil {
		logrus.Fatalf("creating new netns failed: %v", err)
	}
	defer newNS.Close()

	// Switch into the new netns.
	if err := netns.Set(newNS); err != nil {
		logrus.Fatalf("switching to new netns failed: %v", err)
	}*/

	// Initialize CNI library.
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("getting working directory failed: %v", err)
	}
	libcni, err := cni.New(
		cni.WithMinNetworkCount(2),
		cni.WithPluginConfDir(filepath.Join(wd, "net.d")),
		cni.WithPluginDir([]string{cniBinDir, cni.DefaultCNIDir}),
	)
	if err != nil {
		logrus.Fatalf("creating new CNI instance failed: %v", err)
	}

	// Load the CNI configuration.
	if err := libcni.Load(
		cni.WithLoNetwork,
		cni.WithConfFile(filepath.Join(wd, "net.d", "macvlan.conf")),
	); err != nil {
		logrus.Fatalf("loading CNI configuration failed: %v", err)
	}

	// Setup network for namespace.
	//netnsFD := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), syscall.Gettid())
	netnsFD := fmt.Sprintf("/proc/%d/ns/net", pid)
	result, err := libcni.Setup(fmt.Sprintf("%d", pid), netnsFD)
	if err != nil {
		logrus.Fatalf("setting up netns for id (%d) and netns (%s) failed: %v", pid, netnsFD, err)
	}

	// Get the IP of the default interface.
	defaultInterface := cni.DefaultPrefix + "0"
	ip := result.Interfaces[defaultInterface].IPConfigs[0].IP.String()
	fmt.Printf("IP of the default interface %s:%s", defaultInterface, ip)

	// Switch back to the original netns.
	if err := netns.Set(originalNS); err != nil {
		logrus.Fatalf("switching back to original netns failed: %v", err)
	}
}
