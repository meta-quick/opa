# tengo-x

[Tengo](https://github.com/d5/tengo) is a small, dynamic, fast, secure script language for Go.

This package is intended to extend Tengo to make `golang functions and struct instances`
as Tengo `built-in functions and objects` very easily, so calling golang functions or objects from
Tengo is very simple.

### Usage

The package is fully go-getable, So, just type

  `go get github.com/rosbit/tengo-x`

to install.

```go
package main

import "fmt"
import "github.com/rosbit/tengo-x"

func main() {
  ctx := tgx.NewTengo()

  res, _ := ctx.Eval("1 + 2", nil)
  fmt.Println("result is:", res)
}
```

### Tengo calls Go function and objects

Tengo calling Go function is also easy. In the Go code, make a golang function. Just put varibles,
functions, struct instances in an enviroment map when loading Tengo script.

```go
package main

import "github.com/rosbit/tengo-x"
import "fmt"

type M struct {
   Name string
   Age int
}
func (m *M) IncAge(a int) {
   m.Age += a
}

func adder(a1 float64, a2 float64) float64 {
    return a1 + a2
}

func main() {
  vars := map[string]interface{}{
     "m": &M{Name:"rosbit", Age:1}, // to Tengo object
     "adder": adder,                // to Tengo built-in function
     "a": []int{1,2,3}              // to Tengo array
  }

  ctx := js.NewTengo()
  if err := ctx.LoadFile("a.tengo", vars); err != nil {
     fmt.Printf("%v\n", err)
     return
  }

  res, err := ctx.GetGlobals("r") // get the value of var named "r"
  if err != nil {
     fmt.Printf("%v\n", err)
     return
  }
  fmt.Printf("res:", res)
}
```

In Tengo code, one can call the registered name directly. There's the example `a.tengo`.

```go
fmt := import("fmt")
r := adder(1, 100)   // the function "adder" is implemented in Go
fmt.println(r)
fmt.println(m.name, " ", m.age) // m is an object
```

### Status

The package is not fully tested, so be careful.

### Contribution

Pull requests are welcome! Also, if you want to discuss something send a pull request with proposal and changes.
__Convention:__ fork the repository and make changes on your fork in a feature branch.
