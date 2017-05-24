# Grid+ Client
This repo contains the Grid+ client, which is a binary that is meant to live on
the smart agent device (any small system-on-a-chip capable of running this binary).

## Installation

You will need to have Go installed. Once you do, you will need to generate a
setup key. With which your device will be registered. The registration process
will need to interface with our hub API.

*NOTE: This is not yet possible.In the future, we will provide a signup process to test our demo app.*

To generate the setup key, run:

```
bash install.sh
```

This will install the prerequisites (via `go get`) and then it will generate a
private key for your simulated device and put it in the proper config file.

To install the Grid+ client, run:
```
bash run.sh
```

### Troubleshooting

Here are some common issues and solutions.

#### panic: Agent's setup key was not registered by Grid+

You need to register your agent (with its setup key) via our demo web portal.
*NOTE: This is not yet possible. Coming soon.*
