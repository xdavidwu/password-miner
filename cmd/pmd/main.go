package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"

	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/pool"
	"git.cs.nctu.edu.tw/wuph0612/password-miner/internal/stratum"
)

const (
	loggerFlags = log.LstdFlags | log.Lmicroseconds
)

func main() {
	addr := flag.String("address", "0.0.0.0:1234", "Address to listen on")
	hashesFile := flag.String("hashes", "hashes", "File containing list of hases")
	flag.Parse()

	log.SetFlags(loggerFlags)

	list, err := os.Open(*hashesFile)
	if err != nil {
		log.Fatalf("Cannot open hashes list %v: %v", *hashesFile, err)
	}
	defer list.Close()

	works := []pool.Work{}

	for {
		w := pool.Work{}
		_, err := fmt.Fscanf(list, "%s %s\n", &w.Algo, &w.Hash)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Fatalf("Fail to parse hashes list: %v", err)
			}
		}
		works = append(works, w)
	}
	log.Printf("%v hashes loaded\n", len(works))

	sol := make(chan string)

	l, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("Fail to listen %v: %v", *addr, err)
	}

	clientWorks := map[chan pool.Work]struct{}{}
	worksC := make(chan pool.Work)
	reg := make(chan chan pool.Work)
	rm := make(chan chan pool.Work)

	go func() {
		var cw *pool.Work = nil
		for {
			select {
			case c := <-reg:
				clientWorks[c] = struct{}{}
				if cw != nil {
					c <- *cw
				}
			case c := <-rm:
				delete(clientWorks, c)
			case w := <-worksC:
				cw = &w
				for c := range clientWorks {
					c <- w
				}
			}
		}
	}()

	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				log.Fatalf("Failed to accept: %v", err)
			}

			go func() {
				r := bufio.NewReader(conn)
				b, err := r.ReadBytes(byte('\n'))
				if err != nil {
					log.Println("Pre-auth read error")
					return
				}
				a := stratum.StratumLoginRequest{}
				err = json.Unmarshal(b, &a)
				if err != nil {
					log.Fatalf("Cannot unmarshal login request: %v", err)
				}

				// TODO auth?

				result := stratum.StratumLoginResult{
					StratumResult: stratum.StratumResult{Status: "OK"},
				}
				res := stratum.StratumLoginResponse{
					JsonRpcResponse: stratum.JsonRpcResponse{
						JsonRpcMessage: stratum.JsonRpcMessage{
							Version: "2.0",
							Id: a.Id,
						},
					},
					Result: &result,
				}
				b, err = json.Marshal(res)
				if err != nil {
					log.Fatalf("Cannot marshal login response: %v", err)
				}
				_, err = conn.Write(append(b, byte('\n')))
				if err != nil {
					log.Println("Pre-auth write error")
					return
				}

				id := fmt.Sprintf("%s-%x", a.Params.Login, rand.Uint32())
				logger := log.New(os.Stderr, "Client(" + id + "): " , loggerFlags | log.Lmsgprefix)
				logger.Printf("New client\n")
				defer logger.Printf("Lost client\n")
				c := pool.Client{
					R: r,
					W: conn,
					Id: id,
					Log: logger,
				}
				w := make(chan pool.Work)
				reg <- w
				defer func() {
					rm <- w
				}()
				c.Control(w, sol)
			}()
		}
	}()

	for _, w := range works {
		log.Printf("Solving %v in %v\n", w.Hash, w.Algo)
		worksC <- w
		log.Printf("Solved %v in %v: \"%v\"", w.Hash, w.Algo, <-sol)
	}
	log.Println("All hashes solved")
}
