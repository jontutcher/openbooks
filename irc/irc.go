package irc

import (
	"crypto/tls"
	"net"
)

// Conn represents an IRC connection to a server
type Conn struct {
	net.Conn
	channel  string
	Username string
	realname string
}

// New creates a new IRC connection to the server using the supplied username and realname
func New(username, realname string) *Conn {
	irc := &Conn{
		channel:  "",
		Username: username,
		realname: realname,
	}

	return irc
}

// Connect connects to the given server at port 6667
func (i *Conn) Connect(address string, enableTLS bool) error {
	var conn net.Conn
	var err error
	if enableTLS {
		conn, err = tls.Dial("tcp", address, tlsConfig())
	} else {
		conn, err = net.Dial("tcp", address)
	}

	if err != nil {
		return err
	}

	i.Conn = conn

	user := "USER " + i.Username + " 0 * :" + i.Username + "\r\n"
	nick := "NICK " + i.Username + "\r\n"

	i.Write([]byte(user))
	i.Write([]byte(nick))
	return nil
}

// tlsConfig is the TLS client configuration used for IRC connections.
//
// Certificate verification is off because IRCHighway presents a self-signed
// certificate.
//
// The cipher list is the toolchain default plus TLS_RSA_WITH_AES_128_CBC_SHA256.
// Go removed the RSA key exchange suites from its client defaults in 1.22 and
// deleted the tlsrsakex GODEBUG that re-enabled them in 1.27, but several
// irc.irchighway.net servers still run InspIRCd 2.0 and offer nothing except
// TLS_RSA_WITH_AES_128_CBC_SHA256 and TLS_RSA_WITH_AES_256_CBC_SHA256 over
// TLS 1.2. Go has never implemented the 256-bit one, so the 128-bit suite is
// the only overlap available. irc.irchighway.net is a rotating pool with a
// two minute TTL, so without this the connection fails with a bare EOF
// whenever DNS happens to return one of those servers. CipherSuites is
// ignored for TLS 1.3, so servers that support it still negotiate it.
func tlsConfig() *tls.Config {
	suites := make([]uint16, 0, len(tls.CipherSuites())+1)
	for _, suite := range tls.CipherSuites() {
		suites = append(suites, suite.ID)
	}
	suites = append(suites, tls.TLS_RSA_WITH_AES_128_CBC_SHA256)

	return &tls.Config{
		InsecureSkipVerify: true,
		CipherSuites:       suites,
	}
}

// Disconnect closes connection to the IRC server
func (i *Conn) Disconnect() {
	if !i.IsConnected() {
		return
	}
	i.Write([]byte("QUIT :Goodbye\r\n"))
	i.Conn.Close()
	i.Conn = nil
}

// SendMessage sends the given message string to the connected IRC server
func (i *Conn) SendMessage(message string) {
	if !i.IsConnected() {
		return
	}
	i.Write([]byte("PRIVMSG #" + i.channel + " :" + message + "\r\n"))
}

// SendNotice sends a notice message to the specified user
func (i *Conn) SendNotice(user string, message string) {
	if !i.IsConnected() {
		return
	}
	i.Write([]byte("NOTICE " + user + " :" + message + "\r\n"))
}

// JoinChannel joins the channel given by channel string
func (i *Conn) JoinChannel(channel string) {
	if !i.IsConnected() {
		return
	}
	i.channel = channel
	i.Write([]byte("JOIN #" + channel + "\r\n"))
}

// GetUsers sends a NAMES request to the IRC server
func (i *Conn) GetUsers(channel string) {
	if !i.IsConnected() {
		return
	}
	i.Write([]byte("NAMES #" + channel + "\r\n"))
}

// Pong sends a Pong message to the server, often used after a PING request
func (i *Conn) Pong(server string) {
	if !i.IsConnected() {
		return
	}
	i.Write([]byte("PONG " + server + "\r\n"))
}

// IsConnected returns true if the IRC connection is not null
func (i *Conn) IsConnected() bool {
	return i.Conn != nil
}
