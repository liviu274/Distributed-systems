package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Person struct {
	Name     string
	Age      int
	Adresses []string
}

func main() {
	var per1 = Person{"Andrei", 23, []string{
		"123 Main St, Bucharest",
		"45 Calea Victoriei, Bucharest",
		"78 Strada Lipscani, Bucharest",
	}}
	per1B, err := json.Marshal(per1)
	if err != nil {
		panic(err)
	}
	if err := os.WriteFile("per1.json", per1B, 0644); err != nil {
		panic(err)
	}
	per2B, err := os.ReadFile("per1.json")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(per2B))
	// fmt.Println(string(per1B))
	var data = Person{}
	if err := json.Unmarshal(per1B, &data); err != nil {
		panic(err)
	}
	// fmt.Println(data)
}
