package main

import (
	"flag"
	"fmt"
)

type Person struct {
	Name string
	Age  uint
}

func main() {
	intArr := []int{1, 2, 3, 4}
	port := flag.Int("port", 8080, "Port must be a number")
	flag.Parse()
	fmt.Println(*port)
	intArr = append(intArr, 5)

	var intSlice []int = make([]int, 10, 20)

	intSlice = append(intSlice, intArr...)
	fmt.Println(intArr, intSlice)

	var myMap map[string]Person = make(map[string]Person)
	myMap["1"] = Person{Name: "Adam", Age: 12}
	fmt.Printf("myMap: %v\n", myMap)
	fmt.Printf("myMap[\"1\"]: %v\n", myMap["1"])

	for key, person := range myMap {
		fmt.Printf("key: %v\n", key)
		fmt.Printf("name: %v\n", person.Name)
		fmt.Printf("age: %v\n", person.Age)
	}

}
