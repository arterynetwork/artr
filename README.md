# Artery Network 2.0 Testnet

You can learn about the project at [https://artr.network/](https://artr.network/).

The 2.0 Testnet and bug-hunting program are described at [https://artery-testnet.io/](https://artery-testnet.io/)

## Full Node Quick Start
First of all, you can download the Artery Network application for Windows and macOS from 
[our site](https://artery-testnet.io/). It is probably what you want. If you are using another OS or just interested 
what's under the hood, you can start your full node by yourself following these steps.     

0. Make you sure you have [Docker](https://www.docker.com/) installed.
0. Download this repo and checkout 2.0.0-b.1 version with ```
git clone https://github.com/arterynetwork/artr.git -b 2.0.0-b.1```
0. Build the application with `make build-all`
0. Copy a built binary from the `builds` directory to somewhere your OS could find it (i.e. to some directory in a 
`$PATH`)
0. (Optional) Look around with `artrd --help`
0. Initialize your node with `artrd init [moniker]` where `[moniker]` is a name your like.
0. Replace a just created genesis file (`$HOME/.artrd/config/genesis.json` by default) with one downloaded from 
[https://artery-testnet.io/.well-known/genesis/artery_testnet-2/genesis.json](https://artery-testnet.io/.well-known/genesis/artery_testnet-2/genesis.json)
0. In the node configuration file (`$HOME/.artrd/config/config.toml` by default) set peers and consensus parameters:

            peristent_peers = ""
            seeds = "4326ac75d047a3f37efd5f858af0907c27879322@104.248.173.200:26656"
            
            timeout_propose = "3s"
            timeout_propose_delta = "500ms"
            timeout_prevote = "1s"
            timeout_prevote_delta = "500ms"
            timeout_precommit = "1s"
            timeout_precommit_delta = "500ms"
            timeout_commit = "8s"

0. Start your node with `artrd start`
0. When the node reaches the current blockchain height (you can refer with the 
[blockchain explorer](https://artery-testnet.io/blockchain) to be sure), it is ready. Now you can use the `artrd` 
command to send transactions.

## How to Become a Validator

To become a validator, one must have at least Leader status and 10k+ ARTR total team delegation. If you are not using 
[Artery Node](https://artery-testnet.io/) application, you can activate validation via CLI:
```
artrd keys add -i --recover <your_key_name>
artrd tx noding on <your_key_name> $(artrd tendermint show-validator)
``` 
