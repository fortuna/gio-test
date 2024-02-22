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

//go:build android

package sysresolver

/*
#include <stdlib.h>
#include <android/multinetwork.h>
*/
import "C"

import (
	"context"
	"time"
	"unsafe"

	"golang.org/x/net/dns/dnsmessage"
	"golang.org/x/sys/unix"
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
	cQname := C.CString(qname)
	defer C.free(unsafe.Pointer(cQname))
	// https://developer.android.com/ndk/reference/group/networking#android_res_nquery
	fd := C.android_res_nquery(0, cQname, C.int(dnsmessage.ClassINET), C.int(qtype), 0)
	if fd < 0 {
		return nil, unix.Errno(-fd)
	}
	timeout := int(-1)
	if deadline, ok := ctx.Deadline(); ok {
		timeout = int(time.Until(deadline).Milliseconds())
	}
	nReady, err := unix.Poll([]unix.PollFd{unix.PollFd{Fd: int32(fd), Events: unix.EPOLLIN | unix.EPOLLERR}}, timeout)
	if err != nil {
		return nil, err
	}
	if nReady == 0 {
		return nil, context.DeadlineExceeded
	}
	answer := make([]byte, 1500)
	var rcode C.int
	// https://developer.android.com/ndk/reference/group/networking#android_res_nresult
	n := C.android_res_nresult(fd, &rcode, (*C.uint8_t)(&answer[0]), C.size_t(len(answer)))
	if n < 0 {
		return nil, unix.Errno(-n)
	}
	answer = answer[:n]
	var msg dnsmessage.Message
	err = msg.Unpack(answer)
	if err != nil {
		return nil, err
	}
	return &msg, nil
}
