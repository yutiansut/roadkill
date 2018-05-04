package main

import (
	"gitlab.com/CuteQ/roadkill/exchanges/fast/poloniex"
	"gitlab.com/CuteQ/roadkill/exchanges/slow/bitmex"
	"gitlab.com/CuteQ/roadkill/orderbook"
	"gitlab.com/CuteQ/roadkill/orderbook/tectonic"
)

func main() {
	var (
		receiver        = make(chan orderbook.DeltaBatch, 1<<16)
		tConn           = tectonic.DefaultTectonic
		exchangeSymbols = make(map[string][]string, 64)
		//{"XBTUSD", "ETHM18", "XBT7D_U110", "BTC_ETH", "BTC_XMR", "BTC_ETC", "USDT_BTC", "USDT_ETH"}
	)
	exchangeSymbols["bitmex"] = []string{"XBTUSD", "ETHM18", "XBT7D_U110"}
	exchangeSymbols["poloniex"] = []string{"BTC_ETH", "BTC_XMR", "BTC_ETC", "USDT_BTC", "USDT_ETH"}

	tErr := tConn.Connect()

	if tErr != nil {
		panic(tErr)
	}

	for exchange, symbols := range exchangeSymbols {
		for _, symbol := range symbols {
			dbName := exchange + ":" + symbol
			if !tConn.Exists(dbName) {
				tConn.Create(dbName)
			}
		}
	}

	polo := poloniex.DefaultSettings
	polo.Initialize("BTC_ETH", "BTC_XMR", "BTC_ETC", "USDT_BTC", "USDT_ETH")

	bitm := bitmexslow.DefaultSettings
	bitm.ChannelType = []string{"orderBookL2", "trade"}
	bitm.Initialize("XBTUSD", "ETHM18", "XBT7D_U110")

	go bitm.ReceiveMessageLoop(&receiver)
	go polo.ReceiveMessageLoop(&receiver)

	for {
		var (
			tickBatch = <-receiver
			dbName    = tickBatch.Exchange + ":" + tickBatch.Symbol
		)
		insErr := tConn.BulkAddInto(dbName, tickBatch.Deltas)
		// Catch any insertion errors here
		// TODO: Implement some logging here
		if insErr != nil {
			panic(insErr)
		}
	}
}
