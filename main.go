package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"bootdev_pokedex/internal/pokecache"
)

type Config struct {
	Next     string
	Previous string
	Cache    *pokecache.Cache
	Pokedex  map[string]Pokemon
}

type LocationResponse struct {
	Results []struct {
		Name string `json:"name"`
	} `json:"results"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
}

type ExploreResponse struct {
	PokemonEncounters []struct {
		Pokemon struct {
			Name string `json:"name"`
		} `json:"pokemon"`
	} `json:"pokemon_encounters"`
}

type Pokemon struct {
	Name        string `json:"name"`
	Height      int    `json:"height"`
	Weight      int    `json:"weight"`
	BaseExp     int    `json:"base_experience"`
	Stats       []struct {
		Stat struct {
			Name string `json:"name"`
		} `json:"stat"`
		BaseStat int `json:"base_stat"`
	} `json:"stats"`
	Types []struct {
		Type struct {
			Name string `json:"name"`
		} `json:"type"`
	} `json:"types"`
}

func main() {
	config := &Config{Cache: pokecache.NewCache(10 * time.Second), Pokedex: make(map[string]Pokemon)}
	s := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("Pokedex > ")
		s.Scan()
		input := s.Text()
		args := cleanInput(input)
		if len(args) == 0 {
			continue
		}
		switch args[0] {
		case "exit":
			commandExit()
			return
		case "help":
			fmt.Println("Welcome to the Pokedex!\nUsage:\n\nhelp: Displays a help message\nexit: Exit the Pokedex")
		case "map":
			mapCommand(config)
		case "mapb":
			mapbCommand(config)
		case "explore":
			if len(args) < 2 {
				fmt.Println("Usage: explore <area_name>")
				continue
			}
			exploreCommand(config, args[1])
		case "catch":
			if len(args) < 2 {
				fmt.Println("Usage: catch <pokemon_name>")
				continue
			}
			catchCommand(config, args[1])
		case "inspect":
			if len(args) < 2 {
				fmt.Println("Usage: inspect <pokemon_name>")
				continue
			}
			inspectCommand(config, args[1])
		default:
			fmt.Println("Unknown command")
		}
	}
}

func mapCommand(config *Config) {
	url := config.Next
	if url == "" {
		url = "https://pokeapi.co/api/v2/location-area/"
	}

data, err := fetchLocations(url, config.Cache)
	if err != nil {
		fmt.Println("Error fetching locations:", err)
		return
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	config.Next = data.Next
	config.Previous = data.Previous
}

func mapbCommand(config *Config) {
	url := config.Previous
	if url == "" {
		fmt.Println("you're on the first page")
		return
	}
    data, err := fetchLocations(url, config.Cache)
	if err != nil {
		fmt.Println("Error fetching locations:", err)
		return
	}
	for _, location := range data.Results {
		fmt.Println(location.Name)
	}
	config.Next = data.Next
	config.Previous = data.Previous
}

func exploreCommand(config *Config, location string) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/location-area/%s/", location)
	fmt.Printf("Exploring %s...\n", location)
	data, err := fetchExplore(url, config.Cache)
	if err != nil {
		fmt.Println("Error fetching location details:", err)
		return
	}
	fmt.Println("Found Pokemon:")
	for _, encounter := range data.PokemonEncounters {
		fmt.Printf(" - %s\n", encounter.Pokemon.Name)
	}
}

func catchCommand(config *Config, name string) {
	url := fmt.Sprintf("https://pokeapi.co/api/v2/pokemon/%s/", name)
	fmt.Printf("Throwing a Pokeball at %s...\n", name)
	data, err := fetchPokemon(url, config.Cache)
	if err != nil {
		fmt.Println("Error fetching Pokemon details:", err)
		return
	}
	rand.Seed(time.Now().UnixNano())
	catchChance := rand.Intn(data.BaseExp + 1)
	if catchChance%2 == 0 {
		fmt.Printf("%s was caught!\n", name)
		config.Pokedex[name] = *data
	} else {
		fmt.Printf("%s escaped!\n", name)
	}
}

func fetchLocations(url string, cache *pokecache.Cache) (*LocationResponse, error) {
	if cachedData, found := cache.Get(url); found {
		var data LocationResponse
		if err := json.Unmarshal(cachedData, &data); err == nil {
			return &data, nil
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data LocationResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	responseData, _ := json.Marshal(data)
	cache.Add(url, responseData)
	return &data, nil
}

func fetchPokemon(url string, cache *pokecache.Cache) (*Pokemon, error) {
	if cachedData, found := cache.Get(url); found {
		var data Pokemon
		if err := json.Unmarshal(cachedData, &data); err == nil {
			return &data, nil
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data Pokemon
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	responseData, _ := json.Marshal(data)
	cache.Add(url, responseData)
	return &data, nil
}

func fetchExplore(url string, cache *pokecache.Cache) (*ExploreResponse, error) {
	if cachedData, found := cache.Get(url); found {
		var data ExploreResponse
		if err := json.Unmarshal(cachedData, &data); err == nil {
			return &data, nil
		}
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var data ExploreResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}
	responseData, _ := json.Marshal(data)
	cache.Add(url, responseData)
	return &data, nil
}

func inspectCommand(config *Config, name string) {
	pokemon, found := config.Pokedex[name]
	if !found {
		fmt.Println("you have not caught that pokemon")
		return
	}
	fmt.Printf("Name: %s\n", pokemon.Name)
	fmt.Printf("Height: %d\n", pokemon.Height)
	fmt.Printf("Weight: %d\n", pokemon.Weight)
	fmt.Println("Stats:")
	for _, stat := range pokemon.Stats {
		fmt.Printf("  -%s: %d\n", stat.Stat.Name, stat.BaseStat)
	}
	fmt.Println("Types:")
	for _, typ := range pokemon.Types {
		fmt.Printf("  - %s\n", typ.Type.Name)
	}
}

func commandExit() error {
	fmt.Println("Closing the Pokedex... Goodbye!")
	return fmt.Errorf("Exit")
}

func cleanInput(text string) []string {
	text = strings.ToLower(text)
	words := strings.Fields(text)
	return words
}
