from datetime import datetime
from typing import Optional


class Config:
    def __init__(
            self,
            input_filename: str = "genesis.orig.json",
            output_filename: str = "genesis.v2.json",
            chain_id: str = "artery2.0",
            genesis_time: str = None,
            initial_height: str = None,
            time_quotient: str = "1",
            adam: Optional[str] = None,
    ):
        if genesis_time is None:
            genesis_time = datetime.utcnow().isoformat() + "Z"

        self.chain_id = chain_id
        self.genesis_time = genesis_time
        self.input = input_filename
        self.output = output_filename
        self.initial_height: Optional[int] = None  # will be set later
        if initial_height:
            self.initial_height = int(initial_height)
        self.time_quotient = int(time_quotient)
        if not 1 <= self.time_quotient <= 1440:
            raise ValueError('time_quotient is out of range')
        self.adam = adam

    def init(self, genesis: dict):
        if self.initial_height is None:
            self.initial_height = int(genesis["app_state"]["schedule"].get("params", {}).get("initial_height", "0"))

    def get_genesis_time(self) -> datetime:
        return datetime.fromisoformat(self.genesis_time.split(sep="Z")[0])
