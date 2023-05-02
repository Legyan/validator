# Validator

`validator` is a simple module for validating structs in Go. It uses struct field tags to define validation rules and returns a list of validation errors when a struct does not meet the specified requirements.

## Features

- Validate struct fields using tags
- Built-in validators: `len`, `in`, `min`, `max`
- Supports `string` and `int` types, slices of `string` and `int`
- Customizable error messages

## Installation

```sh
go get -u github.com/legyan/validator
```

## Usage

```go
package main

import (
	"fmt"
	"github.com/legyan/validator"
)

type User struct {
	Username string `validate:"len:6"`
	Age      int    `validate:"min:18"`
  Role     string `validate:"in:admin,user"`
}

func main() {
	user := User{Username: "John", Age: 17}
	if err := validator.Validate(user); err != nil {
		fmt.Println(err)
	}
}
```

## Validators

### len

Checks if the length of a `string` is equal to the specified value.

```go
Username string `validate:"len:6"`
```

### in

Checks if the value of a field is in the specified set.

```go
Role string `validate:"in:admin,user"`
```

### min

Checks if the value of a field is greater than or equal to the specified minimum.

```go
Username string `validate:"min:2"`
```

### max

Checks if the value of a field is less than or equal to the specified maximum.

```go
Age int `validate:"max:99"`
```
