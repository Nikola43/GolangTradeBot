package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"github.com/adshao/go-binance/v2"
	linq "github.com/ahmetb/go-linq/v3"
	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

var (
	apiKey    = ""
	secretKey = ""
)

func main() {
	if err := keyboard.Open(); err != nil {
		panic(err)
	}
	defer func() {
		_ = keyboard.Close()
	}()


	ch := make(chan string)

	go func(ch chan string) {
		// disable input buffering
		exec.Command("stty", "-F", "/dev/tty", "cbreak", "min", "1").Run()
		// do not display entered characters on the screen
		exec.Command("stty", "-F", "/dev/tty", "-echo").Run()
		var b []byte = make([]byte, 1)
		for {
			os.Stdin.Read(b)
			ch <- string(b)
		}
	}(ch)


	coinName := "kdm"
	pairCoinName := "usdt"

	// binance.UseTestnet = true
	colorReset := "\033[0m"
	colorRed := "\033[31m"
	colorGreen := "\033[32m"
	colorCyan := "\033[36m"

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
	stepSizeInt := 0
	tickSizeInt := 0
	isProfit := false
	hasCoin := false
	wantBuy := true
	currentPrice := 0.0
	var pricesHistory = [][]string{{"", ""}, {"", ""}}

	var order *binance.CreateOrderResponse
	var initialStopOrder *binance.CreateOrderResponse
	var initialBuyOrder *binance.CreateOrderResponse
	var pairBalance binance.Balance
	var selectedSymbolTicker *binance.BookTicker

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


	/*

		wsDepthHandler := func(event *binance.WsDepthEvent) {
			select {
			case stdin, _ := <-ch:
				fmt.Println("Keys pressed: ", stdin)



			default:
				fmt.Println("Working..")
			}

			depthAsksSum := 0.0
			depthBidsSum := 0.0
			depthAsks := event.Asks
			depthBids := event.Bids

			for i := 0; i < len(depthAsks); i++ {
				depthAsksSum += parsePriceToFloat(depthAsks[i].Quantity)
			}

			for i := 0; i < len(depthBids); i++ {
				depthBidsSum += parsePriceToFloat(depthBids[i].Quantity)
			}

			//depthBids := event.Bids
			fmt.Println("depthAsksSum")
			fmt.Println(depthAsksSum)

			fmt.Println("depthBidsSum")
			fmt.Println(depthBidsSum)
			fmt.Println("")

			if depthAsksSum > depthBidsSum {
				log.Printf(colorRed+"%s"+colorReset, "SELL")
			} else {
				log.Printf(colorGreen+"%s"+colorReset, "BUY")
			}

			if depthAsksSum > depthBidsSum {
				log.Printf(colorRed+"%s"+colorReset, "SELL")
			} else {
				log.Printf(colorGreen+"%s"+colorReset, "BUY")
			}
		}
		DepthErrHandler := func(err error) {
			fmt.Println(err)
		}
		DepthDoneC, _, err := binance.WsDepthServe(selectedSymbol, wsDepthHandler, DepthErrHandler)
		if err != nil {
			fmt.Println(err)
			return
		}

		// remove this if you do not want to be blocked here
		<-DepthDoneC

	*/

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

	pairBalance = getCoinBalance(selectedPair, account.Balances)


	fmt.Println("waiting")
	wsMarketStatHandler := func(event *binance.WsMarketStatEvent) {

		t := time.Now()
		b := []string{t.Format("2006-01-02 15:04:05"), event.BidPrice}
		pricesHistory = append(pricesHistory, b)

		if hasCoin == false && wantBuy == true {
			currentPrice = parsePriceToFloat(event.BidPrice)
			log.Println("currentPrice")
			log.Println(currentPrice)

			symbolFilters := getSymbolFilters(selectedSymbol, exchangeInfo.Symbols)
			var tickSize = symbolFilters.Filters[0]["tickSize"].(string)
			var stepSize = symbolFilters.Filters[2]["stepSize"].(string)

			//fmt.Println(stepSize)
			//fmt.Println(tickSize)

			step := parsePriceToFloat(stepSize)
			s := fmt.Sprintf("%f", step)
			p := parsePriceToFloat(s[0:])
			s2 := fmt.Sprintf("%f", p)

			//fmt.Println(s2)
			//fmt.Println(p)


			if p == 1 {
				stepSizeInt = 0
			} else {
				stepSizeInt = len(s2[0 : strings.Index(s2, "1")-1])
			}
			//fmt.Println(stepSizeInt)
			//fmt.Println("----------")

			tickSizeInt = len(tickSize[0 : strings.Index(tickSize, "1")-1])
			//fmt.Println(tickSizeInt)
			//fmt.Println("----------")


			// Get Pair Balance

			selectedSymbolTicker = getTickersBySymbol(client, selectedSymbol)
			buyQuantity = parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice)
			//buyQuantity = parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice) / usedBalancePercent
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
			fmt.Println("initialBuyPrice")
			fmt.Println(initialBuyPrice)

			fmt.Println("minimumSellPrice")
			fmt.Println(minimumSellPrice)


			// INITIAL BUY
			initialBuyOrder, err = client.NewCreateOrderService().Symbol(selectedSymbol).
				Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
				Quantity(parsePriceToString(buyQuantity, stepSizeInt)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(initialBuyOrder)
			hasCoin = true
			wantBuy = false


			fmt.Println("selectedSymbolTicker.AskPrice")
			fmt.Println(selectedSymbolTicker.AskPrice)



			fmt.Println("initialBuyPrice")
			fmt.Println(initialBuyPrice)

			fmt.Println("minimumSellPrice")
			fmt.Println(minimumSellPrice)

			// set stop loss

			// set stop loss
			stopPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((initialBuyPrice*1.7)/100, tickSizeInt)), tickSizeInt))
			sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((initialBuyPrice*2.0)/100, tickSizeInt)), tickSizeInt))

			fmt.Println("sellPrice")
			fmt.Println(sellPrice)
			fmt.Println("stopPrice")
			fmt.Println(stopPrice)
			fmt.Println("sellPrice")
			fmt.Println(sellPrice)
			fmt.Println("stopPrice")
			fmt.Println(stopPrice)

			// CREATE STOP
			initialStopOrder, err = client.NewCreateOrderService().Symbol(selectedSymbol).
				Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
				TimeInForce(binance.TimeInForceTypeGTC).Quantity(parsePriceToString(sellQuantity, stepSizeInt)).
				Price(parsePriceToString(sellPrice, tickSizeInt)).
				StopPrice(parsePriceToString(stopPrice, tickSizeInt)).
				Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Println("initialStopOrder")
			fmt.Println(initialStopOrder)
			fmt.Println(initialStopOrder.OrderID)
			fmt.Println(initialStopOrder.Status)
		}

		select {
		case stdin, _ := <-ch:
			fmt.Println("Keys pressed: ", stdin)

			if stdin == "s" {
				if initialStopOrder != nil {
					cancelOrder(client, selectedSymbol, initialStopOrder.OrderID)
					initialStopOrder = nil
				}

				if order != nil {
					cancelOrder(client, selectedSymbol, order.OrderID)
					order = nil
				}

				order, err = client.NewCreateOrderService().Symbol(selectedSymbol).
					Side(binance.SideTypeSell).Type(binance.OrderTypeMarket).
					Quantity(parsePriceToString(sellQuantity, stepSizeInt)).
					Do(context.Background())
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
					return
				}
				initialStopOrder = nil
				order = nil
				hasCoin = false
				wantBuy = false
				log.Printf("\033[2K\r"+colorGreen+"%s"+colorReset, "SELL")

				file, err := os.Create(parsePriceToString(currentPrice, stepSizeInt))
				if err != nil {
					log.Fatal("error", err)
				}
				defer file.Close()

				writer := csv.NewWriter(file)
				defer writer.Flush()

				for _, value := range pricesHistory {
					writer.Write(value)
				}

			}

			if stdin == "b" {
				pairBalance = getCoinBalance(selectedPair, account.Balances)
				selectedSymbolTicker = getTickersBySymbol(client, selectedSymbol)

				buyQuantity = parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice)
				// buyQuantity = parsePriceToFloat(pairBalance.Free) / parsePriceToFloat(selectedSymbolTicker.AskPrice) / usedBalancePercent
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
				minimumSellPrice = initialBuyPrice + (initialBuyPrice * 3.0 / 100)
				minimumSellPrice = parsePriceToFloat(parsePriceToString(minimumSellPrice, tickSizeInt))
				highPrice = minimumSellPrice

				initialBuyOrder, err = client.NewCreateOrderService().Symbol(selectedSymbol).
					Side(binance.SideTypeBuy).Type(binance.OrderTypeMarket).
					Quantity(parsePriceToString(buyQuantity, stepSizeInt)).
					Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
				fmt.Println(initialBuyOrder)

				// set stop loss
				stopPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((initialBuyPrice*1.7)/100, tickSizeInt)), tickSizeInt))
				sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(initialBuyPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((initialBuyPrice*2.0)/100, tickSizeInt)), tickSizeInt))

				// stop
				initialStopOrder, err = client.NewCreateOrderService().Symbol(selectedSymbol).
					Side(binance.SideTypeSell).Type(binance.OrderTypeStopLossLimit).
					TimeInForce(binance.TimeInForceTypeGTC).Quantity(parsePriceToString(sellQuantity, stepSizeInt)).
					Price(parsePriceToString(sellPrice, tickSizeInt)).
					StopPrice(parsePriceToString(stopPrice, tickSizeInt)).
					Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
				hasCoin = true
				wantBuy = false
				log.Printf("\033[2K\r"+colorGreen+"%s"+colorReset, "BUY")
			}

		default:
			fmt.Println("")
		}

		if hasCoin == true {

			currentPrice = parsePriceToFloat(event.BidPrice)

			if currentPrice > initialBuyPrice {
				isProfit = true
			} else {
				isProfit = false
			}

			if currentPrice > minimumSellPrice &&
				currentPrice > highPrice {
				highPrice = currentPrice
				isProfit = true
				color.Yellow("Nuevo precio más alto")

				// set stop loss
				stopPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(highPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((highPrice*1.5)/100, tickSizeInt)), tickSizeInt))
				sellPrice = parsePriceToFloat(parsePriceToString(parsePriceToFloat(parsePriceToString(highPrice, tickSizeInt))-parsePriceToFloat(parsePriceToString((highPrice*1.7)/100, tickSizeInt)), tickSizeInt))

				fmt.Println("sellPrice")
				fmt.Println(sellPrice)
				fmt.Println("stopPrice")
				fmt.Println(stopPrice)

				if initialStopOrder != nil {
					cancelOrder(client, selectedSymbol, initialStopOrder.OrderID)
				}

				if order != nil {
					cancelOrder(client, selectedSymbol, order.OrderID)
					order = nil
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

			if isProfit == true {
				profitPercent := PercentageChange(initialBuyPrice, currentPrice)
				log.Printf("\033[2K\r"+colorGreen+"%s"+colorReset, "PROFIT: "+parsePriceToString(profitPercent, 2))
			} else {
				lossPercent := PercentageChange(initialBuyPrice, currentPrice)
				log.Printf("\033[2K\r"+colorRed+"%s"+colorReset, "LOSS: "+parsePriceToString(lossPercent, 2))
			}

			// fmt.Printf("\033[2K\r"+priceColor+"%s "+colorGreen+"%s "+colorRed+"%s "+colorReset, parsePriceToString(currentPrice, tickSizeInt), parsePriceToFloat(event.BidPrice), parsePriceToFloat(event.AskPrice))
			lastPrice = currentPrice

			if initialStopOrder != nil {
				// fmt.Println("initialStopOrder")
				// fmt.Println(initialStopOrder)
				o := getOrder(client, selectedSymbol, initialStopOrder.OrderID)

				if o != nil {
					if o.Status == binance.OrderStatusTypeFilled {
						color.Red("STOP LOSS SELL -> FILLED")
						os.Exit(1)
						return
					}
				}
				/*
					else {
						color.Red("STOP LOSS SELL -> NOT EXIST")
						os.Exit(1)
						return
					}
				*/
			}

			if order != nil {
				//fmt.Println("order")
				//fmt.Println(order)
				o := getOrder(client, selectedSymbol, order.OrderID)

				if o != nil {
					if o.Status == binance.OrderStatusTypeFilled {
						color.Green("PROFIT SELL")
						os.Exit(1)
						return
					}
				}
			}
		}

	}
	errHandler := func(err error) {
		fmt.Println(err)
		return
	}
	doneC, _, err := binance.WsMarketStatServe(selectedSymbol, wsMarketStatHandler, errHandler)
	if err != nil {
		fmt.Println(err)
		return
	}
	<-doneC
}

func PercentageChange(old, new float64) (delta float64) {
	diff := float64(new - old)
	delta = (diff / float64(old)) * 100
	return
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
