package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"

	cni "github.com/containerd/go-cni"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
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
	logrus.Infof("Getting current netns...")
	originalNS, err := netns.Get()
	if err != nil {
		logrus.Fatalf("getting current netns failed: %v", err)
	}
	defer originalNS.Close()

	// Initialize CNI library.
	wd, err := os.Getwd()
	if err != nil {
		logrus.Fatalf("getting working directory failed: %v", err)
	}
	pluginConfDir := filepath.Join(wd, "net.d")
	pluginDirs := []string{cniBinDir, cni.DefaultCNIDir}
	logrus.Infof("Initializing new CNI library instance with configuration directory %s and plugin directories %s...", pluginConfDir, strings.Join(pluginDirs, ", "))
	libcni, err := cni.New(
		cni.WithMinNetworkCount(2),
		cni.WithPluginConfDir(pluginConfDir),
		cni.WithPluginDir(pluginDirs),
	)
	if err != nil {
		logrus.Fatalf("creating new CNI instance failed: %v", err)
	}

	// Create a new network namespace.
	logrus.Info("Creating new netns process...")
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNET}
	if err := cmd.Start(); err != nil {
		logrus.Fatalf("unsharing command failed: %v", err)
	}
	pid := cmd.Process.Pid

	logrus.Infof("Parent process ($this) has PID %d", os.Getpid())
	logrus.Infof("New netns process has PID %d", pid)

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
	logrus.Infof("IP of the default interface in the netns is %s:%s", defaultInterface, ip)

	logrus.Infof("Getting netns file descriptor from the pid %d", pid)
	newNS, err := netns.GetFromPid(pid)
	if err != nil {
		logrus.Fatalf("creating new netns failed: %v", err)
	}
	defer newNS.Close()

	// Switch into the new netns.
	logrus.Infof("Performing setns into netns from pid %d", pid)
	if err := netns.Set(newNS); err != nil {
		logrus.Fatalf("switching to new netns failed: %v", err)
	}

	// Get a list of the links.
	links, err := netlink.LinkList()
	if err != nil {
		logrus.Fatalf("getting list of ip links failed: %v", err)
	}
	l := []string{}
	for _, link := range links {
		l = append(l, link.Type())
	}
	if len(l) > 0 {
		logrus.Infof("Found netns ip links: %s", strings.Join(l, ", "))
	}

	// Switch back to the original netns.
	logrus.Info("Switching back into our original netns...")
	if err := netns.Set(originalNS); err != nil {
		logrus.Fatalf("switching back to original netns failed: %v", err)
	}
}
