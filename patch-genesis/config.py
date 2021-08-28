from datetime import datetime


class Config:
    def __init__(
            self,
            input_filename: str = "genesis.orig.json",
            output_filename: str = "genesis.v2.json",
            chain_id: str = "artery2.0",
            genesis_time: str = None,
            time_quotient: str = "1"
    ):
        if genesis_time is None:
            genesis_time = datetime.utcnow().isoformat() + "Z"

        self.chain_id = chain_id
        self.genesis_time = genesis_time
        self.input = input_filename
        self.output = output_filename
        self.initial_height: int = None  # will be set later
        self.time_quotient = int(time_quotient)
        if not 1 <= self.time_quotient <= 1440:
            raise ValueError('time_quotient is out of range')


    def init(self, genesis: dict):
        self.initial_height = int(genesis["app_state"]["schedule"].get("params", {}).get("initial_height", "0"))

    def get_genesis_time(self) -> datetime:
        return datetime.fromisoformat(self.genesis_time.split(sep="Z")[0])
