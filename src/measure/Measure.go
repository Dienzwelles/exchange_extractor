package measure

import (
	"../models"
	"../utils"
	"encoding/json"
	"math"
	"log"
	"time"
)

type Cluster struct {
	Levels[] float64
	Counts[] int
}

type Measure  struct {
	Exchange string
	Symbol string
	Measures[] models.Trade
	ValueCluster *Cluster
	Tick[] models.Trade
	SlowSpeedDeep int
	MediumSpeedDeep int
	HighSpeedDeep int
}

const DEF_SLOW_SPEED_DEEP = 1000

const DEF_MEDIUM_SPEED_DEEP = 100

const DEF_HIGH_SPEED_DEEP = 10

func NewDefCluster() *Cluster {
	levels := []float64{-100000, -10000, -1500, -500, 0, 500, 1500, 10000, 100000}
	counts := []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	return &Cluster{levels, counts}
}

func NewDefMesure(exchange string, symbol string) *Measure {
	return NewMesure(exchange, symbol, DEF_SLOW_SPEED_DEEP, DEF_MEDIUM_SPEED_DEEP, DEF_HIGH_SPEED_DEEP)
}

func NewMesure(exchange string, symbol string, slowSpeedDeep int, mediumSpeedDeep int, highSpeedDeep int) *Measure {
	return &Measure{exchange, symbol, []models.Trade{}, NewDefCluster(),[]models.Trade{},slowSpeedDeep, mediumSpeedDeep, highSpeedDeep}
}

func (c Cluster) updateData(trade models.Trade) {
	value := trade.Amount * trade.Price

	for i := 0; i < len(c.Counts); i++  {
		if (i < len(c.Counts) -1 && value < c.Levels[i]) || (i == len(c.Counts) -1 && value > c.Levels[i -1]) {
			c.Levels[i] += 1
			break
		}
	}
}

func (c Cluster) clear() {
	c.Counts = []int{0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
}

func (m Measure) getUpdatedMeasure(trade models.Trade) []models.Trade {
	if trade.Exchange_id == m.Exchange && trade.Symbol == m.Symbol {
		if m.Measures == nil || len(m.Measures) < m.SlowSpeedDeep {
			m.Measures = append([]models.Trade{trade}, m.Measures...)
		} else {
			m.Measures = append([]models.Trade{trade}, m.Measures[1:len(m.Measures)-2]...)
		}
	}

	return m.Measures
}

func (m Measure) getUpdatedTick(trade models.Trade) []models.Trade {
	if trade.Exchange_id == m.Exchange && trade.Symbol == m.Symbol {
		if m.Tick == nil || len(m.Tick) < m.SlowSpeedDeep {
			m.Tick = append([]models.Trade{trade}, m.Tick...)
		} else {
			m.Tick = append([]models.Trade{trade}, m.Tick[1:len(m.Tick)-2]...)
		}
	}

	return m.Tick
}

func (m Measure) calculateTick() models.Ticks{
	deltaTickSlow := 0.0
	deltaTickMedium := 0.0
	deltaTickHigh := 0.0

	priceSlow := 0.0
	priceMedium := 0.0
	priceHigh := 0.0

	sumCoeffSlow := 0.0
	sumCoeffMedium := 0.0
	sumCoeffHigh := 0.0

	if (m.Tick != nil && len(m.Tick) > 1) {
		m.ValueCluster.clear()
		for i := 1; i < len(m.Tick); i++  {
			coeff := float64(1.0 / i)//1.0 / (float64(i) + 1.0)

			delta := (m.Tick[i -1].Price - m.Tick[i].Price) / m.Tick[i].Price

			coeffDiff := coeff * delta

			deltaTickSlow += coeffDiff
			priceCoeff := m.Tick[i -1].Price * coeff

			priceSlow += priceCoeff
			sumCoeffSlow += coeff

			m.ValueCluster.updateData(m.Tick[i])

			if i <= m.MediumSpeedDeep {
				deltaTickMedium += coeffDiff
				priceMedium += priceCoeff
				sumCoeffMedium += coeff
			}

			if i <= m.HighSpeedDeep {
				deltaTickHigh += coeffDiff
				priceHigh += priceCoeff
				sumCoeffHigh += coeff
			}
		}
	}

	var momentum = 0.0
	if len(m.Tick) > 2 {
		deltaTick := (m.Tick[0].Price - m.Tick[1].Price) / m.Tick[0].Price
		println("Delta Tick", deltaTick)
		diff := float64(m.Tick[0].Trade_ts.Unix() - m.Tick[1].Trade_ts.Unix()) / 1000.0

		if diff == 0 {
			if deltaTick == 0 {
				momentum = 0
			} else {
				momentum = utils.TernaryFloat64(deltaTick >= 0, 999999999999999999, -999999999999999999)
			}
		} else {
			momentum = deltaTick / diff
		}

		println("Tick Momentum", momentum)
	}

	deltaTickSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, deltaTickSlow / sumCoeffSlow)
	deltaTickMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, deltaTickMedium / sumCoeffMedium)
	deltaTickHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, deltaTickHigh / sumCoeffHigh)

	priceSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, priceSlow / sumCoeffSlow)
	priceMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, priceMedium / sumCoeffMedium)
	priceHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, priceHigh / sumCoeffHigh)

	print("Tick --> ")
	print(deltaTickSlow)
	print(", ", deltaTickMedium)
	println(", ", deltaTickHigh)

	print("Price --> ")
	print(priceSlow)
	print(", ", priceMedium)
	println(", ", priceHigh)

	//TODO Agganciare i tick direttamente a quando arrivano i dati!!! non su ciclo thread e seprarare tabelle a questo punto
	if len(m.Tick) > 2 {
		print("ciao")
	}
	//tickData := models.TickData{momentum,models.MeasureData{deltaTickSlow, deltaTickMedium, deltaTickHigh}}
	return models.Ticks{0, m.Exchange, m.Symbol, time.Now(),momentum, deltaTickSlow, deltaTickMedium, deltaTickHigh}
}
//TODO REFACTOR
func (m Measure) calculateMeasure() models.Measures{
	println(m.Exchange, " - ", m.Symbol)
	/*priceSlow := 0.0
	priceMedium := 0.0
	priceHigh := 0.0*/

	tradeSlow := 0.0
	tradeMedium := 0.0
	tradeHigh := 0.0

	positiveTradeSlow := 0.0
	positiveTradeMedium := 0.0
	positiveTradeHigh := 0.0

	negativeTradeSlow := 0.0
	negativeTradeMedium := 0.0
	negativeTradeHigh := 0.0

	absAmountSlow := 0.0
	absAmountMedium := 0.0
	absAmountHigh := 0.0

	amountSlow := 0.0
	amountMedium := 0.0
	amountHigh := 0.0

	sumCoeffSlow := 0.0
	sumCoeffMedium := 0.0
	sumCoeffHigh := 0.0

	for i, trade := range m.Measures {
		coeff := 1.0 //1.0 / (float64(i) + 1.0)

		//newPrice := coeff * trade.Price

		newAmount := coeff * trade.Amount
		newAbsAmount := coeff * math.Abs(trade.Amount)

		//priceSlow += newPrice
		tradeV := coeff * utils.TernaryFloat64(trade.Amount != 0.0, 1.0, 0.0)

		tradeP := 0.0
		tradeN := 0.0
		if trade.Amount > 0.0 {
			tradeP = coeff * 1.0
		} else if trade.Amount < 0.0 {
			tradeN = coeff * 1.0
		}

		amountSlow += newAmount
		absAmountSlow += newAbsAmount
		sumCoeffSlow += coeff

		tradeSlow += tradeV
		positiveTradeSlow += tradeP
		negativeTradeSlow += tradeN

		if i <= m.MediumSpeedDeep {
			//priceMedium += newPrice
			tradeMedium += tradeV
			positiveTradeMedium += tradeP
			negativeTradeMedium += tradeN
			amountMedium += newAmount
			absAmountMedium += newAbsAmount
			sumCoeffMedium += coeff
		}

		if i <= m.HighSpeedDeep {
			//priceHigh += newPrice
			amountHigh += newAmount
			absAmountHigh += newAbsAmount
			tradeHigh += tradeV
			positiveTradeHigh += tradeP
			negativeTradeHigh += tradeN
			sumCoeffHigh += coeff
		}
	}

	tradeSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, tradeSlow / sumCoeffSlow)
	tradeMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, tradeMedium / sumCoeffMedium)
	tradeHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, tradeHigh / sumCoeffHigh)

	positiveTradeSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, positiveTradeSlow / sumCoeffSlow)
	positiveTradeMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, positiveTradeMedium / sumCoeffMedium)
	positiveTradeHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, positiveTradeHigh / sumCoeffHigh)

	negativeTradeSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, negativeTradeSlow / sumCoeffSlow)
	negativeTradeMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, negativeTradeMedium / sumCoeffMedium)
	negativeTradeHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, negativeTradeHigh / sumCoeffHigh)

	absAmountSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0, 0.0, absAmountSlow / sumCoeffSlow)
	absAmountMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0, 0.0, absAmountMedium / sumCoeffMedium)
	absAmountHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0, 0.0, absAmountHigh / sumCoeffHigh)

	var amountOnAbsSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0 || absAmountSlow == 0.0, 0.0, (amountSlow / sumCoeffSlow) / (absAmountSlow / sumCoeffSlow))
	var amountOnAbsMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0 || absAmountMedium == 0.0, 0.0, (amountMedium / sumCoeffMedium) / (absAmountMedium / sumCoeffMedium))
	var amountOnAbsHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0 || absAmountHigh == 0.0, 0.0, (amountHigh / sumCoeffHigh) / (absAmountHigh / sumCoeffHigh))

	var amountTradeSlow = utils.TernaryFloat64(sumCoeffSlow == 0.0 || tradeSlow == 0.0, 0.0, (amountSlow / sumCoeffSlow) / (tradeSlow / sumCoeffSlow))
	var amountTradeMedium = utils.TernaryFloat64(sumCoeffMedium == 0.0 || tradeMedium == 0.0, 0.0, (amountMedium / sumCoeffMedium) / (tradeMedium / sumCoeffMedium))
	var amountTradeHigh = utils.TernaryFloat64(sumCoeffHigh == 0.0 || tradeHigh == 0.0, 0.0, (amountHigh / sumCoeffHigh) / (tradeHigh / sumCoeffHigh))

	//println("Price ", priceSlow / sumCoeffSlow, ", ", priceMedium / sumCoeffMedium, ", ", priceHigh / sumCoeffHigh)
	println("Trades ", tradeSlow, ", ", tradeMedium, ", ", tradeHigh)
	println("Trades Positive ", positiveTradeSlow, ", ", positiveTradeMedium, ", ", positiveTradeHigh)
	println("Trades Negative ", negativeTradeSlow, ", ", negativeTradeMedium, ", ", negativeTradeHigh)
	println("Abs Amount ", absAmountSlow, ", ", absAmountMedium, ", ", absAmountHigh)
	println("Amount/Abs ", amountOnAbsSlow, ", ", amountOnAbsMedium, ", ", amountOnAbsHigh)
	println("Amount/Trade", amountTradeSlow, ", ", amountTradeMedium, ", ", amountTradeHigh)

	//var measurePrice = models.MeasureData{priceSlow / sumCoeffSlow, priceMedium / sumCoeffMedium, priceHigh / sumCoeffHigh}
	var measureTrades = models.MeasureData{tradeSlow, tradeMedium, tradeHigh}
	var measureNegativeTrades = models.MeasureData{positiveTradeSlow, positiveTradeMedium, positiveTradeHigh}
	var measurePositiveTrades = models.MeasureData{negativeTradeSlow, negativeTradeMedium, negativeTradeHigh}

	var measureAbsAmount = models.MeasureData{absAmountSlow, absAmountMedium, absAmountHigh}
	var measureAmountOnAbs = models.MeasureData{amountOnAbsSlow,amountOnAbsMedium, amountOnAbsHigh}
	var measureAmountOnTrade = models.MeasureData{amountTradeSlow, amountTradeMedium, amountTradeHigh}

	ret := models.Measures{0,m.Exchange, m.Symbol, time.Now(),
		measureTrades.Slow, measureTrades.Medium, measureTrades.High,
		measureNegativeTrades.Slow, measureNegativeTrades.Medium, measureNegativeTrades.High,
		measurePositiveTrades.Slow, measurePositiveTrades.Medium, measurePositiveTrades.High,
		measureAbsAmount.Slow, measureAbsAmount.Medium, measureAbsAmount.High,
		measureAmountOnAbs.Slow, measureAmountOnAbs.Medium, measureAmountOnAbs.High,
		measureAmountOnTrade.Slow, measureAmountOnTrade.Medium, measureAmountOnTrade.High}
	//return models.MeasuresData{m.Exchange, m.Symbol, measureTick.Momentum, measureTick.Ticks, measureTrades, measureNegativeTrades, measurePositiveTrades, measurePositiveTrades, measureAmountOnAbs, measureAmountOnTrade}

	_, jsonErr := json.Marshal(ret)
	if jsonErr != nil {
		log.Fatal(jsonErr)
	}
	return ret
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

