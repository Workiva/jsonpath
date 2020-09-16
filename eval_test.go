package jsonpath

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test struct {
	name     string
	json     string
	path     string
	expected []Result
}

var tests = []test{
	test{`key selection with back to back quotes`, `{"aKey":"32\""}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `"32\""`, JsonString, `aKey`)}},
	test{`key selection with double backslashes at the end`, `{"bkey": "31\\", "aKey":"32"}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `"32"`, JsonString, `aKey`)}},
	test{`key selection`, `{"aKey":32}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `32`, JsonNumber, `aKey`)}},
	test{`nested key selection`, `{"aKey":{"bKey":32}}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `{"bKey":32}`, JsonObject, `aKey`)}},
	test{`empty array`, `{"aKey":[]}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `[]`, JsonArray, `aKey`)}},
	test{`multiple same-level keys, weird spacing`, `{    "aKey" 	: true ,    "bKey":  [	1 , 2	], "cKey" 	: true		} `, `$.bKey+`, []Result{newResult(`$.bKey+`, `[1,2]`, JsonArray, `bKey`)}},

	test{`array index selection`, `{"aKey":[123,456]}`, `$.aKey[1]+`, []Result{newResult(`$.aKey[1]+`, `456`, JsonNumber, `aKey`, 1)}},
	test{`array wild index selection`, `{"aKey":[123,456]}`, `$.aKey[*]+`, []Result{newResult(`$.aKey[*]+`, `123`, JsonNumber, `aKey`, 0), newResult(`$.aKey[*]+`, `456`, JsonNumber, `aKey`, 1)}},
	test{`array range index selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:3]+`, []Result{newResult(`$.aKey[1:3]+`, `22`, JsonNumber, `aKey`, 1), newResult(`$.aKey[1:3]+`, `33`, JsonNumber, `aKey`, 2)}},
	test{`array range (no index) selection`, `{"aKey":[11,22,33,44]}`, `$.aKey[1:1]+`, []Result{}},
	test{`array range (no upper bound) selection`, `{"aKey":[11,22,33]}`, `$.aKey[1:]+`, []Result{newResult(`$.aKey[1:]+`, `22`, JsonNumber, `aKey`, 1), newResult(`$.aKey[1:]+`, `33`, JsonNumber, `aKey`, 2)}},

	test{`empty array - try selection`, `{"aKey":[]}`, `$.aKey[1]+`, []Result{}},
	test{`null selection`, `{"aKey":[null]}`, `$.aKey[0]+`, []Result{newResult(`$.aKey[0]+`, `null`, JsonNull, `aKey`, 0)}},
	test{`empty object`, `{"aKey":{}}`, `$.aKey+`, []Result{newResult(`$.aKey+`, `{}`, JsonObject, `aKey`)}},
	test{`object w/ height=2`, `{"aKey":{"bKey":32}}`, `$.aKey.bKey+`, []Result{newResult(`$.aKey.bKey+`, `32`, JsonNumber, `aKey`, `bKey`)}},
	test{`array of multiple types`, `{"aKey":[1,{"s":true},"asdf"]}`, `$.aKey[1]+`, []Result{newResult(`$.aKey[1]+`, `{"s":true}`, JsonObject, `aKey`, 1)}},
	test{`nested array selection`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey+`, []Result{newResult(`$.aKey.bKey+`, `[123,456]`, JsonArray, `aKey`, `bKey`)}},
	test{`nested array`, `[[[[[]], [true, false, []]]]]`, `$[0][0][1][2]+`, []Result{newResult(`$[0][0][1][2]+`, `[]`, JsonArray, 0, 0, 1, 2)}},
	test{`index of array selection`, `{"aKey":{"bKey":[123, 456, 789]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`$.aKey.bKey[1]+`, `456`, JsonNumber, `aKey`, `bKey`, 1)}},
	test{`index of array selection (more than one)`, `{"aKey":{"bKey":[123,456]}}`, `$.aKey.bKey[1]+`, []Result{newResult(`$.aKey.bKey[1]+`, `456`, JsonNumber, `aKey`, `bKey`, 1)}},
	test{`multi-level object/array`, `{"1Key":{"aKey": null, "bKey":{"trash":[1,2]}, "cKey":[123,456] }, "2Key":false}`, `$.1Key.bKey.trash[0]+`, []Result{newResult(`$.1Key.bKey.trash[0]+`, `1`, JsonNumber, `1Key`, `bKey`, `trash`, 0)}},
	test{`multi-level array`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3"] ]}`, `$.*[*].michael[1]+`, []Result{newResult(`$.*[*].michael[1]+`, `6`, JsonNumber, `aKey`, 3, `michael`, 1)}},
	test{`multi-level array 2`, `{"aKey":[true,false,null,{"michael":[5,6,7]}, ["s", "3\""] ]}`, `$.*[*][1]+`, []Result{newResult(`$.*[*][1]+`, `"3\""`, JsonString, `aKey`, 4, 1)}},

	test{`evaluation literal equality`, `{"items":[ {"name":"alpha", "value":11}]}`, `$.items[*]?("bravo" == "bravo").value+`, []Result{newResult(`$.items[*]?("bravo" == "bravo").value+`, `11`, JsonNumber, `items`, 0, `value`)}},
	test{`evaluation based on string equal to path value`, `{"items":[ {"name":"alpha", "value":11}, {"name":"bravo", "value":22}, {"name":"charlie", "value":33} ]}`, `$.items[*]?(@.name == "bravo").value+`, []Result{newResult(`$.items[*]?(@.name == "bravo").value+`, `22`, JsonNumber, `items`, 1, `value`)}},
}

func TestPathQuery(t *testing.T) {
	as := assert.New(t)

	for _, t := range tests {
		paths, err := ParsePaths(t.path)
		if as.NoError(err) {
			eval, err := EvalPathsInBytes([]byte(t.json), paths)
			if as.NoError(err, "Testing: %s", t.name) {
				res := toResultArray(eval)
				if as.NoError(eval.Error) {
					as.EqualValues(t.expected, res, "Testing of %q", t.name)
				}
			}

			eval_reader, err := EvalPathsInReader(strings.NewReader(t.json), paths)
			if as.NoError(err, "Testing: %s", t.name) {
				res := toResultArray(eval_reader)
				if as.NoError(eval.Error) {
					as.EqualValues(t.expected, res, "Testing of %q", t.name)
				}
			}
		}
	}
}

func newResult(pathString, value string, typ int, keys ...interface{}) Result {
	keysChanged := make([]interface{}, len(keys))
	for i, k := range keys {
		switch v := k.(type) {
		case string:
			keysChanged[i] = []byte(v)
		default:
			keysChanged[i] = v
		}
	}

	return Result{
		PathString: pathString,
		Value:      []byte(value),
		Keys:       keysChanged,
		Type:       typ,
	}
}

func toResultArray(e *Eval) []Result {
	vals := make([]Result, 0)
	for {
		if r, ok := e.Next(); ok {
			if r != nil {
				vals = append(vals, *r)
			}
		} else {
			break
		}
	}
	return vals
}
