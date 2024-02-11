package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/Knetic/govaluate"
	genagents "github.com/greentornado/genagents"

	//"github.com/Knetic/govaluate"
	"io"

	strftime "github.com/itchyny/timefmt-go"
	"github.com/mattn/go-colorable"
	"golang.org/x/crypto/ssh/terminal"
	"golang.org/x/net/proxy"
)

var (
	conf           Config
	proc_set_title uintptr = 0
	clientlist     []*HttpClient
)

func in_sells(storage []Storage, x int64) bool {
	for _, r := range storage {
		for _, s := range r.sell_history {
			for _, o := range s.op_list {
				if o.id == x {
					return true
				}
			}
		}
	}
	return false
}

func do_log(fatal bool, args ...interface{}) {
	//if logfile == nil { return }
	log.Println(args...)
	if fatal {
		os.Exit(1)
	}
}

func get_pos(slice []OpItem, value int64) int {
	for p, v := range slice {
		if v.id == value {
			return p
		}
	}
	return -1
}

func exists(s []OpItem, e int64) bool {
	return get_pos(s, e) != -1
}

func get_domen() string {
	agents := [...]string{
		"msk", "spb", "chelyabinsk", "rostov", "irkutsk", "ekt", "nnov", "barnaul", "arh", "belgorod", "bryansk",
		"vladimir", "volgograd", "vologda", "voronezh", "eao", "ivanovo", "kaliningrad", "kaluga", "kamchatka",
		"kuzbass", "kirov", "kostroma", "krasnodar", "krasnoyarsk", "norilsk", "kurgan", "kursk", "lipetsk",
		"magadan", "murmansk", "novgorod", "novosibirsk", "omsk", "orenburg", "orel", "penza", "perm", "vladivostok",
		"pskov", "altai", "buryatia", "karelia", "komi", "mariel", "mordovia", "kazan", "khakasia", "ryazan",
		"samara", "saratov", "sakhalin", "smolensk", "tambov", "tver", "tomsk", "tula", "tyumen", "izhevsk",
		"uln", "hmao", "chuvashia", "yanao", "yar"}
	var r int = rand.Intn(len(agents))
	return agents[r]
}

func dump_list(slice []OpItem, filename string, desc string) {
	if !pfexists("dumps") {
		return
	}
	var file_name string = "dumps\\" + filename + "_" + desc + "_" + strftime.Format(time.Now(), "%y%m%d_%H%M%S") + ".txt"
	if pfexists(file_name) {
		os.Remove(file_name)
	}
	outfile, err := os.Create(file_name)
	if err != nil {
		return
	}
	outfile.WriteString(desc + "\n")
	for _, v := range slice {
		func() (n int, err error) {
			var s string = strconv.FormatInt(v.id, 10) + "\n"
			b := unsafe.Slice(unsafe.StringData(s), len(s))
			return outfile.Write(b)
		}()
	}
	outfile.Close()
}

func setcustomdns(customdns string) {
	if !strings.Contains(customdns, "://") {
		customdns = "udp://" + customdns
	}
	u, err := url.Parse(customdns)
	if err != nil {
		return
	}
	var (
		dnsResolverIP        = u.Host
		dnsResolverProto     = u.Scheme
		dnsResolverTimeoutMs = 5000
	)

	dialer := &net.Dialer{
		Resolver: &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{
					Timeout: time.Duration(dnsResolverTimeoutMs) * time.Millisecond,
				}
				return d.DialContext(ctx, dnsResolverProto, dnsResolverIP)
			},
		},
	}

	dialContext := func(ctx context.Context, network, addr string) (net.Conn, error) {
		return dialer.DialContext(ctx, network, addr)
	}

	http.DefaultTransport.(*http.Transport).DialContext = dialContext
}

func RandStringRunes(n int) string {

	var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandIP() string {
	var r string
	for i := 0; i < 4; i++ {
		r += strconv.Itoa(rand.Intn(253) + 1)
		if i < 3 {
			r += "."
		}
	}
	return r
}

func getTimestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func reqst_item(sitem *Storage, cookielist []*http.Cookie, cookie2 *http.Cookie, client *HttpClient) {

	_type := sitem._type
	_vol := sitem._vol
	_cost := sitem._cost
	_depth := sitem._depth

	var subdom string = get_domen()
	var url string = fmt.Sprintf(
		"https://%s.tele2.ru/api/exchange/lots?trafficType=%s&volume=%d&cost=%d&offset=0&limit=%d", subdom, _type, _vol, _cost, 2*_depth)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		do_log(false, "wget", err)
		time.Sleep(time.Second * time.Duration(1))
		return
	}
	//req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36")
	req.Header.Set("User-Agent", genagents.GenAgent())
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml")
	req.Header.Set("Accept-Charset", "ISO-8859-1,utf-8")
	req.Header.Set("Accept-Encoding", "none")
	req.Header.Set("Accept-Language", "ru-RU,ru;en-US,en;q=0.8")
	req.Header.Set("cache-control", "max-age=0")
	req.Header.Set("referer", "https://"+subdom+".tele2.ru/stock-exchange/internet")
	req.Header.Set("dnt", "1")
	req.Header.Set("pragma", "no-cache")
	req.Header.Set("X-Special-Proxy-Header", RandIP())
	req.Header.Set("X-Forwarded-For", RandIP())
	//language=ru-RU;
	//req.Header.Set("session-cookie", "1778d3ea1fb278ae06f76f4dbeb261f5e4229eacdbe51eccd04f23584d0a080e98104631c81e68bd2467466324fc25c2")
	//req.Header.Set("JSESSIONID", "_trLft2JUj2lOLZhiGolFiBlWrZt_RVX_vl_m5uKV0nAFfbYd7si!-1602525487")
	req.AddCookie(cookie2)
	for c := range cookielist {
		//req.Header.Set("cookie", "language=ru-RU; session-cookie=1778d3ea1fb278ae06f76f4dbeb261f5e4229eacdbe51eccd04f23584d0a080e98104631c81e68bd2467466324fc25c2; JSESSIONID=_trLft2JUj2lOLZhiGolFiBlWrZt_RVX_vl_m5uKV0nAFfbYd7si!-1602525487")
		//req.Header.Set(v, c)
		req.AddCookie(cookielist[c])
	}
	//ck := make(http.Cookie)

	//proto = clientlist[proto_idx]

	for i := 0; i < 50; i++ {
		cur_ts := getTimestamp()
		delta_time := cur_ts - client.last_accessed
		if delta_time >= conf.PollIntv {
			break
		}
		//time.Sleep(time.Millisecond * time.Duration(100))
		time.Sleep(time.Millisecond * time.Duration(10))
	}
	// s1 := clientlist[proto_idx].last_accessed.String()
	resp, err := client.client.Do(req)
	if err != nil || !strings.Contains(resp.Status, "200") {
		//fmt.Println(err)
		//lp = nil
		do_log(false, "wget", err)
		time.Sleep(time.Second * time.Duration(1))
		return
	}
	cookielist = resp.Cookies() //save cookies
	body, err := io.ReadAll(io.Reader(resp.Body))
	if err != nil {
		do_log(false, "wget", err)
		time.Sleep(time.Second * time.Duration(1))
		return
	}
	sitem.readch <- body
	resp.Body.Close()
	last_time := client.last_accessed
	client.last_accessed = getTimestamp()
	client.last_answer = int(client.last_accessed - last_time)
	resp = nil
}

func wget(storage StorageList, offset int, client *HttpClient) {
	//var clientlist []HttpClient

	var cookielist []*http.Cookie

	cookie2 := &http.Cookie{
		Name:  "session-cookie",
		Value: "",
	}

	for {
		for idx := 0; idx < len(storage); idx++ {

			idx_ofs := (idx + offset) % len(storage)

			sitem := &storage[idx_ofs]

			reqst_item(sitem, cookielist, cookie2, client)

			if sitem.prim_fill {
				sitem.prim_fill = false
				reqst_item(sitem, cookielist, cookie2, client)
			}
			//time.Sleep(time.Millisecond * time.Duration(1000/conf.PollFreq))
		}
	}
}

func whandle(storage []Storage, stor_idx int, _type string, _vol int, _cost int, _depth int, opinfo chan<- OpInfo) {

	var lp []OpItem = nil
	var lc []OpItem = nil
	var lunite []OpItem = nil
	var ldata Tele2Struct
	var hash hash.Hash64 = fnv.New64a()

	for body := range storage[stor_idx].readch {
		json.Unmarshal(body, &ldata)
		if len(ldata.Data) == 2*_depth {
			lc = make([]OpItem, 2*_depth)
			for i := 0; i < 2*_depth; i++ {
				lc[i].id, _ = strconv.ParseInt(ldata.Data[i].ID, 10, 64)
				lc[i].my = ldata.Data[i].My
				lc[i].isPremium = ldata.Data[i].IsPremium
				lc[i].status = ldata.Data[i].Status
				var s string = ldata.Data[i].Seller.Name

				if len(s) == 0 {
					s = "null"
					if !lc[i].my {
						lc[i].my = true
					} else {
						_ = lc[i].my
					}
				} /*else {
				    lc[i].my = true
				}*/
				if len(ldata.Data[i].Seller.Emojis) > 0 {
					for i := 0; i < len(ldata.Data[i].Seller.Emojis); i++ {
						s += ldata.Data[i].Seller.Emojis[i].(string)
					}
				} else {
					s += RandStringRunes(3)
				}
				hash.Write([]byte(s))
				lc[i].sellerhash = int64(hash.Sum64())
			}
			var cur_timestamp time.Time = time.Now()

			if len(lp) > 0 {
				var buys_list []OpItem = nil
				var sells_list []OpItem = nil
				var moved_up []OpItem = nil
				var min_unite_idx_lc int = 2 * conf.Depth
				var min_unite_idx_lp int = 2 * conf.Depth

				dump_list(lp, "lp", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))
				dump_list(lc, "lc", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))

				for idx2, x := range lp[:_depth] {
					if !exists(lc, x.id) && idx2 < conf.Buynumpos && !x.my {
						buys_list = append(buys_list, x)
					} else {
						if !exists(lunite, x.id) {
							lunite = append(lunite, x)
							if idx2 < min_unite_idx_lp {
								min_unite_idx_lp = idx2
							}
						}
					}
				}

				if len(buys_list) > 0 {
					//stritm.add_buy_history(cur_timestamp, buys_list)
					opinfo <- OpInfo{stor_idx, cur_timestamp, _type, _vol, _cost, "B", buys_list}
					dump_list(buys_list, "buys_list", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))
				}

				for idx2, x := range lc[:_depth] {
					if !exists(lp, x.id) {
						if in_sells(storage, x.id) {
							moved_up = append(moved_up, x)
						}
						sells_list = append(sells_list, x)
					} else {
						if !exists(lunite, x.id) {
							lunite = append(lunite, x)
							if idx2 < min_unite_idx_lc {
								min_unite_idx_lc = idx2
							}
						}
						var prev_idx int = get_pos(lp, x.id)
						var cur_idx int = get_pos(lc, x.id)
						if prev_idx > conf.Buynumpos+1 && cur_idx < prev_idx {
							for up_lot_prev_idx := prev_idx - 1; up_lot_prev_idx >= 0; up_lot_prev_idx-- {
								var up_lot int64 = lp[up_lot_prev_idx].id
								if exists(lc, up_lot) {
									var up_lot_cur_idx int = get_pos(lc, up_lot)
									if up_lot_cur_idx > up_lot_prev_idx && cur_idx < up_lot_cur_idx {
										moved_up = append(moved_up, x)
										dump_list(lp, "moved_up_lp", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))
										dump_list(lc, "moved_up_lc", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))
									}
									break
								}
							}
						}
					}
				}

				if len(sells_list) > 0 {
					//stritm.add_sell_history(cur_timestamp, sells_list)
					opinfo <- OpInfo{stor_idx, cur_timestamp, _type, _vol, _cost, "S", sells_list}
					dump_list(sells_list, "sells_list", _type+"_"+strconv.Itoa(_vol)+"_"+strconv.Itoa(_cost))
				}

				if len(moved_up) > 0 {
					//stritm.add_moved_history(cur_timestamp, moved_up)
					opinfo <- OpInfo{stor_idx, cur_timestamp, _type, _vol, _cost, "M", moved_up}
					moved_up = nil
				}
			}
			/*lastupdated = cur_timestamp
			  accesscount ++*/
			lp = nil
			lp = lc
			lc = nil
			lunite = nil
		}
	}
}

func color_to_ansi(color int) string {
	// return Esc + "["+strconv.Itoa(color)+";m"
	return Esc + "[" + strconv.Itoa(color) + Comma + "m"
}

func unique(intSlice []int64) []int64 {
	keys := make(map[int64]bool)
	list := []int64{}
	for _, entry := range intSlice {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func doformdata(storage []Storage, printdata []PrintItm, opinfo <-chan OpInfo) {
	var customformula, showcondition *govaluate.EvaluableExpression
	var err error

	if len(conf.Customformula) > 0 {
		customformula, err = govaluate.NewEvaluableExpression(conf.Customformula)
		if err != nil {
			do_log(false, "doformdata", conf.Customformula, err)
			customformula = nil
		}
	}
	if len(conf.Showcondition) > 0 {
		showcondition, err = govaluate.NewEvaluableExpression(conf.Showcondition)
		if err != nil {
			do_log(false, "doformdata", conf.Showcondition, err)
			showcondition = nil
		}
	}
	parameters := make(map[string]interface{}, 3)

	for {
		operation := <-opinfo
		switch operation.optype {
		case "B":
			storage[operation.stor_idx].add_buy_history(operation.curtime, operation.oplist)
		case "S":
			storage[operation.stor_idx].add_sell_history(operation.curtime, operation.oplist)
		case "M":
			storage[operation.stor_idx].add_moved_history(operation.curtime, operation.oplist)
		}
		var cur_timestamp time.Time = time.Now()
		var start_timestamp time.Time = cur_timestamp.Add(-time.Second * time.Duration(conf.Interval))
		var printitem *PrintItm
		for idx, v := range storage {
			printitem = &printdata[idx]
			var color *Color = nil
			switch v._type {
			case "data":
				color = &conf.Data.Color
			case "voice":
				color = &conf.Voice.Color
			case "sms":
				color = &conf.Sms.Color
			}

			var l []Operation
			for _, b := range v.buy_history {
				if b.timestamp.After(start_timestamp) {
					l = append(l, b)
				}
			}
			v.buy_history = nil
			v.buy_history = l
			l = nil
			for _, s := range v.sell_history {
				if s.timestamp.After(start_timestamp) {
					l = append(l, s)
				}
			}
			v.sell_history = nil
			v.sell_history = l
			l = nil
			for _, m := range v.moved_history {
				if m.timestamp.After(start_timestamp) {
					l = append(l, m)
				}
			}
			v.moved_history = nil
			v.moved_history = l
			l = nil

			// set default colors
			if printitem.time_new_values.Before(cur_timestamp.Add(-time.Second)) {
				printitem.nbuys_color = color_to_ansi(color.BuyColor)
				printitem.nsells_color = color_to_ansi(color.SellColor)
				printitem.nsells_uniq_color = color_to_ansi(color.SellColor)
				printitem.k_color = color_to_ansi(color.K)
				printitem.moved_up_color = color_to_ansi(color.Moves)
			}

			// backup all prevoius print data
			printitem.first_op_prev = printitem.first_op
			printitem.last_op_prev = printitem.last_op
			printitem.nbuys_prev = printitem.nbuys
			printitem.nsells_prev = printitem.nsells
			printitem.nsells_uniq_prev = printitem.nsells_uniq
			printitem.k_prev = printitem.k
			printitem.moved_up_prev = printitem.moved_up

			// now start fill print data
			if len(v.buy_history) > 0 && len(v.sell_history) > 0 {
				if v.buy_history[0].timestamp.Before(v.sell_history[0].timestamp) {
					printitem.first_op = v.buy_history[0].timestamp
					printitem.first_op_color = color_to_ansi(color.BuyColor)
				} else {
					printitem.first_op = v.sell_history[0].timestamp
					printitem.first_op_color = color_to_ansi(color.SellColor)
				}

				if v.buy_history[len(v.buy_history)-1].timestamp.After(v.sell_history[len(v.sell_history)-1].timestamp) {
					printitem.last_op = v.buy_history[len(v.buy_history)-1].timestamp
					printitem.last_op_color = color_to_ansi(color.BuyColor)
				} else {
					printitem.last_op = v.sell_history[len(v.sell_history)-1].timestamp
					printitem.last_op_color = color_to_ansi(color.SellColor)
				}
			} else if len(v.buy_history) > 0 {
				printitem.first_op = v.buy_history[0].timestamp
				printitem.first_op_color = color_to_ansi(color.BuyColor)
				printitem.last_op = v.buy_history[len(v.buy_history)-1].timestamp
				printitem.last_op_color = color_to_ansi(color.BuyColor)
			} else if len(v.sell_history) > 0 {
				printitem.first_op = v.sell_history[0].timestamp
				printitem.first_op_color = color_to_ansi(color.SellColor)
				printitem.last_op = v.sell_history[len(v.sell_history)-1].timestamp
				printitem.last_op_color = color_to_ansi(color.SellColor)
			}

			printitem.nbuys = 0
			printitem.nbuys_ispremium = 0
			printitem.nbuys_my = 0
			printitem.nbuys_status = 0
			for _, b := range v.buy_history {
				printitem.nbuys += b.count
				for _, l := range b.op_list {
					if l.isPremium {
						printitem.nbuys_ispremium += 1
					}
					if l.my {
						printitem.nbuys_my += 1
					}
					if l.status != "active" {
						printitem.nbuys_status += 1
					}
				}
			}
			printitem.nsells = 0
			printitem.nsells_ispremium = 0
			printitem.nsells_my = 0
			printitem.nsells_status = 0
			for _, s := range v.sell_history {
				printitem.nsells += s.count
				for _, l := range s.op_list {
					if l.isPremium {
						printitem.nsells_ispremium += 1
					}
					if l.my {
						printitem.nsells_my += 1
					}
					if l.status != "active" {
						printitem.nsells_status += 1
					}
				}
			}
			printitem.nsells_uniq = 0
			var sells []int64
			for _, s := range v.sell_history {
				for _, o := range s.op_list {
					sells = append(sells, o.sellerhash)
				}
			}
			if len(sells) > 1 {
				printitem.nsells_uniq = len(unique(sells))
			} else if len(sells) == 1 {
				printitem.nsells_uniq = 1
			}
			if printitem.nsells_uniq != printitem.nsells {
				_ = printitem.nsells_uniq
			}
			sells = nil
			printitem.moved_up = 0
			for _, m := range v.moved_history {
				printitem.moved_up += m.count
			}

			printitem.k = -1
			if customformula != nil {
				parameters["b"] = float64(printitem.nbuys)
				parameters["s"] = float64(printitem.nsells)
				parameters["bm"] = float64(printitem.nbuys_my)
				parameters["sm"] = float64(printitem.nsells_my)
				parameters["d"] = float64(conf.Interval)
				parameters["m"] = float64(printitem.moved_up)
				parameters["u"] = float64(printitem.nsells_uniq)
				parameters["i"] = float64(conf.Interval)
				k, err := customformula.Evaluate(parameters)
				if err == nil {
					var kf float64 = k.(float64)
					if !math.IsNaN(kf) {
						printitem.k = kf
						printitem.k_color = color_to_ansi(color.K)
					} else {
						printitem.k = math.NaN()
						printitem.k_color = color_to_ansi(color.K)
					}
				}
			}
			printitem.show_item = true
			if showcondition != nil {
				parameters["b"] = printitem.nbuys
				parameters["s"] = printitem.nsells
				parameters["bm"] = printitem.nbuys_my
				parameters["sm"] = printitem.nsells_my
				parameters["d"] = conf.Interval
				parameters["m"] = printitem.moved_up
				parameters["k"] = printitem.k
				parameters["u"] = printitem.nsells_uniq
				show, err := showcondition.Evaluate(parameters)
				if err == nil {
					printitem.show_item = show.(bool)
				}
			}

			if !printitem.first_op.Equal(printitem.first_op_prev) {
				if printitem.first_op_color == color_to_ansi(color.BuyColor) {
					printitem.first_op_color = color_to_ansi(color.BuyHighlightColor)
				} else if printitem.first_op_color == color_to_ansi(color.SellColor) {
					printitem.first_op_color = color_to_ansi(color.SellHighlightColor)
				}
				printitem.time_new_values = cur_timestamp
			}
			if !printitem.last_op.Equal(printitem.last_op_prev) {
				if printitem.last_op_color == color_to_ansi(color.BuyColor) {
					printitem.last_op_color = color_to_ansi(color.BuyHighlightColor)
				} else if printitem.last_op_color == color_to_ansi(color.SellColor) {
					printitem.last_op_color = color_to_ansi(color.SellHighlightColor)
				}
				printitem.time_new_values = cur_timestamp
			}
			if printitem.nbuys != printitem.nbuys_prev {
				printitem.nbuys_color = color_to_ansi(color.BuyHighlightColor)
				printitem.time_new_values = cur_timestamp
				printitem.first_op_color = color_to_ansi(color.BuyHighlightColor)
			}
			if printitem.nsells != printitem.nsells_prev {
				printitem.nsells_color = color_to_ansi(color.SellHighlightColor)
				printitem.time_new_values = cur_timestamp
				printitem.last_op_color = color_to_ansi(color.SellHighlightColor)
			}
			if printitem.nsells_uniq != printitem.nsells_uniq_prev {
				printitem.nsells_uniq_color = color_to_ansi(color.SellHighlightColor)
				printitem.time_new_values = cur_timestamp
			}
			if printitem.k != printitem.k_prev {
				printitem.k_color = color_to_ansi(color.KHighlightColor)
				printitem.time_new_values = cur_timestamp
			}
			if printitem.moved_up != printitem.moved_up_prev {
				printitem.moved_up_color = color_to_ansi(color.MovesHighlightColor)
				printitem.time_new_values = cur_timestamp
			}
		}

		if len(conf.Sort) > 0 {
			rev_slc := []string{}

			for i := len(conf.Sort) - 1; i >= 0; i-- {
				rev_slc = append(rev_slc, conf.Sort[i])
			}

			for _, v := range rev_slc {
				// f, l, t, v, c, b, s
				var desc bool = false
				v = col_name(v)
				if len(v) > 0 && v[len(v)-1:] == "d" {
					desc = true
					v = v[:len(v)-1]
				}
				switch v {
				case "f":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].first_op.After(printdata[j].first_op)
							} else {
								return printdata[i].first_op.Before(printdata[j].first_op)
							}
						})
				case "l":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].last_op.After(printdata[j].last_op)
							} else {
								return printdata[i].last_op.Before(printdata[j].last_op)
							}
						})
				case "t":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i]._type > printdata[j]._type
							} else {
								return printdata[i]._type < printdata[j]._type
							}
						})
				case "v":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].vol > printdata[j].vol
							} else {
								return printdata[i].vol < printdata[j].vol
							}
						})
				case "c":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].cost > printdata[j].cost
							} else {
								return printdata[i].cost < printdata[j].cost
							}
						})
				case "b":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].nbuys > printdata[j].nbuys
							} else {
								return printdata[i].nbuys < printdata[j].nbuys
							}
						})
				case "s":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].nsells > printdata[j].nsells
							} else {
								return printdata[i].nsells < printdata[j].nsells
							}
						})
				case "sm":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].nsells_my > printdata[j].nsells_my
							} else {
								return printdata[i].nsells_my < printdata[j].nsells_my
							}
						})
				case "u":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].nsells_uniq > printdata[j].nsells_uniq
							} else {
								return printdata[i].nsells_uniq < printdata[j].nsells_uniq
							}
						})
				case "k":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].k > printdata[j].k
							} else {
								return printdata[i].k < printdata[j].k
							}
						})
				case "m":
					sort.Slice(
						printdata, func(i, j int) bool {
							if desc {
								return printdata[i].moved_up > printdata[j].moved_up
							} else {
								return printdata[i].moved_up < printdata[j].moved_up
							}
						})
				}
			}
		}
		//time.Sleep(1 * time.Second)
	}
}

func pfexists(path string) bool {
	_, err := os.Stat(path)
	if err == nil {
		return true
	}
	if os.IsNotExist(err) {
		return false
	}
	return false
}

func anyshow(printdata []PrintItm) bool {
	for _, v := range printdata {
		if v.show_item {
			return true
		}
	}
	return false
}

func dosavedata(printdata []PrintItm, param string) {
	if !pfexists(conf.Savedir) {
		_ = os.Mkdir(conf.Savedir, os.ModePerm)
	}
	var filename string = conf.Savedir + "\\"
	var delim string = "\t"
	for {

		if anyshow(printdata) {
			outfile, err := os.Create(filename + param + "_" + strftime.Format(time.Now(), "%y%m%d_%H%M%S") + "_" + strconv.Itoa(conf.Interval) + ".txt")
			if err == nil {
				for _, v := range printdata {
					if !v.show_item {
						continue
					}
					/*var text string = strftime.Format(v.first_op, "%H:%M:%S")
					  text += delim + strftime.Format(v.last_op, "%H:%M:%S")
					  text += delim + v._type
					  text += delim + strconv.Itoa(v.vol)
					  text += delim + strconv.Itoa(v.cost)
					  text += delim + strconv.Itoa(v.nbuys)
					  text += delim + strconv.Itoa(v.nsells)
					  text += delim + strconv.Itoa(v.nsells_my)
					  text += delim + strconv.Itoa(v.moved_up)
					  if v.k != -1 {
					      if v.k == math.Round(v.k) {
					          text += delim + strconv.Itoa(int(v.k))
					      } else {
					          text += delim + fmt.Sprintf("%3.2f", v.k)
					      }
					  }
					  //text += delim + strconv.Itoa(v.)*/

					var print_text string = ""
					var add_delim string = ""
					for _, a := range conf.Showcols {
						switch column := col_name(a); {
						case column == "f":
							print_text += add_delim + strftime.Format(v.first_op, "%H:%M:%S")
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "l":
							print_text += add_delim + strftime.Format(v.last_op, "%H:%M:%S")
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "t":
							print_text += add_delim + v._type
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "v":
							print_text += add_delim + strconv.Itoa(v.vol)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "c":
							print_text += add_delim + strconv.Itoa(v.cost)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "b":
							print_text += add_delim + strconv.Itoa(v.nbuys)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "bm":
							print_text += add_delim + strconv.Itoa(v.nbuys_my)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "bs":
							print_text += add_delim + strconv.Itoa(v.nbuys_status)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "bp":
							print_text += add_delim + strconv.Itoa(v.nbuys_ispremium)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "s":
							print_text += add_delim + strconv.Itoa(v.nsells)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "sm":
							print_text += add_delim + strconv.Itoa(v.nsells_my)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "ss":
							print_text += add_delim + strconv.Itoa(v.nsells_status)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "sp":
							print_text += add_delim + strconv.Itoa(v.nsells_ispremium)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "u":
							print_text += add_delim + strconv.Itoa(v.nsells_uniq)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "m":
							print_text += add_delim + strconv.Itoa(v.moved_up)
							if len(add_delim) == 0 {
								add_delim = delim
							}
						case column == "k":
							if !math.IsNaN(v.k) {
								if v.k != -1 {
									if v.k != math.Round(v.k) {
										print_text += add_delim + fmt.Sprintf("%3.2f", v.k)
									} else {
										print_text += add_delim + strconv.Itoa(int(v.k))
									}
									if len(add_delim) == 0 {
										add_delim = delim
									}
								}
							}
						}
					}
					outfile.WriteString(print_text + "\n")
				}
				outfile.Close()
			} else {
				do_log(false, "dosavedata", param, err)
			}
		}
		time.Sleep(time.Duration(conf.Interval) * time.Second)
	}
}

func col_name(column string) string {
	for i, r := range column {
		if int(r) >= int('0') && int(r) <= int('9') {
			return column[:i]
		}
	}
	return column
}

func col_num(column string) string {
	for i, r := range column {
		if int(r) >= int('0') && int(r) <= int('9') {
			return column[i:]
		}
	}
	return ""
}

func add_is_shown(column string, source *string, addcolor string, addtext interface{}) {
	var print_text string

	var l int
	var f string
	var e error
	num := col_num(column)
	if strings.Contains(column, ".") {
		f = num
	} else {
		l, e = strconv.Atoi(num)
	}

	switch addval := addtext.(type) {
	case int:
		print_text = strconv.Itoa(addval)
	case string:
		print_text = addval
	case float64:
		print_text = fmt.Sprintf("%"+f+"f", addval)
		l = len(print_text)
	case time.Time:
		print_text = strftime.Format(addval, "%H:%M:%S")
	default:
		print_text = "<unk>"
	}

	if len(num) == 0 {
		*source = *source + addcolor + print_text
	} else {
		if e == nil {
			if len(print_text) == l {
				print_text = print_text[:l]
			} else if len(print_text) > l {
				switch addval := addtext.(type) {
				case int:
					print_text = strings.Repeat("#", l)
				case float64:
					if addtext.(float64) != math.Round(addtext.(float64)) {
						print_text = strings.Repeat("#", l)
					}
				default:
					print_text = print_text[:l]
					_ = addval
				}
			} else {
				for {
					if len(print_text) == l {
						break
					}
					print_text = " " + print_text
				}
			}
			*source = *source + addcolor + print_text
		}
	}
	*source = *source + " "
}

func doprintdata(printdata []PrintItm, param string) {
	fmt.Fprint(stdOut, "\033[2J")
	HideCursor()

	for {
		if conf.Autosize {
			var rows int = 0
			for _, r := range printdata {
				if r.show_item {
					rows++
				}
			}
			if conf.Dynamicsize {
				if rows < 3 {
					rows = 3
				}
				if conf.last_rows != rows {
					conf.last_rows = rows
					conf.need_resize = true
				}
			}

			if conf.need_resize {
				set_term_size(conf.last_cols, conf.last_rows)
				conf.need_resize = false
			}
		}
		var row int = 1
		var max_col int = 0
		for _, v := range printdata {
			if !v.show_item {
				continue
			}

			var posx int = 1
			MoveTo(row, posx)
			//PrintAnsi(fmt.Sprintf("%50s", " "), FORE_RESET, BACK_BLACK)
			//ClearLine(stdOut)

			var print_text string = ""
			for _, a := range conf.Showcols {
				switch column := col_name(a); {
				case column == "f":
					add_is_shown(a, &print_text, v.first_op_color, v.first_op)
				case column == "l":
					add_is_shown(a, &print_text, v.last_op_color, v.last_op)
				case column == "t":
					add_is_shown(a, &print_text, v._type_color, v._type)
				case column == "v":
					add_is_shown(a, &print_text, v._type_color, v.vol)
				case column == "c":
					add_is_shown(a, &print_text, v._type_color, v.cost)
				case column == "b":
					add_is_shown(a, &print_text, v.nbuys_color, v.nbuys)
				case column == "bm":
					add_is_shown(a, &print_text, v.nbuys_color, v.nbuys_my)
				case column == "bs":
					add_is_shown(a, &print_text, v.nbuys_color, v.nbuys_status)
				case column == "bp":
					add_is_shown(a, &print_text, v.nbuys_color, v.nbuys_ispremium)
				case column == "s":
					add_is_shown(a, &print_text, v.nsells_color, v.nsells)
				case column == "sm":
					add_is_shown(a, &print_text, v.nsells_color, v.nsells_my)
				case column == "ss":
					add_is_shown(a, &print_text, v.nsells_color, v.nsells_status)
				case column == "sp":
					add_is_shown(a, &print_text, v.nsells_color, v.nsells_ispremium)
				case column == "u":
					add_is_shown(a, &print_text, v.nsells_uniq_color, v.nsells_uniq)
				case column == "m":
					add_is_shown(a, &print_text, v.moved_up_color, v.moved_up)
				case column == "k":
					if !math.IsNaN(v.k) {
						add_is_shown(a, &print_text, v.k_color, v.k)
					}
				}
			}
			print_text = print_text[:len(print_text)-1]
			PrintAnsi(print_text + FORE_RESET)
			posx = len(print_text) - strings.Count(print_text, Esc)*len(FORE_WHITE) + 1
			if max_col < posx {
				max_col = posx
			}
			row++
		}
		if conf.last_rows > row {
			for i := row; i <= conf.last_rows; i++ {
				MoveTo(i, 1)
				ClearLine()
			}
		}
		if conf.last_cols != max_col {
			conf.last_cols = max_col
			conf.need_resize = true
		}
		stdOut.Flush()
		var title string = strconv.Itoa(conf.Interval) + "s"

		for idx := range clientlist {
			intv := clientlist[idx].last_answer
			if intv > (conf.Timeout)*1000 {
				intv = -1
			}
			title += " " + strconv.Itoa(idx) + "=" + strconv.Itoa(int(intv)) + "ms"
			//conf.add_title_data = strconv.Itoa(client.index) + "=" + strconv.Itoa(int(last_time)) + "ms"
		}

		set_title(strftime.Format(time.Now(), "%H:%M:%S") + " " + param + " " + title)
		time.Sleep(time.Duration(conf.Showrefreshintv) * time.Millisecond)
	}
}

func isvalidcategory(category string) bool {
	switch category {
	case
		"data",
		"voice",
		"sms",
		"lots":
		return true
	}
	return false
}

/*func set_title(args ...string) {
    //cmd := exec.Command("cmd", "/c", "title", args...)
    args = append(args[:1], args[0:]...)
    args[0] = "title"
    args = append(args[:1], args[0:]...)
    args[0] = "/c"
    cmd := exec.Command("cmd", args...)
    _ = cmd.Run()
}*/

func set_title(title string) (int, error) {
	if proc_set_title == 0 {
		handle, err := syscall.LoadLibrary("Kernel32.dll")
		if err != nil {
			return 0, err
		}
		defer syscall.FreeLibrary(handle)
		proc_set_title, err = syscall.GetProcAddress(handle, "SetConsoleTitleW")
		if err != nil {
			return 0, err
		}
	}
	r, _, err := syscall.Syscall(proc_set_title, 1, uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))), 0, 0)
	return int(r), err
}

func set_term_size(cols int, lines int) {
	cmd := exec.Command("mode", "con:", fmt.Sprintf("cols=%d", cols), fmt.Sprintf("lines=%d", lines))
	_ = cmd.Run()
	HideCursor()
}

func main() {

	/*f, err := os.Create("tele2go.prof")
	  if err != nil {
	      fmt.Println(err)
	      return
	  }

	  pprof.StartCPUProfile(f)
	*/

	var logfile *os.File

	if pfexists("logs") {
		var err error
		var log_filename string = "logs\\log_" + strftime.Format(time.Now(), "%y%m%d_%H%M%S") + ".txt"
		logfile, err = os.OpenFile(log_filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		defer logfile.Close()
		log.SetOutput(logfile)
	} else {
		log.SetOutput(ioutil.Discard)
	}

	fmt.Println("Starting " + version + " build " + BuildTime)
	_, err := toml.DecodeFile("tele2go.ini", &conf)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		do_log(false, "main", "toml.DecodeFile", err)
		os.Exit(1)
	}

	if len(conf.ProxyList) == 0 {
		conf.ProxyList = append(conf.ProxyList, "direct")
	}

	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "No paramenter specified")
		fmt.Fprintln(os.Stderr, "Valid options are: data | voice | sms | lots")
		os.Exit(1)
	}
	var param string = os.Args[1]
	if !isvalidcategory(param) {
		fmt.Fprintln(os.Stderr, "Unknown paramenter:", param)
		fmt.Fprintln(os.Stderr, "Valid options are: data | voice | sms | lots")
		os.Exit(1)
	}

	conf.term_desc = int(os.Stdout.Fd())
	orig_w, orig_h, err := terminal.GetSize(conf.term_desc)
	if err != nil {
		conf.term_desc = int(os.Stdin.Fd())
		orig_w, orig_h, _ = terminal.GetSize(conf.term_desc)
	}

	stdOut = bufio.NewWriter(colorable.NewColorableStdout())

	var storage StorageList
	var printdata []PrintItm
	storage = make(StorageList, 0)
	var prm Desc

	rand.Seed(time.Now().UnixNano())

	if param == "lots" {
		if len(conf.Lotslist) == 0 {
			fmt.Fprintln(os.Stderr, "No lots file specified")
			os.Exit(1)
		}
		lots_file, err := os.Open(conf.Lotslist)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Can't open lots file: ", conf.Lotslist)
			os.Exit(1)
		}

		scanner := bufio.NewScanner(lots_file)
		for scanner.Scan() {
			var text string = scanner.Text()
			if len(text) == 0 {
				continue
			}
			if text[:1] == "#" {
				continue
			}
			for {
				text = strings.Replace(text, "  ", " ", -1)
				if !strings.Contains(text, "  ") {
					break
				}
			}
			for {
				text = strings.Replace(text, "\t\t", "\t", -1)
				if !strings.Contains(text, "\t\t") {
					break
				}
			}
			text = strings.Replace(text, "\t", " ", -1)
			text = strings.TrimSpace(text)
			split := strings.Split(text, " ")
			//fmt.Println(split)
			switch split[0] {
			case "data":
				prm = conf.Data.Desc
			case "voice":
				prm = conf.Voice.Desc
			case "sms":
				prm = conf.Sms.Desc
			default:
				continue
			}
			if len(split) == 2 {
				costformula, err := govaluate.NewEvaluableExpression(prm.Costformula)
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				parameters := make(map[string]interface{}, 1)
				v, err := strconv.Atoi(split[1])
				if err != nil {
					continue
				}
				parameters["v"] = v
				cf, err := costformula.Evaluate(parameters)
				var cfi int = int(cf.(float64))
				if err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
				split = append(split, strconv.Itoa(cfi))
			}
			if isvalidcategory(split[0]) {
				v, err := strconv.Atoi(split[1])
				if err != nil {
					continue
				}
				c, err := strconv.Atoi(split[2])
				if err != nil {
					continue
				}
				storage = append(storage, Storage{split[0], v, c, prm.Depth, time.Now(), 0, nil, nil, nil, make(chan []byte), true})
			}
		}
		lots_file.Close()
	} else {
		switch param {
		case "data":
			prm = conf.Data.Desc
		case "voice":
			prm = conf.Voice.Desc
		case "sms":
			prm = conf.Sms.Desc
		}
		costformula, err := govaluate.NewEvaluableExpression(prm.Costformula)
		if err != nil {
			do_log(false, "main", "costformula", prm.Costformula, err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		costformulato, err := govaluate.NewEvaluableExpression(prm.Costformulato)
		if err != nil {
			do_log(false, "main", "costformulato", prm.Costformulato, err)
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		for i := prm.From; i <= prm.To; i += prm.Step {
			parameters := make(map[string]interface{}, 1)
			parameters["v"] = i
			cf, err := costformula.Evaluate(parameters)
			var cff float64 = cf.(float64)
			var cfi int = int(math.Ceil(cff))

			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				do_log(false, "main", "costformula", "int(math.Ceil)", err)
				os.Exit(1)
			}
			cft, err := costformulato.Evaluate(parameters)
			var cfft float64 = cft.(float64)
			var cfti int = int(math.Ceil(cfft))
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				do_log(false, "main", "costformulato", "int(math.Ceil)", err)
				os.Exit(1)
			}
			for j := cfi; j <= cfti; j++ {
				storage = append(storage, Storage{param, i, j, prm.Depth, time.Now(), 0, nil, nil, nil, make(chan []byte), true})
			}
		}
	}

	if len(conf.Customdns) > 0 {
		setcustomdns(conf.Customdns)
	}

	opinfo := make(chan OpInfo, len(storage))

	var timeout = time.Duration(conf.Timeout) * time.Second
	offset := int(math.Ceil(float64(len(storage)) / float64(len(conf.ProxyList))))

	for idx, proxy_name := range conf.ProxyList {

		jar, _ := cookiejar.New(nil)
		httpClient := HttpClient{}
		httpClient.client = &http.Client{Timeout: timeout, Jar: jar}

		proxyUrl, _ := url.Parse(proxy_name)
		transport := http.Transport{}
		if strings.Contains(proxyUrl.Scheme, "socks") {
			dialSocksProxy, err := proxy.SOCKS5("tcp", proxyUrl.Host, nil, proxy.Direct)
			if err == nil {
				transport.Dial = dialSocksProxy.Dial
			} else {
				log.Println("Error connecting to proxy:", err)
				os.Exit(1)
			}
		} else if strings.Contains(proxyUrl.Scheme, "http") {
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
		transport.TLSClientConfig = &tls.Config{
			InsecureSkipVerify: true}
		httpClient.client.Transport = &transport
		httpClient.last_accessed = getTimestamp()
		httpClient.name = proxy_name
		httpClient.index = idx
		clientlist = append(clientlist, &httpClient)
		go wget(storage, offset+idx, &httpClient)
	}

	for idx, sitem := range storage {
		go whandle(storage, idx, sitem._type, sitem._vol, sitem._cost, sitem._depth, opinfo)
		var color *Color = nil
		switch sitem._type {
		case "data":
			color = &conf.Data.Color
		case "voice":
			color = &conf.Voice.Color
		case "sms":
			color = &conf.Sms.Color
		}
		var printitem = PrintItm{
			idx, sitem._vol, sitem._cost, sitem._type, color_to_ansi(color.Color),
			time.Date(9999, 1, 1, 00, 00, 00, 00, time.UTC), time.Date(9999, 1, 1, 00, 00, 00, 00, time.UTC), color_to_ansi(color.Color),
			time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC), time.Date(0001, 1, 1, 00, 00, 00, 00, time.UTC), color_to_ansi(color.Color),
			0, 0, 0, 0, 0, color_to_ansi(color.BuyColor), // byus
			0, 0, 0, 0, 0, 0, color_to_ansi(color.SellColor), // sells
			0, 0, color_to_ansi(color.SellColor), // sells_prev
			0, 0, color_to_ansi(color.K), // k
			true, // show item
			0, 0, color_to_ansi(color.Moves), time.Now()}
		printdata = append(printdata, printitem)
	}

	conf.last_cols = 51
	conf.last_rows = len(storage)
	conf.need_resize = true

	go doformdata(storage, printdata, opinfo)
	HideCursor()
	time.Sleep(time.Second)
	go doprintdata(printdata, param)

	if len(conf.Savedir) > 0 {
		go dosavedata(printdata, param)
	}
	var quit = make(chan bool)

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for sig := range c {
			// sig is a ^C, handle it
			_ = sig
			quit <- true
		}
	}()
	<-quit
	close(quit)
	// close(opinfo)
	set_term_size(orig_w, orig_h)
	fmt.Fprint(stdOut, "\033[2J")
	ShowCursor()

	fmt.Fprintln(os.Stderr, "Done")
	//pprof.StopCPUProfile()
}
