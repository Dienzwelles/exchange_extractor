package measure

import (
	"../models"
	"../utils"
	"math"
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

	return m.Measures
}

func (m Measure) calculate() models.MeasuresData{
	println(m.Exchange, " - ", m.Symbol)

	return m.calculateMeasure()
}

func (m Measure) calculateTick() models.MeasureData{
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

	if len(m.Tick) > 2 {
		deltaTick := (m.Tick[0].Price - m.Tick[0].Price) / m.Tick[0].Price
		println("Delta Tick", deltaTick)
		diff := m.Tick[0].Trade_ts.Sub(m.Tick[1].Trade_ts)
		println("Tick Momentum", deltaTick / diff.Seconds())
	}

	print("Tick --> ")
	print(deltaTickSlow / sumCoeffSlow)
	print(", ", deltaTickMedium / sumCoeffMedium)
	println(", ", deltaTickHigh / sumCoeffHigh)

	print("Price --> ")
	print(priceSlow / sumCoeffSlow)
	print(", ", priceMedium / sumCoeffMedium)
	println(", ", priceHigh / sumCoeffHigh)
	return models.MeasureData{/*tickSlow, tickMedium, tickHigh*/}
}
//TODO REFACTOR
func (m Measure) calculateMeasure() models.MeasuresData{
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

	//println("Price ", priceSlow / sumCoeffSlow, ", ", priceMedium / sumCoeffMedium, ", ", priceHigh / sumCoeffHigh)
	println("Trades ", tradeSlow / sumCoeffSlow, ", ", tradeMedium / sumCoeffMedium, ", ", tradeHigh / sumCoeffHigh)
	println("Trades Positive ", positiveTradeSlow / sumCoeffSlow, ", ", positiveTradeMedium / sumCoeffMedium, ", ", positiveTradeHigh / sumCoeffHigh)
	println("Trades Negative ", negativeTradeSlow / sumCoeffSlow, ", ", negativeTradeMedium / sumCoeffMedium, ", ", negativeTradeHigh / sumCoeffHigh)
	println("Abs Amount ", absAmountSlow / sumCoeffSlow, ", ", absAmountMedium / sumCoeffMedium, ", ", absAmountHigh / sumCoeffHigh)
	println("Amount/Abs ", (amountSlow / sumCoeffSlow) / (absAmountSlow / sumCoeffSlow), ", ", (amountMedium / sumCoeffMedium) / (absAmountMedium / sumCoeffMedium), ", ", (amountHigh / sumCoeffHigh) / (absAmountHigh / sumCoeffHigh))
	println("Amount/Trade", (amountSlow / sumCoeffSlow) / (tradeSlow / sumCoeffSlow), ", ", (amountMedium / sumCoeffMedium) / (tradeMedium / sumCoeffMedium), ", ", (amountHigh / sumCoeffHigh) / (tradeHigh / sumCoeffHigh))

	var measureTick = m.calculateTick()

	//var measurePrice = models.MeasureData{priceSlow / sumCoeffSlow, priceMedium / sumCoeffMedium, priceHigh / sumCoeffHigh}
	var measureTrades = models.MeasureData{tradeSlow / sumCoeffSlow, tradeMedium / sumCoeffMedium, tradeHigh / sumCoeffHigh}
	var measureNegativeTrades = models.MeasureData{positiveTradeSlow / sumCoeffSlow, positiveTradeMedium / sumCoeffMedium, positiveTradeHigh / sumCoeffHigh}
	var measurePositiveTrades = models.MeasureData{negativeTradeSlow / sumCoeffSlow, negativeTradeMedium / sumCoeffMedium, negativeTradeHigh / sumCoeffHigh}

	var measureAbsAmount = models.MeasureData{absAmountSlow / sumCoeffSlow, absAmountMedium / sumCoeffMedium, absAmountHigh / sumCoeffHigh}
	var measureAmountOnAbs = models.MeasureData{(amountSlow / sumCoeffSlow) / (absAmountSlow / sumCoeffSlow),(amountMedium / sumCoeffMedium) / (absAmountMedium / sumCoeffMedium), (amountHigh / sumCoeffHigh) / (absAmountHigh / sumCoeffHigh)}
	var measureAmountOnTrade = models.MeasureData{(amountSlow / sumCoeffSlow) / (tradeSlow / sumCoeffSlow), (amountMedium / sumCoeffMedium) / (tradeMedium / sumCoeffMedium), (amountHigh / sumCoeffHigh) / (tradeHigh / sumCoeffHigh)}

	return models.MeasuresData{m.Exchange, m.Symbol,measureTick, models.MeasureData{}, measureTrades, measureNegativeTrades, measurePositiveTrades, measureAbsAmount, measureAmountOnAbs, measureAmountOnTrade}
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}

