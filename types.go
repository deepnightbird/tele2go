package main

import (
	"net/http"
	"time"
)

type OpItem struct {
	id         int64
	sellerhash int64
	my         bool
	status     string
	isPremium  bool
}

type Operation struct {
	timestamp time.Time
	count     int
	op_list   []OpItem
}

type Storage struct {
	_type         string
	_vol          int
	_cost         int
	_depth        int
	lastupdated   time.Time
	accesscount   int
	buy_history   []Operation
	sell_history  []Operation
	moved_history []Operation
	readch        chan []byte
}

func (s *Storage) add_buy_history(timestamp time.Time, op_list []OpItem) {
	s.buy_history = append(s.buy_history, Operation{timestamp, len(op_list), op_list})
}

func (s *Storage) add_sell_history(timestamp time.Time, op_list []OpItem) {
	s.sell_history = append(s.sell_history, Operation{timestamp, len(op_list), op_list})
}

func (s *Storage) add_moved_history(timestamp time.Time, op_list []OpItem) {
	s.moved_history = append(s.moved_history, Operation{timestamp, len(op_list), op_list})
}

type Tele2Struct struct {
	Meta struct {
		Status  string      `json:"status"`
		Message interface{} `json:"message"`
	} `json:"meta"`
	Data []struct {
		ID     string `json:"id"`
		Seller struct {
			Name   string        `json:"name"`
			Emojis []interface{} `json:"emojis"`
		} `json:"seller"`
		TrafficType string `json:"trafficType"`
		Volume      struct {
			Value int    `json:"value"`
			Uom   string `json:"uom"`
		} `json:"volume"`
		Cost struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"cost"`
		Commission struct {
			Amount   float64 `json:"amount"`
			Currency string  `json:"currency"`
		} `json:"commission"`
		Status    string `json:"status"`
		My        bool   `json:"my"`
		Hash      string `json:"hash"`
		IsPremium bool   `json:"isPremium"`
	} `json:"data"`
}

type PrintItm struct {
	stor_idx          int
	vol               int
	cost              int
	_type             string
	_type_color       string
	first_op          time.Time
	first_op_prev     time.Time
	first_op_color    string
	last_op           time.Time
	last_op_prev      time.Time
	last_op_color     string
	nbuys             int
	nbuys_status      int
	nbuys_my          int
	nbuys_ispremium   int
	nbuys_prev        int
	nbuys_color       string
	nsells            int
	nsells_status     int
	nsells_my         int
	nsells_my_prev    int
	nsells_ispremium  int
	nsells_prev       int
	nsells_color      string
	nsells_uniq       int
	nsells_uniq_prev  int
	nsells_uniq_color string
	k                 float64
	k_prev            float64
	k_color           string
	show_item         bool
	moved_up          int
	moved_up_prev     int
	moved_up_color    string
	time_new_values   time.Time
}

type Config struct {
	Interval        int      `toml:"interval"`
	PollIntv        int64    `toml:"pollintv"`
	ProxyList       []string `toml:"proxylist"`
	Lotslist        string   `toml:"lotslist"`
	Depth           int      `toml:"depth"`
	Timeout         int      `toml:"timeout"`
	Savedir         string   `toml:"savedir"`
	Buynumpos       int      `toml:"buynumpos"`
	Customdns       string   `toml:"customdns"`
	Autosize        bool     `toml:"autosize"`
	Dynamicsize     bool     `toml:"dynamicsize"`
	Showrefreshintv int      `toml:"showrefreshintv"`
	Sort            []string `toml:"sort"`
	Showcols        []string `toml:"showcols"`
	Showcondition   string   `toml:"showcondition"`
	Customformula   string   `toml:"customformula"`
	Data            struct {
		Desc  `toml:"desc"`
		Color `toml:"color"`
	} `toml:"data"`
	Voice struct {
		Desc  `toml:"desc"`
		Color `toml:"color"`
	} `toml:"voice"`
	Sms struct {
		Desc  `toml:"desc"`
		Color `toml:"color"`
	} `toml:"sms"`
	last_rows   int
	last_cols   int
	need_resize bool
	term_desc   int
}

type Desc struct {
	Costformula   string `toml:"costformula"`
	Costformulato string `toml:"costformulato"`
	From          int    `toml:"from"`
	To            int    `toml:"to"`
	Step          int    `toml:"step"`
	Depth         int    `toml:"depth"`
}

type Color struct {
	Color               int `toml:"color"`
	ColorHightlight     int `toml:"color_hightlight"`
	BuyColor            int `toml:"buy_color"`
	BuyHighlightColor   int `toml:"buy_highlight_color"`
	SellColor           int `toml:"sell_color"`
	SellHighlightColor  int `toml:"sell_highlight_color"`
	Moves               int `toml:"moves"`
	MovesHighlightColor int `toml:"moves_highlight_color"`
	K                   int `toml:"k"`
	KHighlightColor     int `toml:"k_highlight_color"`
}

type OpInfo struct {
	stor_idx int
	curtime  time.Time
	ttype    string
	vol      int
	cost     int
	optype   string
	oplist   []OpItem
}

type HttpClient struct {
	client        *http.Client
	name          string
	index         int
	last_accessed int64
	last_answer   int
}
