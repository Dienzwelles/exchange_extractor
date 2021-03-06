package arbitrage

import(
	"../datastorage"
	"../models"
	//"log"
	//"time"
	//"math"
	//"go/constant"
	"math"
	"../utils"
)

type volumesEl struct {
	tradeRatio float64
	trade string
}
type profEl struct {
	exchange      		string
	profitability 		float64
	firstMarketPrice  float64
	firstAmount  float64
	secondMarketPrice float64
	secondAmount float64
	thirdMarketPrice    	float64
	thirdAmount    	float64

	firstSymbol     	string
	secondSymbol   		string
	thirdSymbol    		string
}

func ExtractArbitrage(exchangeId string) ([] models.Arbitrage){

	conn := datastorage.NewConnection()
	db := datastorage.GetConnectionORM(conn)

	defer db.Close()

	selectData := "bookstart.exchange_id, " +
		"IF(bookstart.bid > 0, (inversebookend.price / inversebookdirect.price - inversebookstart.price) - ?/1000, (inversebookstart.price*bookdirect.price/bookend.price - 1) - ?/1000) profitability, " +
			"IF(bookstart.bid > 0, inversebookdirect.price, bookend.price) first_market_price, " +
			"IF(bookstart.bid > 0, -inversebookdirect.amount, -bookend.amount) first_amount, " +
			"inversebookstart.price second_market_price, " +
			"-inversebookstart.amount second_amount, " +
			"IF(bookstart.bid > 0, inversebookend.price, bookdirect.price) third_market_price, " +
			"IF(bookstart.bid > 0, -inversebookend.amount, -bookdirect.amount) third_amount, " +
			"IF(bookstart.bid > 0, inversebookdirect.symbol, bookend.symbol) first_symbol, " +
			"inversebookstart.symbol second_symbol, " +
			"IF(bookstart.bid > 0, inversebookend.symbol, bookdirect.symbol) third_symbol"

	rowsProfittabilities, err := db.Table("extractor.best_books bookstart").
		Joins("JOIN extractor.best_books bookend on SUBSTRING(bookstart.symbol, 1, 3) = SUBSTRING(bookend.symbol, 1, 3) AND bookstart.exchange_id = bookend.exchange_id").
		Joins("JOIN extractor.best_books bookdirect ON SUBSTRING(bookstart.symbol, 4, 3) = SUBSTRING(bookdirect.symbol, 1, 3) AND bookstart.exchange_id = bookdirect.exchange_id").
		Joins("JOIN extractor.best_books inversebookstart ON bookstart.symbol = inversebookstart.symbol AND bookstart.exchange_id = inversebookstart.exchange_id AND bookstart.bid <> inversebookstart.bid").
		Joins("JOIN extractor.best_books inversebookend ON bookend.symbol = inversebookend.symbol AND bookend.exchange_id = inversebookend.exchange_id AND bookend.bid <> inversebookend.bid").
		Joins("JOIN extractor.best_books inversebookdirect ON bookdirect.symbol = inversebookdirect.symbol AND bookdirect.exchange_id = inversebookdirect.exchange_id AND bookdirect.bid <> inversebookdirect.bid").
		Select(selectData, 5, 4).
		Where("bookstart.exchange_id = ? AND SUBSTRING(inversebookdirect.symbol, 4, 3) = ? AND SUBSTRING(bookend.symbol, 4, 3) = ? AND SUBSTRING(bookend.symbol, 4, 3) = SUBSTRING(bookdirect.symbol, 4, 3)" +
		" AND ((bookstart.bid, bookend.bid, bookdirect.bid) = (1,-1, 1) OR (bookstart.bid, bookend.bid, bookdirect.bid) = (-1, -1, 1))", exchangeId, "USD", "USD").
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
		err:=rowsProfittabilities.Scan(&profRecord.exchange, &profRecord.profitability, &profRecord.firstMarketPrice, &profRecord.firstAmount, &profRecord.secondMarketPrice, &profRecord.secondAmount,
			&profRecord.thirdMarketPrice, &profRecord.thirdAmount, &profRecord.firstSymbol, &profRecord.secondSymbol, &profRecord.thirdSymbol)
		if err != nil {
			//log.Fatal(err)
			return nil
		}

		print(profRecord.firstAmount)

		bookStartSymbols = append(bookStartSymbols, profRecord.secondSymbol)
		tradeEndSymbols = append(tradeEndSymbols, profRecord.thirdSymbol)
		bookDirectSymbols = append(bookDirectSymbols, profRecord.firstSymbol)
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


		ret = append(ret, getArbitrage(profittable, volumes))
		/*}*/
	}

	return ret
}

func getArbitrage(profittable profEl, volumes []volumesEl) (models.Arbitrage){

	firstTrade := profittable.firstAmount * profittable.firstMarketPrice

	direct := profittable.firstSymbol[0:3] == profittable.secondSymbol[3:6]
	secondMarketPrice := utils.TernaryFloat64(direct, profittable.firstMarketPrice, profittable.thirdMarketPrice) * profittable.secondMarketPrice
	secondTrade := profittable.secondAmount * secondMarketPrice

	thirdTrade := profittable.thirdAmount * profittable.thirdMarketPrice

	absFirstTrade := math.Abs(firstTrade)
	absSecondTrade := math.Abs(secondTrade)
	absThirdTrade := math.Abs(thirdTrade)

	minPrice := math.Min(absFirstTrade, math.Min(absSecondTrade, absThirdTrade))
	cross := utils.Ternary(firstTrade < secondTrade && firstTrade< absThirdTrade, profittable.firstSymbol,
		utils.Ternary(secondTrade < thirdTrade, profittable.secondSymbol, profittable.thirdSymbol))
	minPriceAmount := utils.TernaryFloat64(firstTrade < secondTrade && firstTrade< absThirdTrade, profittable.firstAmount,
		utils.TernaryFloat64(secondTrade < thirdTrade, profittable.secondAmount, profittable.thirdAmount))


	ratio := calculateRatio(minPriceAmount, cross, volumes)

	return models.Arbitrage{SymbolStart: profittable.firstSymbol, SymbolTransitory: profittable.secondSymbol, SymbolEnd: profittable.thirdSymbol,
		AmountStart: minPrice/profittable.firstMarketPrice * ratio * utils.Sgn(profittable.firstAmount), AmountTransitory: minPrice/secondMarketPrice * ratio * utils.Sgn(profittable.secondAmount),
		AmountEnd: minPrice/profittable.thirdMarketPrice * ratio * utils.Sgn(profittable.thirdAmount),
		PriceStart: profittable.firstMarketPrice, PriceTransitory: profittable.secondMarketPrice, PriceEnd: profittable.thirdMarketPrice}
}

func calculateRatio(amount float64, cross string, volumes []volumesEl) float64{
	volume := getVolume(volumes, cross)

	var volumeRatio float64
	volumeRatio = 0.0

	if volume != nil {
		volumeRatio = volume.tradeRatio
	}

	return 1 / (2 * (1 + math.Max(utils.Sgn(amount) * volumeRatio, 0)))

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