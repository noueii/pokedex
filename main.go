package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type cliCommand struct {
	name        string
	description string
	callback    func() error
}

var pokemands map[string]cliCommand

func main() {
	initialize()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()

		text := scanner.Text()
		cleanText := cleanInput(text)

		command := cleanText[0]
		handleCommand(command)

	}

}

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	strs := strings.Fields(text)
	result := []string{}
	for _, str := range strs {
		res := strings.ReplaceAll(str, " ", "")
		if len(res) == 0 {
			continue
		}

		result = append(result, res)
	}
	return result

}

func handleCommand(command string) error {
	for _, com := range pokemands {
		if command == com.name {
			com.callback()
			return nil
		}

	}

	fmt.Println("Unknown command")
	return nil

}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return fmt.Errorf("Fatal: Could not exit pokedex")
}

func commandHelp() error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:\n")
	for _, cmd := range pokemands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)

	}
	return nil

}

func initialize() {
	pokemands = make(map[string]cliCommand)
	pokemands = map[string]cliCommand{
		"help": {
			name:        "help",
			description: "Displays a help message",
			callback:    commandHelp,
		},
		"exit": {
			name:        "exit",
			description: "Exit the Pokedex",
			callback:    commandExit,
		},
	}

}
