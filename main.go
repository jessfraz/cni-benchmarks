package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
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

	// Walk the configuration directory and get all the configs.
	plugins := []string{}
	if err := filepath.Walk(pluginConfDir, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if f.IsDir() {
			// Skip directories.
			return nil
		}

		plugins = append(plugins, filepath.Base(strings.TrimSuffix(p, ".conf")))
		return nil
	}); err != nil {
		logrus.Fatalf("walking plugin configuration directory %s failed: %v", pluginConfDir, err)
	}
	logrus.Infof("Found plugin configurations for %s", strings.Join(plugins, ", "))

	logrus.Infof("Parent process ($this) has PID %d", os.Getpid())

	// Iterate over the plugin configurations.
	for _, plugin := range plugins {
		// Create a new network namespace.
		logrus.Infof("[%s] creating new netns process...", plugin)
		cmd := exec.Command("sleep", "30")
		cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNET}
		if err := cmd.Start(); err != nil {
			logrus.Fatalf("[%s] unsharing command failed: %v", plugin, err)
		}
		pid := cmd.Process.Pid

		logrus.Infof("[%s] new netns process has PID %d", plugin, pid)

		// Load the CNI configuration.
		if err := libcni.Load(
			cni.WithLoNetwork,
			cni.WithConfFile(filepath.Join(wd, "net.d", plugin+".conf")),
		); err != nil {
			logrus.Fatalf("[%s] loading CNI configuration failed: %v", plugin, err)
		}

		// Setup network for namespace.
		//netnsFD := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), syscall.Gettid())
		netnsFD := fmt.Sprintf("/proc/%d/ns/net", pid)
		result, err := libcni.Setup(fmt.Sprintf("%d", pid), netnsFD)
		if err != nil {
			logrus.Fatalf("[%s] setting up netns for id (%d) and netns (%s) failed: %v", plugin, pid, netnsFD, err)
		}

		// Get the IP of the default interface.
		defaultInterface := cni.DefaultPrefix + "0"
		ip := result.Interfaces[defaultInterface].IPConfigs[0].IP.String()
		logrus.Infof("[%s] IP of the default interface (%s) in the netns is %s", plugin, defaultInterface, ip)

		logrus.Infof("[%s] getting netns file descriptor from the pid %d", plugin, pid)
		newNS, err := netns.GetFromPid(pid)
		if err != nil {
			logrus.Fatalf("[%s] creating new netns failed: %v", plugin, err)
		}
		defer newNS.Close()

		// Switch into the new netns.
		logrus.Infof("[%s] performing setns into netns from pid %d", plugin, pid)
		if err := netns.Set(newNS); err != nil {
			logrus.Fatalf("[%s] switching to new netns failed: %v", plugin, err)
		}

		// Get a list of the links.
		links, err := netlink.LinkList()
		if err != nil {
			logrus.Fatalf("[%s] getting list of ip links failed: %v", plugin, err)
		}
		l := []string{}
		for _, link := range links {
			l = append(l, fmt.Sprintf("%s->%s", link.Type(), link.Attrs().Name))
		}
		if len(l) > 0 {
			logrus.Infof("[%s] found netns ip links: %s", plugin, strings.Join(l, ", "))
		}

		// Try getting an outbound resource.
		resp, err := http.Get("https://httpbin.org/ip")
		if err != nil {
			logrus.Fatalf("[%s] getting an out of network resource failed: %v", plugin, err)
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			logrus.Fatalf("[%s] reading response body failed: %v", plugin, err)
		}
		logrus.Infof("[%s] httpbin returned: %s", plugin, strings.Replace(strings.TrimSpace(string(body)), "\n", "", -1))

		// Switch back to the original netns.
		logrus.Info("Switching back into our original netns...")
		if err := netns.Set(originalNS); err != nil {
			logrus.Fatalf("[%s] switching back to original netns failed: %v", plugin, err)
		}
	}
}
