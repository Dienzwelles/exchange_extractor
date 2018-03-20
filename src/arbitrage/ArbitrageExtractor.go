package arbitrage

import(
	"../datastorage"
	"log"
	//"time"
	//"math"
	//"go/constant"
)

func ExtractArbitrage() {

	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	rowsProfittabilities, err := db.Table("extractor.last_trades tradestart").
					Joins("JOIN extractor.last_trades tradeend on SUBSTRING(tradestart.symbol, 1, 3) = SUBSTRING(tradeend.symbol, 1, 3) AND tradestart.exchange_id = tradeend.exchange_id").
					Joins("JOIN extractor.last_trades tradedirect ON SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 1, 3) AND tradestart.exchange_id = tradedirect.exchange_id").
					Joins("JOIN extractor.trades datatradestart ON datatradestart.id = tradestart.max_id"). /*datatradestart.symbol = tradestart.symbol AND*/
					Joins("JOIN extractor.trades datatradeend ON datatradeend.id = tradeend.max_id"). /*datatradeend.symbol = tradeend.symbol AND */
					Joins("JOIN extractor.trades datatradedirect ON datatradedirect.id = tradedirect.max_id"). /*datatradedirect.symbol = tradedirect.symbol AND*/
					Select("ABS(datatradeend.price / datatradedirect.price - datatradestart.price) /*- FEE*/ profitability, (datatradedirect.price * datatradestart.price) price_limit, tradestart.exchange_id exchange, tradestart.symbol start_symbol, tradeend.symbol end_symbol, tradedirect.symbol direct_symbol").
					Where("tradestart.exchange_id = ? AND SUBSTRING(tradeend.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 4, 3)", "Bitfinex").
						Order("profitability desc").Having("profitability >= ?", 0.0000000000000001).Rows()

	if(err != nil){
		print(err)
	}
	type profEl struct {
		profitability float64
		priceLimit float64
		exchange string
		tradestart string
		tradeend string
		tradedirect string
	}

	var profittabilities []profEl
	var tradeStartSymbols []string
	var tradeEndSymbols []string
	var tradeDirectSymbols []string
	var profRecord profEl

	defer rowsProfittabilities.Close()
	for rowsProfittabilities.Next(){
		err:=rowsProfittabilities.Scan(&profRecord.profitability, &profRecord.priceLimit, &profRecord.exchange, &profRecord.tradestart, &profRecord.tradeend, &profRecord.tradedirect)
		if err != nil {
			log.Fatal(err)
		}
		tradeStartSymbols = append(tradeStartSymbols, profRecord.tradestart)
		tradeEndSymbols = append(tradeEndSymbols, profRecord.tradeend)
		tradeDirectSymbols = append(tradeDirectSymbols, profRecord.tradedirect)
		profittabilities = append(profittabilities, profRecord)
	}

	rowsVolumes, _ := db.Table("extractor.trades tradestart").
		Joins("JOIN extractor.trades tradeend on SUBSTRING(tradestart.symbol, 1, 3) = SUBSTRING(tradeend.symbol, 1, 3) AND tradestart.exchange_id = tradeend.exchange_id").
		Joins("JOIN extractor.trades tradedirect ON SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 1, 3) AND tradestart.exchange_id = tradedirect.exchange_id").
		Select("SUM(tradestart.amount)/SUM(ABS(tradestart.amount)) tradestart_ratio, SUM(tradeend.amount)/SUM(ABS(tradeend.amount)) tradeend_ratio, SUM(tradedirect.amount)/SUM(ABS(tradedirect.amount)) tradedirect_ratio, tradestart.symbol start_symbol, tradeend.symbol end_symbol, tradedirect.symbol direct_symbol").
		Where("tradestart.exchange_id = ? AND tradestart.symbol IN (?) AND tradeend.symbol IN (?) AND tradedirect.symbol IN (?) AND tradestart.trade_ts > date_sub(current_timestamp, INTERVAL ? MINUTE) AND tradeend.trade_ts > date_sub(current_timestamp, INTERVAL ? MINUTE) AND tradedirect.trade_ts > date_sub(current_timestamp, INTERVAL ? MINUTE)", "Bitfinex", tradeStartSymbols, tradeEndSymbols, tradeDirectSymbols, 50000, 50000, 50000).
		Group("tradestart.symbol, tradeend.symbol, tradedirect.symbol").Rows()

	type volumesEl struct {
		tradestartRatio float64
		tradeendRatio float64
		tradedirectRatio string
		tradestart string
		tradeend string
		tradedirect string
	}
	var volumes []volumesEl
	var volumeRecord volumesEl

	for rowsVolumes.Next(){
		err:=rowsVolumes.Scan(&volumeRecord.tradestartRatio, &volumeRecord.tradeendRatio, &volumeRecord.tradedirectRatio, &volumeRecord.tradestart, &volumeRecord.tradeend, &volumeRecord.tradedirect)
		if err != nil {
			log.Fatal(err)
		}

		volumes = append(volumes, volumeRecord)
	}

/*QTA da SCALARE in BASE A FATTORE VOLUMES E CORRETTIVO FISSO */

	type bookEl struct {
		price float64
		amount float64
		exchange string
		symbol string
	}

	var book bookEl

	for i := 0; i < len(profittabilities); i++{
		volume := volumes[i].tradestartRatio
		profittable := profittabilities[i]
		rowsEndAmount, _ := db.Table("extractor.aggregate_books books").
		Joins("extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
		Select("AVG(books.price * books.amount)/SUM(books.amount) booksend_price, sum(books.amount) booksend_amount, books.exchange_id, books.symbol").
		Where("books.exchange_id = ? AND books.symbol = ? AND books.amount*SIGN(?) >= ? AND books.price*SIGN(?) >= SIGN(?)*? ", "Bitfinex", profittable.tradeend, profittable.profitability, 0, profittable.profitability, profittable.priceLimit).
		Rows()

		rowsEndAmount.Next()
		err:=rowsEndAmount.Scan(&book.price, &book.amount, &book.exchange, &book.symbol)
		if err != nil {
			log.Fatal(err)
		}

		rowsStartAmount, _ := db.Table("extractor.aggregate_books books").
		Joins("extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
		Select("AVG(books.price * books.amount)/SUM(books.amount) booksstart_price, sum(books.amount) booksstart_amount, books.exchange_id, books.symbol").
		Where("books.exchange_id = ? AND books.symbol = ? AND books.amount*SIGN(?) >= ? AND books.price*SIGN(?) >= SIGN(?)*? ", "Bitfinex", profittable.tradestart, profittable.profitability, 0, profittable.profitability, profittable.priceLimit / book.price).
		Rows()

		var amount float64
		if(rowsAmount.Next()){
			rowsAmount.Scan(amount)
			print(amount / (2 + 2 * Sgn(profittable.profitability) * volume))
		}
	}

}

func Sgn(a float64) float64 {
	switch {
	case a < 0:
		return -1
	case a > 0:
		return +1
	}
	return 0
}