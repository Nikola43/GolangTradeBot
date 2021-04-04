package main

import (
	"context"
	"fmt"
	"github.com/Kucoin/kucoin-go-sdk"
	"github.com/adshao/go-binance/v2"
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/fatih/color"
	"log"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"
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
	//colorYellow := "\033[33m"

	//initialPrice := 0.0
	initialBuyPrice := 0.0
	highPrice := 0.0
	stopLossPrice := 0.0
	sellQuantity := 0.0
	//buyQuantity := 0.0
	lastPrice := 0.0
	//orderId := ""
	//coinExist := false
	coinName := "wrx"
	pairCoinName := "usdt"
	selectedCoin := strings.ToUpper(coinName)
	selectedPair := strings.ToUpper(pairCoinName)
	selectedSymbol := selectedCoin + "" + selectedPair
	fmt.Println(selectedSymbol)

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
	buyQuantity := math.Trunc(parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice))
	sellQuantity = math.Trunc(buyQuantity - (buyQuantity * 1 / 100))

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
	stopLossPrice = initialBuyPrice - (initialBuyPrice * 1 / 100)
	minimumSellPrice := initialBuyPrice + (initialBuyPrice * 1 / 100)

	fmt.Println("order.Status")
	fmt.Println(order.Status)
	order = nil

	// WEB SOCKET
	wsAggTradeHandler := func(event *binance.WsAggTradeEvent) {
		currentPrice := parsePriceToFloat(event.Price)
		fmt.Println("currentPrice")
		fmt.Println(currentPrice)

		if order != nil && order.Status == binance.OrderStatusTypeFilled {
			color.Yellow("PROFIT SELL")
			return
		}

		if currentPrice > initialBuyPrice &&
			currentPrice > minimumSellPrice &&
			currentPrice > highPrice {
			highPrice = currentPrice
			color.Yellow("Nuevo precio más alto")

			stopPrice := highPrice - (highPrice * 1 / 100)
			sellPrice := highPrice - (highPrice * 1.5 / 100)

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
			fmt.Println(order)
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
				return
			}

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

	}
	errHandler := func(err error) {
		fmt.Println(err)
	}
	doneC, _, err := binance.WsAggTradeServe(selectedSymbol, wsAggTradeHandler, errHandler)
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

func randomString(l int) string {
	bytes := make([]byte, l)
	for i := 0; i < l; i++ {
		bytes[i] = byte(randInt(65, 90))
	}
	return string(bytes)
}

func randInt(min int, max int) int {
	return min + rand.Intn(max-min)
}

func getBalanceByCoin(kucoinService *kucoin.ApiService, currency string) string {
	balance := ""
	accounts := kucoin.AccountsModel{}
	b, err := kucoinService.Accounts(currency, "trade")
	if err != nil {
		fmt.Println(err.Error())
	}

	err = b.ReadData(&accounts)
	if err != nil {
		fmt.Println(err.Error())
	}

	if len(accounts) > 0 {
		balance = accounts[0].Available
		log.Printf("Available balance: %s %s => %s", accounts[0].Type, accounts[0].Currency, accounts[0].Available)
	}

	return balance
}

func trailingZero(n int) int {
	var i int = 5
	var count int = 0

	for i <= n {

		count = count + (n / i)

		i = i * 5
	}
	return count
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

func createMarketOrder(kucoinService *kucoin.ApiService, side, symbol, size string) *kucoin.CreateOrderResultModel {
	rand.Seed(time.Now().UnixNano())
	oid := strconv.FormatInt(int64(rand.Intn(99999999)), 10)

	order := &kucoin.CreateOrderModel{
		ClientOid: oid,
		Side:      side,
		Symbol:    symbol,
		Type:      "market",
		Size:      size,
	}

	createOrderResult, err := kucoinService.CreateOrder(order)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	marketOrder := &kucoin.CreateOrderResultModel{}
	err = createOrderResult.ReadData(marketOrder)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println(marketOrder)
	return marketOrder
}

func createTakeProfitOrder(kucoinService *kucoin.ApiService, symbol, size, stopPrice, price string) *kucoin.CreateOrderResultModel {
	createOrderResultModel := &kucoin.CreateOrderResultModel{}
	oid := strconv.FormatInt(int64(rand.Intn(99999999)), 10)

	order := &kucoin.CreateOrderModel{
		ClientOid: oid,
		Side:      "sell",
		Symbol:    symbol,
		Stop:      "loss",
		StopPrice: stopPrice,
		Price:     price,
		Size:      size,
	}

	createOrderResult, err := kucoinService.CreateOrder(order)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	err = createOrderResult.ReadData(createOrderResultModel)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println(createOrderResultModel)
	return createOrderResultModel
}

func getOrder(kucoinService *kucoin.ApiService, orderId string) *kucoin.OrderModel {
	order := &kucoin.OrderModel{}
	getOrderResult, err := kucoinService.Order(orderId)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
	err = getOrderResult.ReadData(order)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}
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

func getSymbolTicker(kucoinService *kucoin.ApiService, selectedSymbol string) *kucoin.TickerLevel1Model {
	apiResponse, err := kucoinService.TickerLevel1(selectedSymbol)
	if err != nil {
		// fmt.Println(err)
	}

	ticker := &kucoin.TickerLevel1Model{}
	err = apiResponse.ReadData(ticker)
	if err != nil {
		// fmt.Println(err.Error())
		return nil
	}
	return ticker
}
