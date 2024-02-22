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

//go:build darwin && dnssd

package sysresolver

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
	"log"
	"runtime/cgo"
	"time"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
)

func Query(ctx context.Context, qname string, qtype dnsmessage.Type) (string, error) {
	msg, err := queryAnswers(ctx, qname, qtype)
	return formatMessage(msg), err
}

type queryCallbackFunc func(flags C.DNSServiceFlags, interfaceIndex int,
	errorCode C.DNSServiceErrorType, fullname string,
	rrtype dnsmessage.Type, rrclass dnsmessage.Class,
	rdata []byte, ttl uint32)

//export goDNSServiceQueryRecordReply
func goDNSServiceQueryRecordReply(sdRef C.DNSServiceRef, flags C.DNSServiceFlags, interfaceIndex C.uint32_t,
	errorCode C.DNSServiceErrorType, fullname *C.char, rrtype C.uint16_t, rrclass C.uint16_t,
	rdlen C.uint16_t, rdata unsafe.Pointer, ttl C.uint32_t, context unsafe.Pointer) {
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

// Uses dns_sd. It leverages the system cache.
// See https://opensource.apple.com/source/mDNSResponder/mDNSResponder-878.70.2/mDNSShared/dns_sd.h.auto.html
func queryAnswers(ctx context.Context, qname string, qtype dnsmessage.Type) (*dnsmessage.Message, error) {
	defer log.Println("queryAnswers returned")
	var sdRef C.DNSServiceRef
	var response dnsmessage.Message
	response.Header.Response = true
	parsedQname, err := dnsmessage.NewName(qname)
	if err != nil {
		return nil, err
	}
	response.Questions = []dnsmessage.Question{{Name: parsedQname, Type: qtype, Class: dnsmessage.ClassINET}}

	cQname := C.CString(qname)
	defer C.free(unsafe.Pointer(cQname))
	answers := make([]dnsmessage.Resource, 0, 1)
	done := make(chan struct{}, 1)
	var callback queryCallbackFunc = func(flags C.DNSServiceFlags, interfaceIndex int,
		errorCode C.DNSServiceErrorType, fullname string,
		rrtype dnsmessage.Type, rrclass dnsmessage.Class,
		rdata []byte, ttl uint32) {
		// See https://developer.apple.com/documentation/dnssd/1823436-anonymous/kdnsserviceflagsmorecoming?language=objc
		if flags&C.kDNSServiceFlagsMoreComing != C.kDNSServiceFlagsMoreComing &&
			rrtype == qtype {
			// When there's a CNAME chain cached, the system will immediately return that,
			// with kDNSServiceFlagsMoreComing == 0, and return the requested type later,
			// when the resolver responds. So we need to keep waiting until we
			// get the right resource type or an error occurs. We still need to check for
			// kDNSServiceFlagsMoreComing == 0 to make sure we get all responses.
			// TODO: what if the resolver returns a CNAME chain, but no answer with the right type?
			defer func() { done <- struct{}{} }()
		}
		if errorCode != C.kDNSServiceErr_NoError {
			switch errorCode {
			case C.kDNSServiceErr_NoSuchRecord:
				log.Println("No such record (NOERROR, no answers)")
			case C.kDNSServiceErr_NoSuchName:
				response.RCode = dnsmessage.RCodeNameError
				log.Println("NXDOMAIN")

			default:
				log.Println("errorCode:", errorCode)
			}
			return
		}
		parsedName, err := dnsmessage.NewName(fullname)
		if err != nil {
			return
		}
		resource := dnsmessage.Resource{
			Header: dnsmessage.ResourceHeader{
				Name:  parsedName,
				Type:  dnsmessage.Type(rrtype),
				Class: dnsmessage.Class(rrclass),
				TTL:   ttl,
			},
			Body: &dnsmessage.UnknownResource{Type: dnsmessage.Type(rrtype), Data: rdata},
		}
		switch rrtype {
		case dnsmessage.TypeA:
			if len(rdata) == 4 {
				resource.Body = &dnsmessage.AResource{A: [4]byte(rdata)}
			}
		case dnsmessage.TypeAAAA:
			if len(rdata) == 16 {
				resource.Body = &dnsmessage.AAAAResource{AAAA: [16]byte(rdata)}
			}
		case dnsmessage.TypeCNAME:
			var cname string
			for remaining := rdata; len(remaining) > 0; {
				prefixLen := int(remaining[0])
				if prefixLen == 0 {
					break
				}
				if 1+prefixLen > len(rdata) {
					prefixLen = len(rdata) - 1
				}
				prefix := remaining[1 : prefixLen+1]
				cname += string(prefix) + "."
				remaining = remaining[prefixLen+1:]
			}
			name, err := dnsmessage.NewName(cname)
			if err == nil {
				resource.Body = &dnsmessage.CNAMEResource{CNAME: name}
			}
		default:
			resource.Body = &dnsmessage.UnknownResource{Type: rrtype, Data: rdata}
		}
		log.Println("Answer:", flags, interfaceIndex, errorCode, resource.GoString())

		answers = append(answers, resource)
		response.Answers = append(response.Answers, resource)
	}
	cContext := cgo.NewHandle(callback)
	defer cContext.Delete()
	log.Println("starting")
	// https://developer.apple.com/documentation/dnssd/1823436-anonymous/kdnsserviceflagsreturnintermediates?language=objc
	var flags C.DNSServiceFlags = C.kDNSServiceFlagsReturnIntermediates
	// See https://developer.apple.com/documentation/dnssd/1804747-dnsservicequeryrecord?language=objc
	serviceErr := C.DNSServiceQueryRecord(&sdRef, flags, 0,
		cQname, C.uint16_t(qtype), C.uint16_t(dnsmessage.ClassINET),
		C.DNSServiceQueryRecordReply(C.goDNSServiceQueryRecordReply),
		unsafe.Pointer(&cContext))
	log.Println("queryDNS serviceErr:", serviceErr)
	if serviceErr != C.kDNSServiceErr_NoError {
		return nil, fmt.Errorf("failed to start DNS query: %v", serviceErr)
	}
	defer C.DNSServiceRefDeallocate(sdRef)

	// See https://developer.apple.com/documentation/dnssd/1804698-dnsservicerefsockfd
	fd := C.DNSServiceRefSockFD(sdRef)
	if fd < 0 {
		return nil, fmt.Errorf("failed to get DNSServiceRef file descriptor")
	}
	log.Println("For loop, fd:", fd)
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
			return &response, nil
		default:
		}

		log.Println("Poll")
		pollTimeout := int(-1)
		if deadline, ok := ctx.Deadline(); ok {
			timeout := time.Until(deadline)
			pollTimeout = int(timeout.Milliseconds())
		}
		nReady, err := unix.Poll([]unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN | unix.POLLERR | unix.POLLHUP}}, pollTimeout)
		if err != nil {
			return nil, err
		}
		if nReady == 0 {
			return nil, context.DeadlineExceeded
		}

		// See https://developer.apple.com/documentation/dnssd/1804696-dnsserviceprocessresult?language=objc.
		log.Println("DNSServiceProcessResult")
		serviceErr = C.DNSServiceProcessResult(sdRef)
		if serviceErr != C.kDNSServiceErr_NoError {
			return nil, fmt.Errorf("failed to process DNS response: %v", serviceErr)
		}
	}
}
