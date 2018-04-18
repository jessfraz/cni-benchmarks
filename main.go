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

	b, err := newCNIBenchmark()
	if err != nil {
		logrus.Fatal(err)
	}
	defer b.originalNS.Close()

	// Walk the configuration directory and get all the configs.
	plugins := []string{}
	if err := filepath.Walk(b.pluginConfDir, func(p string, f os.FileInfo, err error) error {
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
		logrus.Fatalf("walking plugin configuration directory %s failed: %v", b.pluginConfDir, err)
	}
	logrus.Infof("Found plugin configurations for %s", strings.Join(plugins, ", "))

	logrus.Infof("Parent process ($this) has PID %d", os.Getpid())

	// Iterate over the plugin configurations.
	for _, plugin := range plugins {
		logrus.WithFields(logrus.Fields{"plugin": plugin}).Info("creating new netns process")

		if err := b.createNetwork(plugin); err != nil {
			logrus.WithFields(logrus.Fields{"plugin": plugin}).Error(err)
		}
	}
}

type benchmarkCNI struct {
	originalNS    netns.NsHandle
	libcni        cni.CNI
	pluginConfDir string
}

func newCNIBenchmark() (*benchmarkCNI, error) {
	// Save the current network namespace.
	logrus.Infof("Getting current netns")
	originalNS, err := netns.Get()
	if err != nil {
		return nil, fmt.Errorf("getting current netns failed: %v", err)
	}

	// Initialize CNI library.
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("getting working directory failed: %v", err)
	}
	pluginConfDir := filepath.Join(wd, "net.d")
	pluginDirs := []string{cniBinDir, cni.DefaultCNIDir}
	logrus.Infof("Initializing new CNI library instance with configuration directory %s and plugin directories %s", pluginConfDir, strings.Join(pluginDirs, ", "))
	libcni, err := cni.New(
		cni.WithMinNetworkCount(2),
		cni.WithPluginConfDir(pluginConfDir),
		cni.WithPluginDir(pluginDirs),
	)
	if err != nil {
		return nil, fmt.Errorf("creating new CNI instance failed: %v", err)
	}

	return &benchmarkCNI{
		originalNS:    originalNS,
		libcni:        libcni,
		pluginConfDir: pluginConfDir,
	}, nil
}

func (b benchmarkCNI) createNetwork(plugin string) error {
	// Switch back to the original netns.
	defer netns.Set(b.originalNS)

	// Create a new network namespace.
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNET}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unsharing command failed: %v", err)
	}
	defer cmd.Process.Kill()
	pid := cmd.Process.Pid

	logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("netns process has PID %d", pid)

	// Load the CNI configuration.
	if err := b.libcni.Load(
		cni.WithLoNetwork,
		cni.WithConfFile(filepath.Join(b.pluginConfDir, plugin+".conf")),
	); err != nil {
		return fmt.Errorf("loading CNI configuration failed: %v", err)
	}

	// Setup network for namespace.
	//netnsFD := fmt.Sprintf("/proc/%d/task/%d/ns/net", os.Getpid(), syscall.Gettid())
	netnsFD := fmt.Sprintf("/proc/%d/ns/net", pid)
	result, err := b.libcni.Setup(fmt.Sprintf("%d", pid), netnsFD)
	if err != nil {
		return fmt.Errorf("setting up netns for id (%d) and netns (%s) failed: %v", pid, netnsFD, err)
	}

	// Get the IP of the default interface.
	defaultInterface := cni.DefaultPrefix + "0"
	ip := result.Interfaces[defaultInterface].IPConfigs[0].IP.String()
	logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("IP of the default interface (%s) in the netns is %s", defaultInterface, ip)

	logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("getting netns file descriptor from the pid %d", pid)
	newNS, err := netns.GetFromPid(pid)
	if err != nil {
		return fmt.Errorf("creating new netns failed: %v", err)
	}
	defer newNS.Close()

	// Switch into the new netns.
	logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("[performing setns into netns from pid %d", pid)
	if err := netns.Set(newNS); err != nil {
		return fmt.Errorf("switching to new netns failed: %v", err)
	}

	// Get a list of the links.
	links, err := netlink.LinkList()
	if err != nil {
		return fmt.Errorf("getting list of ip links failed: %v", err)
	}
	l := []string{}
	for _, link := range links {
		l = append(l, fmt.Sprintf("%s->%s", link.Type(), link.Attrs().Name))
	}
	if len(l) > 0 {
		logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("found netns ip links: %s", strings.Join(l, ", "))
	}

	// Try getting an outbound resource.
	resp, err := http.Get("https://httpbin.org/ip")
	if err != nil {
		return fmt.Errorf("getting an out of network resource failed: %v", err)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body failed: %v", err)
	}
	logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof("httpbin returned: %s", strings.Replace(strings.Replace(strings.TrimSpace(string(body)), "\n", "", -1), " ", "", -1))

	return nil
}
