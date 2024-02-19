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

//go:build linux && !android

package main

/*
#include <stdlib.h>
#include <resolv.h>
*/
import "C"

import (
	"context"
	"fmt"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
)

func queryCNAME(ctx context.Context, qname string) (string, error) {
	cQname := C.CString(qname)
	defer C.free(unsafe.Pointer(cQname))
	answer := make([]byte, 1500)
	// This is blocking and returns -1 on DNS errors.
	n := C.res_query(cQname, C.int(dnsmessage.ClassINET), C.int(dnsmessage.TypeCNAME),
		(*C.uint8_t)(&answer[0]), C.int(len(answer)))
	if n < 0 {
		return "", fmt.Errorf("failed to query DNS: %v", n)
	}
	answer = answer[:n]
	var msg dnsmessage.Message
	err := msg.Unpack(answer)
	if err != nil {
		return "", err
	}
	return msg.GoString(), nil
}
