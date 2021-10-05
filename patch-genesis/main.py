import json
from typing import Dict, Optional

from .config import Config
from .x import bank, auth, delegating, earning, noding, profile, referral, schedule, voting


def patch(genesis: dict, config: Config) -> None:
    config.init(genesis)

    genesis["chain_id"] = config.chain_id
    if "app_hash" in genesis:
        del genesis["app_hash"]
    genesis["genesis_time"] = config.genesis_time
    genesis["initial_height"] = str(config.initial_height)

    consensus_params: Optional[Dict] = genesis.get("consensus_params", None)
    if consensus_params:
        consensus_params["evidence"]["max_bytes"] = "1048576"
        consensus_params["version"] = {
            "app_version": "0"
        }
        consensus_params["block"]["max_gas"] = "5000000000"

    app_state: Dict = genesis["app_state"]
    auth_state = app_state.get("auth")
    auth_accounts = (auth_state or {}).get("accounts", [])
    delegating_state = app_state.get("delegating")

    genesis["app_state"] = {
        "artrbank":   bank.patch(app_state.get("artrbank"), auth_accounts, app_state.get("supply")),
        "auth":       auth.patch(auth_state),
        "delegating": delegating.patch(delegating_state, config),
        "earning":    earning.patch(app_state.get("earning")),
        "noding":     noding.patch(app_state.get("noding"), config),
        "profile":    profile.patch(
                            app_state.get("profile"),
                            app_state.get("storage"),
                            app_state.get("vpn"),
                            app_state.get("subscription"),
                            config
                        ),
        "referral":   referral.patch(app_state.get("referral"), config),
        "schedule":   schedule.patch(app_state.get("schedule"), delegating_state, config),
        "voting":     voting.patch(app_state.get("voting"), config)
    }


def main(config: Config) -> None:
    with open(config.input) as genesis_file:
        genesis = json.load(genesis_file)

    patch(genesis, config)

    with open(config.output, 'w') as genesis_file:
        json.dump(genesis, genesis_file, sort_keys=True, indent=2)
