package sdk

import (
	"context"
	"errors"
	"fmt"
	"github.com/kaytu-io/kaytu/pkg/plugin/proto/src/golang"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"sync/atomic"
	"time"
)

type StreamController struct {
	stream golang.Plugin_RegisterClient

	sendChan      chan *golang.PluginMessage
	sendClosed    atomic.Bool
	receiveChan   chan *golang.ServerMessage
	receiveClosed atomic.Bool
}

func NewStreamController(stream golang.Plugin_RegisterClient) *StreamController {
	controller := StreamController{
		stream:        stream,
		sendChan:      make(chan *golang.PluginMessage, 10000),
		sendClosed:    atomic.Bool{},
		receiveChan:   make(chan *golang.ServerMessage, 10000),
		receiveClosed: atomic.Bool{},
	}
	controller.sendClosed.Store(false)
	controller.receiveClosed.Store(false)

	return &controller
}

func (s *StreamController) startReceiver(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("StreamController receiver panic: %v", r)
			time.Sleep(1 * time.Second)
			go s.startReceiver(ctx)
		}
	}()

	for {
		if err := ctx.Err(); err != nil {
			log.Printf("stream controller context error: %v", err)
			return
		}
		msg, err := s.stream.Recv()
		if err != nil && !errors.Is(err, io.EOF) {
			grpcStatus, ok := status.FromError(err)
			if ok && grpcStatus.Code() == codes.Unavailable && grpcStatus.Message() == "error reading from server: EOF" {
				s.CloseRecv()
				return
			}
			log.Printf("receive error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if s.receiveClosed.Load() {
			return
		}
		s.receiveChan <- msg
	}
}

func (s *StreamController) startSender(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("stream controller sender panic: %v", r)
			time.Sleep(1 * time.Second)
			go s.startSender(ctx)
		}
	}()

	for {
		if len(s.sendChan) == 0 && s.sendClosed.Load() {
			return
		}
		if err := ctx.Err(); len(s.sendChan) == 0 && err != nil {
			log.Printf("stream controller context error: %v", err)
			s.CloseSend()
			return
		}
		msg, ok := <-s.sendChan
		if !ok {
			return
		}
		err := s.stream.Send(msg)
		if err != nil && !errors.Is(err, io.EOF) {
			log.Printf("send error: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}
	}
}

func (s *StreamController) Start(ctx context.Context) {
	go s.startReceiver(ctx)
	go s.startSender(ctx)
}

func (s *StreamController) Send(msg *golang.PluginMessage) (err error) {
	defer func() {
		if r := recover(); r != nil {
			s.sendClosed.Store(true)
			err = fmt.Errorf("send is closed: %v", r)
		}
	}()
	s.sendChan <- msg
	return nil
}

func (s *StreamController) Recv(ctx context.Context) (*golang.ServerMessage, error) {
	select {
	case msg, ok := <-s.receiveChan:
		if !ok {
			return nil, fmt.Errorf("receive is closed")
		}
		return msg, nil
	case <-ctx.Done():
		s.receiveClosed.Store(true)
		return nil, fmt.Errorf("receive is closed")
	}
}

func (s *StreamController) CloseSend() {
	s.sendClosed.Store(true)
	close(s.sendChan)
}

func (s *StreamController) CloseRecv() {
	s.receiveClosed.Store(true)
	close(s.receiveChan)
}
