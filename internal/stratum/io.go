package stratum

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
)

func Stratum(r io.Reader, w io.Writer, loginParam StratumLoginParams) {
	reader := bufio.NewReader(r)
	id := 1 // we don't really track this yet
	login, err := json.Marshal(StratumLoginRequest{
		JsonRpcRequest: JsonRpcRequest{
			JsonRpcMessage: JsonRpcMessage{
				Version: "2.0",
				Id: id,
			},
			Method: "login",
		},
		Params: loginParam,
	})
	if err != nil {
		log.Fatalf("unable to marshal login request: %v", err)
	}
	_, err = w.Write(append(login, byte('\n')))
	if err != nil {
		log.Fatalf("unable to send login request: %v", err)
	}

	res, err := reader.ReadBytes('\n')
	if err != nil {
		log.Fatalf("unable to read login response: %v", err)
	}
	response := StratumLoginResponse{}
	err = json.Unmarshal(res, &response)
	if err != nil {
		log.Fatalf("unable to unmarshal response: %v", err)
	}

	if response.Error != nil {
		log.Fatalf("Error on login: %v", response.Error.Message)
	}

	in := make(chan StratumJobParams)
	out := make(chan StratumSubmitParams)
	stop := SwitchedScan(in, out)

	if response.Result.Job != nil {
		in <- *response.Result.Job
	}

	go func() {
		for {
			sub, ok := <-out
			if !ok {
				close(stop)
				return
			}
			request, err := json.Marshal(StratumSubmitRequest{
				JsonRpcRequest: JsonRpcRequest{
					JsonRpcMessage: JsonRpcMessage{
						Version: "2.0",
						Id: id,
					},
					Method: "sumbit",
				},
				Params: sub,
			})
			if err != nil {
				log.Printf("unable to marshal submit request: %v", err)
				close(stop)
				return
			}
			_, err = w.Write(append(request, byte('\n')))
			if err != nil {
				log.Fatalf("unable to send submit request: %v", err)
			}
		}
	}()

	accepted := 0
	total := 0

	for {
		res, err := reader.ReadBytes('\n')
		if err != nil {
			close(stop)
			log.Printf("unable to read response: %v", err)
			return
		}
		asJob := StratumJobNotif{}
		err = json.Unmarshal(res, &asJob)
		if err != nil {
			log.Fatalf("unable to unmarshal response: %v", err)
		}

		if asJob.Method == "job" {
			in <- asJob.Params
		} else {
			asResponse := StratumSubmitResponse{}
			err = json.Unmarshal(res, &asResponse)
			if err != nil {
				log.Fatalf("unable to unmarshal response: %v", err)
			}
			if asResponse.Error != nil {
				total += 1
				log.Printf("Unaccepted share %v/%v: %v\n", accepted, total, asResponse.Error.Message)
			} else if asResponse.Result.Status == "OK" {
				accepted += 1
				total += 1
				log.Printf("Accepted share %v/%v\n", accepted, total)
			} else {
				log.Fatalf("Unrecognized message: %v", res)
			}
		}
	}
}
