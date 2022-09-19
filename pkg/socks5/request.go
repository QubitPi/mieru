package socks5

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/enfein/mieru/pkg/log"
	"github.com/enfein/mieru/pkg/metrics"
	"github.com/enfein/mieru/pkg/netutil"
	"github.com/enfein/mieru/pkg/stderror"
)

const (
	ConnectCommand   = uint8(1)
	BindCommand      = uint8(2)
	AssociateCommand = uint8(3)
	ipv4Address      = uint8(1)
	fqdnAddress      = uint8(3)
	ipv6Address      = uint8(4)
)

const (
	successReply uint8 = iota
	serverFailure
	ruleFailure
	networkUnreachable
	hostUnreachable
	connectionRefused
	ttlExpired
	commandNotSupported
	addrTypeNotSupported
)

var (
	unrecognizedAddrType = fmt.Errorf("unrecognized address type")
)

// AddrSpec is used to return the target AddrSpec
// which may be specified as IPv4, IPv6, or a FQDN.
type AddrSpec struct {
	FQDN string
	IP   net.IP
	Port int
}

func (a *AddrSpec) String() string {
	if a.FQDN != "" {
		return fmt.Sprintf("%s (%s):%d", a.FQDN, a.IP, a.Port)
	}
	return fmt.Sprintf("%s:%d", a.IP, a.Port)
}

// Address returns a string suitable to dial; prefer returning IP-based
// address, fallback to FQDN
func (a AddrSpec) Address() string {
	if 0 != len(a.IP) {
		return net.JoinHostPort(a.IP.String(), strconv.Itoa(a.Port))
	}
	return net.JoinHostPort(a.FQDN, strconv.Itoa(a.Port))
}

// A Request represents request received by a server.
type Request struct {
	// Protocol version.
	Version uint8
	// Requested command.
	Command uint8
	// AuthContext provided during negotiation.
	AuthContext *AuthContext
	// AddrSpec of the the network that sent the request.
	RemoteAddr *AddrSpec
	// AddrSpec of the desired destination.
	DestAddr *AddrSpec
}

// NewRequest creates a new Request from the tcp connection.
func NewRequest(conn io.Reader) (*Request, error) {
	// Read the version byte.
	header := []byte{0, 0, 0}
	if _, err := io.ReadFull(conn, header); err != nil {
		return nil, fmt.Errorf("failed to get command version: %w", err)
	}

	// Ensure we are compatible.
	if header[0] != socks5Version {
		return nil, fmt.Errorf("unsupported command version: %v", header[0])
	}

	// Read in the destination address.
	dest, err := readAddrSpec(conn)
	if err != nil {
		return nil, err
	}

	request := &Request{
		Version:  socks5Version,
		Command:  header[1],
		DestAddr: dest,
	}

	return request, nil
}

// handleRequest is used for request processing after authentication.
func (s *Server) handleRequest(req *Request, conn io.ReadWriteCloser) error {
	ctx := context.Background()

	// Resolve the address if we have a FQDN.
	dest := req.DestAddr
	if dest.FQDN != "" {
		ctx_, addr, err := s.config.Resolver.Resolve(ctx, dest.FQDN)
		if err != nil {
			atomic.AddUint64(&metrics.Socks5DNSResolveErrors, 1)
			if err := sendReply(conn, hostUnreachable, nil); err != nil {
				return fmt.Errorf("failed to send reply: %w", err)
			}
			return fmt.Errorf("failed to resolve destination %q: %w", dest.FQDN, err)
		}
		ctx = ctx_
		dest.IP = addr
	}

	// Return error if access local destination is not allowed.
	if !s.config.AllowLocalDestination && isLocalhostDest(req) {
		return fmt.Errorf("access to localhost resource via proxy is not allowed")
	}

	// Switch on the command.
	switch req.Command {
	case ConnectCommand:
		return s.handleConnect(ctx, conn, req)
	case BindCommand:
		return s.handleBind(ctx, conn, req)
	case AssociateCommand:
		return s.handleAssociate(ctx, conn, req)
	default:
		atomic.AddUint64(&metrics.Socks5UnsupportedCommandErrors, 1)
		if err := sendReply(conn, commandNotSupported, nil); err != nil {
			return fmt.Errorf("failed to send reply: %w", err)
		}
		return fmt.Errorf("unsupported command: %v", req.Command)
	}
}

// handleConnect is used to handle a connect command.
func (s *Server) handleConnect(ctx context.Context, conn io.ReadWriteCloser, req *Request) error {
	var d net.Dialer
	target, err := d.DialContext(ctx, "tcp", req.DestAddr.Address())
	if err != nil {
		msg := err.Error()
		var resp uint8
		if strings.Contains(msg, "refused") {
			resp = connectionRefused
			atomic.AddUint64(&metrics.Socks5ConnectionRefusedErrors, 1)
		} else if strings.Contains(msg, "network is unreachable") {
			resp = networkUnreachable
			atomic.AddUint64(&metrics.Socks5NetworkUnreachableErrors, 1)
		} else {
			resp = hostUnreachable
			atomic.AddUint64(&metrics.Socks5HostUnreachableErrors, 1)
		}
		if err := sendReply(conn, resp, nil); err != nil {
			return fmt.Errorf("failed to send reply: %w", err)
		}
		return fmt.Errorf("connect to %v failed: %w", req.DestAddr, err)
	}
	defer target.Close()

	// Send success.
	local := target.LocalAddr().(*net.TCPAddr)
	bind := AddrSpec{IP: local.IP, Port: local.Port}
	if err := sendReply(conn, successReply, &bind); err != nil {
		atomic.AddUint64(&metrics.Socks5HandshakeErrors, 1)
		return fmt.Errorf("failed to send reply: %w", err)
	}

	return BidiCopy(conn, target, false)
}

// handleBind is used to handle a bind command.
func (s *Server) handleBind(ctx context.Context, conn io.ReadWriteCloser, req *Request) error {
	atomic.AddUint64(&metrics.Socks5UnsupportedCommandErrors, 1)
	if err := sendReply(conn, commandNotSupported, nil); err != nil {
		atomic.AddUint64(&metrics.Socks5HandshakeErrors, 1)
		return fmt.Errorf("failed to send reply: %w", err)
	}
	return nil
}

// handleAssociate is used to handle a associate command.
func (s *Server) handleAssociate(ctx context.Context, conn io.ReadWriteCloser, req *Request) error {
	// Create a UDP listener on a random port.
	// All the requests associated to this connection will go through this port.
	udpListenerAddr, err := net.ResolveUDPAddr("udp", netutil.MaybeDecorateIPv6(netutil.AllIPAddr())+":0")
	if err != nil {
		atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
		return fmt.Errorf("failed to resolve UDP address: %w", err)
	}
	udpConn, err := net.ListenUDP("udp", udpListenerAddr)
	if err != nil {
		atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
		return fmt.Errorf("failed to listen UDP: %w", err)
	}

	// Use 0.0.0.0:<port> as the bind address.
	// This is the port used by the server. Client will rewrite the port number.
	_, udpPortStr, err := net.SplitHostPort(udpConn.LocalAddr().String())
	if err != nil {
		atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
		return fmt.Errorf("net.SplitHostPort() failed: %w", err)
	}
	udpPort, err := strconv.Atoi(udpPortStr)
	if err != nil {
		atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
		return fmt.Errorf("strconv.Atoi() failed: %w", err)
	}
	bind := AddrSpec{IP: net.IP{0, 0, 0, 0}, Port: udpPort}
	if err := sendReply(conn, successReply, &bind); err != nil {
		atomic.AddUint64(&metrics.Socks5HandshakeErrors, 1)
		return fmt.Errorf("failed to send reply: %w", err)
	}

	conn = WrapUDPAssociateTunnel(conn)
	var udpErr atomic.Value

	var wg sync.WaitGroup
	wg.Add(2)

	// Send outbound UDP packets.
	go func() {
		defer wg.Done()
		defer udpConn.Close()
		buf := make([]byte, 1<<16)
		var n int
		var err error
		for {
			n, err = conn.Read(buf)
			if err != nil {
				udpErr.Store(err)
				return
			}

			// Validate received UDP request.
			if n <= 6 {
				udpErr.Store(stderror.ErrNoEnoughData)
				atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				return
			}
			if buf[0] != 0x00 || buf[1] != 0x00 {
				udpErr.Store(stderror.ErrInvalidArgument)
				atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				return
			}
			if buf[2] != 0x00 {
				// UDP fragment is not supported.
				udpErr.Store(stderror.ErrUnsupported)
				atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				return
			}
			addrType := buf[3]
			if addrType != 0x01 && addrType != 0x03 && addrType != 0x04 {
				udpErr.Store(stderror.ErrInvalidArgument)
				atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				return
			}
			if (addrType == 0x01 && n <= 10) || (addrType == 0x03 && n <= int(buf[4])+6) || (addrType == 0x04 && n <= 22) {
				udpErr.Store(stderror.ErrNoEnoughData)
				atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				return
			}

			// Get target address and send data.
			switch addrType {
			case 0x01:
				dstAddr := &net.UDPAddr{
					IP:   net.IP(buf[4:8]),
					Port: int(buf[8])<<8 + int(buf[9]),
				}
				ws, err := udpConn.WriteToUDP(buf[10:n+1], dstAddr)
				if err != nil {
					if log.IsLevelEnabled(log.DebugLevel) {
						log.Debugf("UDP associate [%v - %v] WriteToUDP() failed: %v", udpConn.LocalAddr(), dstAddr, err)
					}
					atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				} else {
					atomic.AddUint64(&metrics.UDPAssociateOutPkts, 1)
					atomic.AddUint64(&metrics.UDPAssociateOutBytes, uint64(ws))
				}
			case 0x03:
				fqdnLen := buf[4]
				fqdn := string(buf[5 : 5+fqdnLen])
				dstAddr, err := net.ResolveUDPAddr("udp", fqdn+":"+strconv.Itoa(int(buf[5+fqdnLen])<<8+int(buf[6+fqdnLen])))
				if err != nil {
					if log.IsLevelEnabled(log.DebugLevel) {
						log.Debugf("UDP associate %v ResolveUDPAddr() failed: %v", udpConn.LocalAddr(), err)
					}
					atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
					break
				}
				ws, err := udpConn.WriteToUDP(buf[7+fqdnLen:n+1], dstAddr)
				if err != nil {
					if log.IsLevelEnabled(log.DebugLevel) {
						log.Debugf("UDP associate [%v - %v] WriteToUDP() failed: %v", udpConn.LocalAddr(), dstAddr, err)
					}
					atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				} else {
					atomic.AddUint64(&metrics.UDPAssociateOutPkts, 1)
					atomic.AddUint64(&metrics.UDPAssociateOutBytes, uint64(ws))
				}
			case 0x04:
				dstAddr := &net.UDPAddr{
					IP:   net.IP(buf[4:20]),
					Port: int(buf[20])<<8 + int(buf[21]),
				}
				ws, err := udpConn.WriteToUDP(buf[22:n+1], dstAddr)
				if err != nil {
					if log.IsLevelEnabled(log.DebugLevel) {
						log.Debugf("UDP associate [%v - %v] WriteToUDP() failed: %v", udpConn.LocalAddr(), dstAddr, err)
					}
					atomic.AddUint64(&metrics.Socks5UDPAssociateErrors, 1)
				} else {
					atomic.AddUint64(&metrics.UDPAssociateOutPkts, 1)
					atomic.AddUint64(&metrics.UDPAssociateOutBytes, uint64(ws))
				}
			}
		}
	}()

	// Receive inbound UDP packets.
	go func() {
		defer wg.Done()
		buf := make([]byte, 1<<16)
		var n int
		var err error
		for {
			n, err = udpConn.Read(buf)
			if err != nil {
				// This is typically due to close of UDP listener.
				// Don't contribute to metrics.Socks5UDPAssociateErrors.
				if log.IsLevelEnabled(log.DebugLevel) {
					log.Debugf("UDP associate %v Read() failed: %v", udpConn.LocalAddr(), err)
				}
				if udpErr.Load() == nil {
					udpErr.Store(err)
				}
				return
			} else {
				atomic.AddUint64(&metrics.UDPAssociateInPkts, 1)
				atomic.AddUint64(&metrics.UDPAssociateInBytes, uint64(n))
			}
			_, err = conn.Write(buf[:n])
			if err != nil {
				if log.IsLevelEnabled(log.DebugLevel) {
					log.Debugf("UDP associate %v Write() to proxy client failed: %v", udpConn.LocalAddr(), err)
				}
				if udpErr.Load() == nil {
					udpErr.Store(err)
				}
				return
			}
		}
	}()

	wg.Wait()
	return udpErr.Load().(error)
}

// proxySocks5AuthReq transfers the socks5 authentication request and response
// between socks5 client and server.
func (s *Server) proxySocks5AuthReq(conn, proxyConn net.Conn) error {
	// Send the version and authtication methods to the server.
	version := []byte{0}
	if _, err := io.ReadFull(conn, version); err != nil {
		return fmt.Errorf("failed to get version byte: %w", err)
	}
	if version[0] != socks5Version {
		return fmt.Errorf("unsupported SOCKS version: %v", version)
	}
	nMethods := []byte{0}
	if _, err := io.ReadFull(conn, nMethods); err != nil {
		return fmt.Errorf("failed to get the length of authentication methods: %w", err)
	}
	methods := make([]byte, int(nMethods[0]))
	if _, err := io.ReadFull(conn, methods); err != nil {
		return fmt.Errorf("failed to get authentication methods: %w", err)
	}
	authReq := []byte{}
	authReq = append(authReq, version...)
	authReq = append(authReq, nMethods...)
	authReq = append(authReq, methods...)
	if _, err := proxyConn.Write(authReq); err != nil {
		return fmt.Errorf("failed to write authentication request to the server: %w", err)
	}

	// Get server authentication response.
	authResp := make([]byte, 2)
	if _, err := io.ReadFull(proxyConn, authResp); err != nil {
		return fmt.Errorf("failed to read authentication response from the socks5 server: %w", err)
	}
	if _, err := conn.Write(authResp); err != nil {
		return fmt.Errorf("failed to write authentication response to the socks5 client: %w", err)
	}

	return nil
}

// proxySocks5ConnReq transfers the socks5 connection request and response
// between socks5 client and server. Optionally, if UDP association is used,
// return the created UDP connection.
func (s *Server) proxySocks5ConnReq(conn, proxyConn net.Conn) (*net.UDPConn, error) {
	// Send the connection request to the server.
	connReq := make([]byte, 4)
	if _, err := io.ReadFull(conn, connReq); err != nil {
		return nil, fmt.Errorf("failed to get socks5 connection request: %w", err)
	}
	cmd := connReq[1]
	reqAddrType := connReq[3]
	var reqFQDNLen []byte
	var dstAddr []byte
	switch reqAddrType {
	case ipv4Address:
		dstAddr = make([]byte, 6)
	case fqdnAddress:
		reqFQDNLen = []byte{0}
		if _, err := io.ReadFull(conn, reqFQDNLen); err != nil {
			return nil, fmt.Errorf("failed to get FQDN length: %w", err)
		}
		dstAddr = make([]byte, reqFQDNLen[0]+2)
	case ipv6Address:
		dstAddr = make([]byte, 18)
	default:
		return nil, fmt.Errorf("unsupported address type: %d", reqAddrType)
	}
	if _, err := io.ReadFull(conn, dstAddr); err != nil {
		return nil, fmt.Errorf("failed to get destination address: %w", err)
	}
	if len(reqFQDNLen) != 0 {
		connReq = append(connReq, reqFQDNLen...)
	}
	connReq = append(connReq, dstAddr...)
	if _, err := proxyConn.Write(connReq); err != nil {
		return nil, fmt.Errorf("failed to write connection request to the server: %w", err)
	}

	// Get server connection response.
	connResp := make([]byte, 4)
	if _, err := io.ReadFull(proxyConn, connResp); err != nil {
		return nil, fmt.Errorf("failed to read connection response from the server: %w", err)
	}
	respAddrType := connResp[3]
	var respFQDNLen []byte
	var bindAddr []byte
	switch respAddrType {
	case ipv4Address:
		bindAddr = make([]byte, 6)
	case fqdnAddress:
		respFQDNLen = []byte{0}
		if _, err := io.ReadFull(proxyConn, respFQDNLen); err != nil {
			return nil, fmt.Errorf("failed to get FQDN length: %w", err)
		}
		bindAddr = make([]byte, respFQDNLen[0]+2)
	case ipv6Address:
		bindAddr = make([]byte, 18)
	default:
		return nil, fmt.Errorf("unsupported address type: %d", respAddrType)
	}
	if _, err := io.ReadFull(proxyConn, bindAddr); err != nil {
		return nil, fmt.Errorf("failed to get bind address: %w", err)
	}
	if len(respFQDNLen) != 0 {
		connResp = append(connResp, respFQDNLen...)
	}
	connResp = append(connResp, bindAddr...)

	var udpConn *net.UDPConn
	if cmd == AssociateCommand {
		// Create a UDP listener on a random port.
		udpAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:0")
		if err != nil {
			return nil, fmt.Errorf("net.ResolveUDPAddr() failed: %w", err)
		}
		udpConn, err = net.ListenUDP("udp", udpAddr)
		if err != nil {
			return nil, fmt.Errorf("net.ListenUDP() failed: %w", err)
		}
		// Get the port number and rewrite the response.
		_, udpPortStr, err := net.SplitHostPort(udpConn.LocalAddr().String())
		if err != nil {
			udpConn.Close()
			return nil, fmt.Errorf("net.SplitHostPort() failed: %w", err)
		}
		udpPort, err := strconv.Atoi(udpPortStr)
		if err != nil {
			udpConn.Close()
			return nil, fmt.Errorf("strconv.Atoi() failed: %w", err)
		}
		lenResp := len(connResp)
		connResp[lenResp-2] = byte(udpPort >> 8)
		connResp[lenResp-1] = byte(udpPort)
	}

	if _, err := conn.Write(connResp); err != nil {
		return nil, fmt.Errorf("failed to write connection response to the socks5 client: %w", err)
	}

	return udpConn, nil
}

// readAddrSpec is used to read AddrSpec.
// Expects an address type byte, follwed by the address and port.
func readAddrSpec(r io.Reader) (*AddrSpec, error) {
	d := &AddrSpec{}

	// Get the address type.
	addrType := []byte{0}
	if _, err := io.ReadFull(r, addrType); err != nil {
		return nil, err
	}

	// Handle on a per type basis.
	switch addrType[0] {
	case ipv4Address:
		addr := make([]byte, 4)
		if _, err := io.ReadFull(r, addr); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case ipv6Address:
		addr := make([]byte, 16)
		if _, err := io.ReadFull(r, addr); err != nil {
			return nil, err
		}
		d.IP = net.IP(addr)

	case fqdnAddress:
		if _, err := io.ReadFull(r, addrType); err != nil {
			return nil, err
		}
		addrLen := int(addrType[0])
		fqdn := make([]byte, addrLen)
		if _, err := io.ReadFull(r, fqdn); err != nil {
			return nil, err
		}
		d.FQDN = string(fqdn)

	default:
		return nil, unrecognizedAddrType
	}

	// Read the port number.
	port := []byte{0, 0}
	if _, err := io.ReadFull(r, port); err != nil {
		return nil, err
	}
	d.Port = (int(port[0]) << 8) | int(port[1])

	return d, nil
}

// sendReply is used to send a reply message.
func sendReply(w io.Writer, resp uint8, addr *AddrSpec) error {
	// Format the address.
	var addrType uint8
	var addrBody []byte
	var addrPort uint16
	switch {
	case addr == nil:
		addrType = ipv4Address
		addrBody = []byte{0, 0, 0, 0}
		addrPort = 0

	case addr.FQDN != "":
		addrType = fqdnAddress
		addrBody = append([]byte{byte(len(addr.FQDN))}, addr.FQDN...)
		addrPort = uint16(addr.Port)

	case addr.IP.To4() != nil:
		addrType = ipv4Address
		addrBody = []byte(addr.IP.To4())
		addrPort = uint16(addr.Port)

	case addr.IP.To16() != nil:
		addrType = ipv6Address
		addrBody = []byte(addr.IP.To16())
		addrPort = uint16(addr.Port)

	default:
		return fmt.Errorf("failed to format address: %v", addr)
	}

	// Format the message.
	msg := make([]byte, 6+len(addrBody))
	msg[0] = socks5Version
	msg[1] = resp
	msg[2] = 0 // reserved byte
	msg[3] = addrType
	copy(msg[4:], addrBody)
	msg[4+len(addrBody)] = byte(addrPort >> 8)
	msg[4+len(addrBody)+1] = byte(addrPort & 0xff)

	// Send the message.
	_, err := w.Write(msg)
	return err
}

func isLocalhostDest(req *Request) bool {
	if req == nil || req.DestAddr == nil {
		return false
	}
	if req.DestAddr.FQDN == "localhost" || req.DestAddr.IP.IsLoopback() {
		return true
	}
	return false
}
