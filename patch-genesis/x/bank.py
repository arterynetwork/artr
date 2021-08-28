from typing import Dict, Iterable, Optional


def patch(state: Optional[Dict], auth_accounts: Iterable[Dict], supply: Optional[Dict]) -> Dict:
    state = state.copy() if state else {}

    if "send_enabled" in state:
        del state["send_enabled"]
    state = {
        "params":   state,
        "balances": list(
            filter(
                lambda x: any(int(coins["amount"]) for coins in x["coins"]),
                ({
                    "address": acc["value"]["address"],
                    "coins":   acc["value"]["coins"],
                } for acc in auth_accounts)
            )
        ),
    }
    if supply:
        val = supply.get("supply", None)
        if val:
            state["supply"] = val

    return state
