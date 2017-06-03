package main;

import (
  "fmt"
  "os"
  "log"
  "config"
  "path/filepath"
)


func main() {
  writeKeyFile()
  writeConfigFile()
}

/**
 * Create a setup key and write it to a config file in plaintext. This will
 * only be used once. For demo only.
 */
func writeKeyFile() {
  // Write key and address to a toml file
  addr, pkey := keygen()
  f, err := os.Create("../src/config/setup_keys.toml")
  if err != nil {
    log.Panic("Could not create or open setup_keys.toml")
  }
  defer f.Close()

  var s = fmt.Sprintf("[agent]\naddr = \"%s\"\npkey = \"%s\"", addr, pkey)
  _, err2 := f.WriteString(s)
  if err2 != nil {
    log.Panic("Could not write your address or pkey")
  }
  return
}


/**
 * Write a config file to be used by the client. This is a router to important
 * services and directories.
 */
func writeConfigFile() {
  f, err := os.Create("../src/config/config.toml")
  if err != nil {
    log.Panic("Could not open config.toml")
  }
  defer f.Close()

  var api = "http://localhost:3000" // Placeholder
  var rpc = "http://localhost:8545" // Placeholder
  dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))

  if len(os.Args) < 2 {
    log.Panic("You must provide a serial number.")
  }

  var s = fmt.Sprintf(`[development]
gridplus_api = "%s"
rpc_provider = "%s"
serial_no = "%s"
[wallet]
key_path = "%s/../src/config"`, api, rpc, os.Args[1], dir)

  _, err2 := f.WriteString(s)
  if err2 != nil {
    log.Panic("Could not write config file")
  }
  return
}

/**
 * Generate a key and an address
 */
func keygen() (string, string) {
  priv, _ := config.GenerateRandomBytes(32)
  addr := config.PrivateToAddress(priv)
  return "0x"+addr, fmt.Sprintf("%x", priv)
}
