#!/usr/bin/python3

from getopt import getopt
import sys

from .config import Config
from .main import main


if __name__ == "__main__":
    args = dict()
    named, _ = getopt(sys.argv[1:],
                      "hi:o:c:t:b:s:",
                      ["help", "input=", "output=", "chain-id=", "time=", "block=", "speed="])
    for k, v in named:
        if k in ("-h", "--help"):
            print('''
Gets genesis exported from v1.x.x Artery blockchain and makes patches, so it might be used by v2.x.x

Usage:
    patch-genesis -h|--help
    patch-genesis [-i|--input <filename>] [-o|--output <filename>] [-c|--chain-id <chain ID>] [-t|--time <genesis time>]
        [-b|--block <initial height>] [-s|--speed <time quotient>]

Options:
    -i, --input <filename>
        Path to a original (v1.x.x) genesis file. Default is 'genesis.orig.json'.
    
    -o, --output <filename>
        Path to a resulting (v2.x.x) genesis file. Default is 'genesis.v2.json'.
    
    -c, --chain-id <chain ID>
        A new genesis chain ID. Default is 'artery2.0'.
    
    -t, --time <genesis time>
        Genesis time, i.e. time when a new, post-fork, blockchain will be run. Default is current time. All scheduled 
        operation time will be calculated assuming a genesis block is committed at that time. Value should be passed 
        using ISO format, f.e.:
        
            2021-11-01T03:00:00.000000Z

    -b, --block <initial height>
        Height, a genesis was exported at. Default is app_state.schedule.params.initial_height from the genesis file.

    -s, --speed <time quotient>
        Time speed multiplier. Must be set to 1 in production, but may be greater during testing. The more the value,
        the more often delegating accrue, referral compression and other regular operations accrue. The minimal value is
        1 (real time), the maximal one is 1440 (one day per a minute). Default is 1.
        ''')
            sys.exit(1)

        arg = {
            "-i":         "input_filename",
            "--input":    "input_filename",
            "-o":         "output_filename",
            "--output":   "output_filename",
            "-c":         "chain_id",
            "--chain-id": "chain_id",
            "-x":         "app_hash",
            "--hash":     "app_hash",
            "-t":         "genesis_time",
            "--time":     "genesis_time",
            "-b":         "initial_height",
            "--block":    "initial_height",
            "-s":         "time_quotient",
            "--speed":    "time_quotient"
        }[k]
        if arg in args:
            print("Duplicating key:", arg)
            sys.exit(1)
        args[arg] = v

    config = Config(**args)
    main(config)
