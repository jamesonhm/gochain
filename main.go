package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"sync"
	"time"

	"github.com/jamesonhm/gochain/internal/dxlink"
	"github.com/jamesonhm/gochain/internal/monitor"
	"github.com/jamesonhm/gochain/internal/strategy"
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

	entries := make(map[string]strategy.EntryCondition)
	entries["days"] = strategy.EntryDayOfWeek(time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday)
	entries["d2"] = strategy.EntryDayOfWeek(1, 2, 3)

	strat := strategy.NewStrategy(
		"test",
		"^XSP",
		[]strategy.Leg{
			strategy.NewLeg(strategy.Put, strategy.Sell, 1, 45, strategy.Delta, 35, 5),
			strategy.NewLeg(strategy.Put, strategy.Buy, 1, 45, strategy.Offset, -5, 0),
		},
		strategy.RiskParams{
			PctPortfolio: 100,
			NumContracts: 1,
		},
		entries,
	)

	fmt.Printf("Strategy: %+v\n", strat)

	tastyClient := tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastySandbox)
	login := tasty.LoginInfo{
		Login:      mustEnv("TASTY_USER"),
		Password:   mustEnv("SB_PASSWORD"),
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

	chains, err := tastyClient.GetOptionCompact(ctx, "XSP")
	if err != nil {
		fmt.Println(err)
	}
	//	else {
	//		//fmt.Printf("%+v\n", chain)
	//		for _, chain := range chains {
	//			fmt.Println(chain.ExpirationType)
	//			fmt.Println(chain.StreamerSymbols)
	//			fmt.Println("=============================================================")
	//		}
	//	}

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

	streamClient := dxlink.New(ctx, streamer.DXLinkURL, streamer.Token)
	for _, c := range chains {
		fmt.Printf("%s - %s\n", c.ExpirationType, c.StreamerSymbols[0:3])
		err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols[0:3], 0)
		if err != nil {
			fmt.Println(err)
		}
	}

	// register callback for setting up channels and feeds (called after Authorized)
	// register calbacks for processing data (called at each msgType)

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
	monitor.AddStrategy(strat)
	monitor.Run(ctx)

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		//time.Sleep(30 * time.Second)
		//vixMove, err := streamClient.VixONMove()
		//if err != nil {
		//	slog.Error("Error getting VIX ON move", "error", err)
		//} else {
		//	slog.Info("VIX ON Move:", "data", vixMove)
		//}

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
