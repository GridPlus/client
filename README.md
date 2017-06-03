# Grid+ Client
This repo contains the Grid+ client, which is a binary meant to live on
the smart agent device (any small system-on-a-chip). Once the agent is registered and claimed
by a human owner, the client will receive data periodically from our hub.
Based on this data, the client may choose to reply to the hub with signed messages that move tokens
owned by the device.

## Setup

Once you have the client installed (instructions below), please go to ~~app.gridplus.io~~(not available yet) and click the button on the menu titled "Get setup key".

![setup key button](images/getAgentButton.png)

This will pop up a window with your setup address, private key, and serial number. **In the production app, these would be pre-loaded onto your device and the serial number would be printed on the box.**

![config.toml information](images/tomlfile.png)

Copy all that information and save it into the following file: `src/config/setup_keys.toml`. An example `setup_keys.toml` file looks like this:

```
[agent]
addr = "0x2a919a8ff288615fb1381ff1a582b826d412dab2"
pkey = "1aec3339a5388d3c165f7d0dd35e5c16acad31eb311f1526b920d410636a6028"
serial_no = "726a686c68f""
```

Now you can run the client:
```
bash run.sh
```

You will see something like this:
<image of console>

Now go back to the web app and register your serial number on Ethereum. This associates your metamask address with the device's address on the blockchain.

<image of first setup tab>

Once that's complete, you can proceed to getting started with Grid+. Navigate to the second tab and enter the serial number again, this time with a nickname for your device (you can call it anything).

<image of second setup tab>

You should now see a chart of your usage data. For the demo, this is randomly generated data, but in the production app you would see how much energy you are consuming and how much it is costing you.

## Prerequisites

You will need a UNIX shell (i.e. OSX or Linux). You will also need to have Go installed. If you are on OSX, you can do this with homebrew (`brew install golang`).

## Installation

The installation process first generates a setup key. This setup key must be registered with the Grid+ registry contract (which is not yet available to the public). To generate your setup key and build your binaries, run:

```
bash install.sh
```

This will install the prerequisites (via `go get`) and then it will generate a
private key for your simulated device and put it in the proper config file.

You can now run the client with:

```
bash run.sh
```


### Troubleshooting
Here are some common issues and solutions. Note that errors are, by default, logged to `src/agent.log`.

#### "Serial number not registered with Grid+"

You need to register your agent (with its setup key) via our demo web portal. *NOTE: This is not yet possible. Coming soon.*

#### install.sh fails to fetch packages
This is likely an issue with fetching go-ethereum. OSX and Ubuntu 16.04 should be fine, but we have run into issues with Ubuntu 14.04. If you have a problem on another OS, please let us know.

#### "Waiting for agent to be claimed..." continues forever
It is waiting for you to claim it. You may do so on our web portal. *NOTE: This is not yet possible. Coming soon.*

#### getsockopt: connection refused
You cannot hit the RPC provider.
