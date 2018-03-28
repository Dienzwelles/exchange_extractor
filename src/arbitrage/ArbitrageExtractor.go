package arbitrage

import(
	"../datastorage"
	"../models"
	//"log"
	//"time"
	//"math"
	//"go/constant"
)

type volumesEl struct {
	tradeRatio float64
	trade string
}

func ExtractArbitrage() (*models.Arbitrage){

	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	rowsProfittabilities, err := db.Table("extractor.last_trades tradestart").
		Joins("JOIN extractor.last_trades tradeend on SUBSTRING(tradestart.symbol, 1, 3) = SUBSTRING(tradeend.symbol, 1, 3) AND tradestart.exchange_id = tradeend.exchange_id").
		Joins("JOIN extractor.last_trades tradedirect ON SUBSTRING(tradestart.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 1, 3) AND tradestart.exchange_id = tradedirect.exchange_id").
		Joins("JOIN extractor.trades datatradestart ON datatradestart.id = tradestart.max_id"). /*datatradestart.symbol = tradestart.symbol AND*/
		Joins("JOIN extractor.trades datatradeend ON datatradeend.id = tradeend.max_id"). /*datatradeend.symbol = tradeend.symbol AND */
		Joins("JOIN extractor.trades datatradedirect ON datatradedirect.id = tradedirect.max_id"). /*datatradedirect.symbol = tradedirect.symbol AND*/
		Select("ABS(datatradeend.price / datatradedirect.price - datatradestart.price) /*- FEE*/ profitability, SIGN(datatradeend.price / datatradedirect.price - datatradestart.price) direction, (datatradedirect.price * datatradestart.price) price_limit, tradestart.exchange_id exchange, tradestart.symbol start_symbol, tradeend.symbol end_symbol, tradedirect.symbol direct_symbol").
		Where("tradestart.exchange_id = ? AND SUBSTRING(tradeend.symbol, 4, 3) = SUBSTRING(tradedirect.symbol, 4, 3)", "Bitfinex").
		Order("profitability desc").Having("profitability >= ?", 0.0000000000000001).Rows()

	if(err != nil){
		print(err)
	}
	type profEl struct {
		profitability float64
		direction float64
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
		err:=rowsProfittabilities.Scan(&profRecord.profitability, &profRecord.direction, &profRecord.priceLimit, &profRecord.exchange, &profRecord.tradestart, &profRecord.tradeend, &profRecord.tradedirect)
		if err != nil {
			//log.Fatal(err)
			return nil
		}
		tradeStartSymbols = append(tradeStartSymbols, profRecord.tradestart)
		tradeEndSymbols = append(tradeEndSymbols, ternary(profRecord.direction == 1, profRecord.tradeend, profRecord.tradedirect))
		tradeDirectSymbols = append(tradeDirectSymbols, ternary(profRecord.direction == 1, profRecord.tradedirect, profRecord.tradeend))
		profittabilities = append(profittabilities, profRecord)
	}
	rowsProfittabilities.Close()

	//girare i cross nel caso di rezione -1!!!!!!!! se non ho il dato metto 0! left join
	rowsVolumes, _ := db.Table("extractor.trades trade").
		Select("IFNULL(SUM(trade.amount), 0)/IFNULL(SUM(ABS(trade.amount)), 1) trade_ratio, trade.symbol symbol").
		Where("trade.exchange_id = ? AND trade.symbol IN (?) AND trade.trade_ts > date_sub(current_timestamp, INTERVAL ? SECOND)", "Bitfinex", append(append(tradeStartSymbols, tradeEndSymbols...), tradeDirectSymbols...), 30).
		Group("trade.symbol").Rows()

	var volumes []volumesEl
	var volumeRecord volumesEl

	for rowsVolumes.Next(){
		err:=rowsVolumes.Scan(&volumeRecord.tradeRatio, &volumeRecord.trade)
		if err != nil {
			//log.Fatal(err)
			return nil
		}

		volumes = append(volumes, volumeRecord)
	}
	rowsVolumes.Close()

	/*QTA da SCALARE in BASE A FATTORE VOLUMES E CORRETTIVO FISSO */

	type bookEl struct {
		price float64
		amount float64
		exchange string
		symbol string
	}

	var bookEnd, bookStart bookEl

	for i := 0; i < len(profittabilities); i++ {
		profittable := profittabilities[i]
		volume := getVolume(volumes, profittable.tradestart)
		tradeEndSymbol := tradeEndSymbols[i]
		tradeStartSymbol := tradeStartSymbols[i]
		tradeDirectSymbol := tradeDirectSymbols[i]

		rowsEndAmount, _ := db.Table("extractor.aggregate_books books").
			Joins("JOIN extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
			Select("SUM(books.price * books.amount)/SUM(books.amount) booksend_price, sum(books.amount) booksend_amount, books.exchange_id, books.symbol").
			Where("books.exchange_id = ? AND books.symbol = ? AND books.amount*? >= ? AND books.price*? >= ?*? ", "Bitfinex", tradeEndSymbol, profittable.direction, 0, profittable.direction, profittable.direction, profittable.priceLimit).
			Rows()

		if rowsEndAmount.Next() {
			err := rowsEndAmount.Scan(&bookEnd.price, &bookEnd.amount, &bookEnd.exchange, &bookEnd.symbol)

			if err != nil {
				//log.Fatal(err)
				return nil
			}
		}
		rowsEndAmount.Close()
		rowsStartAmount, _ := db.Table("extractor.aggregate_books books").
			Joins("JOIN extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
			Select("SUM(books.price * books.amount)/SUM(books.amount) booksstart_price, sum(books.amount) booksstart_amount, books.exchange_id, books.symbol").
			Where("books.exchange_id = ? AND books.symbol = ? AND books.amount*? >= ? AND books.price*? >= ?*? ", "Bitfinex", tradeStartSymbol, profittable.direction, 0, profittable.direction, profittable.direction, profittable.priceLimit / bookEnd.price).
			Rows()

		if rowsStartAmount.Next() {
			err := rowsStartAmount.Scan(&bookStart.price, &bookStart.amount, &bookStart.exchange, &bookStart.symbol)

			if err != nil {
				//log.Fatal(err)
				return nil
			}


			var volumeRatio float64
			volumeRatio = 0.0

			if volume != nil {
				volumeRatio = volume.tradeRatio
			}
			amount := bookStart.amount / (2 + 2 * Sgn(profittable.profitability) * volumeRatio)
			print(amount)
			return &models.Arbitrage{SymbolStart: tradeDirectSymbol, SymbolTransitory:tradeStartSymbol, SymbolEnd:tradeEndSymbol, AmountStart: amount}
		}
	}

	return nil
}

func getVolume(volumes []volumesEl, trade string) *volumesEl{
	if(volumes != nil){
		for i := 0; i < len(volumes); i++ {
			if(volumes[i].trade == trade) {
				return &volumes[i]
			}
		}
	}

	return nil
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

func ternary(test bool, a string, b string) string{
	if(test){
		return a
	}

	return b
}