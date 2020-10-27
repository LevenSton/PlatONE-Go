package main

import (
	"gopkg.in/urfave/cli.v1"
)

var (
	TestnetNodeNumberFlag = cli.UintFlag{
		Name:  "number",
		Usage: "Number of node in testnet configuration (default = 4)",
		Value: 4,
	}

	BinaryDirFlag = cli.StringFlag{
		Name:  "binary.dir",
		Usage: "Binary directory",
		Value: "./",
	}

	DataDirFlag = cli.StringFlag{
		Name:  "datadir",
		Usage: "Testnet Data directory for node data",
		Value: "./localtestnetdata",
	}

	GCModeFlag = cli.StringFlag{
		Name:  "gcmode",
		Usage: `Blockchain garbage collection mode ("full", "archive")`,
		Value: "full",
	}

	RPCPortFlag = cli.IntFlag{
		Name:  "rpcport",
		Usage: "HTTP-RPC server listening port",
		Value: 6780,
	}

	P2PPortFlag = cli.IntFlag{
		Name:  "p2pport",
		Usage: "P2P network listening port",
		Value: 16780,
	}

	WSPortFlag = cli.IntFlag{
		Name:  "wsport",
		Usage: "WS-RPC server listening port",
		Value: 26780,
	}
)
