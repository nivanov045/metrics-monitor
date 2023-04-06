package requester

import (
	"bytes"
	"net/http"

	"github.com/rs/zerolog/log"
)

type Requester struct {
	address string
}

func New(address string) *Requester {
	return &Requester{address: address}
}

func (r *Requester) Send(a []byte) error {
	log.Debug().Interface("string", string(a)).Msg("started send of ")
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, "http://"+r.address+"/update/", bytes.NewBuffer(a))
	request.Close = true
	if err != nil {
		log.Error().Err(err).Stack()
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(err).Stack()
		return err
	}

	defer response.Body.Close()
	return nil
}

func (r *Requester) SendSeveral(a []byte) error {
	log.Debug().Interface("string", string(a)).Msg("started send of several")
	client := &http.Client{}
	request, err := http.NewRequest(http.MethodPost, "http://"+r.address+"/updates/", bytes.NewBuffer(a))
	request.Close = true
	if err != nil {
		log.Error().Err(err).Stack()
		return err
	}

	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		log.Error().Err(err).Stack()
		return nil
	}

	defer response.Body.Close()
	return nil
}
