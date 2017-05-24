/*
Everything related to creating, saving, and loading a private key
as well as recovering an Ethereum address from it
 */
package config

import (
    "crypto/rand"
    "encoding/hex"
    "io/ioutil"
    secp256k1 "github.com/haltingstate/secp256k1-go"
    "github.com/ebfe/keccak"
)

/**
 * Generate a private key and save it to disk
 *
 * @param path {string} - absolute file path from which to get the key
 * @param n {int}       - number of bytes to randomly generate. Should be 32 always
 * @returns (error)
 */
func createKey(path string, n int) (error) {
  b, err := GenerateRandomBytes(n)
  if (err == nil) {
    err2 := keyToFile(b, path)
    if (err2 != nil) { return err2 }
  } else {
    return err
  }
  return nil
}

/**
 * Get the private key from a file
 *
 * @param path {string} - absolute file path from which to get the key
 * @returns (string, error) - private key (hex), error
 */
func getKey(path string) (string, error) {
  b, err := keyFromFile(path)
  if err != nil { return "", err }
  key := hex.EncodeToString(b)
  return key, err
}

func getAddr(path string) (string, error) {
  b, err := keyFromFile(path)
  if err != nil { return "", err }
  addr := PrivateToAddress(b)
  return "0x"+addr, err
}

/**
 * Dump bytes to a file
 *
 * @param b {bytes} - arbitrary byte array
 * @param fpath {string} - absolute file path in which to save the key
 * @returns (error)
 */
func keyToFile(b []byte, fpath string) (error) {
  err := ioutil.WriteFile(fpath+"/wallet.pem", b, 0644)
  return err
}

/**
 * Read a file containing a byte array
 *
 * @param fpath {string} - absolute file path from which to read the key
 * @returns ([]byte, error) - private key, error
 */
func keyFromFile(fpath string) ([]byte, error) {
  b, err := ioutil.ReadFile(fpath+"/wallet.pem")
  return b, err
}

/**
 * Generate some random bytes
 *
 * @param n {int} - number of bytes to generate
 * @returns []bytes, error - byte array and error object
 */
func GenerateRandomBytes(n int) ([]byte, error) {
  b := make([]byte, n)
  _, err := rand.Read(b)
  if (err != nil) { return nil, err }
  return b, nil
}

/**
 * Convert a private key to an Ethereum address
 * @param  {buffer} privateKey - A buffered 32-byte private key
 * @return {string}            - The Ethereum address
 */
func PrivateToAddress(priv []byte) (string){
  // Recover the public key from private key using
  // bitcoin secp256k1 function
  pub := privateToPublic(priv)
  // Generate a new keccak256 hash object
  h := keccak.New256()
  // Add the public key to the has object
  // NOTE: we remove the first byte (for some reason)
  h.Write(pub[1:])
  hash := h.Sum(nil)
  // Encode the hash object to hex string
  // NOTE: we remove the first 12 bytes (again, for some reason)
  hashHex := hex.EncodeToString(hash[12:32])

  return hashHex
};

/**
* Returns the ethereum public key of a given private key
* @method privateToPublic
* @param {[]byte} privateKey A private key must be 256 bits (32 bytes) wide
* @return {[]byte}
*/
func privateToPublic(priv []byte) ([]byte){
  // Using the Bitcoin secp256k1 library, get a public key from
  // our private key.
  // NOTE: This is the uncomressed (65 byte) version of the key
  pub := secp256k1.UncompressedPubkeyFromSeckey(priv)
  return pub
}
