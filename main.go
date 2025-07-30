package main

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"slices"
	"strconv"

	"sync"
	"time"

	"github.com/jamesonhm/gochain/internal/dt"
	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/monitor"
	"github.com/jamesonhm/gochain/internal/options"
	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/jamesonhm/gochain/internal/yahoo"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	// defer functions are processed LIFO, context cancel must run before scheduler shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Env Variable Load
	godotenv.Load()

	yahooClient := yahoo.New(mustEnv("YAHOO_API_KEY"), 10*time.Second, 1*time.Second, 1)
	//histParams := yahoo.HistoryParams{
	//	Symbol:        "^VIX",
	//	Interval:      "1d",
	//	DiffAndSplits: false,
	//}
	move, err := yahooClient.ONMove("^VIX")
	if err != nil {
		fmt.Println(err)
	}
	fmt.Printf("VIX ON MOVE: %.2f\n", move)

	//opt := options.OptionSymbol{
	//	Underlying: "XSP",
	//	Date:       time.Date(2025, 7, 29, 0, 0, 0, 0, time.Local),
	//	Strike:     637.43,
	//	OptionType: options.PutOption,
	//}
	//fmt.Printf("Initial option: %+v\n", opt.DxLinkString())
	//for range 7 {
	//	opt.IncrementStrike(1)
	//	fmt.Printf("Incremented: %+v\n", opt.DxLinkString())
	//}

	strats := loadStrategies()
	// SB USER
	//tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastySandbox)
	//login := tasty.LoginInfo{
	//	Login:      mustEnv("TASTY_USER"),
	//	Password:   mustEnv("SB_PASSWORD"),
	//	RememberMe: true,
	//}
	// PROD USER
	tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastyProd)
	login := tasty.LoginInfo{
		Login:      mustEnv("TASTY_USER"),
		Password:   mustEnv("PW"),
		RememberMe: true,
	}
	err = tastyClient.CreateSession(ctx, login)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Tasty Session", slog.String("error creating session", err.Error()))
	}
	logger.Info("Tasty Session", "tasty user", tastyClient.GetUser())

	//customer, err := tastyClient.GetCustomer(ctx)
	//fmt.Println(customer)

	accts, err := tastyClient.GetAccounts(ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Got Accounts, Day Trader?: %t\n", accts[0].Account.DayTraderStatus)
	}

	//acctNum := accts[0].Account.AccountNumber

	//acct, err := tastyClient.GetAccount(ctx, acctNum)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", acct)
	//}

	//status, err := tastyClient.GetAccountTradingStatus(ctx, acctNum)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", status)
	//}

	//pos, err := tastyClient.GetAccountPositions(ctx, acctNum, nil)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", pos)
	//}

	//bal, err := tastyClient.GetAccountBalances(ctx, acctNum)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", bal)
	//}

	// Get curr market price for each tracked symbol
	mktPrices := make(map[string]float64)
	if tastyClient.Env == tasty.TastyProd {
		mktParams := tasty.MarketDataParams{
			Index: []string{"XSP"},
		}
		mktData, err := tastyClient.GetMarketData(ctx, &mktParams)
		if err != nil {
			logger.Error("error getting Tasty Market Data", "error", err)
		} else {
			for _, item := range mktData {
				flVal, err := strconv.ParseFloat(item.Last, 64)
				if err != nil {
					logger.Error("unable to parse float", "string val", item.Last, "for symbol", item.Symbol)
					flVal = 0.0
				}
				mktPrices[item.Symbol] = flVal
			}
		}
	} else {
		mktPrices["XSP"] = 623.37
	}
	fmt.Printf("Last Market Prices: %+v\n", mktPrices)

	chains, err := tastyClient.GetOptionCompact(ctx, "XSP")
	if err != nil {
		fmt.Println(err)
	} else {
		//fmt.Printf("%+v\n", chain)
		for _, chain := range chains {
			fmt.Println(chain.ExpirationType)
			fmt.Println(chain.StreamerSymbols[0:10])
			fmt.Println("=============================================================")
		}
	}

	//act := true
	//syms := []string{"XSP 250430P00529000"}
	//eoSymbol := tasty.EquityOptionsSymbology{
	//	Symbol:     "XSP",
	//	OptionType: tasty.Call,
	//	Strike:     550,
	//	Expiration: time.Date(2025, 04, 25, 0, 0, 0, 0, time.UTC),
	//}
	//eqOpParams := tasty.EquityOptionsParams{
	//	Symbols: []string{eoSymbol.Build()},
	//}
	//eqOpts, err := tastyClient.GetEquityOptions(ctx, &eqOpParams)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", eqOpts)
	//}

	//eqOpSym := tasty.EquityOptionSymbol{
	//	Active: &act,
	//}
	//eqOpt, err := tastyClient.GetEquityOption(ctx, sym, &eqOpSym)
	//if err != nil {
	//	fmt.Println(err)
	//} else {
	//	fmt.Printf("%+v\n", eqOpt)
	//}

	streamer, err := tastyClient.GetQuoteStreamerToken(ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("%+v\n", streamer)
	}

	dteDates := make(map[int]time.Time)
	for _, strat := range strats {
		dtes := strat.ListDTEs()
		for _, dte := range dtes {
			dteDates[dte] = dt.DTEToDate(dte)
		}
	}
	fmt.Printf("DTE Dates: %+v\n", dteDates)
	datesOnly := slices.Collect(maps.Values(dteDates))
	fmt.Printf("DTE Dates Only: %+v\n", datesOnly)

	streamClient := dxlink.New(ctx, streamer.DXLinkURL, streamer.Token)
	for _, c := range chains {
		fmt.Printf("%s - %s\n", c.ExpirationType, c.StreamerSymbols[0:10])
		//err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols, 5, mktPrices["XSP"], 9)
		//err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols, mktPrices["XSP"], 9, dxlink.FilterOptionsDays(5))
		err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols, mktPrices["XSP"], 9, dxlink.FilterOptionsDates(datesOnly))
		if err != nil {
			fmt.Println(err)
		}
	}

	err = streamClient.Connect()
	if err != nil {
		fmt.Println(err)
	}

	monitor := monitor.NewEngine(
		tastyClient,
		streamClient,
		yahooClient,
		5*time.Second,
	)
	for _, strat := range strats {
		monitor.AddStrategy(strat)
	}
	go monitor.Run(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(1 * time.Second)
		fmt.Println("************* ^^^^^^^^^^^^^^^^^^^^^^ *********************** ^^^^^^^^^^^^^^^^^^^^^")
		opt, err := streamClient.StrikeFromDelta("XSP", mktPrices["XSP"], 7, options.PutOption, 1, -0.20)
		if err != nil {
			logger.Error("Strike From Delta Error", "value", err)
		} else {
			logger.Info("Strike from delta Option found", "value", opt)
		}
		time.Sleep(10 * time.Second)
		streamClient.Close()
	}(&wg)
	wg.Wait()

}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Env variable %s required", key))
	}
	return val
}
