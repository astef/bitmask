# bitmask

[![Go Reference](https://pkg.go.dev/badge/github.com/astef/bitmask.svg)](https://pkg.go.dev/github.com/astef/bitmask) ![Coverage Badge](https://img.shields.io/badge/coverage-97.6%25-green.svg)

Arbitrary size bitmask (aka bitset) with efficient Slice method.

### `.Slice` doesn't create copies of the underlying buffer, just like Go slices

```go
bm := bitmask.New(4)        // [4]{0000}
bm.Set(3)                   // [4]{0001}
bm.Slice(2, 4).ToggleAll()  // [4]{0010}
bm.Slice(1, 3).ToggleAll()  // [4]{0100}
bm.Slice(0, 2).ToggleAll()  // [4]{1000}
bm.ClearAll()               // [4]{0000}
```

### It's safe to copy overlapping bitmasks, which were created by slicing the original one

```go
base := bitmask.New(10)
base.Set(0)
base.Set(9)                 // [10]{1000000001}

src := base.Slice(0, 8)     // [8]{10000000}
dst := base.Slice(2, 10)    // [8]{00000001}

bitmask.Copy(dst, src)

fmt.Println(base)
```

```
[10]{1010000000}
```

### Iteration

```go
bm := bitmask.New(5)
bm.Set(0)
bm.Set(3)

it := bm.Iterator()
for {
    ok, value, index := it.Next()
    if !ok {
        break
    }

    // use the value
    fmt.Printf("%v) %v\n", index, value)
}
```

```
0) true
1) false
2) false
3) true
4) false
```

### `.String()` is O(1), it starts stripping after 512 bits

```go
bm := bitmask.New(513)
bm.Slice(255, 321).SetAll()
fmt.Println(bm)
```

```
[513]{0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000001 <more 64 bits> 1000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0000000000000000000000000000000000000000000000000000000000000000 0}
```
