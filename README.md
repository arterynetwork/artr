# Artery Network

You can learn about the project at [https://artery.network/](https://artery.network/).

## Full Node Quick Start
First of all, you can download the Artery Network application for Windows and macOS from 
[our site](https://artery.network). It is probably what you want. If you are using another OS or just interested 
what's under the hood, you can start your full node by yourself following these steps.     

0. Make you sure you have [Go](https://golang.org/) 1.15+ installed.
0. Download this repo and checkout 1.3.4 version with ```
git clone https://github.com/arterynetwork/artr.git -b 1.3.4```
0. Build and install the daemon and the CLI client with `make all`
0. (Optional) Look around with `artrd --help` and `artrcli --help`
0. Initialize your node with `artrd init [moniker]` where `[moniker]` is a name your like.
0. Replace a just created genesis file (`$HOME/.artrd/config/genesis.json` by default) with one downloaded from 
https://artery.network/.well-known/genesis1.3.4.json
0. In the node configuration file (`$HOME/.artrd/config/config.toml` by default) set peers and consensus parameters:

            peristent_peers = ""
            seeders = "47deee9e7c5c68e077ced2ad2e41cf47d9675c0e@64.227.124.171:26656"
            
            timeout_propose = "3s"
            timeout_propose_delta = "500ms"
            timeout_prevote = "1s"
            timeout_prevote_delta = "500ms"
            timeout_precommit = "1s"
            timeout_precommit_delta = "500ms"
            timeout_commit = "30s"

0. (Optional) Download the latest blockchain data snapshot and place it to the `$HOME/.artrd/data` directory. An actual path can be found in https://blocks.artr-api.com/latest.json
0. Start your node with `artrd start`
0. The node will download and replay blocks, it may take a while. When it reaches a moment of software upgrade, a 
message like `UPGRADE "x.x.x" NEEDED at height yyyy` appears and the daemon stops.
0. Checkout specified version and repeat steps 2 and 8.
0. When the node reaches the current blockchain height (you can refer with the 
[blockchain explorer](https://artery.network/blockchain) to be sure), it is ready. Now you can use the `artrcli` 
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

To become a validator, one must have at least Leader status and 10k+ ARTR total team delegation. If you are not using 
[Artery Node](https://artery.network/node) application, you can activate validation via CLI:
```artrcli tx noding on $(artrd tendermint show-validator) --from [your_key_name]``` 
