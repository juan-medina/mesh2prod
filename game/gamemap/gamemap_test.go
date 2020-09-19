/*
 * Copyright (c) 1010 Juan Medina.
 *
 *  Permission is hereby granted, free of charge, to any person obtaining a copy
 *  of this software and associated documentation files (the "Software"), to deal
 *  in the Software without restriction, including without limitation the rights
 *  to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 *  copies of the Software, and to permit persons to whom the Software is
 *  furnished to do so, subject to the following conditions:
 *
 *  The above copyright notice and this permission notice shall be included in
 *  all copies or substantial portions of the Software.
 *
 *  THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 *  IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 *  FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 *  AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 *  LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 *  OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 *  THE SOFTWARE.
 */

package gamemap

import (
	"fmt"
	"testing"
)

func TestFromString(t *testing.T) {
	str := "" +
		"       " + "\n" +
		"   333 " + "\n" +
		"    33 " + "\n" +
		"   333 " + "\n" +
		"       " + "\n"

	gm := fromString(str)
	got := gm.String()

	expect := "" +
		"       " + "\n" +
		"   333 " + "\n" +
		"    33 " + "\n" +
		"   333 " + "\n" +
		"       " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}

func TestGameMap_Set(t *testing.T) {
	str := "" +
		"       " + "\n" +
		"   333 " + "\n" +
		"    33 " + "\n" +
		"   333 " + "\n" +
		"       " + "\n"

	gm := fromString(str)

	gm.set(3, 2, placed)

	got := gm.String()
	expect := "" +
		"       " + "\n" +
		"   333 " + "\n" +
		"   133 " + "\n" +
		"   333 " + "\n" +
		"       " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}

func TestGameMap_Place(t *testing.T) {
	type tc struct {
		given  string
		placeR int
		placeC int
		expect string
	}

	cases := []tc{
		{
			given: "" +
				"        " + "\n" +
				"   3333 " + "\n" +
				"    333 " + "\n" +
				"   3333 " + "\n" +
				"        " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"        " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"        " + "\n",
		},
		{
			given: "" +
				"        " + "\n" +
				"    333 " + "\n" +
				"   3333 " + "\n" +
				"   3333 " + "\n" +
				"        " + "\n",
			placeC: 3,
			placeR: 1,
			expect: "" +
				"        " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"        " + "\n",
		},
		{
			given: "" +
				"        " + "\n" +
				"   3333 " + "\n" +
				"   3333 " + "\n" +
				"    333 " + "\n" +
				"        " + "\n",
			placeC: 3,
			placeR: 3,
			expect: "" +
				"        " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"   2222 " + "\n" +
				"        " + "\n",
		},
		{
			given: "" +
				"        " + "\n" +
				"   3333 " + "\n" +
				"   3333 " + "\n" +
				"    333 " + "\n" +
				"        " + "\n",
			placeC: 2,
			placeR: 1,
			expect: "" +
				"        " + "\n" +
				"  13333 " + "\n" +
				"   3333 " + "\n" +
				"    333 " + "\n" +
				"        " + "\n",
		},
		{
			given: "" +
				"                  " + "\n" +
				"    3333333333333 " + "\n" +
				"   33333333333333 " + "\n" +
				"                  " + "\n",
			placeC: 3,
			placeR: 1,
			expect: "" +
				"                  " + "\n" +
				"   22222222222222 " + "\n" +
				"   22222222222222 " + "\n" +
				"                  " + "\n",
		},
		{
			given: "" +
				"                  " + "\n" +
				"   33333333333333 " + "\n" +
				"    3333333333333 " + "\n" +
				"                  " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"                  " + "\n" +
				"   22222222222222 " + "\n" +
				"   22222222222222 " + "\n" +
				"                  " + "\n",
		},
		{
			given: "" +
				"      " + "\n" +
				"   33 " + "\n" +
				"    3 " + "\n" +
				"   33 " + "\n" +
				"      " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"      " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"      " + "\n",
		},
		{
			given: "" +
				"      " + "\n" +
				"   33 " + "\n" +
				"    3 " + "\n" +
				"   33 " + "\n" +
				"   33 " + "\n" +
				"   33 " + "\n" +
				"   33 " + "\n" +
				"      " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"      " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"   22 " + "\n" +
				"      " + "\n",
		},
		{
			given: "" +
				"        " + "\n" +
				"   3333333 " + "\n" +
				"    333333 " + "\n" +
				"   3333    " + "\n" +
				"           " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"           " + "\n" +
				"   2222222 " + "\n" +
				"   2222222 " + "\n" +
				"   2222    " + "\n" +
				"           " + "\n",
		},
		{
			given: "" +
				"           " + "\n" +
				"    333333 " + "\n" +
				"   3333333 " + "\n" +
				"   3333    " + "\n" +
				"           " + "\n",
			placeC: 3,
			placeR: 1,
			expect: "" +
				"           " + "\n" +
				"   2222222 " + "\n" +
				"   2222222 " + "\n" +
				"   2222    " + "\n" +
				"           " + "\n",
		},
		{
			given: "" +
				"           " + "\n" +
				"    333333 " + "\n" +
				"   3333333 " + "\n" +
				"   3333    " + "\n" +
				"           " + "\n",
			placeC: 3,
			placeR: 1,
			expect: "" +
				"           " + "\n" +
				"   2222222 " + "\n" +
				"   2222222 " + "\n" +
				"   2222    " + "\n" +
				"           " + "\n",
		},
		{
			given: "" +
				"           " + "\n" +
				"    333333 " + "\n" +
				"   3333333 " + "\n" +
				"    333    " + "\n" +
				"           " + "\n",
			placeC: 3,
			placeR: 3,
			expect: "" +
				"           " + "\n" +
				"    333333 " + "\n" +
				"   2222333 " + "\n" +
				"   2222    " + "\n" +
				"           " + "\n",
		},
		{
			given: "" +
				"           " + "\n" +
				"   3333333 " + "\n" +
				"    3333   " + "\n" +
				"   3333333 " + "\n" +
				"   3333    " + "\n" +
				"           " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"           " + "\n" +
				"   2222233 " + "\n" +
				"   22222   " + "\n" +
				"   2222233 " + "\n" +
				"   2222    " + "\n" +
				"           " + "\n",
		},
		{
			given: "" +
				"                         " + "\n" +
				"   333333333333333333333 " + "\n" +
				"    3333                 " + "\n" +
				"   333333333333333333333 " + "\n" +
				"                         " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"                         " + "\n" +
				"   222223333333333333333 " + "\n" +
				"   22222                 " + "\n" +
				"   222223333333333333333 " + "\n" +
				"                         " + "\n",
		},
		{
			given: "" +
				"                         " + "\n" +
				"   333333333333333333333 " + "\n" +
				"    3                  3 " + "\n" +
				"   333333333333333333333 " + "\n" +
				"                         " + "\n",
			placeC: 3,
			placeR: 2,
			expect: "" +
				"                         " + "\n" +
				"   223333333333333333333 " + "\n" +
				"   22                  3 " + "\n" +
				"   223333333333333333333 " + "\n" +
				"                         " + "\n",
		},
		{
			given: "" +
				"                         " + "\n" +
				"   33333                 " + "\n" +
				"     3333333333333333333 " + "\n" +
				"   33333                 " + "\n" +
				"                         " + "\n",
			placeC: 4,
			placeR: 2,
			expect: "" +
				"                         " + "\n" +
				"   32222                 " + "\n" +
				"    22223333333333333333 " + "\n" +
				"   32222                 " + "\n" +
				"                         " + "\n",
		},
		{
			given: "" +
				"                         " + "\n" +
				"   33333                 " + "\n" +
				"   33333                 " + "\n" +
				"   33333                 " + "\n" +
				"      333                " + "\n" +
				"     3333                " + "\n" +
				"                         " + "\n",
			placeC: 5,
			placeR: 4,
			expect: "" +
				"                         " + "\n" +
				"   33222                 " + "\n" +
				"   33222                 " + "\n" +
				"   33222                 " + "\n" +
				"     2222                " + "\n" +
				"     2222                " + "\n" +
				"                         " + "\n",
		},
	}

	for i, c := range cases {
		t.Run(fmt.Sprintf("case %d", i+1), func(t *testing.T) {
			gm := fromString(c.given)

			gm.place(c.placeC, c.placeR)

			got := gm.String()

			if got != c.expect {
				t.Fatalf("from string error, got %v, expect %v", got, c.expect)
			}
		})
	}
}

func TestGameMap_Add(t *testing.T) {
	gm := newGameMap(10, 10)

	piece := [][]blocState{
		{0, 0, 3},
		{0, 3, 3},
		{0, 3, 3},
		{0, 0, 3},
	}

	gm.add(3, 5, piece, 1)

	got := gm.String()
	expect := "" +
		"          " + "\n" +
		"          " + "\n" +
		"          " + "\n" +
		"          " + "\n" +
		"          " + "\n" +
		"     4    " + "\n" +
		"    44    " + "\n" +
		"    44    " + "\n" +
		"     4    " + "\n" +
		"          " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}
