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

package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fortuna/gio-test/sysresolver"
	"golang.org/x/net/dns/dnsmessage"
)

var debugLog log.Logger = *log.New(io.Discard, "", 0)

func init() {
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s [flags...] <domain>\n", path.Base(os.Args[0]))
		flag.PrintDefaults()
	}
}

func rcodeToString(rcode dnsmessage.RCode) string {
	rcodeStr, _ := strings.CutPrefix(strings.ToUpper(rcode.String()), "RCODE")
	return rcodeStr
}

func main() {
	verboseFlag := flag.Bool("v", false, "Enable debug output")
	typeFlag := flag.String("type", "A", "The type of the query (A, AAAA, CNAME, NS or TXT).")
	timeoutFlag := flag.Int("timeout", 3, "Timeout in seconds")

	flag.Parse()
	if *verboseFlag {
		debugLog = *log.New(os.Stderr, "[DEBUG] ", log.LstdFlags|log.Lmicroseconds|log.Lshortfile)
	}

	domain := strings.TrimSpace(flag.Arg(0))
	if domain == "" {
		log.Fatal("Need to pass the domain to resolve in the command-line")
	}

	var qtype dnsmessage.Type
	switch strings.ToUpper(*typeFlag) {
	case "A":
		qtype = dnsmessage.TypeA
	case "AAAA":
		qtype = dnsmessage.TypeAAAA
	case "CNAME":
		qtype = dnsmessage.TypeCNAME
	case "NS":
		qtype = dnsmessage.TypeNS
	case "SOA":
		qtype = dnsmessage.TypeSOA
	case "TXT":
		qtype = dnsmessage.TypeTXT
	default:
		log.Fatalf("Unsupported query type %v", *typeFlag)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeoutFlag)*time.Second)
	defer cancel()
	response, err := sysresolver.Query(ctx, domain, qtype)

	if err != nil {
		log.Fatalf("Query failed: %v", err)
	}
	fmt.Println(response)

	// if response.RCode != dnsmessage.RCodeSuccess {
	// 	log.Fatalf("Got response code %v", rcodeToString(response.RCode))
	// }
	// debugLog.Println(response.GoString())
	// for _, answer := range response.Answers {
	// 	if answer.Header.Type != qtype {
	// 		continue
	// 	}
	// 	switch answer.Header.Type {
	// 	case dnsmessage.TypeA:
	// 		fmt.Println(net.IP(answer.Body.(*dnsmessage.AResource).A[:]))
	// 	case dnsmessage.TypeAAAA:
	// 		fmt.Println(net.IP(answer.Body.(*dnsmessage.AAAAResource).AAAA[:]))
	// 	case dnsmessage.TypeCNAME:
	// 		fmt.Println(answer.Body.(*dnsmessage.CNAMEResource).CNAME.String())
	// 	case dnsmessage.TypeNS:
	// 		fmt.Println(answer.Body.(*dnsmessage.NSResource).NS.String())
	// 	case dnsmessage.TypeSOA:
	// 		soa := answer.Body.(*dnsmessage.SOAResource)
	// 		fmt.Printf("ns: %v email: %v minTTL: %v\n", soa.NS, soa.MBox, soa.MinTTL)
	// 	case dnsmessage.TypeTXT:
	// 		fmt.Println(strings.Join(answer.Body.(*dnsmessage.TXTResource).TXT, ", "))
	// 	default:
	// 		fmt.Println(answer.Body.GoString())
	// 	}
	// }
}
