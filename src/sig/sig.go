// Util for signing transactions and messages
package sig;

import (
  "crypto/ecdsa"
  "github.com/ethereum/go-ethereum/common/math"
  "github.com/ethereum/go-ethereum/crypto/secp256k1"
  "github.com/ethereum/go-ethereum/crypto/sha3"
  "encoding/hex"
  "math/big"
  "strconv"
)
import "fmt"
import "github.com/ethereum/go-ethereum/core/types"
import "github.com/ethereum/go-ethereum/common"
import "github.com/ethereum/go-ethereum/crypto"


type ChannelMsg struct {
  MsgHash string `json:"msg_hash"`
  Value string `json:"value"`
  V string `json:"v"`
  R string `json:"r"`
  S string `json:"s"`
}

/**
 * Hash a byte array. This is a keccak 256 sha3 hash.
 */
func Keccak256Hash(data []byte) []byte {
	d := sha3.NewKeccak256()
	d.Write(data)
	return d.Sum(nil)
}


/**
 * Sign a byte array with a private key instance.
 *
 * @param  hash    Byte array representing a hashed message.
 * @param  prv    Instance of ecdsa PrivateKey object
 * @return        (string, error)
 *                  Where the string is a byte array of form [R || S || V] format
 *                  where V is 0 or 1.
 */
func Ecsign(hash []byte, prv *ecdsa.PrivateKey) (string, error) {
	seckey := math.PaddedBigBytes(prv.D, prv.Params().BitSize/8)
	sig, err := secp256k1.Sign(hash, seckey)
  if err != nil { return "", err }
  return hex.EncodeToString(sig), nil
}

/**
 * Convert transaction parameters into a raw transaction string that can be sent
 * to our RPC provider directly.
 */
func GetRawTx(
    chainID int64,
    from string,
    _to string,
    data string,
    nonce uint64,
    value int64,
    gasLimit *big.Int,
    gasPrice *big.Int,
    privkey *ecdsa.PrivateKey) (string, error) {

    var amount = big.NewInt(value)
    var bytesto [20]byte
    _bytesto, _ := hex.DecodeString(_to[2:])
    copy(bytesto[:], _bytesto)
    to := common.Address([20]byte(bytesto))

    // Create a new signer with the chain id (this is net.version from web3)
    signer := types.NewEIP155Signer(big.NewInt(chainID))
    // Note the recasting of our data string to a geth common data type
    tx := types.NewTransaction(nonce, to, amount, gasLimit, gasPrice, common.FromHex(data))
    // Sign the tx with our private key and transform the Transaction object
    signature, _ := crypto.Sign(tx.SigHash(signer).Bytes(), privkey)
    signed_tx, _ := tx.WithSignature(signer, signature)
    // Recast to a "Transactions" object and get the RLP payload (raw transaction)
    t := types.Transactions{signed_tx}
    return fmt.Sprintf("0x%x", t.GetRlp(0)), nil
}


/**
 * Sign a message that will be sent to a payment channel.
 *
 * @param  channel_id    0x-prefixed bytes32 id of payment channel
 * @param  amount        amount to send (hex string)
 * @param  pkey          Private key of signer
 * @return               Message, signature, and amount
 */
func SignPayment(channel_id string, amount string, pkey string) (*ChannelMsg) {
  var resp = ChannelMsg{}

  // Form the message to be signed sha3(channel_id, value)
  var str = zfill(channel_id) + zfill(amount)
  msg, _ := hex.DecodeString(str)
  msg_hash := Keccak256Hash(msg)

  // Instantiate a private key oject for signature
  privkey, _ := crypto.HexToECDSA(pkey)

  // Sign the message and deconstruct the signature
  sig, _ := Ecsign(msg_hash, privkey)
  resp.R = sig[:64]
  resp.S = sig[64:128]
  v, _ := strconv.ParseUint(sig[129:], 0, 64)
  resp.V = fmt.Sprintf("%x", v + 27)
  resp.MsgHash = fmt.Sprintf("%x",msg_hash)
  resp.Value = amount
  return &resp
}


// Same as rpc.Zfill, but rpc import isn't allowed in this module
func zfill(s string) (string) {
  // Cut off any rouge 0x prefixes
  if (s[:2] == "0x") { s = s[2:]}
  var pad = ""
  for i := 0; i < (64-len(s)); i++ {
		pad += "0"
	}
  return pad + s
}
