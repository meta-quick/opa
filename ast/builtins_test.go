// Copyright 2017 The OPA Authors.  All rights reserved.
// Use of this source code is governed by an Apache2
// license that can be found in the LICENSE file.

package ast

import (
	"github.com/bytedance/sonic"
	"testing"

	"github.com/meta-quick/opa/types"
)

func TestBuiltinDeclRoundtrip(t *testing.T) {

	bs, err := sonic.Marshal(Plus)
	if err != nil {
		t.Fatal(err)
	}

	var cpy Builtin

	if err := sonic.Unmarshal(bs, &cpy); err != nil {
		t.Fatal(err)
	}

	if types.Compare(cpy.Decl, Plus.Decl) != 0 || cpy.Name != Plus.Name || cpy.Infix != Plus.Infix || cpy.Relation != Plus.Relation {
		t.Fatal("expected:", Plus, "got:", cpy)
	}
}
