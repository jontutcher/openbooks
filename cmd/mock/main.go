package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/evan-buss/openbooks/mock"
)

func main() {
	// The IRC server itself is started separately ("task irc") so that it can
	// keep running across restarts of this operator.
	config := &mock.Config{}
	flag.StringVar(&config.Server, "server", "localhost:6667", "Address of the IRC server to connect to.")
	flag.StringVar(&config.SearchResultPath, "search-results",
		"SearchBot_results_for_ the great gatsby.txt.zip",
		"Archive served in response to an @search request.")
	flag.StringVar(&config.BookResultPath, "book-result", "ebook.epub",
		"File served in response to a book download request.")
	flag.Parse()

	// Fail now rather than when the first request comes in.
	for _, path := range []string{config.SearchResultPath, config.BookResultPath} {
		if _, err := os.Stat(path); err != nil {
			log.Fatalf("unable to read the mock response file: %v", err)
		}
	}

	// Ctrl-C or SIGTERM closes the operator's IRC connection. SIGKILL cannot be
	// caught, so it is deliberately not listed here.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	operator := mock.NewOperator(config)

	if err := operator.StartListening(ctx); err != nil && !errors.Is(err, context.Canceled) {
		log.Fatal(err)
	}
}
