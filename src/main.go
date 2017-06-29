package main;

import (
  "setup"
)

func main() {
  // Initialize the program
  data := setup.Init()
  // Run program
  setup.Run(data[0], data[1], data[2], data[3], data[4], data[5])
}
