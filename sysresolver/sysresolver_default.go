// Copyright 2024 Vinicius Fortuna
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build defaultresolver

// Uses default resolver for addresses, getaddrinfo for CNAME.
//
// Benefits:
// - Leverages the system resolver and caching.
//
// Issues:
// - Doesn't return the message
// - getaddrinfo for CNAME is blocking, no way to pass a timeout

package sysresolver

/*
#include <stdlib.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <netdb.h>
*/
import "C"

import (
	"context"
	"fmt"
	"net"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
)

func Query(ctx context.Context, qname string, qtype dnsmessage.Type) (string, error) {
	switch qtype {
	case dnsmessage.TypeA:
		ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip4", qname)
		return formatIPs(ips), err
	case dnsmessage.TypeAAAA:
		ips, err := net.DefaultResolver.LookupIP(context.Background(), "ip6", qname)
		return formatIPs(ips), err
	case dnsmessage.TypeCNAME:
		return queryCNAME(ctx, qname)
	default:
		return "", fmt.Errorf("Query type %v not implemented", qtype)
	}
}

func queryCNAME(ctx context.Context, qname string) (string, error) {
	// The default LookupCNAME uses libresolv, which doesn't work on mobile and
	// doesn't leverage the system cache.
	// return net.DefaultResolver.LookupCNAME(ctx, qname)
	type result struct {
		cname string
		err   error
	}

	results := make(chan result)
	go func() {
		cname, err := lookupCNAMEBlocking(qname)
		results <- result{cname, err}
	}()

	select {
	case r := <-results:
		return r.cname, r.err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

func lookupCNAMEBlocking(host string) (string, error) {
	var hints C.struct_addrinfo
	var result *C.struct_addrinfo

	chost := C.CString(host)
	defer C.free(unsafe.Pointer(chost))

	hints.ai_family = C.AF_UNSPEC
	hints.ai_flags = C.AI_CANONNAME

	// Call getaddrinfo
	res := C.getaddrinfo(chost, nil, &hints, &result)
	if res != 0 {
		return "", fmt.Errorf("getaddrinfo error: %s", C.GoString(C.gai_strerror(res)))
	}
	defer C.freeaddrinfo(result)

	// Extract canonical name
	cname := C.GoString(result.ai_canonname)
	return cname, nil
}
