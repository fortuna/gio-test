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
*/
import "C"

import (
	"fmt"
	"runtime/cgo"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
)

//export goCallback
func goCallback(sdRef C.DNSServiceRef, flags C.DNSServiceFlags, interfaceIndex C.uint32_t,
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
