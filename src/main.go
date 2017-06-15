package main;

import (
  "api"
  "channels"
  "config"
  "fmt"
  "log"
  "math"
  "os"
  "rpc"
  "time"
  "sig"
)

func main() {
  // Setup logging
  f, err := os.OpenFile("agent.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    fmt.Printf("\x1b[31;1mWARNING: Could not start logging process (%s)\x1b[0m\n", err)
  }
  defer f.Close()
  log.SetOutput(f)

  conf := config.Load()
  log.Println("Starting system. Agent serial number: ", conf.SerialNo)
  fmt.Printf("%s Starting system. Agent serial number: \x1b[4;49;33m%s\x1b[0m\n", DateStr(), conf.SerialNo)
  rpc.ConnectToRPC(conf.Provider)

  var registry_addr = ""
  var usdx_addr = ""
  for registry_addr == "" || usdx_addr == "" {
    _registry_addr, _ := api.GetRegistry(conf.API)
    registry_addr = _registry_addr
    _usdx_addr, _ := api.GetUSDX(conf.API)
    usdx_addr = _usdx_addr
    if usdx_addr == "" || registry_addr == "" {
      time.Sleep(time.Second*10)
    }
  }

  // If the setup keypair was not registered, something fishy is going on
  check_registered(conf.HashedSerialNo, conf.WalletAddr, registry_addr)

  // Add the wallet address to the registrar
  add_wallet(conf.WalletAddr, conf.HashedSerialNo, conf.SetupAddr, conf.SetupPkey, registry_addr, conf.API)
  // System cannot proceed until agent is registered
  check_claimed(conf.HashedSerialNo, registry_addr)

  // Authenticate the agent to use the API
  auth_token := authenticate(conf.WalletAddr, conf.WalletPkey, conf.API)

  // Get the ether balance
  balance := rpc.EtherBalance(conf.WalletAddr)
  fmt.Printf("%s Balance: \x1b[32m%d\x1b[0m wei\n", DateStr(), balance)
  fmt.Printf("\x1b[32m%s Setup complete. Running.\x1b[0m\n", DateStr())
  // Run program
  run(auth_token, conf.WalletAddr, conf.HashedSerialNo, usdx_addr, conf.API, conf.WalletPkey)
}


/**
 * Main event loop. Periodically check API for data.
 *
 * @param auth_token    Used to query authenticated routes
 * @param wallet        Wallet address (identifier of the device)
 * @param serial_hash   Hash of agent's serial number
 * @param usdx          Address of USDX token contract
 * @param hub           Full base url of the hub
 * @param pkey          Private key of the wallet
 */
func run(auth_token string, wallet string, serial_hash string, usdx string, hub string, pkey string) {
  var hub_addr = ""
  var channels_addr = ""
  var channel_balance = 0

  for hub_addr == "" || channels_addr == "" {
    // Get the addresses from the API
    _hub_addr, _ := api.GetHubAddr(hub)
    hub_addr = _hub_addr
    _channels_addr, _ := api.GetChannelsAddr(hub)
    channels_addr = _channels_addr
    if hub_addr == "" || channels_addr == "" {
      time.Sleep(time.Second*10)
    }
  }

  channel_id := channels.CheckForChanneId(wallet, hub_addr, channels_addr)
  if channel_id != "" {
    fmt.Printf("%s Found existing payment channel: \x1b[32m%s\x1b[0m \n", DateStr(), channel_id)
  }

  for true {
    // Make sure ether balance is high enough to send a transaction.
    // NOTE: We won't be sending a transaction, but we need to make sure if
    // the tx gets played by the hub, it will go through
    gas, gasPrice := rpc.DefaultGas(hub)
    needed := gas.Uint64()*gasPrice.Uint64()
    check_ether(needed, wallet, serial_hash, auth_token, hub)

    // Open a payment channel if one is needed. This will skip if the existing
    // channel is still good.
    _channel_id := handle_channel(wallet, channels_addr, hub_addr, usdx, hub, pkey)
    channel_id = _channel_id

    // 1. Ping the hub and ask if there are any unpaid bills. This will return
    //    amounts and ids for the bills.
    bills, err := api.GetBills(serial_hash, hub, auth_token)
    if err != nil {
      fmt.Printf("\x1b[91m%s ERROR: Failed to get unpaid bills (%e)\x1b[0m\n", DateStr(), err)
      log.Println("Encountered error getting bills (%s)", err)
    } else {

      // 3. Get the total amount committed to the channel
      channel_sum, err3 := api.GetChannelSum(channel_id, hub, auth_token) // Total amount already commited to channel
      if err3 != nil {
        fmt.Printf("\x1b[91m%s ERROR: Failed to get channel sum (%e)\x1b[0m\n", DateStr(), err3)
        log.Println("Encountered error getting bills (%s)", err)
      } else {
        // 2. Total the unpaid bills and sign a message that will move that many
        //    tokens to the address provided by the hub.
        var unpaid_sum float64
        var unpaid_bill_ids []int
        for _, bill := range *bills {
          unpaid_sum += bill.Amount
          unpaid_bill_ids = append(unpaid_bill_ids, bill.BillId)
        }

        if unpaid_sum > 0 {
          // ascii colors: http://misc.flogisoft.com/_media/bash/colors_format/colors_and_formatting.sh.png
          fmt.Printf("%s Unpaid amount: \x1b[91m$%.6f\x1b[0m\n", DateStr(), unpaid_sum)

          // 3. Get balance in the channel
          decimals := float64(rpc.TokenDecimals(wallet, usdx))
          // Total amount available to channel
          channel_deposit := float64(channels.GetDeposit())
          // Balance of the device (external to channel)
          token_balance := float64(rpc.TokenBalance(wallet, usdx)) / math.Pow(10, decimals)
          // Total remainder (in dollars) of the channel
          var usd_balance = (channel_deposit-channel_sum)/(math.Pow(10, decimals))


          if usd_balance >= unpaid_sum {
            // Round to the nearest USDX atomic unit
            var to_pay = int(math.Ceil(channel_sum + (unpaid_sum * math.Pow(10, decimals)) ))
            // Sign message that will be sent to the payment channel by the hub
            to_pay_hex := fmt.Sprintf("%x", int64(to_pay))
            proof := sig.SignPayment(channel_id, to_pay_hex, pkey)

            // Load up the request payload
            var payload = api.BillPayReq{}
            payload.BillIds = unpaid_bill_ids
            payload.Msg = proof.MsgHash
            payload.V = proof.V
            payload.R = proof.R
            payload.S = proof.S
            payload.Value = proof.Value

            err, ids, remaining := api.PayBills(&payload, hub, auth_token)
            if err != nil {
              fmt.Printf("\x1b[91m%s ERROR: Failed to pay bills.\x1b[0m\n", DateStr())
            } else {
              channel_balance = remaining
              var channel_bal_disp = float64(channel_balance)/(math.Pow(10, decimals))
              fmt.Printf("\x1b[32m%s Successfully paid %d bills.\x1b[0m\n", DateStr(), len(ids))
              fmt.Printf("%s Channel balance: \x1b[32m$%.6f\x1b[0m USDX reserve: \x1b[32m$%.6f\x1b[0m\n", DateStr(), channel_bal_disp, token_balance)
            }
          } else {
            fmt.Println("unpaid sum", unpaid_sum)
            fmt.Printf("\x1b[91m%s ERROR: Insufficient balance to pay bills.\x1b[0m\n", DateStr())
          }
        }

      }
    }

    // Wait 10 seconds and execute again
    time.Sleep(time.Second*10)
  }
}

/**
 * Set up a payment channel if one does not exist. Load it up with a default
 * amount of USDX tokens.
 *
 * @param wallet              Address of this device's wallet
 * @param channels_addr       Address of the payment channel contract
 * @param hub_addr            Address of the admin to pay
 * @param usdx                Address of token contract
 * @param hub                 Full base URI of the hub API
 * @param pkey                Private key of wallet
 */
func handle_channel(wallet string, channels_addr string, hub_addr string, usdx string,
hub string, pkey string) (string) {
  id := channels.CheckForChanneId(wallet, hub_addr, channels_addr)
  // Open a channel with the existing token balance
  balance := rpc.TokenBalance(wallet, usdx)
  HARD_MIN := uint64(500000000)  // Minimum of $5 deposited to open a channel
  err_disp := false
  if id == "" {
    // Make sure the balance is high enough
    for balance < HARD_MIN {
      _balance := rpc.TokenBalance(wallet, usdx)
      if _balance < HARD_MIN {
        if err_disp == false {
          fmt.Printf("\x1b[31;1mInsufficient token balance to open channel. Need at least %d, have %d. Please deposit funds.\x1b[0m\n", HARD_MIN, _balance)
          err_disp = true
        }
        time.Sleep(time.Second*10)
      }  else {
        balance = _balance
      }
    }
    // If the balance is high enough, open a channel
    id := channels.OpenChannel(wallet, channels_addr, usdx, hub_addr, balance, pkey, hub)
    fmt.Printf("%s Opened new payment channel: \x1b[32m%s\x1b[0m \n", DateStr(), id)
  }
  return id
}

/**
 * Sanity check to make sure the device was actually registered.
 *
 * @param serial_hash    Serial number.
 * @param wallet         Address of the device's wallet
 * @param registry       Address of the registry contract
 */
func check_registered(serial_hash string, wallet string, registry string) {
  // Check if the setup key is registered
  reg := rpc.CheckRegistered(wallet, serial_hash, registry)
  if reg == false {
    // If it isn't registered, someone is probably trying to spoof some data.
    log.Panic("Serial number not registered with Grid+")
  }
  return
}


/**
 * Add a wallet address to the registry contract and wait until the transaction
 * has been succesfully mined.
 *
 * @param wallet_addr    Address of the wallet we want to register
 * @param hashed_serial  Keccak256 hash of the serial number
 * @param setup_addr     Address of the setup keypair
 * @param setup_pkey     Private key of the currently registered address
 * @param registry       Address of the registry contract
 * @param _api           Full base URI for the API
 */
func add_wallet(wallet_addr string, hashed_serial string, setup_addr string,
setup_pkey string, registry string, _api string) {
  added := rpc.CheckRegistry(setup_addr, hashed_serial, wallet_addr, registry)
  if added == false {
    log.Println("Adding wallet...")
    fmt.Printf("%s Adding wallet...\n", DateStr())

    // Form a transaction to add the wallet
    var data = "0xb993b3f5"+rpc.Zfill(wallet_addr)+hashed_serial
    err, txhash := rpc.AddWallet(setup_addr, registry, data, _api, setup_pkey)
    if err != nil {
      log.Panic("Unable to add wallet to registry", err)
    }

    // Wait until the tx is mined
    var mined = false
    for mined == false {
      success, err2 := rpc.CheckReceipt(txhash)
      if err2 != nil {
        log.Panic("Unable to get receipt ", err2)
      }
      if success == 1 {
        mined = true
        fmt.Printf("%s Wallet successfully added.\n", DateStr())
        log.Println("Wallet successfully added.")
      } else if success == -1 {
        // Recursive call if the tx threw. Not sure what else to do here, since
        // we can't proceed without a wallet
        mined = true
        log.Println("Wallet could not be added (tx threw). Reattempting in 10 seconds...")
        time.Sleep(time.Second*10)
        add_wallet(wallet_addr, hashed_serial, setup_addr, setup_pkey, registry, _api)
      } else {
        time.Sleep(time.Second*10)
      }
    }
  } else {
    log.Println("Wallet already registered. Skipping.")
  }

  return
}

/**
 * Check if the agent has been claimed by an owner.
 *
 * @param  serial_hash    Keccak256 hash of the serial number
 * @param  registry      Address of the registry contract
 */
func check_claimed(serial_hash string, registry string) {
  log.Println("Waiting for agent to be claimed...")
  fmt.Printf("%s Waiting for agent to be claimed...\n", DateStr())
  var reg = false
  for reg == false {
    _reg := rpc.CheckClaimed(serial_hash, registry)
    if _reg != true {
      time.Sleep(time.Second*10)
    } else {
      log.Println("Agent claimed and fully registered.")
      fmt.Printf("%s Agent claimed and fully registered.\n", DateStr())
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

func DateStr() (string) {
  return time.Now().UTC().Format(time.UnixDate)+": "
}


/**
 * Make sure the ether balance is high enough to cover what is needed.
 * If it isn't call the faucet and wait until it is
 *
 * @param  needed        Number of wei needed to proceed
 * @param  wallet        Address to check and call the faucet for
 * @param serial_hash   Hash of agent's serial number
 * @param  auth_token    JSON web token to call the faucet with
 * @param  API           Full base URI of API
 */
func check_ether(needed uint64, wallet string, serial_hash string, auth_token string, API string) {
  balance := rpc.EtherBalance(wallet)
  if balance < needed {
    fmt.Printf("%s Balance: \x1b[91m%d\x1b[0m wei. Calling faucet.\n", DateStr(), balance)
  }
  for balance < needed {
    // Call the faucet and wait for the transaction to clear
    var done = false
    txhash, err := api.Faucet(serial_hash, wallet, auth_token, API)
    if err != nil {
      fmt.Printf("\x1b[91m%d\x1b[0m %s Error encountered calling /Faucet.\n", DateStr(), balance)
      time.Sleep(time.Second*30)
      check_ether(needed, wallet, serial_hash, auth_token, API)
    }
    for done == false {
      _success, _ := rpc.CheckReceipt(txhash)
      if _success == 0 {
        time.Sleep(time.Second * 5)
      } else {
        done = true
      }
    }
    // Update the balance and see if we need more faucet (we shouldn't)
    balance = rpc.EtherBalance(wallet)
    fmt.Printf("%s New balance: \x1b[32m%d\x1b[0m wei\n", DateStr(), balance)
  }
}
