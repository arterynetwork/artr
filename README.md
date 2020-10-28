# Artery Network

You can learn about the project at [https://artery.network/](https://artery.network/).

## Full Node Quick Start
First of all, you can download the Artery Node application for Windows and macOS from 
[our site](https://artery.network/node). It is probably what you want. If you are using another OS or just interested 
what's under the hood, you can start your full node by yourself following these steps.     

0. Make you sure you have [Go](https://golang.org/) 1.15+ installed.
0. Download this repo and checkout 1.1.0 version with ```
git clone https://github.com/arterynetwork/artr.git -b 1.1.0```
0. Build and install the daemon and the CLI client with `make all`
0. (Optional) Look around with `artrd --help` and `artrcli --help`
0. Initialize your node with `artrd init [moniker]` where `[moniker]` is a name your like.
0. Replace a just created genesis file (`$HOME/.artrd/config/genesis.json` by default) with one downloaded from 
https://artery.network/.well-known/genesis.json 
0. In the node configuration file (`$HOME/.artrd/config/config.toml` by default) set persistent peers and consensus 
parameters: 

            peristent_peers = "3f7d1d07d708546caf6a8a97754c0bbe7e52df52@167.172.60.181:26656,9f92b61e3ccc1bc301f236bbd95e0d83faecdf0a@165.22.118.160:26656,f1bf2da0f0b77db4223b337ccf727f4611d10c52@178.62.83.249:26656,12c42d1a14894bc1e249ee267d0d993d9649e51d@165.22.124.15:26656"
            
            timeout_propose = "3s"
            timeout_propose_delta = "500ms"
            timeout_prevote = "1s"
            timeout_prevote_delta = "500ms"
            timeout_precommit = "1s"
            timeout_precommit_delta = "500ms"
            timeout_commit = "30s"

0. Start your node with `artrd start`
0. The node will download and replay blocks, it may take a while. When it reaches a moment of software upgrade, a 
message like `UPGRADE "x.x.x" NEEDED at height yyyy` appears and the daemon stops.
0. Checkout specified version and repeat steps 3 and 8.
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