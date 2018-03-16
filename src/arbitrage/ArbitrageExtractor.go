package arbitrage

import(
	"../datastorage"
)

func extractArbitrage() {

	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	rowsProfittabilities, _ := db.Table("extractor.last_trades tradestart").
					Joins("extractor.last_trades tradeend on SUBSTRING(tradestart.symbol, 1, 3) = SUBSTRING(tradeend.symbol, 1, 3)").
					Joins("extractor.last_trades tradedirect ON SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 1, 3)").
					Joins("extractor.trades datatradestart ON datatradestart.symbol = tradestart.symbol AND datatradestart.id = tradestart.max_id").
					Joins("extractor.trades datatradeend ON datatradeend.symbol = tradeend.symbol AND datatradeend.id = tradeend.max_id").
					Joins("extractor.trades datatradedirect ON datatradedirect.symbol = tradedirect.symbol AND datatradedirect.id = tradedirect.max_id").
					Select("ABS(datatradeend.price / datatradedirect.price - datatradestart.price) /*- FEE*/ profitability, (datatradedirect.price * datatradestart.price) price_limit").
					Where("exchange_id = ? AND SUBSTRING(tradeend.symbol, 4, 3) = ? AND SUBSTRING(tradedirect.symbol, 4, 3) = ?", "Bitfinex", "USD", "USD").Rows()

	volumes, _ := db.Table("extractor.trades").
		Joins("extractor.last_trades tradeend on SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradeend.symbol, 1, 3)").
		Joins("extractor.last_trades tradedirect ON SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradedirect, 1, 3)").
		Select("SUM(trades.amount)/SUM(ABS(trades.amount))").Where("extractor.trades where exchange_id = ? AND symbol = ? and  trade_ts < date_sub(current_timestamp, INTERVAL ? MINUTE)", "Bitfinex", "BTCUSD", 5).Rows()

/*QTA da SCALARE in BASE A FATTORE VOLUMES E CORRETTIVO FISSO */
	rowsAmount, _ := db.Table("extractor.aggregate_books books").
					Joins("extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
					Select("AVG(books.price * books.amount)/SUM(books.amount), sum(books.amount), books.exchange_id, books.symbol").Where(" books.price >=|<= 0  and books.amount >=|<= 0 ", 0).
						Group("books.exchange_id, books.symbol").Rows()


	print(rowsProfittabilities)
	print(rowsAmount)
	print(volumes)

}