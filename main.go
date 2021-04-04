package main

import (
	"context"
	"fmt"
	"github.com/adshao/go-binance/v2"
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/fatih/color"
	"math"
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
	//colorReset := "\033[0m"

	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	colorCyan := "\033[36m"

	priceColor := colorRed
	// colorYellow := "\033[33m"

	//initialPrice := 0.0
	initialBuyPrice := 0.0
	highPrice := 0.0
	stopLossPrice := 0.0
	sellQuantity := 0.0
	buyQuantity := 0.0
	lastPrice := 0.0
	// orderId := ""
	// coinExist := false
	coinName := "ctsi"
	pairCoinName := "usdt"


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

	// Get Pair Balance
	pairBalance := getCoinBalance(selectedPair, account.Balances)
	selectedSymbolTicker := getTickersBySymbol(client, selectedSymbol)
	buyQuantity = math.Trunc(parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice))
	buyQuantity = math.Trunc(buyQuantity - (buyQuantity * 1 / 100))

	sellQuantity = math.Trunc(buyQuantity - (buyQuantity * 0.5 / 100))

	// INITIAL BUY
	order, err := client.NewCreateOrderService().Symbol(selectedSymbol).
		Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
		Quantity(parsePriceToString(buyQuantity)).
		Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}

	initialBuyPrice = parsePriceToFloat(order.Price)
	stopLossPrice = initialBuyPrice - (initialBuyPrice * 0.5 / 100)
	minimumSellPrice := initialBuyPrice + (initialBuyPrice * 1 / 100)
	highPrice = minimumSellPrice

	fmt.Println("order.Status")
	fmt.Println(order.Status)
	order = nil


	wsMarketStatHandler := func(event *binance.WsMarketStatEvent) {
		fmt.Println(event.AskPrice)
		fmt.Println(event.BidPrice)

		currentPrice := parsePriceToFloat(event.BidPrice)
		fmt.Println("currentPrice")
		fmt.Println(currentPrice)

		if currentPrice > minimumSellPrice &&
			currentPrice > highPrice {
			highPrice = currentPrice
			color.Yellow("Nuevo precio más alto")

			stopPrice := highPrice - (highPrice * 0.5 / 100)
			sellPrice := highPrice - (highPrice * 0.7 / 100)

			fmt.Println("sellPrice")
			fmt.Println(sellPrice)
			fmt.Println("stopPrice")
			fmt.Println(stopPrice)

			if order != nil {
				cancelOrder(client, selectedSymbol, order.OrderID)
			}

			order, err = client.NewCreateOrderService().Symbol(selectedSymbol).
				Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
				TimeInForce(binance.TimeInForceTypeGTC).Quantity(parsePriceToString(sellQuantity)).
				Price(parsePriceToString(sellPrice)).
				StopPrice(parsePriceToString(stopPrice)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
		}

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
				Quantity(parsePriceToString(buyQuantity)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
				return
			}
			os.Exit(1)
			return

		}

		if lastPrice > highPrice {
			priceColor = colorGreen
		} else if lastPrice < currentPrice {
			priceColor = colorRed
		} else {
			priceColor = colorCyan
		}

		//fmt.Printf("\033[2K\r"+priceColor+"%s "+colorGreen+"%s "+colorRed+"%s "+colorReset, parsePriceToString(currentPrice), ticker.BestBid, ticker.BestAsk)
		lastPrice = currentPrice

		fmt.Println("order")
		fmt.Println(order)

		if order != nil {
			//o := getOrder(client, selectedSymbol, order.OrderID)
			//if o.Status == binance.OrderStatusTypeFilled {
				//color.Yellow("PROFIT SELL")
				//os.Exit(1)
				//return
			//}
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
	return f2
}
func parsePriceToString(price float64) string {
	s := fmt.Sprintf("%.4f", price)
	return s
}

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
