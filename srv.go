package paymail

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
)

var (
	// ErrSRVInvalidDomainName is returned when domain name is invalid
	ErrSRVInvalidDomainName = errors.New("invalid parameter: domainName")
	// ErrSRVInvalidCNAME is returned when CNAME is invalid
	ErrSRVInvalidCNAME = errors.New("srv cname was invalid or not found")
	// ErrSRVMissing is returned when SRV is missing or nil
	ErrSRVMissing = errors.New("invalid parameter: srv is missing or nil")
	// ErrSRVTargetInvalid is returned when SRV target is invalid or empty
	ErrSRVTargetInvalid = errors.New("srv target is invalid or empty")
	// ErrSRVPortMismatch is returned when SRV port does not match expected
	ErrSRVPortMismatch = errors.New("srv port does not match")
	// ErrSRVPriorityMismatch is returned when SRV priority does not match expected
	ErrSRVPriorityMismatch = errors.New("srv priority does not match")
	// ErrSRVWeightMismatch is returned when SRV weight does not match expected
	ErrSRVWeightMismatch = errors.New("srv weight does not match")
	// ErrSRVTargetNoHost is returned when SRV target could not resolve a host
	ErrSRVTargetNoHost = errors.New("srv target could not resolve a host")
)

// defaultResolver will return a custom dns resolver
//
// This uses client options to set the network and port
func (c *Client) defaultResolver() net.Resolver {
	return net.Resolver{
		PreferGo:     true,
		StrictErrors: false,
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			d := net.Dialer{
				Timeout: c.options.dnsTimeout,
			}
			return d.DialContext(
				ctx, c.options.nameServerNetwork, c.options.nameServer+":"+c.options.dnsPort,
			)
		},
	}
}

// GetSRVRecord will get the SRV record for a given domain name
//
// Specs: http://bsvalias.org/02-01-host-discovery.html
func (c *Client) GetSRVRecord(service, protocol, domainName string) (srv *net.SRV, err error) {
	// Invalid parameters?
	if len(service) == 0 { // Use the default from paymail specs
		service = DefaultServiceName
	}
	if len(protocol) == 0 { // Use the default from paymail specs
		protocol = DefaultProtocol
	}
	if len(domainName) == 0 || len(domainName) > 255 {
		err = ErrSRVInvalidDomainName
		return srv, err
	}

	// Force the case
	protocol = strings.TrimSpace(strings.ToLower(protocol))

	// The computed cname to check against
	cnameCheck := fmt.Sprintf("_%s._%s.%s.", service, protocol, domainName)

	// Lookup the SRV record
	var cname string
	var records []*net.SRV
	if cname, records, err = c.resolver.LookupSRV(
		context.Background(), service, protocol, domainName,
	); err != nil || len(records) == 0 {
		// @rohenaz: Paymail spec says if SRV record doesn't exist, assume it is <domain>.<tld> and port of 443
		err = nil
		cname = cnameCheck
		records = append(records, &net.SRV{
			Port:     DefaultPort,
			Priority: DefaultPriority,
			Target:   domainName,
			Weight:   DefaultWeight,
		})
	}

	// Basic CNAME check (sanity check!)
	if cname != cnameCheck {
		err = fmt.Errorf(
			"using: %s and expected: %s: %w",
			cnameCheck, cname, ErrSRVInvalidCNAME,
		)
		return srv, err
	}

	// Only return the first record (in case multiple are returned)
	srv = records[0]

	// Remove any period on the end
	srv.Target = strings.TrimSuffix(srv.Target, ".")

	return srv, err
}

// ValidateSRVRecord will check for a valid SRV record for paymail following specifications
//
// Specs: http://bsvalias.org/02-01-host-discovery.html
func (c *Client) ValidateSRVRecord(ctx context.Context, srv *net.SRV, port, priority, weight uint16) error {
	// Check the parameters
	if srv == nil {
		return ErrSRVMissing
	}
	if port <= 0 { // Use the default(s) from paymail specs
		port = uint16(DefaultPort)
	}
	if priority <= 0 {
		priority = uint16(DefaultPriority)
	}
	if weight <= 0 {
		weight = uint16(DefaultWeight)
	}

	// Check the basics of the SRV record
	if len(srv.Target) == 0 {
		return ErrSRVTargetInvalid
	} else if srv.Port != port {
		return fmt.Errorf("srv port %d does not match %d: %w", srv.Port, port, ErrSRVPortMismatch)
	} else if srv.Priority != priority {
		return fmt.Errorf("srv priority %d does not match %d: %w", srv.Priority, priority, ErrSRVPriorityMismatch)
	} else if srv.Weight != weight {
		return fmt.Errorf("srv weight %d does not match %d: %w", srv.Weight, weight, ErrSRVWeightMismatch)
	}

	// Test resolving the target
	if addresses, err := c.resolver.LookupHost(ctx, srv.Target); err != nil {
		return err
	} else if len(addresses) == 0 {
		return fmt.Errorf("target %s: %w", srv.Target, ErrSRVTargetNoHost)
	}

	return nil
}
