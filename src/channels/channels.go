// Functions for dealing with payment channels
package channels

import "log"
import "rpc"
import "time"

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
to string, amount uint64, pkey string, API string) (string) {
  var data = "0xcfa40e4f" + rpc.Zfill(token) + rpc.Zfill(to) + rpc.Zfill(string(amount))
  var gas = uint64(150000)
  _, _gasPrice := rpc.DefaultGas(API)
  var gasPrice = _gasPrice.Uint64()
  rawtx := rpc.RawTx(from, to, data, pkey, gas, gasPrice, 0)
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
  var mined = false
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
      channel.Deposit = amount
    } else if success == -1 {
      log.Panic("Error: Could not open payment channel")
    } else {
      time.Sleep(time.Second*10)
    }
  }
  return channel.Id
}

func GetChannelId() (string) {
  return channel.Id
}
