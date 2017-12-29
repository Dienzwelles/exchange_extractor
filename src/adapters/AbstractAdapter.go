package adapters

import "../models"

import "fmt"
import "math"

type geometry interface {
	area() float64
	perim() []models.Trade
}
type rect struct {
	width, height float64
}
type circle struct {
	radius float64
}

func (r rect) area() float64 {
	return r.width * r.height
}
func (r rect) perim() float64 {
	return 2*r.width + 2*r.height
}

func (c circle) area() float64 {
	return math.Pi * c.radius * c.radius
}
func (c circle) perim() float64 {
	return 2 * math.Pi * c.radius
}

func measure(g geometry) {
	fmt.Println(g)
	fmt.Println(g.area())
	fmt.Println(g.perim())
	g.perim()
}

func measure2(a AbstractAdapterInterface) {
	a.getData()
}


type AbstractAdapterInterface interface {
	instantiate(symbol string, fetchSize int, reloadInterval int) AbstractAdapter
	instantiateDefault(symbol string) AbstractAdapter
	getData() []models.Trade
}

type AbstractAdapter struct{
	ExchangeId string
	Symbol string
	FetchSize int
	ReloadInterval int
}


func (aa *AbstractAdapter) abstractInstantiateDefault(symbol string) *AbstractAdapter {
	return aa.abstractInstantiate(symbol, -1, 0)
}

func (aa *AbstractAdapter) abstractInstantiate(Symbol string, FetchSize int, ReloadInterval int) *AbstractAdapter {
	aa.Symbol = Symbol
	aa.FetchSize = FetchSize
	aa.ReloadInterval = ReloadInterval

	return aa;
}
