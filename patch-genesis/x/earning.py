from typing import Dict, Optional


def patch(state: Optional[Dict]) -> Dict:
    if not state:
        return {}

    state = state.copy()
    if "earners" in state:
        state["earners"] = [{
            "account": val["account"],
            **val["Points"]
        } for val in state["earners"]]

    return state
