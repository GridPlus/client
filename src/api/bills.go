// Get bills from the API
package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Bill struct {
  BillId int `json:"bill_id"`
  Amount float64 `json:"amount"`
}

type GetBillReq struct {
	SerialHash string `json:"serial_hash"`
}

type ChannelSumReq struct  {
	ChannelId string `json:"channel_id"`
}

type ChannelSumRes struct {
	Result float64 `json:"result"`
}

type GetBillRes struct {
  Result []Bill `json:"result"`
}

type PayBillsData struct {
	PaidIds []int `json:"paid_ids"`
	BalanceRemaining int `json:"bal_remaining"`
}

type PayBillsRes struct {
	Result PayBillsData `json:"result"`
}

type BillPayReq struct {
	BillIds []int `json:"bill_ids"`
	Msg string `json:"msg"`
	V string `json:"v"`
	R string `json:"r"`
	S string `json:"s"`
	Value string `json:"value"`
}

/**
 * Get an array of Bill objects from the API. This is an authenticated request,
 * so a valid JSON web token must be included
 *
 * @param  serial_hash    Needed for request
 * @param  api            Base URI for the hub API
 * @param  token          JSON web token for the agent
 * @return                (array of bills, error)
 */
func GetBills(serial_hash string, api string, token string) (*[]Bill, error) {
	var result = new(GetBillRes)

	payload := GetBillReq{serial_hash}
	b, _ := json.Marshal(payload)

  client := &http.Client{}
	req, _ := http.NewRequest("POST", api+"/Bills", bytes.NewBuffer(b))
  req.Header.Set("x-access-token", token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return nil, fmt.Errorf("Could not read response body (%s)", err)
  } else {
    err2 := json.Unmarshal(body, &result)
    if err2 != nil {
      return nil, fmt.Errorf("Could not unmarshal body (%s)", err)
    }
  }
  return &result.Result, nil
}

/**
 * Get an array of Bill objects from the API. This is an authenticated request,
 * so a valid JSON web token must be included
 *
 * @param  payload       Filled in BillPayReq object
 * @param  api           Base URI for the hub API
 * @param  auth_token    JSON web token for the agent
 * @return               (array of bill ids, error)
 */
func PayBills(payload *BillPayReq, api string, auth_token string) (error, []int, int) {
	b, _ := json.Marshal(payload)
	var result = new(PayBillsRes)
	client := &http.Client{}
	req, _ := http.NewRequest("POST", api+"/PayBills", bytes.NewBuffer(b))
  req.Header.Set("x-access-token", auth_token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := client.Do(req)
  body, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return fmt.Errorf("Could not read response body (%s)", err), nil, 0
  } else {
    err2 := json.Unmarshal(body, &result)
    if err2 != nil {
      return fmt.Errorf("Could not unmarshal body (%s)", err2), nil, 0
    }
  }
  return nil, result.Result.PaidIds, result.Result.BalanceRemaining
}



/**
 * Get the total amount of USD that has been commited to the channel.
 * Note that this is in USD, not tokens
 *
 * @param  id            bytes32 id of the payment channel in question
 * @param  api           Full base uri of hub API
 * @param  auth_token    JSON web token
 * @return
 */
func GetChannelSum(id string, api string, auth_token string) (float64, error) {
	var result = new(ChannelSumRes)

	payload := ChannelSumReq{id}
	b, _ := json.Marshal(payload)

	client := &http.Client{}
	req, _ := http.NewRequest("POST", api+"/ChannelSum", bytes.NewBuffer(b))
	req.Header.Set("x-access-token", auth_token)
	req.Header.Set("Content-Type", "application/json")
	res, _ := client.Do(req)
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return 0, fmt.Errorf("Could not read response body (%s)", err)
	} else {
		err2 := json.Unmarshal(body, &result)
		if err2 != nil {
			return 0, fmt.Errorf("Could not unmarshal body (%s)", err)
		}
	}
	return result.Result, nil

}
