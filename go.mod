module github.com/armon/go-socks5

go 1.16

require (
	golang.org/x/net v0.0.0-20210908191846-a5e095526f91
	sigs.k8s.io/apiserver-network-proxy v0.0.24
)

replace sigs.k8s.io/apiserver-network-proxy/konnectivity-client => sigs.k8s.io/apiserver-network-proxy/konnectivity-client v0.0.24
