# Crafting Interpreters and Compilers in Go (P.S ChatGPT did that)

Welcome to my learning journey in crafting interpreters and compilers in Go! This project is inspired by the books "Crafting Interpreters" by Robert Nystrom and "Writing an Interpreter in Go" by Thorsten Ball. My goal is to apply the concepts and techniques from these books to build a robust interpreter and eventually a compiler using Go.

## Overview

This repository contains my implementation of a programming language interpreter written in Go. The project is divided into several stages, each focusing on different aspects of interpreter and compiler design, including lexical analysis, parsing, semantic analysis, and interpretation.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Project Structure](#project-structure)
3. [Features](#features)
4. [Examples](#examples)
5. [References](#references)
6. [Contributing](#contributing)
7. [License](#license)

## Getting Started

To get started with the project, you'll need to have Go installed on your machine. You can download and install Go from the official [Go website](https://golang.org/dl/).

Clone the repository:

```sh
git clone https://github.com/arsenzairov/crafting-interpreters.git
cd crafting-interpreters
```

Build and run the interpreter:

```sh
go build -o interpreter .
./interpreter
```

## Project Structure

The project is structured as follows:

- `scanner/`: Contains the lexical analyzer which tokenizes the input source code.
- `parser/`: Contains the parser which generates the Abstract Syntax Tree (AST) from tokens.
- `ast/`: Defines the AST node types and visitor interfaces.
- `interpreter/`: Contains the interpreter that traverses the AST and executes the program.
- `main.go`: The entry point of the interpreter.

## Features

- **Lexical Analysis**: Tokenizes the source code into meaningful tokens.
- **Parsing**: Constructs an AST from the tokens.
- **Interpretation**: Executes the program by traversing the AST.
- **Error Handling**: Provides meaningful error messages with line numbers for easy debugging.

## Examples

Here are some examples of what you can do with the interpreter:

### Example 1: Boolean Literals

```go
true && false
```

### Example 2: Arithmetic Expressions

```go
5 + (3 * 2)
```

### Example 3: String Concatenation

```go
"Hello, " + "world!"
```

## References

This project is heavily inspired by the following books:

- [Crafting Interpreters](http://craftinginterpreters.com) by Robert Nystrom
- [Writing an Interpreter in Go](https://interpreterbook.com) by Thorsten Ball

## Contributing

Contributions are welcome! If you have any suggestions or improvements, feel free to open an issue or submit a pull request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

---

Happy coding!
