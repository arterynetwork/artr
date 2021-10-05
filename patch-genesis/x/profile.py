from copy import deepcopy
from typing import Dict, Optional

from ..config import Config
from ..util import height_to_time


def patch(
        state:        Optional[Dict],
        storage:      Optional[Dict],
        vpn:          Optional[Dict],
        subscription: Optional[Dict],
        config:       Config
) -> Dict:
    state = deepcopy(state) if state else {"params": {}}

    if "fee" in state["params"]:
        state["params"]["rename_price"] = state["params"].pop("fee")

    storage_limits, storage_current = {}, {}
    if storage:
        for x in storage.get("limits") or []:
            storage_limits[x["account"]] = x["volume"]
        for x in storage.get("current") or []:
            storage_current[x["account"]] = x["volume"]

    vpn_limits, vpn_current = {}, {}
    if vpn:
        for x in vpn.get("vpn_statuses") or []:
            k, v = x["address"], x["VpnInfo"]
            vpn_limits[k] = v["limit"]
            vpn_current[k] = v["current"]
        vpn_params = vpn["params"]
        if vpn_params:
            state["params"].update({
                "storage_signers": vpn_params["signers"],
                "vpn_signers":     vpn_params["signers"],
            })

    active_until = {}
    if subscription:
        for x in subscription.get("activity", []):
            active_until[x["address"]] = height_to_time(int(x["info"]["expire_at"]), config)
        subscription_params = subscription.get("params", None)
        if subscription_params:
            state["params"].update({
                "token_rate":         str(subscription_params["token_course"]),
                "subscription_price": subscription_params["subscription_price"],
                "base_storage_gb":    subscription_params["base_storage_gb"],
                "base_vpn_gb":        subscription_params["base_vpn_gb"],
                "storage_gb_price":   subscription_params["storage_gb_price"],
                "vpn_gb_price":       subscription_params["vpn_gb_price"],
                "token_rate_signers": subscription_params["course_change_signers"],
            })

    for x in state.get("profiles", []):
        xa, xp = x["address"], x["profile"]
        new_p: Dict = {}

        val = xp.get("nickname", None)
        if val:
            new_p["nickname"] = val

        val = xp.get("autopay", None)
        if val:
            new_p["auto_pay"] = val

        val = xp.get("storage", None)
        if val:
            new_p["storage"] = val

        val = xp.get("VPN", None)
        if val:
            new_p["vpn"] = val

        val = storage_limits.get(xa)
        if val:
            new_p["storage_limit"] = val

        val = storage_current.get(xa)
        if val:
            new_p["storage_current"] = val

        val = vpn_limits.get(xa)
        if val:
            new_p["vpn_limit"] = val

        val = vpn_current.get(xa)
        if val:
            new_p["vpn_current"] = val

        val = active_until.get(xa)
        if val:
            new_p["active_until"] = val

        x["profile"] = new_p

    return state
