package main

import (
    "./pexsh"
    "os"
    "fmt"
    "strings"
)

func getEnvironmentVariable(variable string) (value string) {
	for _, env := range os.Environ() {
    	if strings.HasPrefix(env, variable + "=") {
			value = env[len(variable)+1:]
    	}
    }
    return value
}

func main() {
	if len(os.Args) <= 1 {
		fmt.Println("Pex Shell  (Copyright @2014)")
		fmt.Println("usage: pexsh conference@medianode (example: pexsh meet@10.47.2.43)")
		return
	}
	displayname := getEnvironmentVariable("USER") + "@pexsh"
	destination := os.Args[1]
    shell := pexsh.NewShell(displayname, destination)
    shell.Run()
}