# Grid+ Client
This repo contains the Grid+ client, which is a binary meant to live on
the smart agent device (any small system-on-a-chip). Once the agent is registered and claimed
by a human owner, the client will receive data periodically from our hub.
Based on this data, the client may choose to reply to the hub with signed messages that move tokens
owned by the device.

## Setup

Once you have the client installed (instructions below), please go to app.gridplus.io and click the button on the menu titled "Get Serial Number":

<image of button on web app>

Once you have that, navigate to your install directory and type:

```
bash install.sh <serial number>
```

Where `<serial number>` is your pasted serial number exactly as it is shown on the web app.
*Note that in the future, your device will come pre-loaded with a serial number. Every serial number must be whitelisted by Grid+ before it can be used.*

Once the install script is done, run:

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
