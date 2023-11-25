package main

import "os"

func main() {
	defer os.Exit(1) // want "avoid direct os.Exit calls in main"
	func() {
		os.Exit(1) // want "avoid direct os.Exit calls in main"
	}()
	if false {
		os.Exit(1) // want "avoid direct os.Exit calls in main"
	}
	go os.Exit(1) // want "avoid direct os.Exit calls in main"

	notmain()

	// indirect call
	f := os.Exit
	f(1)
	os.Exit(1) // want "avoid direct os.Exit calls in main"

}

func notmain() {
	os.Exit(1)
}
