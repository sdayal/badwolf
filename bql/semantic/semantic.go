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

// Package semantic contains the semantic analysis required to have a
// senantically valid parser. It includes the data conversion required to
// turn tokens into valid BadWolf structures. It also provides the hooks
// implementations required for buliding an actionable execution plan.
package semantic

import (
	"sort"
	"time"

	"github.com/google/badwolf/triple"
	"github.com/google/badwolf/triple/node"
	"github.com/google/badwolf/triple/predicate"
)

// StatementType describes the type of statement being represented.
type StatementType int8

const (
	// Query statement.
	Query StatementType = iota
	// Insert statemrnt.
	Insert
	// Delete statement.
	Delete
	// Create statement.
	Create
	// Drop statement.
	Drop
)

// String provides a readable version of the StatementType.
func (t StatementType) String() string {
	switch t {
	case Query:
		return "QUERY"
	case Insert:
		return "INSERT"
	case Delete:
		return "DELETE"
	case Create:
		return "CREATE"
	case Drop:
		return "DROP"
	default:
		return "UNKNOWN"
	}
}

// Statement contains all the semantic information extract from the parsing
type Statement struct {
	sType         StatementType
	graphs        []string
	data          []*triple.Triple
	pattern       []*GraphClause
	workingClause *GraphClause
}

// GraphClause represents a clause of a graph pattern in a where clause.
type GraphClause struct {
	S                *node.Node
	SBinding         string
	SAlias           string
	STypeAlias       string
	SIDAlias         string
	P                *predicate.Predicate
	PAlias           string
	PID              string
	PAnchorBinding   string
	PBinding         string
	PLowerBound      *time.Time
	PUpperBound      *time.Time
	PLowerBoundAlias string
	PUpperBoundAlias string
	PIDAlias         string
	PAnchorAlias     string
	O                *triple.Object
	OBinding         string
	OID              string
	OAlias           string
	OTypeAlias       string
	OIDAlias         string
	OAnchorAlias     string
	OAnchorBinding   string
	OLowerBound      *time.Time
	OUpperBound      *time.Time
	OLowerBoundAlias string
	OUpperBoundAlias string
}

// Specificity return
func (c *GraphClause) Specificity() int {
	s := 0
	if c.S != nil {
		s++
	}
	if c.P != nil {
		s++
	}
	if c.O != nil {
		s++
	}
	return s
}

// BindType set he type of a statement.
func (s *Statement) BindType(st StatementType) {
	s.sType = st
}

// Type returns the type of the statement.
func (s *Statement) Type() StatementType {
	return s.sType
}

// AddGraph adds a graph to a given https://critique.corp.google.com/#review/101398527statement.
func (s *Statement) AddGraph(g string) {
	s.graphs = append(s.graphs, g)
}

// Graphs returns the list of graphs listed on the statement.
func (s *Statement) Graphs() []string {
	return s.graphs
}

// AddData adds a triple to a given statement's data.
func (s *Statement) AddData(d *triple.Triple) {
	s.data = append(s.data, d)
}

// Data returns the data available for the given statement.
func (s *Statement) Data() []*triple.Triple {
	return s.data
}

// GraphPatternClauses return the list of graph pattern clauses
func (s *Statement) GraphPatternClauses() []*GraphClause {
	return s.pattern
}

// ResetWorkingGraphClause resets the current working graph clause.
func (s *Statement) ResetWorkingGraphClause() {
	s.workingClause = &GraphClause{}
}

// WorkingClause returns the current working clause.
func (s *Statement) WorkingClause() *GraphClause {
	return s.workingClause
}

// AddWorkingGrpahClause add the current working graph clause to the set of
// clauses that form the graph pattern.
func (s *Statement) AddWorkingGrpahClause() {
	s.pattern = append(s.pattern, s.workingClause)
	s.ResetWorkingGraphClause()
}

// addtoBindings add the binding if not empty.
func addToBindings(bs map[string]int, b string) {
	if b != "" {
		bs[b]++
	}
}

// Bindings retuns the set of bindings available on the graph clauses for he
// statement.
func (s *Statement) Bindings() map[string]int {
	bs := make(map[string]int)
	for _, cls := range s.pattern {
		addToBindings(bs, cls.SBinding)
		addToBindings(bs, cls.SAlias)
		addToBindings(bs, cls.STypeAlias)
		addToBindings(bs, cls.SIDAlias)
		addToBindings(bs, cls.PAlias)
		addToBindings(bs, cls.PAnchorBinding)
		addToBindings(bs, cls.PBinding)
		addToBindings(bs, cls.PLowerBoundAlias)
		addToBindings(bs, cls.PUpperBoundAlias)
		addToBindings(bs, cls.PIDAlias)
		addToBindings(bs, cls.PAnchorAlias)
		addToBindings(bs, cls.OBinding)
		addToBindings(bs, cls.OID)
		addToBindings(bs, cls.OAlias)
		addToBindings(bs, cls.OTypeAlias)
		addToBindings(bs, cls.OIDAlias)
		addToBindings(bs, cls.OAnchorAlias)
		addToBindings(bs, cls.OAnchorBinding)
		addToBindings(bs, cls.OLowerBoundAlias)
		addToBindings(bs, cls.OUpperBoundAlias)
	}
	return bs
}

// bySpecificity type helps sort clauses by Specificity.
type bySpecificity []*GraphClause

// Len returns the lenght of the clauses array.
func (s bySpecificity) Len() int {
	return len(s)
}

// Swap exchange the i and j elements in the clauses array.
func (s bySpecificity) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Less returns true if the i element is less specific than j one.
func (s bySpecificity) Less(i, j int) bool {
	return s[i].Specificity() > s[j].Specificity()
}

// SortedGraphPatternClauses return the list of graph pattern clauses
func (s *Statement) SortedGraphPatternClauses() []*GraphClause {
	ptrns := s.pattern
	sort.Sort(bySpecificity(ptrns))
	return ptrns
}
