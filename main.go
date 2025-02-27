package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	"github.com/noueii/pokedex/internal"
)

type cliCommand struct {
	name        string
	description string
	callback    func(*URLconfig, *string) error
}

type URLconfig struct {
	Next string
	Prev string
}

type Location struct {
	Name string `json:"name"`
	Url  string `json:"url"`
}

type Pokemon struct {
	Name       string
	Experience int
	Height     int
	Weight     int
	Stats      struct {
		Hp             int
		Attack         int
		Defense        int
		SpecialAttack  int
		SpecialDefense int
		Speed          int
	}
	Types []string
}

var urlConfig = URLconfig{
	Next: "",
	Prev: "",
}

var pokemands map[string]cliCommand
var cache *pokecache.Cache
var userPokedex []Pokemon

func main() {
	initialize()

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		scanner.Scan()

		text := scanner.Text()
		cleanText := cleanInput(text)

		command := cleanText[0]
		argument := ""
		if len(cleanText) > 1 {
			argument = cleanText[1]
		}

		err := handleCommand(command, argument)
		if err != nil {
			fmt.Println(err)
		}
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

func handleCommand(command string, argument string) error {
	for _, com := range pokemands {
		if command == com.name {
			com.callback(&urlConfig, &argument)
			return nil
		}

	}

	fmt.Println("Unknown command")
	return nil
}

func commandExit(config *URLconfig, argument *string) error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	os.Exit(0)
	return fmt.Errorf("Fatal: Could not exit pokedex")
}

func commandHelp(config *URLconfig, argument *string) error {
	fmt.Println("Welcome to the Pokedex!")
	fmt.Println("Usage:")
	fmt.Println()
	for _, cmd := range pokemands {
		fmt.Printf("%s: %s\n", cmd.name, cmd.description)

	}
	return nil
}

func commandMap(config *URLconfig, argument *string) error {

	locations, err := fetchLocations(config, true)
	if err != nil {

		fmt.Println(err)
		return err
	}
	for _, location := range locations {
		fmt.Println(location.Name)
	}

	return nil

}

func commandExplore(config *URLconfig, argument *string) error {
	if len(*argument) == 0 {
		fmt.Println("No map provided. Please provide a map after 'explore'")
		fmt.Println("Example: explore [map-name]")
		return nil
	}

	fmt.Printf("Exploring %s...\n", *argument)

	pokemons, err := fetchLocation(*argument)

	if err != nil {
		return err
	}

	fmt.Println("Found Pokemon:")
	for _, pokemon := range pokemons {
		fmt.Printf(" - %s\n", pokemon)

	}
	return nil
}

func commandMapBack(config *URLconfig, argument *string) error {
	locations, err := fetchLocations(config, false)
	if err != nil {

		fmt.Println(err)
		return err
	}
	for _, location := range locations {
		fmt.Println(location.Name)
	}

	return nil

}

func commandCatch(config *URLconfig, argument *string) error {
	if len(*argument) == 0 {
		fmt.Println("Missing Pokemon name. Try again!")
		return nil
	}
	fmt.Printf("Throwing a Pokeball at %s...\n", *argument)
	pokemon, err := getPokemon(*argument)
	if err != nil {
		return err
	}

	chance := 100 - pokemon.Experience/8

	if rand.Intn(100) < chance {
		fmt.Printf("%s was caught!\n", *argument)
		userPokedex = append(userPokedex, pokemon)
		return nil
	}

	fmt.Printf("%s escaped!\n", *argument)
	return nil

}

func commandInspect(config *URLconfig, argument *string) error {
	if len(*argument) == 0 {
		fmt.Println("Please provide a Pokemon name in your pokedex.")
		return nil
	}

	for _, val := range userPokedex {
		if val.Name == *argument {
			val.Print()
			return nil
		}
	}

	fmt.Println("Pokemon not found inside your pokedex. Are you sure you caught it?")
	return nil

}

func commandPokedex(config *URLconfig, argument *string) error {
	fmt.Println("Your Pokedex:")
	for _, pokemon := range userPokedex {

		fmt.Printf("\t- %s\n", pokemon.Name)
	}

	return nil
}

func getPokemon(name string) (Pokemon, error) {
	pokemon := Pokemon{}
	url := "https://pokeapi.co/api/v2/pokemon/" + name
	raw, err := fetchPOKEAPI(url)
	if err != nil {
		return pokemon, err
	}

	type response struct {
		Experience int    `json:"base_experience"`
		Name       string `json:"name"`
		Weight     int    `json:"weight"`
		Height     int    `json:"height"`
		Stats      []struct {
			BaseStat int `json:"base_stat"`
			Stat     struct {
				Name string `json:"name"`
			}
		} `json:"stats"`
		Types []struct {
			Type struct {
				Name string `json:"name"`
			}
		} `json:"types"`
	}

	resp := response{}
	err = json.Unmarshal(raw, &resp)
	if err != nil {
		return pokemon, err
	}

	pokemon = Pokemon{
		Name:       resp.Name,
		Experience: resp.Experience,
		Weight:     resp.Weight,
		Height:     resp.Height,
	}

	for _, field := range reflect.VisibleFields(reflect.TypeOf(pokemon.Stats)) {
		for _, stat := range resp.Stats {
			if strings.ReplaceAll(stat.Stat.Name, "-", "") == strings.ToLower(field.Name) {
				reflect.ValueOf(&pokemon.Stats).Elem().FieldByName(field.Name).SetInt(int64(stat.BaseStat))
				break
			}
		}
	}

	for _, val := range resp.Types {
		pokemon.Types = append(pokemon.Types, val.Type.Name)
	}

	return pokemon, nil

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
		"map": {
			name:        "map",
			description: "Pokedex map - next 20 locations",
			callback:    commandMap,
		},
		"mapb": {
			name:        "mapb",
			description: "Pokedex map - previous 20 locations",
			callback:    commandMapBack,
		},

		"explore": {
			name:        "explore",
			description: "Explore a map to see which pokemons can be found",
			callback:    commandExplore,
		},

		"catch": {
			name:        "catch",
			description: "Catch a pokemon! type their name after the 'catch' command and let's go!",
			callback:    commandCatch,
		},

		"inspect": {
			name:        "inspect",
			description: "Inspect your Pokemon! Type its name after the command.",
			callback:    commandInspect,
		},

		"pokedex": {
			name:        "pokedex",
			description: "List your caught Pokemon",
			callback:    commandPokedex,
		},
	}

	cache = pokecache.NewCache(1 * time.Minute)

}

func fetchPOKEAPI(url string) ([]byte, error) {
	val, ok := cache.Get(url)
	if ok {
		return val, nil
	}

	res, err := http.Get(url)
	if err != nil {

		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()

	if err != nil {
		return nil, err
	}

	cache.Add(url, body)
	return body, nil
}

func fetchLocations(config *URLconfig, next bool) ([]Location, error) {
	url := "https://pokeapi.co/api/v2/location-area"
	if next && len(config.Next) > 0 {
		url = config.Next
	}

	if !next && len(config.Prev) > 0 {
		url = config.Prev
	}
	raw, err := fetchPOKEAPI(url)
	type locationsResponse struct {
		Count    int        `json:"count"`
		Next     string     `json:"next"`
		Previous string     `json:"previous"`
		Results  []Location `json:"results"`
	}

	if err != nil {
		return nil, err
	}

	response := locationsResponse{}

	err = json.Unmarshal(raw, &response)
	if err != nil {
		return nil, err
	}
	config.Next = response.Next
	config.Prev = response.Previous
	return response.Results, nil
}

func fetchLocation(name string) ([]string, error) {
	url := "https://pokeapi.co/api/v2/location-area/" + name
	raw, err := fetchPOKEAPI(url)
	if err != nil {
		return nil, err
	}

	type pokemon struct {
		Name string `json:"name"`
		Url  string `json:"url"`
	}

	type response struct {
		PokemonEncounters []struct {
			Pokemon pokemon `json:"pokemon"`
		} `json:"pokemon_encounters"`
	}

	resp := response{}
	err = json.Unmarshal(raw, &resp)
	if err != nil {
		return nil, err
	}

	pokemons := []string{}
	for _, item := range resp.PokemonEncounters {
		pokemons = append(pokemons, item.Pokemon.Name)
	}

	return pokemons, nil
}

func (pokemon *Pokemon) Print() {
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %v\n", pokemon.Height)
	fmt.Printf("Weight: %v\n", pokemon.Weight)
	fmt.Println("Stats: ")
	fmt.Printf("\t-hp: %v\n", pokemon.Stats.Hp)
	fmt.Printf("\tattack: %v\n", pokemon.Stats.Attack)
	fmt.Printf("\tdefense: %v\n", pokemon.Stats.Defense)
	fmt.Printf("\tspecial-attack: %v\n", pokemon.Stats.SpecialAttack)
	fmt.Printf("\t-special-defense: %v\n", pokemon.Stats.SpecialDefense)
	fmt.Printf("\t-speed: %v\n", pokemon.Stats.Speed)
	fmt.Println("Types: ")
	for i := range pokemon.Types {
		fmt.Printf("\t- %s\n", pokemon.Types[i])
	}

}
