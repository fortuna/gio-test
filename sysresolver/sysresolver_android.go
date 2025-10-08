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

import (
	"context"

	"github.com/fortuna/gio-test/androidresolver"
	"golang.org/x/net/dns/dnsmessage"
)

func Query(ctx context.Context, qname string, qtype dnsmessage.Type) (string, error) {
	resolver := &androidresolver.Resolver{}
	q, err := dnsmessage.NewName(qname)
	if err != nil {
		return "", err
	}
	msg, err := resolver.Query(ctx, dnsmessage.Question{Name: q, Type: qtype, Class: dnsmessage.ClassINET})
	if err != nil {
		return "", err
	}
	return formatMessage(msg), nil
}
