// Copyright 2015 Google Inc. All rights reserved.
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

// Package triple implements and allows to manipulate Badwolf triples.
package triple

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/google/badwolf/triple/literal"
	"github.com/google/badwolf/triple/node"
	"github.com/google/badwolf/triple/predicate"
)

// Object is the box that either contains a literal or a node.
type Object struct {
	n *node.Node
	p *predicate.Predicate
	l *literal.Literal
}

// String pretty prints the object.
func (o *Object) String() string {
	if o.n != nil {
		return o.n.String()
	}
	if o.l != nil {
		return o.l.String()
	}
	if o.p != nil {
		return o.p.String()
	}
	return "@@@INVALID_OBJECT@@@"
}

// GUID returns a global unique identifier for the given object. It is
// implemented as the base64 encoded stringified version of the node.
func (o *Object) GUID() string {
	fo := "@@@INVALID_OBJECT@@@"
	if o.n != nil {
		fo = "node"
	}
	if o.l != nil {
		fo = "literal"
	}
	if o.p != nil {
		fo = "predicate"
	}
	return base64.StdEncoding.EncodeToString([]byte(strings.Join([]string{fo, o.String()}, ":")))
}

// Node attempts to the return the boxed node.
func (o *Object) Node() (*node.Node, error) {
	if o.n == nil {
		return nil, fmt.Errorf("triple.Literal does not box a node in %s", o)
	}
	return o.n, nil
}

// Predicate attempts to the return the boxed predicate.
func (o *Object) Predicate() (*predicate.Predicate, error) {
	if o.p == nil {
		return nil, fmt.Errorf("triple.Literal does not box a predicate in %s", o)
	}
	return o.p, nil
}

// Literal attempts to the return the boxed literal.
func (o *Object) Literal() (*literal.Literal, error) {
	if o.l == nil {
		return nil, fmt.Errorf("triple.Literal does not box a literal in %s", o)
	}
	return o.l, nil
}

// ParseObject attempts to parse and object.
func ParseObject(s string, b literal.Builder) (*Object, error) {
	n, err := node.Parse(s)
	if err != nil {
		l, err := b.Parse(s)
		if err != nil {
			o, err := predicate.Parse(s)
			if err != nil {
				return nil, err
			}
			return NewPredicateObject(o), nil
		}
		return NewLiteralObject(l), nil
	}
	return NewNodeObject(n), nil
}

// NewNodeObject returns a new object that boxes a node.
func NewNodeObject(n *node.Node) *Object {
	return &Object{
		n: n,
	}
}

// NewPredicateObject returns a new object that boxes a predicate.
func NewPredicateObject(p *predicate.Predicate) *Object {
	return &Object{
		p: p,
	}
}

// NewLiteralObject returns a new object that boxes a literal.
func NewLiteralObject(l *literal.Literal) *Object {
	return &Object{
		l: l,
	}
}

// Triple describes a the <subject predicate object> used by BadWolf.
type Triple struct {
	s *node.Node
	p *predicate.Predicate
	o *Object
}

// NewTriple creates a new triple.
func NewTriple(s *node.Node, p *predicate.Predicate, o *Object) (*Triple, error) {
	if s == nil || p == nil || o == nil {
		return nil, fmt.Errorf("triple.NewTriple cannot create triples from nil components in <%v %v %v>", s, p, o)
	}
	return &Triple{
		s: s,
		p: p,
		o: o,
	}, nil
}

// S returns the subject of the triple.
func (t *Triple) S() *node.Node {
	return t.s
}

// P returns the predicate of the triple.
func (t *Triple) P() *predicate.Predicate {
	return t.p
}

// O returns the object of the tirple.
func (t *Triple) O() *Object {
	return t.o
}

// String marshals the triple into pretty string.
func (t *Triple) String() string {
	return fmt.Sprintf("%s\t%s\t%s", t.s, t.p, t.o)
}

var (
	pSplit *regexp.Regexp
	oSplit *regexp.Regexp
)

func init() {
	pSplit = regexp.MustCompile(">\\s+\"")
	oSplit = regexp.MustCompile("(]\\s+/)|(]\\s+\")")
}

// ParseTriple process the provided text and tries to create a triple. It asumes
// that the provided text contains only one triple.
func ParseTriple(line string, b literal.Builder) (*Triple, error) {
	raw := strings.TrimSpace(line)
	idxp := pSplit.FindIndex([]byte(raw))
	idxo := oSplit.FindIndex([]byte(raw))
	if len(idxp) == 0 || len(idxo) == 0 {
		return nil, fmt.Errorf("triple.Parse could not split s p o  out of %s", raw)
	}
	ss, sp, so := raw[0:idxp[0]+1], raw[idxp[1]-1:idxo[0]+1], raw[idxo[1]-1:]
	s, err := node.Parse(ss)
	if err != nil {
		return nil, fmt.Errorf("triple.Parse failed to parse subject %s with error %v", ss, err)
	}
	p, err := predicate.Parse(sp)
	if err != nil {
		return nil, fmt.Errorf("triple.Parse failed to parse predicate %s with error %v", sp, err)
	}
	o, err := ParseObject(so, b)
	if err != nil {
		return nil, fmt.Errorf("triple.Parse failed to parse object %s with error %v", so, err)
	}
	return NewTriple(s, p, o)
}

// Reify given the current triple it returns the original triple and the newly
// reified ones. It also returns the newly created blank node.
func (t *Triple) Reify() ([]*Triple, *node.Node, error) {
	// Function that create the proper reification predicates.
	rp := func(id string, p *predicate.Predicate) (*predicate.Predicate, error) {
		if p.Type() == predicate.Temporal {
			ta, _ := p.TimeAnchor()
			return predicate.NewTemporal(string(p.ID()), *ta)
		}
		return predicate.NewImmutable(id)
	}

	fmt.Println(t.String())
	b := node.NewBlankNode()
	s, err := rp("_subject", t.p)
	if err != nil {
		return nil, nil, err
	}
	ts, _ := NewTriple(b, s, NewNodeObject(t.s))
	p, err := rp("_predicate", t.p)
	if err != nil {
		return nil, nil, err
	}
	tp, _ := NewTriple(b, p, NewPredicateObject(t.p))
	var to *Triple
	if t.o.l != nil {
		o, err := rp("_object", t.p)
		if err != nil {
			return nil, nil, err
		}
		to, _ = NewTriple(b, o, NewLiteralObject(t.o.l))
	}
	if t.o.n != nil {
		o, err := rp("_object", t.p)
		if err != nil {
			return nil, nil, err
		}
		to, _ = NewTriple(b, o, NewNodeObject(t.o.n))
	}
	if t.o.p != nil {
		o, err := rp("_object", t.p)
		if err != nil {
			return nil, nil, err
		}
		to, _ = NewTriple(b, o, NewPredicateObject(t.o.p))
	}

	return []*Triple{t, ts, tp, to}, b, nil
}

// GUID returns a global unique identifier for the given triple. It is
// implemented as the base64 encoded stringified version of the triple.
func (t *Triple) GUID() string {
	return base64.StdEncoding.EncodeToString([]byte(t.String()))
}
