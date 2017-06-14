// Functions for dealing with payment channels
package channels

import "log"
import "rpc"
import "time"
import "fmt"
type Channel struct {
  Id string `json:"id"`
  Token string `json:"token"`
  Recipient string `json:"recipient"`
  Deposit uint64 `json:"deposit"`
}

var channel = Channel{}

/**
 * Make initial connection to RPC provider. Save that connection in memory.
 *
 * @param provider    Full URI of the RPC provider, including the protocol
 *                    and port
 */
func OpenChannel(from string, channel_addr string, token string,
to string, _amount uint64, pkey string, API string) (string) {
  amount := fmt.Sprintf("%x", _amount)

  // 1. Set an allowance
  var allowance_data = "0x095ea7b3" + rpc.Zfill(channel_addr) + rpc.Zfill(string(amount))
  allowance_tx := rpc.DefaultRawTx(from, token, allowance_data, pkey, API)
  var allowance_txhash = ""
  for allowance_txhash == "" {
    err, _allow_txhash := rpc.SendRaw(allowance_tx)
    if err != nil {
      log.Print("Error setting allowance (%s)", err)
      time.Sleep(time.Second*10)
    } else {
      allowance_txhash = _allow_txhash
    }
  }
  // Wait until the tx is mined
  var mined = false
  for mined == false {
    success, _ := rpc.CheckReceipt(allowance_txhash)
    if success == 1 {
      mined = true
    } else if success == -1 {
      log.Panic("Error: Could not set allowance")
    } else {
      time.Sleep(time.Second*10)
    }
  }

  // 2. Open the channel
  var data = "0xcfa40e4f" + rpc.Zfill(token) + rpc.Zfill(to) + rpc.Zfill(string(amount))
  var gas = uint64(150000)
  _, _gasPrice := rpc.DefaultGas(API)
  var gasPrice = _gasPrice.Uint64()
  rawtx := rpc.RawTx(from, channel_addr, data, pkey, gas, gasPrice, 0)

  var txhash = ""
  for txhash == "" {
    err, _txhash := rpc.SendRaw(rawtx)
    if err != nil {
      log.Print("Error opening channel (%s)", err)
      time.Sleep(time.Second*10)
    } else {
      txhash = _txhash
    }
  }
  // Wait until the tx is mined
  mined = false
  for mined == false {
    success, _ := rpc.CheckReceipt(txhash)
    if success == 1 {
      mined = true
      // Get the channel id and record it
      for channel.Id == "" {
        var data = "0x2460ee73" + rpc.Zfill(from) + rpc.Zfill(to)
        _, id := rpc.MakeCall(from, channel_addr, data)
        channel.Id = id
        if channel.Id == "" {
          time.Sleep(time.Second*10)
        }
      }
      // Fill in the rest of the channel info
      channel.Token = token
      channel.Recipient = to
      channel.Deposit = _amount
    } else if success == -1 {
      log.Panic("Error: Could not open payment channel")
    } else {
      time.Sleep(time.Second*10)
    }
  }

  return channel.Id
}


/**
 * Check the blockchain for an existing channel
 *
 * @param from              Channel spender
 * @param to                Channel recipient
 * @param channels_addr     Channel contract address
 * @return                  Id of existing channel or ""
 */
func CheckForChanneId(from string, to string, channels_addr string) (string) {
  var data = "0x2460ee73" + rpc.Zfill(from) + rpc.Zfill(to)
  err, id := rpc.MakeCall(from, channels_addr, data)
  if err != nil {
    log.Print("Could not get channel", err)
    return ""
  } else if id == "" {
    return ""
  } else if id == "0x0000000000000000000000000000000000000000000000000000000000000000" {
    return ""
  } else {
    channel.Id = id
    return id
  }
}

func GetChannelId() (string) {
  return channel.Id
}
