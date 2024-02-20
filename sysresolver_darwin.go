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

extern void goDNSServiceQueryRecordReply(DNSServiceRef sdRef, DNSServiceFlags flags, uint32_t interfaceIndex,
	DNSServiceErrorType errorCode, char* fullname, uint16_t rrtype, uint16_t rrclass,
	uint16_t rdlen, void* rdata, uint32_t ttl, void* context);
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

//export goDNSServiceQueryRecordReply
func goDNSServiceQueryRecordReply(sdRef C.DNSServiceRef, flags C.DNSServiceFlags, interfaceIndex C.uint32_t,
	errorCode C.DNSServiceErrorType, fullname *C.char, rrtype C.uint16_t, rrclass C.uint16_t,
	rdlen C.uint16_t, rdata unsafe.Pointer, ttl C.uint32_t, context unsafe.Pointer) {
	fmt.Println("goCallback", errorCode)
	h := *(*cgo.Handle)(context)
	doneFunc := h.Value().(queryCallbackFunc)

	var goFullname string
	var goRData []byte
	if errorCode == 0 {
		goFullname = C.GoString(fullname)
		goRData = C.GoBytes(rdata, C.int(rdlen))
	}
	goRRType := dnsmessage.Type(rrtype)
	goRRClass := dnsmessage.Class(rrclass)

	doneFunc(flags, int(interfaceIndex), errorCode, goFullname, goRRType, goRRClass, goRData, uint32(ttl))
}

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
	serviceErr := C.DNSServiceQueryRecord(&sdRef, 0, 0,
		cQname, C.uint16_t(dnsmessage.TypeCNAME), C.uint16_t(dnsmessage.ClassINET),
		C.DNSServiceQueryRecordReply(C.goDNSServiceQueryRecordReply),
		unsafe.Pointer(&cContext))
	fmt.Println("queryDNS serviceErr:", serviceErr)
	if serviceErr != C.kDNSServiceErr_NoError {
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

		/*
			fmt.Println("Select")
			var selectTimeout *unix.Timeval
			if deadline, ok := ctx.Deadline(); ok {
				timeout := time.Until(deadline)
				selectTimeout = &unix.Timeval{
					Sec:  timeout.Milliseconds() / 1000,
					Usec: int32(timeout.Milliseconds() % 1000 * 1000),
				}
			}
			var fds unix.FdSet
			fds.Set(int(fd))
			nReady, err := unix.Select(int(fd+1), &fds, nil, &fds, selectTimeout)
			if err != nil {
				return "", err
			}
			if nReady == 0 {
				return "", context.DeadlineExceeded
			}
		*/

		fmt.Println("Poll")
		pollTimeout := int(-1)
		if deadline, ok := ctx.Deadline(); ok {
			timeout := time.Until(deadline)
			pollTimeout = int(timeout.Milliseconds())
		}
		nReady, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN | unix.POLLERR | unix.POLLHUP}}, pollTimeout)
		if err != nil {
			return "", err
		}
		if nReady == 0 {
			return "", context.DeadlineExceeded
		}

		// See https://developer.apple.com/documentation/dnssd/1804696-dnsserviceprocessresult?language=objc.
		fmt.Println("DNSServiceProcessResult")
		serviceErr = C.DNSServiceProcessResult(sdRef)
		if serviceErr != C.kDNSServiceErr_NoError {
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
