package main

import (
	"flag"
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
	"github.com/jessfraz/cni-benchmarks/version"
	"github.com/sirupsen/logrus"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netns"
)

const (
	// BANNER is what is printed for help/info output
	BANNER = `cni-benchmarks

 version: %s

`
)

var (
	debug bool
	vrsn  bool
)

func init() {
	flag.BoolVar(&vrsn, "version", false, "print version and exit")
	flag.BoolVar(&vrsn, "v", false, "print version and exit (shorthand)")
	flag.BoolVar(&debug, "d", false, "run in debug mode")

	flag.Usage = func() {
		fmt.Fprint(os.Stderr, fmt.Sprintf(BANNER, version.VERSION))
		flag.PrintDefaults()
	}

	flag.Parse()

	if vrsn {
		fmt.Printf("cni-benchmarks version %s, build %s", version.VERSION, version.GITCOMMIT)
		os.Exit(0)
	}

	// set log level
	if debug {
		logrus.SetLevel(logrus.DebugLevel)
	}
}

func main() {
	// Lock the OS Thread so we don't accidentally switch namespaces.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	b, err := newCNIBenchmark(true)
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
	doLog         bool
	process       *os.Process
	netnsFD       string
	nsHandle      netns.NsHandle
}

func newCNIBenchmark(doLog bool) (*benchmarkCNI, error) {
	// Save the current network namespace.
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
	binDir := filepath.Join(wd, "bin")
	pluginDirs := []string{binDir, cni.DefaultCNIDir}
	logrus.Debugf("Initializing new CNI library instance with configuration directory %s and plugin directories %s", pluginConfDir, strings.Join(pluginDirs, ", "))
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
		doLog:         doLog,
	}, nil
}

func (b *benchmarkCNI) createNetwork(plugin string) error {
	if err := b.createProcess(plugin); err != nil {
		return err
	}
	defer b.process.Kill()

	if err := b.loadCNIConfig(plugin); err != nil {
		return err
	}

	result, err := b.setupNetNS()
	if err != nil {
		return err
	}
	defer b.libcni.Remove(fmt.Sprintf("%d", b.process.Pid), b.netnsFD)
	defer b.nsHandle.Close()

	// Get the IP of the default interface.
	defaultInterface := cni.DefaultPrefix + "0"
	ip := result.Interfaces[defaultInterface].IPConfigs[0].IP.String()
	b.log(plugin, "IP of the default interface (%s) in the netns is %s", defaultInterface, ip)

	// Switch into the new netns.
	b.log(plugin, "performing setns into netns from pid %d", b.process.Pid)
	if err := b.setNS(); err != nil {
		return err
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
		b.log(plugin, "found netns ip links: %s", strings.Join(l, ", "))
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
	b.log(plugin, "httpbin returned: %s", strings.Replace(strings.Replace(strings.TrimSpace(string(body)), "\n", "", -1), " ", "", -1))

	if err := netns.Set(b.originalNS); err != nil {
		return fmt.Errorf("returning to original namespace failed: %v", err)
	}

	return nil
}

func (b *benchmarkCNI) createProcess(plugin string) error {
	// Create a process in a new network namespace.
	cmd := exec.Command("sleep", "30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Unshareflags: syscall.CLONE_NEWNET}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("unsharing command failed: %v", err)
	}
	b.process = cmd.Process
	b.netnsFD = fmt.Sprintf("/proc/%d/ns/net", cmd.Process.Pid)

	newNS, err := netns.GetFromPid(cmd.Process.Pid)
	if err != nil {
		return fmt.Errorf("creating new netns failed: %v", err)
	}
	b.nsHandle = newNS

	b.log(plugin, "netns process has PID %d", cmd.Process.Pid)

	return nil
}

func (b *benchmarkCNI) loadCNIConfig(plugin string) error {
	// Load the CNI configuration.
	if err := b.libcni.Load(
		cni.WithLoNetwork,
		cni.WithConfFile(filepath.Join(b.pluginConfDir, plugin+".conf")),
	); err != nil {
		return fmt.Errorf("loading CNI configuration failed: %v", err)
	}

	return nil
}

func (b *benchmarkCNI) setupNetNS() (*cni.CNIResult, error) {
	// Setup network for namespace.
	result, err := b.libcni.Setup(fmt.Sprintf("%d", b.process.Pid), b.netnsFD)
	if err != nil {
		return nil, fmt.Errorf("setting up netns for id (%d) and netns (%s) failed: %v", b.process.Pid, b.netnsFD, err)
	}

	return result, nil
}

func (b *benchmarkCNI) setNS() error {
	if err := netns.Set(b.nsHandle); err != nil {
		return fmt.Errorf("switching to new netns failed: %v", err)
	}

	return nil
}

func (b *benchmarkCNI) log(plugin, fmt string, args ...interface{}) {
	if b.doLog {
		logrus.WithFields(logrus.Fields{"plugin": plugin}).Infof(fmt, args...)
	}
}
