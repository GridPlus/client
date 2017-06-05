// Parse config.toml and export params
package config

import (
  "encoding/hex"
  "github.com/spf13/viper"
  "log"
  "sig"
)

type Config struct {
  API string                    // Host of the Grid+ API
  Provider string               // RPC provider (including port)
  SerialNo string               // Serial number of the agent
  HashedSerialNo string         // Keccak256 hash of SerialNo
  SetupPkey string              // Agent's private key (for setup)
  SetupAddr string              // Ethereum address corresponding to private key
  WalletKeyPath string          // Absolute file path for wallet key file
  WalletPkey string             // Agent's permanent wallet key (for moving tokens)
  WalletAddr string             // Agent's wallet address
}

// Load the config file and get system-level parameters
//
func Load() (Config) {
  viper.SetConfigName("config")
  viper.AddConfigPath("config")

  _config := Config{}
  err := viper.ReadInConfig()
  if err != nil {
    log.Panic("Error reading config file", err)
  } else {
    // Get normal config data
    _config.API = viper.GetString("development.gridplus_api")
    _config.Provider = viper.GetString("development.rpc_provider")
    _config.WalletKeyPath = viper.GetString("wallet.key_path")

    // Get setup key
    viper.SetConfigName("setup_keys")
    viper.AddConfigPath("config")
    err2 := viper.ReadInConfig()
    if err2 != nil {
      log.Fatal("Could not find crypto keypair at 'config/setup_keys.toml'")
    } else {
      _config.SetupPkey = viper.GetString("agent.pkey")
      _config.SetupAddr = viper.GetString("agent.addr")
      _config.SerialNo = viper.GetString("agent.serial_no")
      if _config.SerialNo == "" {
        log.Panic("No serial number detected.")
      }
      hash := sig.Keccak256Hash([]byte(_config.SerialNo))
      hash_str := hex.EncodeToString(hash)
      _config.HashedSerialNo = hash_str
      log.Println("hashed serial", _config.HashedSerialNo)
    }

    // Create (or get) wallet key
    wallet_key, err := getKey(_config.WalletKeyPath)
    if err != nil || wallet_key == "" {
      // We can assume that any read errors mean there's no key. Let's create one.
      err2 := createKey(_config.WalletKeyPath, 32)
      if err2 != nil {
        log.Panic("Could not create wallet key:", err2)
      } else {
        wallet_key, err3 := getKey(_config.WalletKeyPath)
        if err3 != nil {
          log.Panic("Could not retrieve newly created wallet key:", err3)
        } else {
          _config.WalletPkey = wallet_key
          wallet_addr, err4 := getAddr(_config.WalletKeyPath)
          if err4 != nil {
            log.Panic("Could not derive address from wallet key", err4)
          }
          _config.WalletAddr = wallet_addr
        }
      }
    } else {
      _config.WalletPkey = wallet_key
      wallet_addr, err5 := getAddr(_config.WalletKeyPath)
      if err5 != nil {
        log.Panic("Could not derive address from wallet key", err5)
      }
      _config.WalletAddr = wallet_addr
    }

  };
  return _config
}
