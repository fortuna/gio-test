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

package sysresolver

import (
	"fmt"
	"net"
	"net/netip"
	"strings"

	"golang.org/x/net/dns/dnsmessage"
)

func formatIPs(ips []net.IP) string {
	if ips == nil {
		return ""
	}
	result := ""
	for _, ip := range ips {
		result += ip.String() + "\n"
	}
	return result
}

func formatMessage(msg *dnsmessage.Message) string {
	if msg == nil {
		return ""
	}
	return formatAnswers(msg.Answers)
}

func formatAnswers(answers []dnsmessage.Resource) string {
	if answers == nil {
		return ""
	}
	result := ""
	for _, answer := range answers {
		result += fmt.Sprintf("%v %v %d %v\n",
			answer.Header.Name,
			strings.TrimPrefix(answer.Header.Type.String(), "Type"),
			answer.Header.TTL,
			formatResourceBody(answer.Body))
	}
	return result
}

func formatResourceBody(rrbody dnsmessage.ResourceBody) string {
	switch b := rrbody.(type) {
	case *dnsmessage.AResource:
		return netip.AddrFrom4(b.A).String()
	case *dnsmessage.AAAAResource:
		return netip.AddrFrom16(b.AAAA).String()
	case *dnsmessage.CNAMEResource:
		return b.CNAME.String()
	default:
		return b.GoString()
	}
}
