package mock

import (
	"context"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"gopkg.in/irc.v4"
)

type Config struct {
	// Server is the address of the IRC server to connect to, as host:port.
	Server string
	// SearchResultPath is the archive served in reply to an @search request.
	SearchResultPath string
	// BookResultPath is the file served in reply to a book download request.
	BookResultPath string
}

// Operator listens to an IRC channel and responds to search / download requests
type Operator struct {
	config     *Config
	client     *irc.Client
	log        *log.Logger
	dccManager *DccManager
}

func NewOperator(config *Config) *Operator {
	conn, err := net.Dial("tcp", config.Server)
	if err != nil {
		panic(err)
	}

	operator := &Operator{
		config:     config,
		dccManager: NewDccManager(),
		log:        log.New(os.Stdout, "OPERATOR: ", 0),
	}

	client := irc.NewClient(conn, irc.ClientConfig{
		Nick:           "search",
		Pass:           "",
		User:           "search",
		Name:           "search",
		EnableISupport: true,
		EnableTracker:  true,
		PingFrequency:  time.Second * 30,
		PingTimeout:    time.Second * 30,
		SendLimit:      0,
		SendBurst:      0,
		Handler:        irc.HandlerFunc(operator.Handler),
	})

	operator.client = client

	return operator
}

func (o *Operator) Handler(client *irc.Client, message *irc.Message) {
	if strings.HasPrefix(message.Trailing(), "@search") {
		queryParts := strings.SplitN(message.Trailing(), " ", 2)
		if len(queryParts) == 1 {
			return
		}
		o.log.Printf("Search for '%s'\n", queryParts[1])

		dccString, err := o.dccManager.ServeFile(o.config.SearchResultPath)
		if err != nil {
			o.log.Printf("Unable to serve %q: %v\n", o.config.SearchResultPath, err)
			return
		}

		time.Sleep(time.Second * time.Duration(rand.Intn(5)))
		if err = client.Writef(`:search PRIVMSG %s :%s`, message.Name, dccString); err != nil {
			o.log.Println(err)
		}
	}

	if strings.HasPrefix(message.Trailing(), "!") {
		queryParts := strings.SplitN(message.Trailing(), " ", 2)
		if len(queryParts) == 1 {
			return
		}
		o.log.Printf("Download request for '%s'\n", queryParts[1])

		dccString, err := o.dccManager.ServeFile(o.config.BookResultPath)
		if err != nil {
			o.log.Printf("Unable to serve %q: %v\n", o.config.BookResultPath, err)
			return
		}

		time.Sleep(time.Second * time.Duration(rand.Intn(5)))
		if err = client.Writef(`:search PRIVMSG %s :%s`, message.Name, dccString); err != nil {
			o.log.Println(err)
		}
	}
}

func (o *Operator) StartListening(ctx context.Context) error {
	go func() {
		time.Sleep(time.Second * 2)
		o.client.Write("JOIN #ebooks")
		o.log.Println("Joined #ebooks channel")
	}()

	o.log.Printf("Connected to %s\n", o.config.Server)
	return o.client.RunContext(ctx)
}
