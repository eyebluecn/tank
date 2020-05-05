// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package webdav

import (
	"reflect"
	"strings"
	"testing"
)

func TestParseIfHeader(t *testing.T) {
	// The "section x.y.z" test cases come from section x.y.z of the spec at
	// http://www.webdav.org/specs/rfc4918.html
	testCases := []struct {
		desc  string
		input string
		want  IfHeader
	}{{
		"bad: empty",
		``,
		IfHeader{},
	}, {
		"bad: no parens",
		`foobar`,
		IfHeader{},
	}, {
		"bad: empty list #1",
		`()`,
		IfHeader{},
	}, {
		"bad: empty list #2",
		`(a) (b c) () (d)`,
		IfHeader{},
	}, {
		"bad: no list after resource #1",
		`<foo>`,
		IfHeader{},
	}, {
		"bad: no list after resource #2",
		`<foo> <bar> (a)`,
		IfHeader{},
	}, {
		"bad: no list after resource #3",
		`<foo> (a) (b) <bar>`,
		IfHeader{},
	}, {
		"bad: no-tag-list followed by tagged-list",
		`(a) (b) <foo> (c)`,
		IfHeader{},
	}, {
		"bad: unfinished list",
		`(a`,
		IfHeader{},
	}, {
		"bad: unfinished ETag",
		`([b`,
		IfHeader{},
	}, {
		"bad: unfinished Notted list",
		`(Not a`,
		IfHeader{},
	}, {
		"bad: double Not",
		`(Not Not a)`,
		IfHeader{},
	}, {
		"good: one list with a Token",
		`(a)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `a`,
				}},
			}},
		},
	}, {
		"good: one list with an ETag",
		`([a])`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					ETag: `a`,
				}},
			}},
		},
	}, {
		"good: one list with three Nots",
		`(Not a Not b Not [d])`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Not:   true,
					Token: `a`,
				}, {
					Not:   true,
					Token: `b`,
				}, {
					Not:  true,
					ETag: `d`,
				}},
			}},
		},
	}, {
		"good: two lists",
		`(a) (b)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `a`,
				}},
			}, {
				Conditions: []Condition{{
					Token: `b`,
				}},
			}},
		},
	}, {
		"good: two Notted lists",
		`(Not a) (Not b)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Not:   true,
					Token: `a`,
				}},
			}, {
				Conditions: []Condition{{
					Not:   true,
					Token: `b`,
				}},
			}},
		},
	}, {
		"section 7.5.1",
		`<http://www.example.com/users/f/fielding/index.html> 
			(<urn:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6>)`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `http://www.example.com/users/f/fielding/index.html`,
				Conditions: []Condition{{
					Token: `urn:uuid:f81d4fae-7dec-11d0-a765-00a0c91e6bf6`,
				}},
			}},
		},
	}, {
		"section 7.5.2 #1",
		`(<urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf>)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf`,
				}},
			}},
		},
	}, {
		"section 7.5.2 #2",
		`<http://example.com/locked/>
			(<urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf>)`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `http://example.com/locked/`,
				Conditions: []Condition{{
					Token: `urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf`,
				}},
			}},
		},
	}, {
		"section 7.5.2 #3",
		`<http://example.com/locked/member>
			(<urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf>)`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `http://example.com/locked/member`,
				Conditions: []Condition{{
					Token: `urn:uuid:150852e2-3847-42d5-8cbe-0f4f296f26cf`,
				}},
			}},
		},
	}, {
		"section 9.9.6",
		`(<urn:uuid:fe184f2e-6eec-41d0-c765-01adc56e6bb4>) 
			(<urn:uuid:e454f3f3-acdc-452a-56c7-00a5c91e4b77>)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `urn:uuid:fe184f2e-6eec-41d0-c765-01adc56e6bb4`,
				}},
			}, {
				Conditions: []Condition{{
					Token: `urn:uuid:e454f3f3-acdc-452a-56c7-00a5c91e4b77`,
				}},
			}},
		},
	}, {
		"section 9.10.8",
		`(<urn:uuid:e71d4fae-5dec-22d6-fea5-00a0c91e6be4>)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `urn:uuid:e71d4fae-5dec-22d6-fea5-00a0c91e6be4`,
				}},
			}},
		},
	}, {
		"section 10.4.6",
		`(<urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2> 
			["I am an ETag"])
			(["I am another ETag"])`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2`,
				}, {
					ETag: `"I am an ETag"`,
				}},
			}, {
				Conditions: []Condition{{
					ETag: `"I am another ETag"`,
				}},
			}},
		},
	}, {
		"section 10.4.7",
		`(Not <urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2> 
			<urn:uuid:58f202ac-22cf-11d1-b12d-002035b29092>)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Not:   true,
					Token: `urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2`,
				}, {
					Token: `urn:uuid:58f202ac-22cf-11d1-b12d-002035b29092`,
				}},
			}},
		},
	}, {
		"section 10.4.8",
		`(<urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2>) 
			(Not <DAV:no-lock>)`,
		IfHeader{
			Lists: []IfList{{
				Conditions: []Condition{{
					Token: `urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2`,
				}},
			}, {
				Conditions: []Condition{{
					Not:   true,
					Token: `DAV:no-lock`,
				}},
			}},
		},
	}, {
		"section 10.4.9",
		`</resource1> 
			(<urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2> 
			[W/"A weak ETag"]) (["strong ETag"])`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `/resource1`,
				Conditions: []Condition{{
					Token: `urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2`,
				}, {
					ETag: `W/"A weak ETag"`,
				}},
			}, {
				ResourceTag: `/resource1`,
				Conditions: []Condition{{
					ETag: `"strong ETag"`,
				}},
			}},
		},
	}, {
		"section 10.4.10",
		`<http://www.example.com/specs/> 
			(<urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2>)`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `http://www.example.com/specs/`,
				Conditions: []Condition{{
					Token: `urn:uuid:181d4fae-7d8c-11d0-a765-00a0c91e6bf2`,
				}},
			}},
		},
	}, {
		"section 10.4.11 #1",
		`</specs/rfc2518.doc> (["4217"])`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `/specs/rfc2518.doc`,
				Conditions: []Condition{{
					ETag: `"4217"`,
				}},
			}},
		},
	}, {
		"section 10.4.11 #2",
		`</specs/rfc2518.doc> (Not ["4217"])`,
		IfHeader{
			Lists: []IfList{{
				ResourceTag: `/specs/rfc2518.doc`,
				Conditions: []Condition{{
					Not:  true,
					ETag: `"4217"`,
				}},
			}},
		},
	}}

	for _, tc := range testCases {
		got, ok := ParseIfHeader(strings.Replace(tc.input, "\n", "", -1))
		if gotEmpty := reflect.DeepEqual(got, IfHeader{}); gotEmpty == ok {
			t.Errorf("%s: should be different: empty header == %t, ok == %t", tc.desc, gotEmpty, ok)
			continue
		}
		if !reflect.DeepEqual(got, tc.want) {
			t.Errorf("%s:\ngot  %v\nwant %v", tc.desc, got, tc.want)
			continue
		}
	}
}
