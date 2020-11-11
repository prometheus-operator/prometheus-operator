// Copyright 2020 The prometheus-operator Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package assets

// BasicAuthCredentials represents a username password pair to be used with
// basic http authentication, see https://tools.ietf.org/html/rfc7617.
type BasicAuthCredentials struct {
	Username string
	Password string
}

// BearerToken represents a bearer token, see
// https://tools.ietf.org/html/rfc6750.
type BearerToken string

// TLSAsset represents any TLS related opaque string, e.g. CA files, client
// certificates.
type TLSAsset string
