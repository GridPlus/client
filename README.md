![Logo](images/color-logo.png)

# Grid+ Agent Client

The Grid+ Agent Client is the official way to connect to the Grid+ hub and pay for your home's electricity on the agent device. The following guide allows a developer to setup the agent client. 

This process is automated in production and the below guide is only relevant for the Grid+ testing environment.

## 1. Generate a setup key pair

The first step is set up a simulated agent device. This consists of three pieces of information:
* **Address:** the agent's setup address. This is seeded with a small amount of ether and used to overwrite itself in the Registry contract with whatever address the agent generates on its own once it first boots up
* **Private key:** A random string from which the setup address is derived. Used to sign the message overwriting the address.
* **Serial:** The serial number. This would come printed on the agent and is assigned by Grid+. This number cannot be changed and is associated with the physical agent device.
* **Serial Hash:** A keccak-256 hash of the serial number, provided for convenience. This is typically the identified needed to update the registry.

To generate this information, call the following endpoint:

```
https://app.gridplus.io:3001/SetupKey/:user
```

**NOTE: This endpoint requires two blocks to process on the Ropsten network. Depending on the network's latency, this request may time out or fail. If it does, wait ~5 minutes and try it again. Please be gentle.**

`:user` can be any string you wish to use as an identifier. If you lose your setup information, you can call this endpoint to retrieve it at any time.

## 2. Setup your agent client

Before installing the agent client, you will need to have Go (programming language) installed. If you are on OSX, you can do this with homebrew (`brew install golang`).

The next step is to create a file in the repo directory with the information you got from step 1: `src/config/setup_keys.toml`. An example `setup_keys.toml` file looks like this:

```
[agent]
addr = "0x2a919a8ff288615fb1381ff1a582b826d412dab2"
pkey = "1aec3339a5388d3c165f7d0dd35e5c16acad31eb311f1526b920d410636a6028"
serial_no = "726a686c68f"
```

Once that file is saved, you can install the client with:

```
bash install.sh
```

This will install the prerequisites (via `go get`) and then it will generate a
private key for your simulated device and put it in the proper config file.

Now you can run the agent:
```
bash run.sh
```

## 3. Claim ownership of your device

Once your agent is ready, it will print `Waiting for agent to be claimed...` to your console.

You will now need to generate an Ethereum wallet and determine your address. Make sure you are connected to the Ropsten network. You can either use [MyEtherWallet](https://myetherwallet.com) or connect to a local node (such as [geth](https://github.com/ethereum/go-ethereum) or [parity](https://github.com/paritytech/parity)) that is synced to the Ropsten network.

You will need to make a transaction to the Ethereum network telling the Registry contract that you own the device you are claiming.

First, get the address of the Registry contract by calling the Grid+ endpoint `https://app.gridplus.io/Registry`.

Next, form the data for the transaction you will send:

```
0xbd66528a[your serial hash]
```

Include this in a transaction from your Ethereum wallet and send it to the registry address provided by the Grid+ API endpoint.

If the transaction is formed correctly, once it gets mined your agent should print the following to its console:

```
Setup complete. Running.
```

Your agent is now set up.


# Grid+ API Documentation

The following is a list of endpoints that are used to connect to the Grid+ hub. The agent client makes a connection with this API.

All routes originate from `https://app.gridplus.io:3001`

## Authentication

Authenticated endpoints can only be reached by registered (whitelisted) agent devices.

Once your device is registered on the Registry contract, you may request an authentication token from the Grid+ hub:

#### GET /AuthDatum

Query this endpoint to get the data string that needs to be signed by a registered key pair.

Returns:
```
{
  "result" : <String>
}
```

#### POST /Authenticate

Send a decomposed ECDSA signature of the data returned from `/AuthDatum` and receive an authentication token. An example signature decomposition is:

```
var ethutil = require('ethereumjs-util')

var message = ethutil.toBuffer(auth_datum)
let datum_hash = ethutil.sha3(message);
let pkey = new Buffer(myPrivateKeyString, 'hex');
let sig = ethutil.ecsign(datum_hash, pkey);
sig.r = sig.r.toString('hex')
sig.s = sig.s.toString('hex')
```

Request:
```
{
  "owner": <String> # The address that made the signature
  "sig": {
    "v": <Integer> # 27 or 28
    "r": <String>
    "s": <String>
  },
  "personal": <Boolean> # OPTIONAL, true if this signature came from metamask's signing library
}
```

Returns:

```
{
  "reslt": <String> # Authentication token
}
```

### Making an Authenticated Request

Once the user has an authentication token, it can be included either in the body of the request (for `POST` requests) or as a header with title `x-access-token`.

### Token Expiration

If a token becomes invalid, it may have expired. A new one can be created through the same process (`/AuthDatum` + `/Authenticate`) at any time.

## Address Routes

Various routes exist that return the Ethereum address of the contract in question. They are all unauthenticated `GET` requests that return the following data:

```
{
  "result": <String> # Address of the contract in question
}
```

The following routes exist:

* `GET /Registry`  # The registry contract address
* `GET /BOLT`      # The BOLT token contract address
* `GET /Hub`       # The address of the Grid+ token recipient (i.e. the counterparty on all transactions)
* `GET /Channels`  # The address of the payment channel contract

All addresses are currently on the **Ropsten test network**.

## Default Constants

The Grid+ hub may make requests to the Ethereum chain on the user's behalf. There are certain default parameters that can be overwritten. The defaults can be queried from these endpoints.

#### GET /Gas

This returns the default `gasPrice` and `gas` for transactions.

Returns:
```
{
  "gas": <Number> # Amount of gas to spend (decimal format, e.g. 100)
  "gasPrice": <Number> # Price of gas to spend (decimal format, e.g. 100)
}
```

## Faucets

Grid+ offers a small amount of ether to authenticated customers upon request. This endpoint is only available to whitelisted devices or registered Grid+ customers and any faucet drips are recorded by Grid+ and billed after the fact at market rate if applicable.

#### POST /Faucet (Authenticated)

This endpoint returns a default amount of ether (0.1 ETH right now) to the requesting party if the recieving address has below a requisite amount of ether.

Request:
```
{
  "serial_hash": <String> # A keccak-256 hash of the serial number assigned to
                          # the whitelisted agent
}
```

Returns:
```
{
  "result": <String> # The transaction receipt hash.
}
```

## Getting BOLT Tokens

BOLT tokens may be purchased with cryptocurrency or with a credit card.

**NOTE: Because this is still an early-alpha release, only one endpoint is available that functions as a BOLT faucet.**

#### POST /BuyBOLTCC (Authenticated)

**NOTE: This may take up to a minute to process and may return an error regardless of success**

Request:
```
{
  "token": <String> # Authentication token (may also be provided as x-access-token header)
  "recipient": <String> # OPTIONAL, receiving address if different from authenticated address
}
```

## Billing and Usage

Once a channel is opened by the client (done automatically on the official Grid+ agent client), a number of endpoints can be used to query for bills or send signatures that pay outstanding bills.

#### POST /Bills

Get a list of unpaid bills based on your agents consumption.

Request:
```
{
  "serial_hash": <String> # A keccak-256 hash of your agent's serial number
  "all": <Boolean> # OPTIONAL, if true, include paid bills
}
```

Returns:
```
{
  "result": [
    {
      bill_id: <Number> # Id for reference
      amount: <Number> # Amount of USD required to pay this bill (USD === BOLT)
    }
  ]
}
```

#### POST /ChannelSum

An agent may request the latest total that has been committed to the hub for a particular payment channel. For instance, if two bills worth $10 each were paid previously, the channel sum would be $20.

Request:
```
{
  token: <String> # Auth token, may also be sent as the x-access-token header
  channel_id: <String> # bytes32 string of the channel id
}
```

Response:
```
{
  result: <Number> # The amount (in units of 10^-8 BOLT) currently committed to
                   # this payment channel
}
```

#### POST /PayBills

Pay a set of bills (based on an array of `bill_id` values) with a signed message that can be checked against an open state channel. This requires a state channel to be open with the Grid+ hub.

The signed message is a keccak-256 hash of the following concatenated data:
1. Channel Id (`bytes32` value)
2. Value (32-byte padded hex integer)

Note that value is a hex integer and it should be the sum of the bills being paid for and the previous channel sum (i.e. these bills are added to the existing "tab").

A decomposed ECDSA signature (producing `v`, `r`, `s` values) is required. For an example, please see the `/Authenticate` section.

Request:
```
{
  "bill_ids": <Array>
  "msg": <String> # Hash of the message (see above)
  "v": <Integer> # 27 or 28
  "r": <String>
  "s": <String>
  "value": <Integer> # Amount of BOLT to be committed
}
```

NOTE: The BOLT `value` listed above is denominated in atomic units. BOLT tokens (in their current design) have 8 decimals, which means 1 USD = 1 BOLT = 100,000,000 atomic BOLT units.
