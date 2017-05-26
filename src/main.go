package main;

import (
  "api"
  "log"
  "config"
  "time"
  "rpc"
)

func main() {
  conf := config.Load()
  log.Println("Starting system. Battery serial number: ", conf.SerialNo)
  rpc.ConnectToRPC(conf.Provider)
  registry_addr, _ := api.GetRegistry(conf.API)
  log.Println("Got registry address: ", registry_addr)

  // If the setup keypair was not registered, something fishy is going on
  check_registered(conf.SetupAddr, conf.WalletAddr, registry_addr)

  // Add the wallet address to the registrar
  add_wallet(conf.WalletAddr, conf.SetupAddr, conf.SetupPkey, registry_addr, conf.API)

  // System cannot proceed until agent is registered
  check_claimed(conf.WalletAddr, registry_addr)

  // Authenticate the agent to use the API
  auth_token := authenticate(conf.WalletAddr, conf.WalletPkey, conf.API)

  // Run program
  run(auth_token, conf.WalletAddr, conf.API)
}


/**
 * Main event loop. Periodically check API for data.
 *
 * @param auth_token    Used to query authenticated routes
 * @param wallet        Wallet address (identifier of the device)
 * @param hub           Full base url of the hub
 */
func run(auth_token string, wallet string, hub string) {
  for true {
    // 1. Ping the hub and ask if there are any unpaid bills. This will return
    //    amounts and ids for the bills.
    //
    // 2. Total the unpaid bills and sign a message that will move that many
    //    tokens to the address provided by the hub.
    //
    // 3. Send back the ids as well as the signed message.
    log.Println("oh hello")
    time.Sleep(time.Second*10)
  }
}


/**
 * Sanity check to make sure the device was actually registered.
 *
 * @param setup       Setup address
 * @param wallet      Address of the device's wallet
 * @param registry    Address of the registry contract
 */
func check_registered(setup string, wallet string, registry string) {
  // Check if the setup key is registered
  reg := rpc.CheckRegistered(setup, registry)
  if reg == false {
    // If a wallet has been added, we need to check to see if that one is
    // registered.
    wallet_reg := rpc.CheckRegistered(wallet, registry)
    if  wallet_reg == false {
      // If neither one is registered, someone is probably trying to spoof
      // some data. Not on our watch!
      log.Panic("Agent's setup key was not registered by Grid+")
    }
  }
  return
}


/**
 * Add a wallet address to the registry contract and wait until the transaction
 * has been succesfully mined.
 *
 * @param wallet_addr    Address of the wallet we want to register
 * @param setup_addr     Address of the setup keypair
 * @param setup_pkey     Private key of the currently registered address
 * @param registry       Address of the registry contract
 * @param _api           Full base URI for the API
 */
func add_wallet(wallet_addr string, setup_addr string, setup_pkey string, registry string, _api string) {
  // Check if the wallet is already registered. We can skip this if it is.
  reg := rpc.CheckRegistered(wallet_addr, registry)
  if reg == false {
    log.Println("Adding wallet...")
    // Form a transaction to add the wallet
    var data = "0xdeaa59df"+rpc.Zfill(wallet_addr)
    err, txhash, gas := rpc.AddWallet(setup_addr, registry, data, _api, setup_pkey)
    if err != nil {
      log.Panic("Unable to add wallet to registry", err)
    }

    // Wait until the tx is mined
    var mined = false
    for mined == false {
      gasUsed, err2 := rpc.CheckReceipt(txhash)
      if err2 != nil {
        log.Panic("Unable to get receipt ", err2)
      }
      if gasUsed < gas && gasUsed != 0 {
        mined = true
        log.Println("Wallet successfully added.")
      } else if gasUsed == gas {
        // Recursive call if the tx threw. Not sure what else to do here, since
        // we can't proceed without a wallet
        mined = true
        log.Println("Wallet could not be added (tx threw). Reattempting in 10 seconds...")
        time.Sleep(time.Second*10)
        add_wallet(wallet_addr, setup_addr, setup_pkey, registry, _api)
      }
    }
  } else {
    log.Println("Wallet already added.")
  }
  return
}

/**
 * Check if the agent has been claimed by an owner.
 *
 * @param  _agent      Address of the wallet, which must have been added to
 *                      the registry
 * @param  _registry   Address of the registry contract
 */
func check_claimed(_agent string, _registry string) {
  log.Println("Waiting for registration...")
  var reg = false
  for reg == false {
    _reg := rpc.CheckClaimed(_agent, _registry)
    if _reg != true {
      time.Sleep(time.Second*10)
    } else {
      log.Println("Agent registration confirmed.")
      reg = true
    }
  }
}


/**
 * Authenticate the device with the API. This should be called with the wallet
 * address and key.
 *
 * @param _agent    Address of the agent's wallet
 * @param _pkey     Private key for the agent's wallet
 * @param _api      Full base URI of the API
 * @return          JSON web token used for authenticated API endpoints
 */
func authenticate(_agent string, _pkey string, _api string) (string) {
  token := ""
  log.Println("Waiting for authentication...")
  for token == "" {
    _token, err := api.GetAuthToken(_agent, _pkey, _api)
    if err != nil {
      log.Println(err)
      time.Sleep(time.Second*10)
    } else {
      log.Println("Authentication successful.")
      token = _token
    }
  }
  return token
}
