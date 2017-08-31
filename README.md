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
/SetupKey/:user
```

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

First, get the address of the Registry contract by calling the Grid+ endpoint `/Registry`.

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
