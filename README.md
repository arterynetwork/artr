# Artery Network

You can learn about the project at [https://artery-network.io/](https://artery-network.io/).

## Full Node Quick Start
First of all, you can download the Artery Network application for Windows and macOS from 
[our site](https://artery-network.io). It is probably what you want. If you are using another OS or just interested 
what's under the hood, you can start your full node by yourself following these steps.     

0. Make you sure you have [Docker](https://www.docker.com/) installed.
0. Download this repo and checkout 2.0.1 version with ```
git clone https://github.com/arterynetwork/artr.git -b 2.0.1```
0. Build the application with `make build-all`
0. Copy a built binary from the `builds` directory to somewhere your OS could find it (i.e. some directory in a `$PATH`)
0. (Optional) Look around with `artrd --help`
0. Initialize your node with `artrd init [moniker]` where `[moniker]` is a name you like.
0. Replace a just created genesis file (`$HOME/.artrd/config/genesis.json` by default) with one downloaded from https://artery.network/.well-known/genesis/artery_network-9/genesis.json 
0. In the node configuration file (`$HOME/.artrd/config/config.toml` by default) set peers and consensus parameters:

            peristent_peers = ""
            seeds = "47deee9e7c5c68e077ced2ad2e41cf47d9675c0e@64.227.124.171:26656"
            
            timeout_propose = "3s"
            timeout_propose_delta = "500ms"
            timeout_prevote = "1s"
            timeout_prevote_delta = "500ms"
            timeout_precommit = "1s"
            timeout_precommit_delta = "500ms"
            timeout_commit = "14s"

0. (Optional) Download the latest blockchain data snapshot and place it to the `$HOME/.artrd/data` directory. An actual path can be found in https://blocks.artery.network/latest.json
0. Start your node with `artrd start`
0. The node will download and replay blocks, it may take a while. When it reaches a moment of software upgrade, a 
message like `UPGRADE "x.x.x" NEEDED ...` appears and the daemon stops.
0. Checkout a specified version and repeat steps 2, 3 and 9 -- 11.
0. When the node reaches the current blockchain height (you can refer with the 
[blockchain explorer](https://artery-network.io/blockchain) to be sure), it is ready. Now you can use the `artrd` 
command to send transactions.

### Upgrade Manager

Artery honors [Upgradeable Binary 
Specification](https://github.com/regen-network/cosmosd#upgradeable-binary-specification), so you can use [Cosmos 
Upgrade Manager](https://github.com/regen-network/cosmosd#cosmos-upgrade-manager) for easy blockchain upgrade, as 
ourselves do. 

We publish daemon binaries for Windows, macOS and Ubuntu, but we cannot guarantee that they are compatible to your 
system. If your node cannot start after upgrade, you probably should build a binary by yourself and replace a 
downloaded one with it. 

## How to Become a Validator

To become a validator, one must have at least Champion status (can be further changed by voting) and 10k+ ARTR total team delegation. If you are not using 
[Artery Network](https://artery-network.io) application, you can activate validation via CLI:
```
artrd keys add -i --recover <your-key_name>
artrd tx noding on <your_key_name> $(artrd tendermint show-validator)
``` 
