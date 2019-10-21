package measure

import(
	"../models"
	//"../adapters"
	"time"
	//"golang.org/x/sync/syncmap"
	"../datastorage"
	"../queuemanager"
	"github.com/jinzhu/copier"
	"strings"
)

const BITFINEX  = "Bitfinex"
//var measureExcMap sync.Map
//var lastMeasureExcMap sync.Map

var measureExcMap = map[string]map[string]*Measure{}
var lastMeasureExcMap = map[string]map[string]models.Trade{}

func UpdateData(trades []models.Trade){
	//TODO create storage function

	for _, trade := range trades {
		var measureBitfinexMap = measureExcMap[BITFINEX]
		//var measureBitfinexMapCast = measureBitfinexMap.(sync.Map)

		/*
		measureBitfinexMapCast.Range(func(ki, vi interface{}) bool {
			if (ki == trade.Symbol) {
				measure = vi.(Measure);
				return false;
			}
			return true;
		});
		*/

		 var measure = measureBitfinexMap[trade.Symbol]
		 measure.Measures = measure.getUpdatedMeasure(trade)
		 measure.Tick = measure.getUpdatedTick(trade)

		var tickData = measure.calculateTick()
		queuemanager.TicksEnqueue(&tickData)
		//var measure, _ = measureBitfinexMapCast.Load(trade.Symbol)

		/*
		if measure != nil {
			var measureCast= measure.(Measure)
			measureCast.updateData(trade)
		}
		*/
	}

}

func Calculate(){
	//measureExcMap.Range(func(ki, vi interface{}) bool {
	for _, value := range measureExcMap {
		///*k*/ _, v := ki.(string), vi.(sync.Map)
		//var lastMeasureMap, _ = lastMeasureExcMap.Load(k);
		CalculateExchange(value /*, lastMeasureMap.(*syncmap.Map)*/)
		//return true
	}
	//})
}

func CalculateExchange(measureMap map[string]*Measure){
	//measureMap.Range(func(ki, vi interface{}) bool {
	for key, value := range measureMap {
		//k, v := ki.(string), vi.(*Measure)
		//var lastMeasure interface{} //, _ = lastMeasureExcMap.Load(k)

		/*lastMeasureExcMap.Range(func(ki, vi interface{}) bool {
			if (ki == k) {
				lastMeasure = vi
				return false
			}
			return true
		});*/

		var lastMeasure = lastMeasureExcMap[BITFINEX][key]

		if len(value.Measures) > 0 {
			//var lastMeasureCast= lastMeasure.(models.Trade)

			if lastMeasure.Symbol == key && lastMeasure == value.Measures[0] {
				var zeroTrade, err = GetZeroTrade(lastMeasure)

				if err != nil {
					panic(err)
				}

				value.Measures = value.getUpdatedMeasure(*zeroTrade)
			}

			lastMeasureExcMap[BITFINEX][key] = value.Measures[0]
		}
		var measureData = value.calculateMeasure()
		queuemanager.MeasureEnqueue(&measureData)
		//return true
	}
	//})
}

func GetZeroTrade(lastMeasure models.Trade) (*models.Trade, error){
	var zeroMeasure = new(models.Trade)
	var err = copier.Copy(&zeroMeasure, &lastMeasure)

	zeroMeasure.Tid = "-1"
	zeroMeasure.Price = 0.0
	zeroMeasure.Amount = 0.0

	return zeroMeasure, err
}

func CreateCalcThread(){

	//forever := make(chan bool)

	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			Calculate()
		}

	}()

	//<- forever
}

func InitMap(){
	//TODOcreate storage function
	var measureBitfinexMap = map[string]*Measure{}
	var lastMeasureBitfinexMap = map[string]models.Trade{}

	//var measureBitfinexMap sync.Map
	//var lastMeasureBitfinexMap sync.Map

	lastMeasureExcMap[BITFINEX] = lastMeasureBitfinexMap
	measureExcMap[BITFINEX] = measureBitfinexMap

	var markets = datastorage.GetMarkets(BITFINEX)
	for _, market := range markets {
		var marketUpper = strings.ToUpper(market)
		measureBitfinexMap[marketUpper] = NewDefMesure(BITFINEX, marketUpper)
		lastMeasureBitfinexMap[marketUpper] = models.Trade{}
	}

}

func calculate(){

}
