package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/fatih/color"
	"os"
	"strconv"
	"strings"
)

var (
	apiKey    = "z8oJ86HRXRKHppUeLZMOY8564f3gnNueSrmOL1455SXtkTmyHwusLc1XCjjGBKZt"
	secretKey = "UZggnxZ7moBpHw74iGK9SkXHlnci6RAsajO7x1wptsGvgr2qs5lRNu6y5WvJZvDJ"
)

func main() {
	// binance.UseTestnet = true
	colorReset := "\033[0m"

	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	colorCyan := "\033[36m"

	usedBalancePercent := 20.0 // 10%
	priceColor := colorRed
	// colorYellow := "\033[33m"
	stopPrice := 0.0
	sellPrice := 0.0
	//initialPrice := 0.0
	initialBuyPrice := 0.0
	highPrice := 0.0
	sellQuantity := 0.0
	buyQuantity := 0.0
	lastPrice := 0.0
	minimumSellPrice := 0.0
	// orderId := ""
	// coinExist := false
	coinName := "btt"
	pairCoinName := "usdt"

	if len(os.Args) == 2 {
		coinName = os.Args[1]
	}

	if len(os.Args) == 3 {
		coinName = os.Args[1]
		pairCoinName = os.Args[2]
	}

	selectedCoin := strings.ToUpper(coinName)
	selectedPair := strings.ToUpper(pairCoinName)
	selectedSymbol := selectedCoin + "" + selectedPair

	// API key version 2.0
	client := binance.NewClient(apiKey, secretKey)
	account, err := client.NewGetAccountService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	exchangeInfo, err := client.NewExchangeInfoService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
	}
	// fmt.Println(exchangeInfo.Symbols)

	symbolFilters := getSymbolFilters(selectedSymbol, exchangeInfo.Symbols)
	var tickSize = symbolFilters.Filters[0]["tickSize"].(string)
	var stepSize = symbolFilters.Filters[2]["stepSize"].(string)
	//fmt.Println(symbolFilters.Filters)

	fmt.Println(stepSize)
	fmt.Println(tickSize)

	step := parsePriceToFloat(stepSize)
	s := fmt.Sprintf("%f", step)
	p := parsePriceToFloat(s[0:])
	s2 := fmt.Sprintf("%f", p)

	fmt.Println(s2)
	fmt.Println(p)

	stepSizeInt := 0
	if p == 1 {
		stepSizeInt = 0
	} else {
		stepSizeInt = len(s2[0:strings.Index(s2, "1") - 1])
	}
	fmt.Println(stepSizeInt)
	fmt.Println("----------")



	tickSizeInt := len(tickSize[0:strings.Index(tickSize, "1") - 1])
	fmt.Println(tickSizeInt)
	fmt.Println("----------")

	// Get Pair Balance
	pairBalance := getCoinBalance(selectedPair, account.Balances)
	selectedSymbolTicker := getTickersBySymbol(client, selectedSymbol)

	buyQuantity = parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice) / usedBalancePercent
	buyQuantity = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(buyQuantity, stepSizeInt))-parsePriceToFloat(parsePriceToString((buyQuantity*1.0)/100, stepSizeInt)), stepSizeInt))
	sellQuantity = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(buyQuantity, stepSizeInt))-parsePriceToFloat(parsePriceToString((buyQuantity*0.5)/100, stepSizeInt)), stepSizeInt))

	fmt.Println("buyQuantity")
	fmt.Println(buyQuantity)

	fmt.Println("sellQuantity")
	fmt.Println(sellQuantity)

	initialBuyPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(selectedSymbolTicker.AskPrice), tickSizeInt))
	fmt.Println("initialBuyPrice")
	fmt.Println(initialBuyPrice)

	initialBuyPrice = parsePriceToFloat(selectedSymbolTicker.AskPrice)
	minimumSellPrice = initialBuyPrice + (initialBuyPrice * 1.2 / 100)
	minimumSellPrice = parsePriceToFloat(parsePriceToString(minimumSellPrice, tickSizeInt))

	highPrice = minimumSellPrice

	fmt.Println("initialBuyPrice")
	fmt.Println(initialBuyPrice)

	fmt.Println("minimumSellPrice")
	fmt.Println(minimumSellPrice)

	// set stop loss
	stopPrice =  parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((initialBuyPrice*0.5)/100, tickSizeInt)), tickSizeInt))
	sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((initialBuyPrice*0.8)/100, tickSizeInt)), tickSizeInt))

	fmt.Println("sellPrice")
	fmt.Println(sellPrice)
	fmt.Println("stopPrice")
	fmt.Println(stopPrice)

	// INITIAL BUY
	initialBuyOrder, err := client.NewCreateOrderService().Symbol(selectedSymbol).
		Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
		Quantity(parsePriceToString(buyQuantity, stepSizeInt)).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println(initialBuyOrder)

	fmt.Println("selectedSymbolTicker.AskPrice")
	fmt.Println(selectedSymbolTicker.AskPrice)

	initialBuyPrice = parsePriceToFloat(selectedSymbolTicker.AskPrice)
	minimumSellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt))+parsePriceToFloat(parsePriceToString((initialBuyPrice*1.2)/100, tickSizeInt)), tickSizeInt))
	highPrice = minimumSellPrice

	fmt.Println("initialBuyPrice")
	fmt.Println(initialBuyPrice)

	fmt.Println("minimumSellPrice")
	fmt.Println(minimumSellPrice)

	// set stop loss
	stopPrice =  parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((initialBuyPrice*0.5)/100, tickSizeInt)), tickSizeInt))
	sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((initialBuyPrice*0.8)/100, tickSizeInt)), tickSizeInt))

	fmt.Println("sellPrice")
	fmt.Println(sellPrice)
	fmt.Println("stopPrice")
	fmt.Println(stopPrice)

	// CREATE STOP
	initialStopOrder, err := client.NewCreateOrderService().Symbol(selectedSymbol).
		Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
		TimeInForce(binance.TimeInForceTypeGTC).Quantity(parsePriceToString(sellQuantity, stepSizeInt)).
		Price(parsePriceToString(sellPrice, tickSizeInt)).
		StopPrice(parsePriceToString(stopPrice, tickSizeInt)).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	var order *binance.CreateOrderResponse

	wsMarketStatHandler := func(event *binance.WsMarketStatEvent) {
		currentPrice := parsePriceToFloat(event.BidPrice)

		if currentPrice > minimumSellPrice &&
			currentPrice > highPrice {
			highPrice = currentPrice
			color.Yellow("Nuevo precio más alto")

			// set stop loss
			stopPrice =  parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(highPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((highPrice*0.5)/100, tickSizeInt)), tickSizeInt))
			sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(highPrice, tickSizeInt)) - parsePriceToFloat(parsePriceToString((highPrice*0.8)/100, tickSizeInt)), tickSizeInt))

			fmt.Println("sellPrice")
			fmt.Println(sellPrice)
			fmt.Println("stopPrice")
			fmt.Println(stopPrice)

			if initialStopOrder != nil {
				cancelOrder(client, selectedSymbol, initialStopOrder.OrderID)
			}

			if order != nil {
				cancelOrder(client, selectedSymbol, order.OrderID)
			}

			order, err = client.NewCreateOrderService().Symbol(selectedSymbol).
				Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
				TimeInForce(binance.TimeInForceTypeGTC).Quantity(parsePriceToString(sellQuantity, stepSizeInt)).
				Price(parsePriceToString(sellPrice, tickSizeInt)).
				StopPrice(parsePriceToString(stopPrice, tickSizeInt)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(order)
		}

		if lastPrice > highPrice {
			priceColor = colorGreen
		} else if lastPrice < currentPrice {
			priceColor = colorRed
		} else {
			priceColor = colorCyan
		}

		fmt.Printf("\033[2K\r"+priceColor+"%s "+colorGreen+"%s "+colorRed+"%s "+colorReset, parsePriceToString(currentPrice, tickSizeInt), parsePriceToFloat(event.BidPrice), parsePriceToFloat(event.AskPrice))
		lastPrice = currentPrice

		fmt.Println("order")
		fmt.Println(order)

		if order != nil {
			o := getOrder(client, selectedSymbol, order.OrderID)

			if o != nil {
				if o.Status == binance.OrderStatusTypeFilled {
					color.Yellow("PROFIT SELL")
					os.Exit(1)
					return
				}
			} else {
				color.Yellow("PROFIT SELL")
				os.Exit(1)
				return
			}
		}
	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := binance.WsMarketStatServe(selectedSymbol, wsMarketStatHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}

func getTickersBySymbol(client *binance.Client, symbol string) *binance.BookTicker {
	prices, err := client.NewListBookTickersService().Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil
	}

	for _, p := range prices {
		if p.Symbol == symbol {
			return p
		}
	}
	return nil
}

func getSymbolFilters(symbol string, symbols []binance.Symbol) binance.Symbol {
	symbolInfo := linq.From(symbols).Where(func(b interface{}) bool {
		return b.(binance.Symbol).Symbol == symbol
	}).Select(func(b interface{}) interface{} {
		return b.(binance.Symbol)
	}).First()
	return symbolInfo.(binance.Symbol)
}

func getCoinBalance(coinName string, balances []binance.Balance) binance.Balance {
	coinBalance := linq.From(balances).Where(func(b interface{}) bool {
		return b.(binance.Balance).Asset == coinName
	}).Select(func(b interface{}) interface{} {
		return b.(binance.Balance)
	}).First()
	return coinBalance.(binance.Balance)
}

func parsePriceToFloat(price string) float64 {
	f1, _ := strconv.ParseFloat(price, 8)
	price = strconv.FormatFloat(f1, 'f', -1, 64) // 10.9
	f2, _ := strconv.ParseFloat(price, 8)
	//f2 = math.Round(f2*1000)/1000
	return f2
}

func parsePriceToString(price float64, size int) string {
	s := strconv.FormatFloat(price, 'f', size, 64) // 10.9
	//s := fmt.Sprintf("%.5f", price)
	return s
}

/*

	/*
		if currentPrice <= stopPrice {
			color.Green("TAKE PROFIT")
			color.Red("currentPrice")
			color.Red(parsePriceToString(currentPrice))

			color.Red("stopLossPrice")
			color.Red(parsePriceToString(stopLossPrice))

			if order != nil {
				cancelOrder(client, selectedSymbol, order.OrderID)
			}

			order, err = client.NewCreateOrderService().Symbol(selectedSymbol).
				Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
				Quantity(parsePriceToString(sellQuantity)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
				return
			}
			os.Exit(1)
			return

		}*/
/*
	if currentPrice <= stopLossPrice {
		color.Red("STOP LOSS")
		color.Red("currentPrice")
		color.Red(parsePriceToString(currentPrice))

		color.Red("stopLossPrice")
		color.Red(parsePriceToString(stopLossPrice))

		if order != nil {
			cancelOrder(client, selectedSymbol, order.OrderID)
		}

		order, err = client.NewCreateOrderService().Symbol(selectedSymbol).
			Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
			Quantity(parsePriceToString(sellQuantity)).
			Do(context.Background())
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
			return
		}
		os.Exit(1)
		return

	}
*/

/*
func parsePriceToFloat(str string) float64 {
	val, err := strconv.ParseFloat(str, 64)
	if err == nil {
		return val
	}

	//Some number may be seperated by comma, for example, 23,120,123, so remove the comma firstly
	str = strings.Replace(str, ",", "", -1)

	//Some number is specifed in scientific notation
	pos := strings.IndexAny(str, "eE")
	if pos < 0 {
		e, _ := strconv.ParseFloat(str, 64)
		return e
	}

	var baseVal float64
	var expVal int64

	baseStr := str[0:pos]
	baseVal, err = strconv.ParseFloat(baseStr, 64)
	if err != nil {
		return 0
	}

	expStr := str[(pos + 1):]
	expVal, err = strconv.ParseInt(expStr, 10, 64)
	if err != nil {
		return 0
	}

	return baseVal * math.Pow10(int(expVal))

	//f1, _ := strconv.ParseFloat(price, 64)
	//price = strconv.FormatFloat(f1, 'f', -1, 64) // 10.9
	//f2, _ := strconv.ParseFloat(price, 64)
}

*/

func getOrder(client *binance.Client, symbol string, orderID int64) *binance.Order {
	order, err := client.NewGetOrderService().Symbol(symbol).
		OrderID(orderID).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return nil
	}
	fmt.Println(order)
	return order
}

func cancelOrder(client *binance.Client, symbol string, orderID int64) bool {
	_, err := client.NewCancelOrderService().Symbol(symbol).
		OrderID(orderID).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return false
	}

	return true
}
