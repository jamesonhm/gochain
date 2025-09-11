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
	"github.com/jamesonhm/gochain/internal/executor"
	"github.com/jamesonhm/gochain/internal/monitor"
	//"github.com/jamesonhm/gochain/internal/options"
	"github.com/jamesonhm/gochain/internal/strategy"
	"github.com/jamesonhm/gochain/internal/tasty"
	"github.com/jamesonhm/gochain/internal/yahoo"
	"github.com/joho/godotenv"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//slog.SetDefault(logger)

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

	strats, err := loadStrategies()
	if err != nil {
		logger.Error("", "unable to load strategies", err)
		return
	}
	stratStates := strategy.NewStratStates("teststates.json")
	stratStates.PPrint()

	// SB USER
	tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastySandbox)
	login := tasty.LoginInfo{
		Login:      mustEnv("TASTY_USER"),
		Password:   mustEnv("SB_PASSWORD"),
		RememberMe: true,
	}
	// PROD USER
	//tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastyProd)
	//login := tasty.LoginInfo{
	//	Login:      mustEnv("TASTY_USER"),
	//	Password:   mustEnv("PW"),
	//	RememberMe: true,
	//}
	err = tastyClient.CreateSession(ctx, login)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Tasty Session", slog.String("error creating session", err.Error()))
	}
	logger.Info("Tasty Session", "tasty user", tastyClient.GetUser())

	accts, err := tastyClient.GetAccounts(ctx)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Printf("Got Accounts, Day Trader?: %t\n", accts[0].Account.DayTraderStatus)
	}

	acctNum := accts[0].Account.AccountNumber

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
	// To be used in filtering subscribed option symbols
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
		mktPrices["XSP"] = 648.37
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
		//fmt.Printf("%s - %s\n", c.ExpirationType, c.StreamerSymbols[0:10])
		err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols, mktPrices["XSP"], 9, dxlink.FilterOptionsDates(datesOnly))
		if err != nil {
			fmt.Println(err)
		}
	}

	err = streamClient.Connect()
	if err != nil {
		fmt.Println(err)
		return
	}

	executor := executor.NewEngine(tastyClient, acctNum, streamClient, stratStates, 1, ctx)

	monitor := monitor.NewEngine(
		tastyClient,
		streamClient,
		yahooClient,
		executor,
		stratStates,
		5*time.Second,
	)
	for _, strat := range strats {
		monitor.AddStrategy(strat)
	}
	go monitor.Run(ctx)
	fmt.Printf("-------Monitor Started------")

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		time.Sleep(120 * time.Second)
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
