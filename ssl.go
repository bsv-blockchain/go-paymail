package paymail

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"
)

// CheckSSL will do a basic check on the host to see if there is a valid SSL cert
//
// All paymail requests should be via HTTPS and have a valid certificate
func (c *Client) CheckSSL(host string) (valid bool, err error) {
	// Lookup the host
	var ips []net.IPAddr
	if ips, err = c.resolver.LookupIPAddr(context.Background(), host); err != nil {
		return valid, err
	}

	// Loop through all found ip addresses
	if len(ips) > 0 {
		for _, ip := range ips {

			// Set the dialer
			dialer := &tls.Dialer{
				NetDialer: &net.Dialer{
					Timeout:  c.options.sslTimeout,
					Deadline: time.Now().Add(c.options.sslDeadline),
				},
				Config: &tls.Config{
					ServerName: host,
					MinVersion: tls.VersionTLS12,
				},
			}

			// Set the connection
			ctx := context.Background()
			conn, dialErr := dialer.DialContext(ctx, DefaultProtocol, fmt.Sprintf("[%s]:%d", ip.String(), DefaultPort))
			if dialErr != nil {
				// catch missing ipv6 connectivity
				// if the ip is ipv6 and the resulting error is "no route to host", the record is skipped
				// otherwise the check will switch to critical
				/*
					if validate.IsValidIPv6(ip.String()) {
						switch dialErr.(type) {
						case *net.OpError:
							// https://stackoverflow.com/questions/38764084/proper-way-to-handle-missing-ipv6-connectivity
							if dialErr.(*net.OpError).Err.(*os.SyscallError).Err == syscall.EHOSTUNREACH {
								// log.Printf("%-15s - ignoring unreachable IPv6 address", ip)
								continue
							}
						}
					}
				*/
				continue
			}

			// remember the checked certs based on their Signature
			checkedCerts := make(map[string]struct{})

			// loop to all certs we get
			// there might be multiple chains, as there may be one or more CAs present on the current system,
			// so we have multiple possible chains
			tlsConn := conn.(*tls.Conn)
			for _, chain := range tlsConn.ConnectionState().VerifiedChains {
				for _, cert := range chain {
					if _, checked := checkedCerts[string(cert.Signature)]; checked {
						continue
					}
					checkedCerts[string(cert.Signature)] = struct{}{}

					// Filter out CA certificates
					if cert.IsCA {
						// log.Printf("ignoring CA certificate on ip %s by %s", ip, cert.Subject.CommonName)
						continue
					}

					// Fail if less than 1 day for expiration
					// remainingValidity := cert.NotAfter.Sub(time.Now())
					if time.Until(cert.NotAfter) > 24*time.Hour {
						valid = true
					}
				}
			}
			_ = conn.Close()
		}
	}

	return valid, err
}
