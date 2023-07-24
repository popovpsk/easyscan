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

go get -u github.com/popovpsk/easyscan
```

## Usage
Import the easyscan package in your Go code:

```go
import "github.com/popovpsk/easyscan"
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
BenchmarkGet_PurePgx-12                                         19189694               310.7 ns/op           920 B/op          3 allocs/op
BenchmarkGet_Easyscan-12                                         3206949              1872 ns/op             952 B/op          5 allocs/op
BenchmarkGet_GeorgysavaScany-12                                   266756             22162 ns/op            8635 B/op        231 allocs/op

BenchmarkSelect20RowsToValues_PurePgx-12                          447489             13084 ns/op           44344 B/op         47 allocs/op
BenchmarkSelect20RowsToValues_Easyscan-12                        1000000              5258 ns/op            9152 B/op          6 allocs/op
BenchmarkSelect20RowsToValues_GeorgysavaScany-12                   81523             74438 ns/op           53281 B/op        316 allocs/op

BenchmarkSelect20RowsToPointers_PurePgx-12                        898324              6459 ns/op           18448 B/op         47 allocs/op
BenchmarkSelect20RowsToPointers_Easyscan-12                       719828              7571 ns/op            9512 B/op         36 allocs/op
BenchmarkSelect20RowsToPointers_GeorgysavaScany-12                 94191             64261 ns/op           27269 B/op        316 allocs/op

BenchmarkSelect1000RowsWithoutPrealloc_PurePgx-12                  16788            357468 ns/op         1340741 B/op         15 allocs/op
BenchmarkSelect1000RowsWithout_PreallocEasyscan-12                 14096            458766 ns/op         1341061 B/op         29 allocs/op
BenchmarkSelect1000RowsWithoutPrealloc_GeorgysavaScany-12           2154           2870733 ns/op         2304837 B/op       4245 allocs/op

BenchmarkSelect1000RowsWithPrealloc_PurePgx-12                     56071            114775 ns/op          402330 B/op          4 allocs/op
BenchmarkSelect1000RowsWithPrealloc_Easyscan-12                    43491            130274 ns/op          402369 B/op          6 allocs/op
BenchmarkSelect1000RowsWithPrealloc_GeorgysavaScany-12              2594           2331992 ns/op         1357855 B/op       4232 allocs/op
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

