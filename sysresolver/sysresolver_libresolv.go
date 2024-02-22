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

//go:build unix && libresolv

// Uses libresolv.
//
// Benefits:
// - Returns a full message, though only on success.
//
// Issues:
// - Doesn't work on mobile
// - Doesn't leverage system cache
// - Blocking, no way to pass a timeout

package sysresolver

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

func Query(ctx context.Context, qname string, qtype dnsmessage.Type) (string, error) {
	msg, err := queryMsg(ctx, qname, qtype)
	var text string
	if msg != nil {
		text = formatMessage(msg)
	}
	return text, err
}

func queryMsg(ctx context.Context, qname string, qtype dnsmessage.Type) (*dnsmessage.Message, error) {
	type Result struct {
		msg *dnsmessage.Message
		err error
	}
	resultCh := make(chan Result, 1)
	go func() {
		cQname := C.CString(qname)
		defer C.free(unsafe.Pointer(cQname))
		answer := make([]byte, 1500)
		// This is blocking and returns -1 on DNS errors.
		n := C.res_query(cQname, C.int(dnsmessage.ClassINET), C.int(qtype),
			(*C.uint8_t)(&answer[0]), C.int(len(answer)))
		if n < 0 {
			resultCh <- Result{nil, fmt.Errorf("failed to query DNS: %v", n)}
			return
		}
		answer = answer[:n]
		var msg dnsmessage.Message
		err := msg.Unpack(answer)
		if err != nil {
			resultCh <- Result{nil, err}
			return
		}
		resultCh <- Result{&msg, nil}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case res := <-resultCh:
		return res.msg, res.err
	}
}
