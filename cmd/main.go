package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"

	socks5 "github.com/armon/go-socks5"
	"sigs.k8s.io/apiserver-network-proxy/pkg/util"
)

func main() {
	fmt.Println("Starting proxy...")
	// Create a SOCKS5 server
	conf := &socks5.Config{
		Dial: dialKonnectivity,
	}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", ":8080"); err != nil {
		panic(err)
	}
}

func dialKonnectivity(ctx context.Context, network, addr string) (net.Conn, error) {
	caCert := "/tmp/kc-ca.crt"
	clientCert := "/tmp/kc-tls.crt"
	clientKey := "/tmp/kc-tls.key"
	proxyHost := "konnectivity-server-local"
	proxyPort := 8090
	tlsConfig, err := util.GetClientTLSConfig(caCert, clientCert, clientKey, proxyHost, nil)
	if err != nil {
		return nil, err
	}
	var proxyConn net.Conn

	proxyAddress := fmt.Sprintf("%s:%d", proxyHost, proxyPort)
	requestAddress := addr

	proxyConn, err = tls.Dial("tcp", proxyAddress, tlsConfig)
	if err != nil {
		return nil, fmt.Errorf("dialing proxy %q failed: %v", proxyAddress, err)
	}
	fmt.Fprintf(proxyConn, "CONNECT %s HTTP/1.1\r\nHost: %s\r\n\r\n", requestAddress, "127.0.0.1")
	br := bufio.NewReader(proxyConn)
	res, err := http.ReadResponse(br, nil)
	if err != nil {
		return nil, fmt.Errorf("reading HTTP response from CONNECT to %s via proxy %s failed: %v",
			requestAddress, proxyAddress, err)
	}
	if res.StatusCode != 200 {
		return nil, fmt.Errorf("proxy error from %s while dialing %s: %v", proxyAddress, requestAddress, res.Status)
	}

	// It's safe to discard the bufio.Reader here and return the
	// original TCP conn directly because we only use this for
	// TLS, and in TLS the client speaks first, so we know there's
	// no unbuffered data. But we can double-check.
	if br.Buffered() > 0 {
		return nil, fmt.Errorf("unexpected %d bytes of buffered data from CONNECT proxy %q",
			br.Buffered(), proxyAddress)
	}
	return proxyConn, nil
}
