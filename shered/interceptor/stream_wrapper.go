package interceptor

import (
	"context"
	"log"

	"google.golang.org/grpc"
)

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func NewWrappedServerStream(ss grpc.ServerStream, ctx context.Context) *wrappedServerStream {
	return &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

//НУЖНО ЛИ НАМ ПРОПИСЫВАТЬ RecvMsg и SendMsg? нужна ли нам инфа в логах?
func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)
	if err == nil {
		log.Printf("[Wrapper Recv] Перехвачено сообщение: %+v", m)
	}
	return err
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	log.Printf("[Wrapper Send] Перехвачено сообщение перед отправкой: %+v", m)
	return w.ServerStream.SendMsg(m)
}