# Grid+ Client
This repo contains the Grid+ client, which is a binary meant to live on
the smart agent device (any small system-on-a-chip). Once the agent is registered and claimed
by a human owner, the client will receive data periodically from our hub. 
Based on this data, the client may choose to reply to the hub with signed messages that move tokens
owned by the device.

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

Here are some common issues and solutions.

#### panic: Agent's setup key was not registered by Grid+

You need to register your agent (with its setup key) via our demo web portal.
*NOTE: This is not yet possible. Coming soon.*
