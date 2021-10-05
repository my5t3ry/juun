package main

import (
	"testing"
	"time"

	"github.com/jackdoe/juun/common"
)

func TestTime(t *testing.T) {
	got := common.TimeToNamespace("i_abc", time.Unix(0, 0)).ToVW()
	expected := "|i_abc year_1970 day_1 month_1 hour_1 "
	if got != expected {
		t.Fatalf("wrong time features, expecte: '%s', got '%s'", expected, got)
	}
}
func TestUpDownUp(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", nil)
	h.add("ps 2", nil)
	h.add("ps 3", nil)

	must(t, h.up("incomplete-before-up"), "ps 3") // "" -> ps 3
	must(t, h.up(""), "ps 2")                     // ps3 -> ps 2
	must(t, h.down(""), "ps 3")                   // ps 2 -> ps 3
	must(t, h.down(""), "incomplete-before-up")
	must(t, h.up(""), "ps 3")
}

func TestHistory(t *testing.T) {
	h := NewHistory()
	h.add("make", nil)
	h.add("make", nil)
	h.add("make", nil)

	if len(h.Lines) != 1 {
		t.Fatalf("expected only 1 line")
	}
	if h.search("m", nil)[0].Line != "make" {
		t.Fatalf("make not found")
	}

	if h.search("make", nil)[0].Line != "make" {
		t.Fatalf("make not found")
	}
}

func must(t *testing.T, a, b string) {
	if a != b {
		t.Logf("result: %s != expected: %s", a, b)
		panic("a")
	}
}

func TestHistoryChange(t *testing.T) {
	h := NewHistory()
	h.add("first-terminal-ps 1", nil)                            // global uuid 0
	h.add("ps 2", nil)                                           // global uuid 1
	h.add("ps 3", nil)                                           // global uuid 1
	must(t, h.up("incomplete-before-up"), "ps 3")                // global uuid  cursor 2
	must(t, h.up("incomplete-before-up"), "ps 2")                // global uuid  cursor 0
	must(t, h.up("incomplete-before-up"), "first-terminal-ps 1") // global uuid  cursor 0
}

func TestGlobalHistory2(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", nil) // -> 0
	h.add("ps 3", nil) // -> 1
	h.add("ps 4", nil) // -> 2

	h.add("zs 1", nil)
	h.add("zs 2", nil)
	h.add("zs 3", nil)

	must(t, h.up("incomplete-before-up"), "zs 3")
	must(t, h.up(""), "zs 2")
	for i := 0; i < 100; i++ {
		h.up("")
	}
	must(t, h.up(""), "ps 2")
	must(t, h.down(""), "ps 3")
	must(t, h.down(""), "ps 4")
	must(t, h.down(""), "zs 1")
	must(t, h.down(""), "zs 2")
	must(t, h.down(""), "zs 3")
	must(t, h.down(""), "incomplete-before-up")
}

func TestGlobalHistory(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", nil)
	h.add("ps 3", nil)
	h.add("ps 4", nil)

	h.add("zs 1", nil)
	h.add("zs 2", nil)

	h.add("zs 3", nil)
	h.add("zs x", nil)

	must(t, h.up("incomplete-before-up"), "zs x")
	must(t, h.up(""), "zs 3")
	for i := 0; i < 100; i++ {
		h.up("")
	}
	must(t, h.up(""), "ps 2")

	must(t, h.down(""), "ps 3")
	must(t, h.down(""), "ps 4")
	must(t, h.down(""), "zs 1")
	must(t, h.down(""), "zs 2")
	must(t, h.down(""), "zs 3")
	must(t, h.down(""), "zs x")
	must(t, h.down(""), "incomplete-before-up")
}

func TestUpDownUpGlobal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", nil) //0
	h.add("ps 2", nil) //1
	h.add("ps 3", nil) //2

	must(t, h.up("incomplete-before-up"), "ps 3") // 2 -> 1
	must(t, h.up(""), "ps 2")                     // 1 -> 0
	must(t, h.down(""), "ps 3")                   // 0 -> 1
	must(t, h.down(""), "incomplete-before-up")   // 1 -> 2

	must(t, h.down(""), "incomplete-before-up")
	must(t, h.up(""), "ps 3")
}

func TestHistoryUpDown(t *testing.T) {
	h := NewHistory()

	must(t, h.up("incomplete"), "")
	must(t, h.down(""), "incomplete")

	h.add("ps 1", nil)

	must(t, h.up("incomplete"), "ps 1")
	must(t, h.down(""), "incomplete")

	h.add("ps 2", nil)
	h.add("ps 3", nil)
	h.add("ps 4", nil)

	must(t, h.up("incomplete-before-up"), "ps 4")
	for i := 0; i < 100; i++ {
		h.up("")
	}
	must(t, h.down(""), "ps 2")
	for i := 0; i < 100; i++ {
		h.down("")
	}

	must(t, h.down(""), "incomplete-before-up")
}

func TestUpDownUpUpGlobal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", nil) //0
	must(t, h.up("a"), "ps 1")
	must(t, h.down(""), "a")
	must(t, h.up("a"), "ps 1")
	must(t, h.down(""), "a")
	must(t, h.up("a"), "ps 1")
	must(t, h.down(""), "a")
	h.add("ps 2", nil) //0
	must(t, h.up("a"), "ps 1")
	must(t, h.down(""), "a")
}

func TestUpDownUpLocal(t *testing.T) {
	h := NewHistory()
	h.add("ps 1", nil) //0
	h.add("ps 2", nil) //1
	h.add("ps 3", nil) //2

	must(t, h.up("incomplete-before-up"), "ps 3") // 2 -> 1
	must(t, h.up(""), "ps 2")                     // 1 -> 0
	must(t, h.down(""), "ps 3")                   // 0 -> 1

	must(t, h.down(""), "incomplete-before-up")
	must(t, h.up(""), "ps 3")
}

func TestHistoryUpDownMany(t *testing.T) {
	h := NewHistory()
	h.add("ps 2", nil)
	h.add("ps 3", nil)
	h.add("ps 4", nil)

	must(t, h.up("incomplete-before-up"), "ps 4")
	for i := 0; i < 10; i++ {
		h.up("")
	}

	must(t, h.down(""), "ps 3")

	for i := 0; i < 10; i++ {
		h.down("")
	}

	must(t, h.down(""), "incomplete-before-up")
}
