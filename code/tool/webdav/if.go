// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

// The If header is covered by Section 10.4.
// http://www.webdav.org/specs/rfc4918.html#HEADER_If

import (
	"strings"
)

// IfHeader is a disjunction (OR) of ifLists.
type IfHeader struct {
	Lists []IfList
}

// IfList is a conjunction (AND) of Conditions, and an optional resource tag.
type IfList struct {
	ResourceTag string
	Conditions  []Condition
}

// ParseIfHeader parses the "If: foo bar" HTTP header. The httpHeader string
// should omit the "If:" prefix and have any "\r\n"s collapsed to a " ", as is
// returned by req.Header.Get("If") for a http.Request req.
func ParseIfHeader(httpHeader string) (h IfHeader, ok bool) {
	s := strings.TrimSpace(httpHeader)
	switch tokenType, _, _ := lex(s); tokenType {
	case '(':
		return ParseNoTagLists(s)
	case AngleTokenType:
		return ParseTaggedLists(s)
	default:
		return IfHeader{}, false
	}
}

func ParseNoTagLists(s string) (h IfHeader, ok bool) {
	for {
		l, remaining, ok := ParseList(s)
		if !ok {
			return IfHeader{}, false
		}
		h.Lists = append(h.Lists, l)
		if remaining == "" {
			return h, true
		}
		s = remaining
	}
}

func ParseTaggedLists(s string) (h IfHeader, ok bool) {
	resourceTag, n := "", 0
	for first := true; ; first = false {
		tokenType, tokenStr, remaining := lex(s)
		switch tokenType {
		case AngleTokenType:
			if !first && n == 0 {
				return IfHeader{}, false
			}
			resourceTag, n = tokenStr, 0
			s = remaining
		case '(':
			n++
			l, remaining, ok := ParseList(s)
			if !ok {
				return IfHeader{}, false
			}
			l.ResourceTag = resourceTag
			h.Lists = append(h.Lists, l)
			if remaining == "" {
				return h, true
			}
			s = remaining
		default:
			return IfHeader{}, false
		}
	}
}

func ParseList(s string) (l IfList, remaining string, ok bool) {
	tokenType, _, s := lex(s)
	if tokenType != '(' {
		return IfList{}, "", false
	}
	for {
		tokenType, _, remaining = lex(s)
		if tokenType == ')' {
			if len(l.Conditions) == 0 {
				return IfList{}, "", false
			}
			return l, remaining, true
		}
		c, remaining, ok := ParseCondition(s)
		if !ok {
			return IfList{}, "", false
		}
		l.Conditions = append(l.Conditions, c)
		s = remaining
	}
}

func ParseCondition(s string) (c Condition, remaining string, ok bool) {
	tokenType, tokenStr, s := lex(s)
	if tokenType == NotTokenType {
		c.Not = true
		tokenType, tokenStr, s = lex(s)
	}
	switch tokenType {
	case StrTokenType, AngleTokenType:
		c.Token = tokenStr
	case SquareTokenType:
		c.ETag = tokenStr
	default:
		return Condition{}, "", false
	}
	return c, s, true
}

// Single-rune tokens like '(' or ')' have a token type equal to their rune.
// All other tokens have a negative token type.
const (
	ErrTokenType    = rune(-1)
	EofTokenType    = rune(-2)
	StrTokenType    = rune(-3)
	NotTokenType    = rune(-4)
	AngleTokenType  = rune(-5)
	SquareTokenType = rune(-6)
)

func lex(s string) (tokenType rune, tokenStr string, remaining string) {
	// The net/textproto Reader that parses the HTTP header will collapse
	// Linear White Space that spans multiple "\r\n" lines to a single " ",
	// so we don't need to look for '\r' or '\n'.
	for len(s) > 0 && (s[0] == '\t' || s[0] == ' ') {
		s = s[1:]
	}
	if len(s) == 0 {
		return EofTokenType, "", ""
	}
	i := 0
loop:
	for ; i < len(s); i++ {
		switch s[i] {
		case '\t', ' ', '(', ')', '<', '>', '[', ']':
			break loop
		}
	}

	if i != 0 {
		tokenStr, remaining = s[:i], s[i:]
		if tokenStr == "Not" {
			return NotTokenType, "", remaining
		}
		return StrTokenType, tokenStr, remaining
	}

	j := 0
	switch s[0] {
	case '<':
		j, tokenType = strings.IndexByte(s, '>'), AngleTokenType
	case '[':
		j, tokenType = strings.IndexByte(s, ']'), SquareTokenType
	default:
		return rune(s[0]), "", s[1:]
	}
	if j < 0 {
		return ErrTokenType, "", ""
	}
	return tokenType, s[1:j], s[j+1:]
}
