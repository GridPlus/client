// Get the automated trading profile
package api

import (
  "encoding/json"
  "io/ioutil"
  "fmt"
  "net/http"
)

type Profile struct {
  Trading bool `json:"trading"`
  Trade_on_price bool `json:"trade_on_price"`
  Min_price float32 `json:"min_price"`
  Max_price float32 `json:"max_price"`
  Numeraire string `json:"numeraire"`
}

var DEFAULT_GAS = 100000
var DEFAULT_GASPRICE = 2000000000

// Load the automated trading profile
// @returns Profile object
//
func GetProfile(token string, API string) (*Profile, error){
  client := &http.Client{}
  var profile = new(Profile)
  // Query the API for a trading profile
  req, err := http.NewRequest("GET", API+"/Profile", nil)
  if err != nil { return nil, fmt.Errorf("Could not hit GET /Profile (%s)", err) }
  req.Header.Set("x-access-token", token)
  res, err2 := client.Do(req)
  if err2 != nil { return nil, fmt.Errorf("Could not GET /Profile (%s)", err2) }
  body, err3 := ioutil.ReadAll(res.Body)
  if err3 != nil { return nil, fmt.Errorf("Could not read GET /Profile res (%s)", err3) }
  if res.StatusCode != 200 { return nil, fmt.Errorf("(%d): Error in GET /Profile", res.StatusCode) }
  defer res.Body.Close()
  err4 := json.Unmarshal(body, &profile)
  if err4 != nil { return nil, fmt.Errorf("Could not unmarshal (%s)", err4)}
  fmt.Println("numeraire", profile.Numeraire)
  return profile, nil
}
