# cni-benchmarks

[![Travis CI](https://travis-ci.org/jessfraz/cni-benchmarks.svg?branch=master)](https://travis-ci.org/jessfraz/cni-benchmarks)

## What this does...

The `main.go` resulting binary loads all the cni plugin configurations from
[`net.d`](net.d) performs the following on each:

1. Unshares a new network namespace with a `sleep` process.
2. Sets up networking for the process via the specific plugin passed.
3. Enters the network namespace and calls get to `https://httpbin.org/ip` 
    just to make sure network works.
4.  Returns to the original namespace. Kills the process and cleans up the
    network.

## Running

Running the benchmarks is just done with go.
You will need to use `sudo` since it requires creating network namespaces.

### Setup

Before testing the cilium, calico, flannel, and weave plugins you will want to run the
following command which will start etcd, calico, cilium, and weave containers:

```
$ make run-containers
```

**NOTE:** both cilium and flannel use vxlan devices so you cannot run both at
the same time. You will need to test those separately.

### Running the benchmarks

```console
$ sudo go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                 1        1139952630 ns/op
BenchmarkCreateNetworkCalico-8                 1        1200416607 ns/op
BenchmarkCreateNetworkCilium-8                 1        1329559748 ns/op
BenchmarkCreateNetworkFlannel-8                1        1207025536 ns/op
BenchmarkCreateNetworkIPvlan-8                 2        1135102997 ns/op
BenchmarkCreateNetworkMacvlan-8                1        1138442676 ns/op
BenchmarkCreateNetworkPTP-8                    1        1258083693 ns/op
BenchmarkCreateNetworkWeave-8                  1        1557366751 ns/op
PASS
ok      github.com/jessfraz/cni-benchmarks      10.514s


$ sudo go test -bench=. -benchtime=20s
goos: linux
goarch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                30        1540904853 ns/op
BenchmarkCreateNetworkCalico-8                20        1280029392 ns/op
BenchmarkCreateNetworkCilium-8                20        1347408329 ns/op
BenchmarkCreateNetworkFlannel-8               30        1219560662 ns/op
BenchmarkCreateNetworkIPvlan-8                50        1196531970 ns/op
BenchmarkCreateNetworkMacvlan-8               20        1238331945 ns/op
BenchmarkCreateNetworkPTP-8                   20        1319141412 ns/op
BenchmarkCreateNetworkWeave-8                 20        1467203727 ns/op
PASS
ok      github.com/jessfraz/cni-benchmarks      248.320s
```

### Running the main program

The `main.go` program just runs all the plugins.

```console
$ make

# You have to sudo the resulting binary since it creates new network
# namespaces.
$ sudo ./cni-benchmarks
INFO[0000] Found plugin configurations for bridge, calico, cilium, ipvlan, macvlan, ptp, weave 
INFO[0000] Parent process ($this) has PID 1837          
INFO[0000] creating new netns process                    plugin=bridge
INFO[0000] netns process has PID 1842                    plugin=bridge
INFO[0000] IP of the default interface (eth0) in the netns is 10.10.0.111  plugin=bridge
INFO[0000] getting netns file descriptor from the pid 1842  plugin=bridge
INFO[0000] [performing setns into netns from pid 1842    plugin=bridge
INFO[0000] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=bridge
INFO[0000] httpbin returned: {"origin":"69.203.154.19"}  plugin=bridge
INFO[0001] creating new netns process                    plugin=calico
INFO[0002] netns process has PID 1962                    plugin=calico
Calico CNI IPAM request count IPv4=1 IPv6=0
Calico CNI IPAM handle=calico-benchmark.1962
Calico CNI IPAM assigned addresses IPv4=[192.168.245.204] IPv6=[]
Calico CNI using IPs: [192.168.245.204/32]
INFO[0002] IP of the default interface (eth0) in the netns is 192.168.245.204  plugin=calico
INFO[0002] getting netns file descriptor from the pid 1962  plugin=calico
INFO[0002] [performing setns into netns from pid 1962    plugin=calico
INFO[0002] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=calico
INFO[0002] httpbin returned: {"origin":"69.203.154.19"}  plugin=calico
INFO[0012] creating new netns process                    plugin=cilium
INFO[0014] netns process has PID 2109                    plugin=cilium
level=debug msg="Processing CNI ADD request" args="&{2109 /proc/2109/ns/net eth0  /home/jessie/.go/src/github.com/jessfraz/cni-benchmarks/bin:/opt/cni/bin [123 34 99 110 105 86 101 114 115 105 111 110 34 58 34 34 44 34 109 116 117 34 58 49 52 53 48 44 34 110 97 109 101 34 58 34 99 105 108 105 117 109 34 44 34 116 121 112 101 34 58 34 99 105 108 105 117 109 45 99 110 105 34 125]}"
level=debug msg="Created veth pair" vethPair="[tmp2109 lxc2109]"
level=debug msg="Configuring link" interface=eth0 ipAddr=10.25.220.6 netLink="&{LinkAttrs:{Index:1422 MTU:1450 TxQLen:0 Name:eth0 HardwareAddr:46:e5:08:3c:25:56 Flags:broadcast|multicast RawFlags:4098 ParentIndex:1423 MasterIndex:0 Namespace:<nil> Alias: Statistics:0xc42159c0e8 Promisc:0 Xdp:0xc4215762e0 EncapType:ether Protinfo:<nil> OperState:down} PeerName:}"
level=debug msg="Adding route" route="{Prefix:{IP:10.25.28.238 Mask:ffffffff} Nexthop:<nil>}"
level=debug msg="Adding route" route="{Prefix:{IP:0.0.0.0 Mask:00000000} Nexthop:10.25.28.238}"
level=debug msg="Configuring link" interface=eth0 ipAddr="f00d::a19:0:0:815b" netLink="&{LinkAttrs:{Index:1422 MTU:1450 TxQLen:0 Name:eth0 HardwareAddr:46:e5:08:3c:25:56 Flags:broadcast|multicast RawFlags:4098 ParentIndex:1423 MasterIndex:0 Namespace:<nil> Alias: Statistics:0xc42159c0e8 Promisc:0 Xdp:0xc4215762e0 EncapType:ether Protinfo:<nil> OperState:down} PeerName:}"
level=debug msg="Adding route" route="{Prefix:{IP:f00d::a19:0:0:8ad6 Mask:ffffffffffffffffffffffffffffffff} Nexthop:<nil>}"
level=debug msg="Adding route" route="{Prefix:{IP::: Mask:00000000000000000000000000000000} Nexthop:f00d::a19:0:0:8ad6}"
INFO[0014] IP of the default interface (eth0) in the netns is 10.25.220.6  plugin=cilium
INFO[0014] getting netns file descriptor from the pid 2109  plugin=cilium
INFO[0014] [performing setns into netns from pid 2109    plugin=cilium
INFO[0014] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=cilium
INFO[0014] httpbin returned: {"origin":"69.203.154.19"}  plugin=cilium
level=debug msg="Processing CNI DEL request" args="&{2109 /proc/2109/ns/net eth0  /home/jessie/.go/src/github.com/jessfraz/cni-benchmarks/bin:/opt/cni/bin [123 34 99 110 105 86 101 114 115 105 111 110 34 58 34 34 44 34 109 116 117 34 58 49 52 53 48 44 34 110 97 109 101 34 58 34 99 105 108 105 117 109 34 44 34 116 121 112 101 34 58 34 99 105 108 105 117 109 45 99 110 105 34 125]}"
INFO[0004] creating new netns process                    plugin=flannel
INFO[0004] netns process has PID 9897                    plugin=flannel
INFO[0004] IP of the default interface (eth0) in the netns is 10.6.50.2  plugin=flannel
INFO[0004] getting netns file descriptor from the pid 9897  plugin=flannel
INFO[0004] [performing setns into netns from pid 9897    plugin=flannel
INFO[0004] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, ipvlan->eth0  plugin=flannel
INFO[0004] httpbin returned: {"origin":"69.203.154.19"}  plugin=flannel
INFO[0014] creating new netns process                    plugin=ipvlan
INFO[0015] netns process has PID 2235                    plugin=ipvlan
INFO[0015] IP of the default interface (eth0) in the netns is 10.1.2.163  plugin=ipvlan
INFO[0015] getting netns file descriptor from the pid 2235  plugin=ipvlan
INFO[0015] [performing setns into netns from pid 2235    plugin=ipvlan
INFO[0015] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, ipvlan->eth0  plugin=ipvlan
INFO[0015] httpbin returned: {"origin":"69.203.154.19"}  plugin=ipvlan
INFO[0015] creating new netns process                    plugin=macvlan
INFO[0016] netns process has PID 2287                    plugin=macvlan
INFO[0016] IP of the default interface (eth0) in the netns is 20.0.0.101  plugin=macvlan
INFO[0016] getting netns file descriptor from the pid 2287  plugin=macvlan
INFO[0016] [performing setns into netns from pid 2287    plugin=macvlan
INFO[0016] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, macvlan->eth0  plugin=macvlan
INFO[0016] httpbin returned: {"origin":"69.203.154.19"}  plugin=macvlan
INFO[0016] creating new netns process                    plugin=ptp
INFO[0017] netns process has PID 2337                    plugin=ptp
INFO[0017] IP of the default interface (eth0) in the netns is 10.1.1.89  plugin=ptp
INFO[0017] getting netns file descriptor from the pid 2337  plugin=ptp
INFO[0017] [performing setns into netns from pid 2337    plugin=ptp
INFO[0017] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=ptp
INFO[0017] httpbin returned: {"origin":"69.203.154.19"}  plugin=ptp
INFO[0017] creating new netns process                    plugin=weave
INFO[0018] netns process has PID 2423                    plugin=weave
INFO[0019] IP of the default interface (eth0) in the netns is 10.32.0.1  plugin=weave
INFO[0019] getting netns file descriptor from the pid 2423  plugin=weave
INFO[0019] [performing setns into netns from pid 2423    plugin=weave
INFO[0019] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=weave
INFO[0019] httpbin returned: {"origin":"69.203.154.19"}  plugin=weave
```

## Using the Makefile to update the CNI binaries, etc

```console
all                            Runs a clean, build, fmt, lint, test, staticcheck, vet and install
build                          Builds a dynamic executable or package
bump-version                   Bump the version in the version file. Set BUMP to [ patch | major | minor ]
clean                          Cleanup any build binaries or packages
cover                          Runs go test with coverage
cross                          Builds the cross-compiled binaries, creating a clean directory structure (eg. GOOS/GOARCH/binary)
fmt                            Verifies all files have men `gofmt`ed
install                        Installs the executable or package
lint                           Verifies `golint` passes
release                        Builds the cross-compiled binaries, naming them in such a way for release (eg. binary-GOOS-GOARCH)
run-calico                     Run calico in a container for testing calico against.
run-cilium                     Run cilium in a container for testing cilium against.
run-containers                 Runs the etcd, calico, cilium, flannel, and weave containers.
run-etcd                       Run etcd in a container for testing calico and cilium against.
run-flannel                    Run flannel in a container for testing flannel against.
run-weave                      Run weave in a container for testing weave against.
static                         Builds a static executable
staticcheck                    Verifies `staticcheck` passes
stop-containers                Stops all the running containers.
tag                            Create a new git tag to prepare to build a release
test                           Runs the go tests
update-binaries                Run the dev dockerfile which builds all the cni binaries for testing.
vet                            Verifies `go vet` passes
```
