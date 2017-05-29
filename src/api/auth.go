// Authenticate the battery and receive a JSON web token in return
// This JWT will expire after some period of time so this may need
// to be called periodically.
package api

import (
  "bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
  "sig"
)
import "github.com/ethereum/go-ethereum/crypto"


type StringRes struct {
  Result string
}

type FaucetReq struct {
  Addr string `json:"agent"`
}

// Authenticate the battery with the API.
// @returns - JSON web token string that must be included in authenticated endpoints.
//            This token will only be valid for a finite period of time. And once it
//            expires, this function will need to be called again for a new one.
func GetAuthToken(address string, pkey string, API string) (string, error) {
  var data = new(StringRes)
  // 1: Get the auth data to sign
  // ----------------------------
  res_data, err := http.Get(API+"/AuthDatum")
  // Data will need to be hashed
  if err != nil { return "", fmt.Errorf("Could not get authentication data: (%s)", err) }
  body, err1 := ioutil.ReadAll(res_data.Body)
  if err != nil { return "", fmt.Errorf("Could not parse authentication data: (%s)", err1) }
  err2 := json.Unmarshal(body, &data)
  if err2 != nil { return "", fmt.Errorf("Could not unmarshal authentication data: (%s)", err2) }
  // Hash the data. Keep the byte array
  data_hash := sig.Keccak256Hash([]byte(data.Result))
  // Sign the data with the private key
  privkey, err3 := crypto.HexToECDSA(pkey)
  if err3 != nil { return "", fmt.Errorf("Could not parse private key: (%s)", err3) }
  // Sign the auth data
  _sig, err4 := sig.Ecsign(data_hash, privkey)
  if err4 != nil { return "", fmt.Errorf("Could not sign with private key: (%s)", err4) }

  // 2: Send sigature, get token
  // ---------------------
  var authdata = new(StringRes)
  var jsonStr = []byte(`{"owner":"`+address+`","sig":"`+_sig+`"}`)
  res, err5 := http.Post(API+"/Authenticate", "application/json", bytes.NewBuffer(jsonStr))
  if err5 != nil { return "", fmt.Errorf("Could not hit POST /Authenticate: (%s)", err5) }
  if res.StatusCode != 200 { return "", fmt.Errorf("(%s): Error in POST /Authenticate", res.StatusCode)}
  body, err6 := ioutil.ReadAll(res.Body)
  if err6 != nil { return "" , fmt.Errorf("Could not read /Authenticate body: (%s)", err6)}
  err7 := json.Unmarshal(body, &authdata)
  if err7 != nil { return "", fmt.Errorf("Could not unmarshal /Authenticate body: (%s)", err7) }

  // Return the JSON web token
  return string(authdata.Result), nil
}



/**
 * Ask the faucet for some ether
 *
 * @param wallet        Address to send ether to
 * @param auth_token    JSON web token
 * @param api           Full base URI of api
 * @return              Transaction hash, error
 */
func Faucet(wallet string, auth_token string, api string) (string, error) {
  var result = new(StringRes)
  payload := FaucetReq{wallet}
  b, _ := json.Marshal(payload)
  client := &http.Client{}
  req, _ := http.NewRequest("POST", api+"/Faucet", bytes.NewBuffer(b))
  req.Header.Set("x-access-token", auth_token)
  req.Header.Set("Content-Type", "application/json")
  res, _ := client.Do(req)
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return "", fmt.Errorf("Could not read response body (%s)", err)
  } else {
    err2 := json.Unmarshal(body, &result)
    if err2 != nil {
      return "", fmt.Errorf("Could not unmarshal body (%s)", err)
    }
  }
  return result.Result, nil

}
