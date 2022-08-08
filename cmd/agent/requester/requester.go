package requester

import (
	"bytes"
	"log"
	"net/http"
)

type Requester struct {
	address string
}

func (r *Requester) Send(a []byte) error {
	log.Println("requester::Send: started", string(a))
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, "http://"+r.address+"/update/", bytes.NewBuffer(a))
	request.Close = true
	if err != nil {
		log.Panicln("requester::Send: can't create request with", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Println("requester::Send: error in request execution:", err)
		return nil
	}
	defer response.Body.Close()
	return nil
}

func New(address string) *Requester {
	return &Requester{address: address}
}
