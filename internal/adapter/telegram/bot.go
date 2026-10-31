package telegram

import (
	"fmt"
	"net/http"
	"time"

	"golang.org/x/net/proxy"
	tele "gopkg.in/telebot.v4"
)

type Bot struct {
	*tele.Bot
}

func NewBot(token string) (*Bot, error) {
	dialer, err := proxy.SOCKS5(
		"tcp",
		"127.0.0.1:12334",
		nil,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("create socks5 dialer: %w", err)
	}

	contextDialer, ok := dialer.(proxy.ContextDialer)
	if !ok {
		return nil, fmt.Errorf("socks5 dialer does not support context")
	}

	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = contextDialer.DialContext

	client := &http.Client{
		Transport: transport,
		Timeout:   time.Minute,
	}

	b, err := tele.NewBot(tele.Settings{
		Token:  token,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
		Client: client,
	})
	if err != nil {
		return nil, fmt.Errorf("create bot: %w", err)
	}

	return &Bot{Bot: b}, nil
}
