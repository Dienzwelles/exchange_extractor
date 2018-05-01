package arbitrage

import(
	"../datastorage"
	"../models"
	//"log"
	//"time"
	//"math"
	//"go/constant"
	"math"
)

type volumesEl struct {
	tradeRatio float64
	trade string
}
type profEl struct {
	profitability 		float64
	firstBuyPriceLimit  float64
	firstBuyAmount  float64
	secondBuyPriceLimit float64
	secondBuyAmount float64
	sellPriceLimit    	float64
	sellAmount    	float64
	exchange      		string
	bookstart     		string
	bookend       		string
	bookdirect    		string
}

func ExtractArbitrage(exchangeId string) ([] models.Arbitrage){

	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	defer db.Close()

	rowsProfittabilities, err := db.Table("extractor.best_books bookstart").
		Joins("JOIN extractor.best_books bookend on SUBSTRING(bookstart.symbol, 1, 3) = SUBSTRING(bookend.symbol, 1, 3) AND bookstart.exchange_id = bookend.exchange_id").
		Joins("JOIN extractor.best_books bookdirect ON SUBSTRING(bookstart.symbol, 4, 3) = SUBSTRING(bookdirect.symbol, 1, 3) AND bookstart.exchange_id = bookdirect.exchange_id").
		Joins("JOIN extractor.best_books inversebookstart ON bookstart.symbol = inversebookstart.symbol AND bookstart.exchange_id = inversebookstart.exchange_id AND bookstart.bid <> inversebookstart.bid").
		Joins("JOIN extractor.best_books inversebookend ON bookend.symbol = inversebookend.symbol AND bookend.exchange_id = inversebookend.exchange_id AND bookend.bid <> inversebookend.bid").
		Joins("JOIN extractor.best_books inversebookdirect ON bookdirect.symbol = inversebookdirect.symbol AND bookdirect.exchange_id = inversebookdirect.exchange_id AND bookdirect.bid <> inversebookdirect.bid").
		Select("IF(bookstart.bid > 0, (bookend.price / bookdirect.price - bookstart.price), (bookdirect.price/bookend.price - (1/bookstart.price))) - ?/1000 profitability," +
		" IF(bookstart.bid > 0, inversebookdirect.price, inversebookend.price) first_buy_price_limit, inversebookend.bid * inversebookend.amount first_buy_amount * (1 - ?/1000), " +
		" bookstart.price second_buy_price_limit, bookstart.bid * bookstart.amount second_buy_amount * (1 - ?/1000), " +
		" IF(bookstart.bid > 0, inversebookend.price, inversebookdirect.price) sell_price_limit, inversebookdirect.bid * inversebookdirect.amount sell_amount * (1 - ?/1000), " +
		" bookstart.exchange_id, bookstart.symbol start_symbol, IF(bookstart.bid > 0, bookend.symbol, bookdirect.symbol) end_symbol, IF(bookstart.bid > 0, bookdirect.symbol, bookend.symbol) direct_symbol", 5, 2, 2, 1).
		Where("bookstart.exchange_id = ? AND SUBSTRING(bookend.symbol, 4, 3) = SUBSTRING(bookdirect.symbol, 4, 3)" +
		" AND ((bookstart.bid, bookend.bid, bookdirect.bid) = (1,-1, 1) OR (bookstart.bid, bookend.bid, bookdirect.bid) = (-1, 1, -1))", exchangeId).
		Order("profitability desc").
		Having("profitability >= ?", 0.00000000000000000001).Rows()

	if(err != nil){
		print(err)
	}

	profittabilities :=	[]profEl{}
	bookStartSymbols :=	[]string{}
	tradeEndSymbols	:=	[]string{}
	bookDirectSymbols := []string{}
	var profRecord profEl

	defer rowsProfittabilities.Close()
	for rowsProfittabilities.Next(){
		err:=rowsProfittabilities.Scan(&profRecord.profitability, &profRecord.firstBuyPriceLimit, &profRecord.firstBuyAmount, &profRecord.secondBuyPriceLimit, &profRecord.secondBuyAmount,
			&profRecord.sellPriceLimit, &profRecord.sellAmount, &profRecord.exchange, &profRecord.bookstart, &profRecord.bookend, &profRecord.bookdirect)
		if err != nil {
			//log.Fatal(err)
			return nil
		}
		bookStartSymbols = append(bookStartSymbols, profRecord.bookstart)
		tradeEndSymbols = append(tradeEndSymbols, profRecord.bookend)
		bookDirectSymbols = append(bookDirectSymbols, profRecord.bookdirect)
		profittabilities = append(profittabilities, profRecord)
	}

	//girare i cross nel caso di rezione -1!!!!!!!! se non ho il dato metto 0! left join
	rowsVolumes, _ := db.Table("extractor.trades trade").
		Select("IFNULL(SUM(trade.amount), 0)/IFNULL(SUM(ABS(trade.amount)), 1) trade_ratio, trade.symbol symbol").
		Where("trade.exchange_id = ? AND trade.symbol IN (?) AND trade.trade_ts > date_sub(current_timestamp, INTERVAL ? SECOND)", "Bitfinex", append(append(bookStartSymbols, tradeEndSymbols...), bookDirectSymbols...), 30).
		Group("trade.symbol").Rows()

	defer rowsVolumes.Close()

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

	/*QTA da SCALARE in BASE A FATTORE VOLUMES E CORRETTIVO FISSO */
	/*
	type bookEl struct {
		price float64
		amount float64
		exchange string
		symbol string
	}

	var bookEnd, bookStart, bookDirect bookEl
	*/
	var ret [] models.Arbitrage
	for i := 0; i < len(profittabilities); i++ {
		profittable := profittabilities[i]
		volume := getVolume(volumes, profittable.bookdirect)
		/*tradeEndSymbol := tradeEndSymbols[i]
		tradeStartSymbol := bookStartSymbols[i]
		tradeDirectSymbol := bookDirectSymbols[i]*/
		/*
		rowsEndAmount, _ := db.Table("extractor.aggregate_books books").
			Joins("JOIN extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
			Select("SUM(books.price * books.amount)/SUM(books.amount) booksend_price, sum(books.amount) booksend_amount, books.exchange_id, books.symbol").
			Where("books.exchange_id = ? AND books.symbol = ? AND books.amount >= ? AND books.price >= ? ", "Bitfinex", tradeEndSymbol, 0, profittable.sellPriceLimit).
			Rows()

		defer rowsEndAmount.Close()

		if rowsEndAmount.Next() {
			err := rowsEndAmount.Scan(&bookEnd.price, &bookEnd.amount, &bookEnd.exchange, &bookEnd.symbol)

			if err != nil {
				//log.Fatal(err)
				return nil
			}
		}

		rowsStartAmount, _ := db.Table("extractor.aggregate_books books").
			Joins("JOIN extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
			Select("SUM(books.price * books.amount)/SUM(books.amount) booksstart_price, sum(books.amount) booksstart_amount, books.exchange_id, books.symbol").
			Where("books.exchange_id = ? AND books.symbol = ? AND books.amount <= ? AND books.price <= ?", "Bitfinex", tradeStartSymbol, 0, profittable.secondBuyPriceLimit).
			Rows()

		defer rowsStartAmount.Close()
		if rowsStartAmount.Next() {
			err := rowsStartAmount.Scan(&bookStart.price, &bookStart.amount, &bookStart.exchange, &bookStart.symbol)

			if err != nil {
				//log.Fatal(err)
				return nil
			}

			rowsDirectAmount, _ := db.Table("extractor.aggregate_books books").
				Joins("JOIN extractor.actual_lots lots ON books.exchange_id = lots.exchange_id and books.symbol = lots.symbol and books.lot = lots.actual_lot").
				Select("SUM(books.price * books.amount)/SUM(books.amount) booksstart_price, sum(books.amount) booksstart_amount, books.exchange_id, books.symbol").
				Where("books.exchange_id = ? AND books.symbol = ? AND books.amount <= ? AND books.price <= ?", "Bitfinex", tradeDirectSymbol, 0, profittable.firstBuyPriceLimit).
				Rows()

			defer rowsDirectAmount.Close()

			if rowsDirectAmount.Next() {
				err := rowsStartAmount.Scan(&bookDirect.price, &bookDirect.amount, &bookDirect.exchange, &bookDirect.symbol)

				if err != nil {
					//log.Fatal(err)
					return nil
				}
			}
			*/

		var volumeRatio float64
		volumeRatio = 0.0

		if volume != nil {
			volumeRatio = volume.tradeRatio
		}

		firstAmount := profittable.firstBuyAmount / (2 + 2 * Sgn(profittable.profitability) * volumeRatio)
		print(firstAmount)
		ret = append(ret, getArbitrage(profittable, firstAmount))
		/*}*/
	}

	return ret
}

func ProvaGetArbitrage() models.Arbitrage{
	profittable := profEl{
		profitability: -0.004953110746587458,
		firstBuyPriceLimit: 8976.9000000000,
		firstBuyAmount: -1.9778950000,
		secondBuyPriceLimit: 0.0732770000,
		secondBuyAmount: 16.3383951400,
		sellPriceLimit:656.9700000000,
		sellAmount: 34.8146422300,
		exchange: "Bitfinex",
		bookdirect: "BTCUSD",
		bookstart: "ETHBTC",
		bookend: "ETHUSD",
	}

	arbitrage := getArbitrage(profittable, 0.005570)

	return arbitrage

}

func getArbitrage(profittable profEl, firstAmount float64) (models.Arbitrage){

	firstBuy := firstAmount * profittable.firstBuyPriceLimit


	direct := profittable.bookdirect[0:3] == profittable.bookstart[3:6]
	secondBuyPrice := ternaryFloat64(direct, profittable.firstBuyPriceLimit, profittable.sellPriceLimit) * profittable.secondBuyPriceLimit
	secondBuy := profittable.secondBuyAmount * secondBuyPrice

	sell := profittable.sellAmount * profittable.sellPriceLimit

	minPrice := math.Min(firstBuy, math.Min(secondBuy, sell))

	return models.Arbitrage{SymbolStart: profittable.bookdirect, SymbolTransitory: profittable.bookstart, SymbolEnd: profittable.bookend,
		AmountStart: minPrice/profittable.firstBuyPriceLimit, AmountTransitory: minPrice/secondBuyPrice, AmountEnd: minPrice/profittable.sellPriceLimit,
		PriceStart: profittable.firstBuyPriceLimit, PriceTransitory: profittable.secondBuyPriceLimit, PriceEnd: profittable.sellPriceLimit}
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

func ternaryFloat64(test bool, a float64, b float64) float64{
	if(test){
		return a
	}

	return b
}

func ternary(test bool, a string, b string) string{
	if(test){
		return a
	}

	return b
}