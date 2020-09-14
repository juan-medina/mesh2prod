/*
 * Copyright (c) 2020 Juan Medina.
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

import "testing"

func TestFromString(t *testing.T) {
	str := "" +
		"       " + "\n" +
		"   111 " + "\n" +
		"    11 " + "\n" +
		"   111 " + "\n" +
		"       " + "\n"

	gm := fromString(str)
	got := gm.String()

	expect := "" +
		"       " + "\n" +
		"   111 " + "\n" +
		"    11 " + "\n" +
		"   111 " + "\n" +
		"       " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}

func TestGameMap_Set(t *testing.T) {
	str := "" +
		"       " + "\n" +
		"   111 " + "\n" +
		"    11 " + "\n" +
		"   111 " + "\n" +
		"       " + "\n"

	gm := fromString(str)

	gm.Set(3, 2, placed)

	got := gm.String()
	expect := "" +
		"       " + "\n" +
		"   111 " + "\n" +
		"   211 " + "\n" +
		"   111 " + "\n" +
		"       " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}

func TestGameMap_Place(t *testing.T) {
	str := "" +
		"        " + "\n" +
		"   1111 " + "\n" +
		"    111 " + "\n" +
		"   1111 " + "\n" +
		"        " + "\n"

	gm := fromString(str)

	gm.Place(3, 2)

	got := gm.String()
	expect := "" +
		"        " + "\n" +
		"   3333 " + "\n" +
		"   3333 " + "\n" +
		"   3333 " + "\n" +
		"        " + "\n"

	if got != expect {
		t.Fatalf("from string error, got %v, expect %v", got, expect)
	}
}
