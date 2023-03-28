# bitmask

[![Go Reference](https://pkg.go.dev/badge/github.com/astef/bitmask.svg)](https://pkg.go.dev/github.com/astef/bitmask)

Arbitrary size bitmask (aka bitset) with efficient Slice method.

	bm := bitmask.New(4)        // [4]{0000}
	bm.Set(3)                   // [4]{0001}
	bm.Slice(2, 4).ToggleAll()  // [4]{0010}
	bm.Slice(1, 3).ToggleAll()  // [4]{0100}
	bm.Slice(0, 2).ToggleAll()  // [4]{1000}
	bm.Clear()                  // [4]{0000}

