package main

import "fmt"

func main() {
	chunk := newChunk()
	constant := int(chunk.WriteValue(1.2))
	chunk.WriteChunk(OP_CONSTANT, 1, constant)

	constant = int(chunk.WriteValue(3.4))
	chunk.WriteChunk(OP_CONSTANT, 1, constant)
	chunk.WriteChunk(OP_ADD, 1)

	constant = int(chunk.WriteValue(5.6))
	chunk.WriteChunk(OP_CONSTANT, 1, constant)

	chunk.WriteChunk(OP_DIVIDE, 1)
	chunk.WriteChunk(OP_NEGATE, 1)
	chunk.WriteChunk(OP_RETURN, 1)

	vm := newVM(chunk)
	err := vm.run()
	if err != nil {
		fmt.Println(err)
	}
}

// func main() {
// 	var err error

// 	if len(os.Args) > 2 {
// 		fmt.Fprint(os.Stderr, "Usage: jlox [script]\n")
// 		os.Exit(64)
// 	} else if len(os.Args) == 2 {
// 		err = runFile(os.Args[0])
// 	} else {
// 		err = runPrompt()
// 	}

// 	if err != nil {
// 		fmt.Fprint(os.Stderr, err)
// 	}
// }

// func runFile(file string) error {
// 	fp, err := os.Open(file)
// 	if err != nil {
// 		return err
// 	}

// 	input, err := io.ReadAll(fp)
// 	if err != nil {
// 		return err
// 	}

// 	env := NewEnvironment()
// 	run(input[:], env)
// 	return nil
// }

// func runPrompt() error {
// 	scanner := bufio.NewScanner(os.Stdin)
// 	fmt.Fprint(os.Stdin, ">>> ")
// 	env := NewEnvironment()

// 	for scanner.Scan() {
// 		text := scanner.Text()
// 		if text == "" {
// 			break
// 		}
// 		run([]byte(text), env)
// 		fmt.Fprint(os.Stdin, ">>> ")
// 	}

// 	if err := scanner.Err(); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func run(input []byte, env *Environment) {
// 	scanner := NewScanner(input)
// 	scanner.scanTokens()

// 	if scanner.Errors.HasErrors() {
// 		scanner.Errors.PrintErrors()
// 		return
// 	}

// 	// fmt.Println(scanner.Tokens)
// 	parser := NewParser(scanner.Tokens)
// 	expr := parser.parse()

// 	if parser.Errors.HasErrors() {
// 		parser.Errors.PrintErrors()
// 		return
// 	}

// 	interpreter := NewInterpreter()
// 	result := interpreter.Interpret(expr, env)
// 	if result != nil {
// 		fmt.Println(result.Inspect())
// 	}
// }
