package requester

import (
	"bytes"
	"log"
	"net/http"
)

type Requester struct {
	address string
}

func New(address string) *Requester {
	return &Requester{address: address}
}

func (r *Requester) Send(a []byte) error {
	log.Println("requester::Send::info: started", string(a))
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, "http://"+r.address+"/update/", bytes.NewBuffer(a))
	request.Close = true
	if err != nil {
		log.Println("requester::Send::error: can't create request with:", err)
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Println("requester::Send::error: can't do request:", err)
		return err
	}
	defer response.Body.Close()
	return nil
}

func (r *Requester) SendSeveral(a []byte) error {
	log.Println("requester::SendSeveral::info: started", string(a))
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, "http://"+r.address+"/updates/", bytes.NewBuffer(a))
	request.Close = true
	if err != nil {
		log.Println("requester::SendSeveral::error: can't create request with:", err)
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Println("requester::SendSeveral::error: can't do request:", err)
		return nil
	}
	defer response.Body.Close()
	return nil
}
