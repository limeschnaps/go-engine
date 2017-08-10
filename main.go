package main

import (
	"fmt"
	"os"
	"runtime"

	"github.com/limeschnaps/go-engine/controller"
	"github.com/limeschnaps/go-engine/editor"
	"github.com/limeschnaps/go-engine/glfwController"
)

func init() {
	// Use all cpu cores
	runtime.GOMAXPROCS(runtime.NumCPU())
	//Set default glfw controller
	controller.SetDefaultConstructor(glfwController.NewActionMap)
}

func main() {
	assetDir := "."
	if len(os.Args) >= 2 {
		assetDir = os.Args[1]
	}

	fmt.Println("Using assetDir: ", assetDir)
	editor.New(assetDir).Start()
}
