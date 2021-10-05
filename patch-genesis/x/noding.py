from copy import deepcopy
from typing import Dict, Optional

from ..config import Config
from ..util import patch_status


_INFRACTIONS_MAPPING = {
    "duplicate/vote": "DUPLICATE_VOTE"
}


def patch(state: Optional[Dict], config: Config) -> Dict:
    state = deepcopy(state) if state else {}

    if "params" not in state:
        state["params"] = {}
    if "min_status" in state["params"]:
        state["params"]["min_status"] = patch_status(state["params"]["min_status"])
    state["params"]["voting_power"] = {
        "slices": [
            {
                "part":         "15%",
                "voting_power": 15
            },
            {
                "part":         "85%",
                "voting_power": 10
            }
        ],
        "luckies_voting_power": 10
    }

    for k in ["active", "non_active"]:
        for x in state.get(k) or []:
            key = x.pop("pubkey", None)
            if key:
                x["pub_key"] = key
            for inf in x.get("infractions", []):
                inf["type"] = _INFRACTIONS_MAPPING[inf["type"]]

            # Purge lifetime bans. It's been decided so.
            if "banned" in x:
                del x["banned"]

    if config.adam:
        # Switch off all validators but one
        active, non_active = [], []
        for x in state["active"]:
            if x["account"] == config.adam:
                active.append(x)
            else:
                non_active.append(x)
        for x in state["non_active"]:
            if x["account"] == config.adam:
                if "jailed" in x:
                    del x["jailed"]
                if "switched_on" in x:
                    del x["switched_on"]
                active.append(x)
            else:
                non_active.append(x)

        assert len(active) == 1

        state["active"], state["non_active"] = active, non_active

    return state
