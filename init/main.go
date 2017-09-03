// Write a config file with a wallet key path as well as URIs of services.
package main;

import (
  "fmt"
  "os"
  "log"
  "config"
  "path/filepath"
)


func main() {
  writeConfigFile()
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

  var api = "https://app.gridplus.io:3001"
  var rpc = "http://app.gridplus.io:8545"
  if len(os.Args) > 1 && os.Args[1] == "dev" {
    api = "http://localhost:3000"
    rpc = "http://localhost:8545"
  }
  dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))


  var s = fmt.Sprintf(`[development]
gridplus_api = "%s"
rpc_provider = "%s"
[wallet]
key_path = "%s/../src/config"`, api, rpc, dir)

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
