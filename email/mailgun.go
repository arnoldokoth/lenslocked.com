package email

import (
	"context"
	"fmt"
	"log"

	mailgun "github.com/mailgun/mailgun-go/v4"
)

const (
	welcomeSubject = `Welcome to LensLocked.com`

	welcomeText = `
	Hi there!

	Welcome to Lenslocked.com! We really hope you enjoy using our
	application!

	Best,
	LensLocked Team.
	`

	welcomeHTML = `
	Hi there! <br />
	Welcome to <a href="https://lenslocked.com">LensLocked.com</a> We really hope you enjoy using
	our application!<br /> <br />

	Best, <br />
	Arnold`
)

// WithMailgun ...
func WithMailgun(domain, apiKey, publicKey string) ClientConfig {
	return func(c *Client) {
		mg := mailgun.NewMailgun(domain, apiKey)
		c.mg = mg
	}
}

// WithSender ...
func WithSender(name, email string) ClientConfig {
	return func(c *Client) {
		c.from = buildEmail(name, email)
	}
}

// ClientConfig ...
type ClientConfig func(*Client)

// NewClient ...
func NewClient(opts ...ClientConfig) *Client {
	client := Client{}
	for _, opt := range opts {
		opt(&client)
	}
	return &client
}

// Client ...
type Client struct {
	from string
	mg   mailgun.Mailgun
}

// Welcome ...
func (c *Client) Welcome(toName, toEmail string) error {
	message := c.mg.NewMessage(c.from, welcomeSubject, welcomeText, buildEmail(toName, toEmail))
	message.SetHtml(welcomeHTML)
	_, _, err := c.mg.Send(context.TODO(), message)
	if err != nil {
		log.Println("email.Welcome() ERROR: ", err)
		return err
	}

	return nil
}

func buildEmail(name, email string) string {
	if name == "" {
		return email
	}

	return fmt.Sprintf("%s %s", name, email)
}
