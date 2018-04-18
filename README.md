# cni-benchmarks

[![Travis CI](https://travis-ci.org/jessfraz/cni-benchmarks.svg?branch=master)](https://travis-ci.org/jessfraz/cni-benchmarks)

## Running

Running the benchmarks is just done with go.
You will need to use sudo since it requires creating network namespaces.

```console
$ sudo go test -bench=.
goos: linux
goarch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                 2        1188596449 ns/op
BenchmarkCreateNetworkIPvlan-8                 1        1154609347 ns/op
BenchmarkCreateNetworkMacvlan-8                1        1018058236 ns/op
BenchmarkCreateNetworkPTP-8                    1        1111937856 ns/op
PASS
ok      github.com/jessfraz/cni-benchmarks      6.524s

$ sudo go test -bench=. -benchtime=20s
goos: linux
goarch: amd64
pkg: github.com/jessfraz/cni-benchmarks
BenchmarkCreateNetworkBridge-8                30        1365005810 ns/op
BenchmarkCreateNetworkIPvlan-8                30        1128743945 ns/op
BenchmarkCreateNetworkMacvlan-8               20        1146909318 ns/op
BenchmarkCreateNetworkPTP-8                   20        1185070353 ns/op
PASS
ok      github.com/jessfraz/cni-benchmarks      125.870s
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
