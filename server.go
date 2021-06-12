package jurpc

import (
	"encoding/json"
	"fmt"
	"io"
	"jurpc/codec"
	"jurpc/request"
	"log"
	"net"
	"reflect"
	"sync"
)

const MagicNumber = 0x3bef5c

type GRpcOption struct {
	MagicNumber int
	CodecType   codec.Type
}

var DefaultOption = &GRpcOption{
	MagicNumber: MagicNumber,
	CodecType:   codec.GobType,
}

type Server struct{}

func NewServer() *Server {
	return &Server{}
}

var DefaultServer = NewServer()

func (svc *Server) Accept(listener net.Listener) {
	for {
		connect, err := listener.Accept()
		if err != nil {
			log.Println("rpc server accept error:", err)
		}
		go svc.ServerConnect(connect)
	}
}

func Accept(listener net.Listener) {
	DefaultServer.Accept(listener)
}

func (svc *Server) ServerConnect(connect io.ReadWriteCloser) {
	defer func() { _ = connect.Close() }()

	var opt GRpcOption
	if err := json.NewDecoder(connect).Decode(&opt); err != nil {
		log.Println("rpc server option error", err)
		return
	}

	if opt.MagicNumber != MagicNumber {
		log.Println("rpc server invalid magic number %i", opt.MagicNumber)
		return
	}

	f := codec.NewCodecFuncMap[opt.CodecType]
	if f == nil {
		log.Println("rpc server invalid code type %s", opt.CodecType)
		return
	}
	svc.serverCodec(f(connect))
}

var invalidRequest = struct{}{}

func (svc *Server) serverCodec(cc codec.Codec) {
	sending := new(sync.Mutex)
	wg := new(sync.WaitGroup)

	for {
		req, err := svc.readRequest(cc)
		if err != nil {
			if req == nil {
				break
			}

			req.Header.Error = err.Error()
			svc.sendResponse(cc, req.Header, invalidRequest, sending)
			continue
		}
		wg.Add(1)
		go svc.handleRequest(cc, req, sending, wg)
	}

	wg.Wait()
	_ = cc.Close()
}

func (svc *Server) readRequestHeader(cc codec.Codec) (*codec.Header, error) {
	var res codec.Header
	if err := cc.ReadHeader(&res); err != nil {
		if err != io.EOF && err != io.ErrUnexpectedEOF {
			log.Println("rpc server read header error", err, res)
		}
		return nil, err
	}
	return &res, nil
}

func (svc *Server) readRequest(cc codec.Codec) (*request.Request, error) {
	h, err := svc.readRequestHeader(cc)
	if err != nil {
		return nil, err
	}

	req := &request.Request{
		Header: h,
	}

	req.Argv = reflect.New(reflect.TypeOf(""))

	if err = cc.ReadBody(req.Argv.Interface()); err != nil {
		log.Println("rpc server: read argv err:", err)
	}

	return req, nil
}

func (svc *Server) sendResponse(cc codec.Codec, h *codec.Header, body interface{}, sending *sync.Mutex) {
	sending.Lock()
	defer sending.Unlock()
	if err := cc.Write(h, body); err != nil {
		log.Println("rpc servier write response error", err)
	}
}

func (svc *Server) handleRequest(cc codec.Codec, req *request.Request, sending *sync.Mutex, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Println(req.Header, req.Argv.Elem())

	req.Replyv = reflect.ValueOf(fmt.Sprintf("jurpc response %d", req.Header.Seq))
	svc.sendResponse(cc, req.Header, req.Replyv.Interface(), sending)
}
