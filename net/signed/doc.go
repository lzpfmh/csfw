// Copyright 2015-2016, Cyrill @ Schumacher.fm and the CoreStore contributors
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

// Package signed (TODO) provides a middleware to sign responses and adds the signature
// to the header or trailer.
//
// With the use of HTTPS this package might not be needed, except theoretically
// MITM attacks ...
//
// https://tools.ietf.org/html/draft-cavage-http-signatures-00
// https://tools.ietf.org/html/draft-burke-content-signature-00
package signed
