// Copyright 2014 Matt Martz <matt@sivel.net>
// All Rights Reserved.
//
//    Licensed under the Apache License, Version 2.0 (the "License"); you may
//    not use this file except in compliance with the License. You may obtain
//    a copy of the License at
//
//         http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
//    WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
//    License for the specific language governing permissions and limitations
//    under the License.

package main

import (
	"os"
	"fmt"
	"flag"
	"path"
	"net/http"
	"io/ioutil"
	"crypto/rsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"encoding/json"
	"encoding/base64"
)

type Key struct {
	Key string
}

func usage() {
	fmt.Printf(`usage: %s --repo=owner/name string
    --repo REPO           Repository slug (:owner/:name)
    string                String to encrypt
`, path.Base(os.Args[0]))
}

func main() {
	flag.Usage = usage
	var repo string
	flag.StringVar(&repo, "repo", "", "Repository slug (:owner/:name)")
	flag.Parse()
	if repo == "" {
		fmt.Println("ERROR: No --repo provided\n")
		usage()
		os.Exit(2)
	}
	keyurl := fmt.Sprintf("https://api.travis-ci.org/repos/%s/key", repo)

	stringToEncrypt := flag.Arg(0)
	if stringToEncrypt == "" {
		fmt.Println("ERROR: No string to encrypt\n")
		usage()
		os.Exit(2)
	}

	resp, err := http.Get(keyurl)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	var key Key
	json.Unmarshal(body, &key)

	block, _ := pem.Decode([]byte(key.Key))
	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		fmt.Printf("Failed to parse RSA public key: %s\n", err)
		os.Exit(2)
	}

	rsaPub, ok := pub.(*rsa.PublicKey)
	if !ok {
		fmt.Println("Value returned from ParsePKIXPublicKey was not an RSA public key")
		os.Exit(2)
	}

	encrypted, err := rsa.EncryptPKCS1v15(rand.Reader, rsaPub, []byte(stringToEncrypt))
	if err != nil {
		fmt.Println(err)
		os.Exit(2)
	}

	encoded := base64.StdEncoding.EncodeToString(encrypted)
	fmt.Printf("secure: \"%s\"\n", encoded)
}
