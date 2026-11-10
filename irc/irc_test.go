package irc

import (
	"crypto/tls"
	"testing"
)

// Guards the workaround documented on tlsConfig. Several irc.irchighway.net
// servers accept nothing but RSA key exchange suites, which are no longer in
// Go's client defaults, and dropping this one turns those servers back into a
// bare EOF at dial time.
func TestTLSConfigOffersRSAKeyExchange(t *testing.T) {
	cfg := tlsConfig()

	for _, id := range cfg.CipherSuites {
		if id == tls.TLS_RSA_WITH_AES_128_CBC_SHA256 {
			return
		}
	}

	t.Error("TLS_RSA_WITH_AES_128_CBC_SHA256 is not offered; TLS 1.2 only IRCHighway servers will fail to handshake")
}

// The cipher list must not narrow what the toolchain would otherwise offer.
func TestTLSConfigKeepsDefaultCipherSuites(t *testing.T) {
	offered := make(map[uint16]bool)
	for _, id := range tlsConfig().CipherSuites {
		offered[id] = true
	}

	for _, suite := range tls.CipherSuites() {
		if !offered[suite.ID] {
			t.Errorf("default cipher suite %s is missing from the offered list", suite.Name)
		}
	}
}
