# cni-benchmarks

[![Travis CI](https://travis-ci.org/jessfraz/cni-benchmarks.svg?branch=master)](https://travis-ci.org/jessfraz/cni-benchmarks)

## What this does...

The `main.go` resulting binary loads all the cni plugin configurations from
[`net.d`](net.d) performs the following on each:

1. Unshares a new network namespace with a `sleep` process.
2. Sets up networking for the process via the specific plugin passed.
3. Enters the network namespace and calls get to `https://httpbin/ip` 
    just to make sure network works.
4.  Returns to the original namespace. Kills the process and cleans up the
    network.

## Running

Running the benchmarks is just done with go.
You will need to use `sudo` since it requires creating network namespaces.

**Setup**

Before testing the cilium, calico, and weave plugins you will want to run the
following command which will start etcd, calico, cilium, and weave containers:

```
$ make run-containers
```

**Running the benchmarks**

```console
$ sudo go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                 1        1103701882 ns/op
BenchmarkCreateNetworkCalico-8                 1        11406779278 ns/op
BenchmarkCreateNetworkCilium-8                 1        2974889818 ns/op
BenchmarkCreateNetworkIPvlan-8                 2        1245111887 ns/op
BenchmarkCreateNetworkMacvlan-8                1        1333958217 ns/op
BenchmarkCreateNetworkPTP-8                    1        1262308289 ns/op
BenchmarkCreateNetworkWeave-8                  1        31650053959 ns/op
PASS
ok      github.com/jessfraz/cni-benchmarks      21.221s

$ sudo go test -bench=. -benchtime=20s
goos: linux
arch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                30        1574208392  ns/op
BenchmarkCreateNetworkCalico-8                 2        11603167194 ns/op
BenchmarkCreateNetworkCilium-8                10        2026048715  ns/op
BenchmarkCreateNetworkIPvlan-8                50        1189171868  ns/op
BenchmarkCreateNetworkMacvlan-8               20        1176936944  ns/op
BenchmarkCreateNetworkPTP-8                   20        1315428103  ns/op
BenchmarkCreateNetworkWeave-8                  1        31601342289 ns/op
PASS
```

The `main.go` program just runs all the plugins.

```console
$ make

# You have to sudo the resulting binary since it creates new network
# namespaces.
$ sudo ./cni-benchmarks
INFO[0000] Getting current netns                        
INFO[0000] Initializing new CNI library instance with configuration directory /home/jessie/.go/src/github.com/jessfraz/cni-benchmarks/net.d and plugin directories /home/jessie/.go/src/github.com/containernetworking/plugins/bin, /opt/cni/bin 
INFO[0000] Found plugin configurations for bridge, ipvlan, macvlan, ptp 
INFO[0000] Parent process ($this) has PID 16824         
INFO[0000] creating new netns process                    plugin=bridge
INFO[0000] netns process has PID 16830                   plugin=bridge
INFO[0000] IP of the default interface (eth0) in the netns is 10.10.0.2  plugin=bridge
INFO[0000] getting netns file descriptor from the pid 16830  plugin=bridge
INFO[0000] [performing setns into netns from pid 16830   plugin=bridge
INFO[0000] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=bridge
INFO[0000] httpbin returned: {"origin":"69.203.154.19"}  plugin=bridge
INFO[0000] creating new netns process                    plugin=ipvlan
INFO[0001] netns process has PID 16904                   plugin=ipvlan
INFO[0001] IP of the default interface (eth0) in the netns is 10.1.2.2  plugin=ipvlan
INFO[0001] getting netns file descriptor from the pid 16904  plugin=ipvlan
INFO[0001] [performing setns into netns from pid 16904   plugin=ipvlan
INFO[0001] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, ipvlan->eth0  plugin=ipvlan
INFO[0001] httpbin returned: {"origin":"69.203.154.19"}  plugin=ipvlan
INFO[0001] creating new netns process                    plugin=macvlan
INFO[0002] netns process has PID 16940                   plugin=macvlan
INFO[0002] IP of the default interface (eth0) in the netns is 20.0.0.2  plugin=macvlan
INFO[0002] getting netns file descriptor from the pid 16940  plugin=macvlan
INFO[0002] [performing setns into netns from pid 16940   plugin=macvlan
INFO[0002] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, macvlan->eth0  plugin=macvlan
INFO[0002] httpbin returned: {"origin":"69.203.154.19"}  plugin=macvlan
INFO[0002] creating new netns process                    plugin=ptp
INFO[0003] netns process has PID 16970                   plugin=ptp
INFO[0003] IP of the default interface (eth0) in the netns is 10.1.1.2  plugin=ptp
INFO[0003] getting netns file descriptor from the pid 16970  plugin=ptp
INFO[0003] [performing setns into netns from pid 16970   plugin=ptp
INFO[0003] found netns ip links: device->lo, ipip->tunl0, ip6gre->gre0, ip6gretap->gretap0, erspan->erspan0, vti->ip_vti0, vti6->ip6_vti0, sit->sit0, ip6tnl->ip6tnl0, ip6gre->ip6gre0, veth->eth0  plugin=ptp
INFO[0003] httpbin returned: {"origin":"69.203.154.19"}  plugin=ptp
```
