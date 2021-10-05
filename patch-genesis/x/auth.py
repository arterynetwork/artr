from typing import Dict, Iterable, List


_PUBKEY_MAPPING = {
    "tendermint/PubKeySecp256k1": ("/cosmos.crypto.secp256k1.PubKey", lambda x: {"key": x}),
}


def _patch_public_key(pubkey: Dict) -> Dict:
    t, f = _PUBKEY_MAPPING[pubkey["type"]]
    return {
        "@type": t,
        **f(pubkey["value"]),
    }


def _patch_base_account(acc: Dict) -> Dict:
    result = {
        "address": acc["address"]
    }
    v = acc.get("public_key", None)
    if v:
        result["pubKey"] = _patch_public_key(v)
    v = acc["account_number"]
    if v:
        result["accountNumber"] = v
    v = acc["sequence"]
    if v:
        result["sequence"] = v
    return result


def _patch_module_account(acc: Dict) -> Dict:
    return {
        "name":         acc["name"],
        "permissions":  acc["permissions"],
        "base_account": _patch_base_account(acc)
    }


_ACCOUNT_MAPPING = {
    "cosmos-sdk/Account":       ("/cosmos.auth.v1beta1.BaseAccount",   _patch_base_account),
    "cosmos-sdk/ModuleAccount": ("/cosmos.auth.v1beta1.ModuleAccount", _patch_module_account),
}


def _patch_account(account: Dict) -> Dict:
    t, f = _ACCOUNT_MAPPING[account["type"]]
    return {
        "@type": t,
        **f(account["value"])
    }


def patch_accounts(state: Iterable[Dict]) -> List[Dict]:
    return list(map(_patch_account, state))


def patch(state: Dict) -> Dict:
    return {
        "params":   state.get("params", []),
        "accounts": patch_accounts(state.get("accounts"))
    }
