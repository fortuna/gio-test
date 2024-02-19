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

//go:build darwin

package main

/*
#include <stdlib.h>
#include <dns_sd.h>
#include <stdio.h>

extern void goCallback(DNSServiceRef sdRef, DNSServiceFlags flags, uint32_t interfaceIndex,
	DNSServiceErrorType errorCode, char* fullname, uint16_t rrtype, uint16_t rrclass,
	uint16_t rdlen, void* rdata, uint32_t ttl, void* context);

void cCallback(DNSServiceRef sdRef, DNSServiceFlags flags, uint32_t interfaceIndex,
	DNSServiceErrorType errorCode, const char *fullname, uint16_t rrtype, uint16_t rrclass,
	uint16_t rdlen, const void *rdata, uint32_t ttl, void *context) {
	printf("cCallback\n");
    goCallback(sdRef, flags, interfaceIndex, errorCode, (char*)fullname, rrtype, rrclass,
		rdlen, (void*)rdata, ttl, context);
}

DNSServiceErrorType queryDNS(DNSServiceRef *sdRef, DNSServiceFlags flags, uint32_t interfaceIndex,
	const char *fullname, uint16_t rrtype, uint16_t rrclass,
	void *context) {
	printf("queryDNS\n");
    return DNSServiceQueryRecord(sdRef, flags, interfaceIndex, fullname, rrtype, rrclass,
	cCallback, context);
	// (DNSServiceQueryRecordReply)goCallback, context);
}
*/
import "C"

import (
	"context"
	"fmt"
	"runtime/cgo"
	"time"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
)

type queryCallbackFunc func(flags C.DNSServiceFlags, interfaceIndex int,
	errorCode C.DNSServiceErrorType, fullname string,
	rrtype dnsmessage.Type, rrclass dnsmessage.Class,
	rdata []byte, ttl uint32)

func queryCNAME(ctx context.Context, qname string) (string, error) {
	defer fmt.Println("cname returned")
	var sdRef C.DNSServiceRef

	cQname := C.CString(qname)
	defer C.free(unsafe.Pointer(cQname))
	var answer string
	done := make(chan struct{}, 1)
	var callback queryCallbackFunc = func(flags C.DNSServiceFlags, interfaceIndex int,
		errorCode C.DNSServiceErrorType, fullname string,
		rrtype dnsmessage.Type, rrclass dnsmessage.Class,
		rdata []byte, ttl uint32) {
		if flags&C.kDNSServiceFlagsMoreComing != C.kDNSServiceFlagsMoreComing {
			defer func() { done <- struct{}{} }()
		}
		answer += fmt.Sprintln(flags, interfaceIndex, errorCode, fullname, rrtype, rrclass, rdata, ttl)
		fmt.Println("Answer:", answer)
	}
	cContext := cgo.NewHandle(callback)
	defer cContext.Delete()
	fmt.Println("starting")
	// See https://developer.apple.com/documentation/dnssd/1804747-dnsservicequeryrecord?language=objc
	serviceErr := C.queryDNS(&sdRef, 0, 0,
		cQname, C.uint16_t(dnsmessage.TypeA), C.uint16_t(dnsmessage.ClassINET),
		unsafe.Pointer(&cContext))
	fmt.Println("queryDNS serviceErr:", serviceErr)
	if serviceErr != 0 {
		return "", fmt.Errorf("failed to start DNS query: %v", serviceErr)
	}
	defer C.DNSServiceRefDeallocate(sdRef)

	// See https://developer.apple.com/documentation/dnssd/1804698-dnsservicerefsockfd
	fd := C.DNSServiceRefSockFD(sdRef)
	if fd < 0 {
		return "", fmt.Errorf("failed to get DNSServiceRef file descriptor")
	}
	fmt.Println("For loop, fd:", fd)
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-done:
			return answer, nil
		default:
		}
		timeout := int(-1)
		if deadline, ok := ctx.Deadline(); ok {
			timeout = int(time.Until(deadline).Milliseconds())
		}
		fmt.Println("Poll")
		nReady, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN | unix.POLLERR | unix.POLLHUP}}, timeout)
		if err != nil {
			return "", err
		}
		if nReady == 0 {
			return "", context.DeadlineExceeded
		}
		// See https://developer.apple.com/documentation/dnssd/1804696-dnsserviceprocessresult?language=objc.
		fmt.Println("DNSServiceProcessResult")
		serviceErr = C.DNSServiceProcessResult(sdRef)
		if serviceErr != 0 {
			return "", fmt.Errorf("failed to process DNS response: %v", serviceErr)
		}
	}

	// answer = answer[:n]
	// var msg dnsmessage.Message
	// err := msg.Unpack(answer)
	// if err != nil {
	// 	return "", err
	// }
	// return msg.GoString(), nil
}
