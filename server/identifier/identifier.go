// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

// Package identifier provides handlers for associating identity information with incoming requests.
package identifier

import (
	"context"
	"net/http"
)

type identityKey string

const identity = identityKey("org.openpolicyagent/identity")

// Identity returns the identity of the caller associated with ctx.
func Identity(r *http.Request) (string, bool) {
	ctx := r.Context()
	v, ok := ctx.Value(identity).(string)
	if ok {
		return v, true
	}
	return "", false
}

// SetIdentity returns a new http.Request with the identity set to v.
func SetIdentity(r *http.Request, v string) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), identity, v))
}
