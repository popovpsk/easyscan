[![Go](https://github.com/popovpsk/easyscan/actions/workflows/go.yml/badge.svg)](https://github.com/popovpsk/easyscan/actions/workflows/go.yml)
# easyscan

## Overview
easyscan is a lightweight and efficient Golang library that provides easy-to-use functions for performing operations on PostgreSQL databases using the pgx/v4 driver. The library offers two main functions: Get and Select, inspired by the popular jmoiron/sqlx library.

## Features

- Efficiency: easyscan is designed to perform operations with minimal overhead, ensuring fast and efficient database interactions.
- Tag-based Scanning: The library works with tagged structs or primitive types like int, string, or sql.Scanner. This allows seamless scanning of query results into Go types.
- Convenience.


## Installation

Make sure you have Go installed and set up correctly. To install easyscan, simply run:

```bash

go get -u github.com/yourusername/easyscan
```

## Usage
Import the easyscan package in your Go code:

```go
import "github.com/yourusername/easyscan"
```

### Connect to the Database
Before using easyscan, you need to establish a connection to your PostgreSQL database using the pgx/v4 driver. Refer to the documentation of pgx/v4 for connection details.

Scanning into Structs
To fetch a single row from the database and scan it into a struct, use the Get function:

### Scanning into Slices
```go
var ids []int
err := easyscan.Select(ctx, conn, &ids, "SELECT id FROM users WHERE active=true")
```

### Scanning into Structs
To fetch a single row from the database and scan it into a struct, use the Get function:
```go
type User struct {
    ID       int    `db:"id"`
    Username string `db:"username"`
    Email    string `db:"email"`
}

user := User{}
err := easyscan.Get(ctx, conn, &user, "SELECT * FROM users WHERE id=$1", 1)
```

## Supported Types
The Get and Select functions support scanning into the following types:

Tagged structs with fields matching the columns in the SELECT statement.
Primitive types such as int, string, and types implementing sql.Scanner.

## Benchmarks 
```shell
BenchmarkGet_PurePgx-12                                         10327090               582.8 ns/op          1232 B/op          6 allocs/op
BenchmarkGet_Easyscan-12                                         2203040              2709 ns/op            1296 B/op         10 allocs/op
BenchmarkGet_GeorgysavaScany-12                                   167077             36841 ns/op           12090 B/op        356 allocs/op

BenchmarkSelect20RowsToValues_PurePgx-12                          255427             23471 ns/op           59456 B/op         94 allocs/op
BenchmarkSelect20RowsToValues_Easyscan-12                         729572              8395 ns/op           12576 B/op         12 allocs/op
BenchmarkSelect20RowsToValues_GeorgysavaScany-12                   51952            117658 ns/op           72754 B/op        526 allocs/op

BenchmarkSelect20RowsToPointers_PurePgx-12                        510237             12189 ns/op           24736 B/op         94 allocs/op
BenchmarkSelect20RowsToPointers_Easyscan-12                       425806             13654 ns/op           13248 B/op         72 allocs/op
BenchmarkSelect20RowsToPointers_GeorgysavaScany-12                 61044             98633 ns/op           37781 B/op        526 allocs/op

BenchmarkSelect1000RowsWithoutPrealloc_PurePgx-12                  10000            580567 ns/op         1831151 B/op         30 allocs/op
BenchmarkSelect1000RowsWithout_PreallocEasyscan-12                  9400            593476 ns/op         1698429 B/op         58 allocs/op
BenchmarkSelect1000RowsWithoutPrealloc_GeorgysavaScany-12           1480           4094572 ns/op         3005140 B/op       8382 allocs/op

BenchmarkSelect1000RowsWithPrealloc_PurePgx-12                     35184            171437 ns/op          541908 B/op          8 allocs/op
BenchmarkSelect1000RowsWithPrealloc_Easyscan-12                    27218            220415 ns/op          541986 B/op         12 allocs/op
BenchmarkSelect1000RowsWithPrealloc_GeorgysavaScany-12              1710           3495441 ns/op         1842449 B/op       8359 allocs/op
```

To find the benchmarks in the benchmarks branch and run them using make bench, follow these steps:

```shell
git checkout benchmarks
make bench
```

## Example
To run a usage example, execute the following command in the terminal:

```shell
make up && go run ./example
```

## Contributing
Contributions are welcome! If you find a bug, have an enhancement suggestion, or want to contribute in any other way, please open an issue or a pull request.

