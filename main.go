package main

import (
	"context"
	//"encoding/json"
	"fmt"
	"log/slog"
	"maps"
	"os"
	"os/signal"
	"slices"
	"strconv"
	"syscall"

	//"sync"
	"time"

	"github.com/jamesonhm/gochain/internal/acctstream"
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
	var ACCT_STREAM bool = true
	var MKT_STREAM bool = true
	// determines wether an order is actually posted
	var LIVE_ORDER bool = false
	var PROD_ACCT bool = false

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	//slog.SetDefault(logger)

	// defer functions are processed LIFO, context cancel must run before scheduler shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Env Variable Load
	godotenv.Load()

	yahooClient := yahoo.New(mustEnv("YAHOO_API_KEY"), 10*time.Second, 1*time.Second, 1, 10*time.Second)
	move, err := yahooClient.ONMovePct("^VIX")
	if err != nil {
		logger.Error("unable to get overnight move for `^VIX`", "error", err)
	}
	fmt.Printf("VIX ON MOVE: %.2f\n", move)
	intraday_move, err := yahooClient.IntradayMove("^VIX")
	if err != nil {
		logger.Error("unable to get intraday move for `^VIX`", "error", err)
	}
	fmt.Printf("VIX Intraday MOVE: %.2f\n", intraday_move)

	var tastyClient *tasty.TastyAPI
	var login tasty.LoginInfo
	if PROD_ACCT {
		// PROD USER
		tastyClient = tasty.New(10*time.Second, 60*time.Second, 60, tasty.TastyProd)
		login = tasty.LoginInfo{
			Login:      mustEnv("TASTY_USER"),
			Password:   mustEnv("PW"),
			RememberMe: true,
		}
	} else {
		// SB USER
		tastyClient = tasty.New(30*time.Second, 60*time.Second, 60, tasty.TastySandbox)
		login = tasty.LoginInfo{
			Login:      mustEnv("TASTY_USER"),
			Password:   mustEnv("SB_PASSWORD"),
			RememberMe: true,
		}
	}
	err = tastyClient.CreateSession(ctx, login)
	if err != nil {
		logger.LogAttrs(ctx, slog.LevelError, "Tasty Session", slog.String("error creating session", err.Error()))
	}
	logger.Info("Tasty Session", "tasty user", tastyClient.GetUser())

	accts, err := tastyClient.GetAccounts(ctx)
	if err != nil {
		logger.Error("unable to get tasty accounts", "error", err)
	} else {
		logger.Info("Got Accounts", "Day Trader?:", accts[0].Account.DayTraderStatus)
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	acctErrChan := make(chan error, 1)
	marketErrChan := make(chan error, 1)

	acctNum := accts[0].Account.AccountNumber
	//balance, err := tastyClient.GetAccountBalances(ctx, acctNum)
	//if err != nil {
	//	logger.Error("unable to get balance info", "error", err)
	//} else {
	//	bytes, _ := json.MarshalIndent(balance, "", "  ")
	//	fmt.Println("Balance")
	//	fmt.Println(string(bytes))
	//}

	// start account streamer
	acctStreamer := acctstream.NewAccountStreamer(ctx, acctNum, tastyClient.GetToken(), tastyClient.Env == tasty.TastyProd)

	startAcctStream := func() {
		err = acctStreamer.Connect()
		if err != nil {
			logger.Error("unable to connect to acccount streamer", "error", err)
			acctErrChan <- err
		}
	}
	if ACCT_STREAM {
		go startAcctStream()
	}

	strats, err := loadStrategies()
	if err != nil {
		logger.Error("unable to load strategies", "error", err)
		return
	}

	streamer, err := tastyClient.GetQuoteStreamerToken(ctx)
	if err != nil {
		logger.Error("unable to get streamer token", "error", err)
		MKT_STREAM = false
	}
	streamClient := dxlink.New(ctx, streamer.DXLinkURL, streamer.Token)

	// setup and run option streamer
	startMarketStream := func() {
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
				marketErrChan <- err
				return
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
			mktPrices["XSP"] = 658.75
		}
		fmt.Printf("Last Market Prices: %+v\n", mktPrices)

		dteDates := make(map[int]time.Time)
		for _, strat := range strats {
			dtes := strat.ListDTEs()
			for _, dte := range dtes {
				dteDates[dte] = dt.DTEToDate(dte)
			}
		}
		datesOnly := slices.Collect(maps.Values(dteDates))

		chains, err := tastyClient.GetOptionCompact(ctx, "XSP")
		if err != nil {
			logger.Error("error getting option chains", "error", err)
			marketErrChan <- err
			return
		}
		for _, c := range chains {
			fmt.Println(c.ExpirationType)
			fmt.Println(c.StreamerSymbols[0:10])
			fmt.Println("=============================================================")
			err = streamClient.UpdateOptionSubs("XSP", c.StreamerSymbols, mktPrices["XSP"], 9, dxlink.FilterOptionsDates(datesOnly))
			if err != nil {
				logger.Error("unable to update option subs", "error", err)
				marketErrChan <- err
				return
			}
		}
		if streamClient.LenOptionSubs() == 0 {
			logger.Error("No Options after filtering")
			marketErrChan <- fmt.Errorf("No options after filtering")
			return
		}

		err = streamClient.Connect()
		if err != nil {
			logger.Error("error connecting to streaming client", "error", err)
			marketErrChan <- err
			return
		}
	}
	if MKT_STREAM {
		go startMarketStream()
	}

	stratStates := strategy.NewStatus("teststates.json")
	//stratStates.PPrint()

	executor := executor.NewEngine(tastyClient, acctNum, streamClient, stratStates, 1, ctx, LIVE_ORDER)

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

	select {
	case sig := <-sigChan:
		logger.Info("Gracefully shutting down", "Received signal:", sig)
		cancel()
	case acctErr := <-acctErrChan:
		logger.Info("Account Streamer Shutting down...", "Error:", acctErr)
		cancel()
	case mktErr := <-marketErrChan:
		logger.Info("Market Streamer Shutting down...", "Error:", mktErr)
		cancel()
	}

	streamClient.Close()
	acctStreamer.Close()
}

func mustEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		panic(fmt.Sprintf("Env variable %s required", key))
	}
	return val
}
