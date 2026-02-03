package main

import (
	"fmt"
	"reflect"

	"github.com/ory/fosite"
	"github.com/ory/fosite/compose"
)

func main() {
	// 1. Compose Signature
	t := reflect.TypeOf(compose.Compose)
	fmt.Printf("Compose Signature: %v\n", t)

	// 2. Hasher Interface
	t2 := reflect.TypeOf((*fosite.Hasher)(nil)).Elem()
	fmt.Println("\nfosite.Hasher Methods:")
	for i := 0; i < t2.NumMethod(); i++ {
		m := t2.Method(i)
		fmt.Printf(" - %s: %v\n", m.Name, m.Type)
	}

	// 3. Config Fields
	tConfig := reflect.TypeOf(fosite.Config{})
	fmt.Println("\nfosite.Config Fields:")
	for i := 0; i < tConfig.NumField(); i++ {
		f := tConfig.Field(i)
		fmt.Printf(" - %s: %v\n", f.Name, f.Type)
	}

	// 4. Fosite Fields
	tFosite := reflect.TypeOf(fosite.Fosite{})
	fmt.Println("\nfosite.Fosite Fields:")
	for i := 0; i < tFosite.NumField(); i++ {
		f := tFosite.Field(i)
		fmt.Printf(" - %s: %v\n", f.Name, f.Type)
	}
}
