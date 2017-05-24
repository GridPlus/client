// Util for signing transactions and messages
package sig

import (
  "crypto/ecdsa"
  "github.com/ethereum/go-ethereum/common/math"
  "github.com/ethereum/go-ethereum/crypto/secp256k1"
  "github.com/ethereum/go-ethereum/crypto/sha3"
  "encoding/hex"
  "math/big"
)
import "fmt"
// import "encoding/json"
import "github.com/ethereum/go-ethereum/core/types"
import "github.com/ethereum/go-ethereum/common"
import "github.com/ethereum/go-ethereum/crypto"

// import "math/rand"
// import "log"

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
