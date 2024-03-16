package notify

import (
	"encoding/json"
	"net/url"
	"time"
)

type Priority int8

const (
	Unspecified Priority = iota
	Min
	Low
	Default
	High
	Max
)

type Message struct {
	Topic    string        `json:"topic"`
	Message  string        `json:"message,omitempty"`
	Title    string        `json:"title,omitempty"`
	Tags     []string      `json:"tags,omitempty"`
	Priority Priority      `json:"priority,omitempty"`
	Actions  []Action      `json:"actions,omitempty"`
	ClickUrl *url.URL      `json:"click,omitempty"`
	IconUrl  *url.URL      `json:"icon,omitempty"`
	Delay    time.Duration `json:"delay,omitempty"`
	Email    string        `json:"email,omitempty"`

	// Attach   string `json:"attach,omitempty"`
	// Filename string `json:"filename,omitempty"`
}

// Action represents a custom user action button for notifications
type Action struct {
	Action string `json:"action"`
	Label  string `json:"label"`
	URL    string `json:"url"`
}

type MsgOptFunc func(*Message)

func defaultMsgOpts() *Message {
	return &Message{
		Priority: Default,
		Actions:  []Action{},
		Delay:    0,
	}
}

func newMessage(topic string, title string, body string, tags []string, opts ...MsgOptFunc) *Message {
	o := defaultMsgOpts()
	for _, fn := range opts {
		fn(o)
	}

	return &Message{
		Topic:    topic,
		Message:  body,
		Title:    title,
		Tags:     tags,
		Priority: o.Priority,
		Actions:  o.Actions,
		ClickUrl: o.ClickUrl,
		IconUrl:  o.IconUrl,
		Delay:    o.Delay,
		Email:    o.Email,
	}
}

func (message *Message) Marshal() ([]byte, error) {
	jsonData, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return jsonData, nil
}

func WithPriority(priority Priority) MsgOptFunc {
	return func(message *Message) {
		message.Priority = priority
	}
}

func WithActions(actions ...Action) MsgOptFunc {
	return func(message *Message) {
		message.Actions = append(message.Actions, actions...)
	}
}

func WithClickUrl(clickUrl *url.URL) MsgOptFunc {
	return func(message *Message) {
		message.ClickUrl = clickUrl
	}
}

func WithIcon(iconUrl *url.URL) MsgOptFunc {
	return func(message *Message) {
		message.IconUrl = iconUrl
	}
}

func WithDelay(seconds int) MsgOptFunc {
	return func(message *Message) {
		message.Delay = time.Second * time.Duration(seconds)
	}
}

func WithEmail(email string) MsgOptFunc {
	return func(message *Message) {
		message.Email = email
	}
}
