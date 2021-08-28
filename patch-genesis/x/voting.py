from copy import deepcopy
from math import ceil
from typing import Dict, Optional

from ..config import Config
from ..util import patch_status, height_to_time


_TYPE_MAPPING = {
    0:  "PROPOSAL_TYPE_UNSPECIFIED",
    1:  "PROPOSAL_TYPE_ENTER_PRICE",
    2:  "PROPOSAL_TYPE_DELEGATION_AWARD",
    3:  "PROPOSAL_TYPE_DELEGATION_NETWORK_AWARD",
    4:  "PROPOSAL_TYPE_PRODUCT_NETWORK_AWARD",
    5:  "PROPOSAL_TYPE_GOVERNMENT_ADD",
    6:  "PROPOSAL_TYPE_GOVERNMENT_REMOVE",
    7:  "PROPOSAL_TYPE_PRODUCT_VPN_BASE_PRICE",
    9:  "PROPOSAL_TYPE_PRODUCT_STORAGE_BASE_PRICE",
    10: "PROPOSAL_TYPE_FREE_CREATOR_ADD",
    11: "PROPOSAL_TYPE_FREE_CREATOR_REMOVE",
    12: "PROPOSAL_TYPE_SOFTWARE_UPGRADE",
    13: "PROPOSAL_TYPE_CANCEL_SOFTWARE_UPGRADE",
    14: "PROPOSAL_TYPE_STAFF_VALIDATOR_ADD",
    15: "PROPOSAL_TYPE_STAFF_VALIDATOR_REMOVE",
    16: "PROPOSAL_TYPE_EARNING_SIGNER_ADD",
    17: "PROPOSAL_TYPE_EARNING_SIGNER_REMOVE",
    18: "PROPOSAL_TYPE_TOKEN_RATE_SIGNER_ADD",
    19: "PROPOSAL_TYPE_TOKEN_RATE_SIGNER_REMOVE",
    20: "PROPOSAL_TYPE_VPN_SIGNER_ADD",
    21: "PROPOSAL_TYPE_VPN_SIGNER_REMOVE",
    22: "PROPOSAL_TYPE_TRANSITION_PRICE",
    23: "PROPOSAL_TYPE_MIN_SEND",
    24: "PROPOSAL_TYPE_MIN_DELEGATE",
    25: "PROPOSAL_TYPE_MAX_VALIDATORS",
    26: "PROPOSAL_TYPE_GENERAL_AMNESTY",
    27: "PROPOSAL_TYPE_LUCKY_VALIDATORS",
    28: "PROPOSAL_TYPE_VALIDATOR_MINIMAL_STATUS",
}
_PARAMS_MAPPING = {
    "voting/EmptyProposalParams":           (None, None),
    "voting/PriceProposalParams":           ("price", lambda x: x),
    "voting/DelegationAwardProposalParams": ("delegation_award", lambda x: {"award": x}),
    "voting/NetworkAwardProposalParams":    ("network_award", lambda x: x),
    "voting/AddressProposalParams":         ("address", lambda x: x),
    "voting/SoftwareUpgradeProposalParams": ("software_upgrade", lambda x: {
        "name":   x["name"],
        "height": x["height"],
        "info":   x["binaries"]
    }),
    "voting/MinAmountProposalParams":       ("min_amount", lambda x: x),
    "voting/ShortCountProposalParams":      ("count", lambda x: x),
    "voting/StatusProposalParams":          ("status", lambda x: {"status": patch_status(x["status"])}),
}


def _patch_proposal(proposal: Dict) -> None:
    proposal["type"] = _TYPE_MAPPING[proposal.pop("type_code")]
    params: Dict = proposal.pop("params", None)
    if params:
        key, f = _PARAMS_MAPPING[params["type"]]
        if key:
            proposal[key] = f(params["value"])


def patch(state: Optional[Dict], config: Config) -> Dict:
    state = deepcopy(state) if state else {}

    # blocks to hours
    state["params"]["voting_period"] = ceil(state["params"]["voting_period"] / 120)

    for x in state.get("history", []):
        x["finished"] = x.pop("ended")
        _patch_proposal(x["proposal"])

    current_proposal = state.get("current_proposal")
    if current_proposal:
        _patch_proposal(current_proposal)
        current_proposal["end_time"] = height_to_time(int(current_proposal.pop("end_block")), config)

    return state
