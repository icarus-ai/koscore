package highway

// from https://github.com/Mrs4s/MiraiGo/tree/master/client/internal/highway/bdh.go

import (
	"crypto/md5"
	"fmt"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/fumiama/gofastTEA"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	"github.com/RomiChan/protobuf/proto"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/highway"
	"github.com/kernel-ai/koscore/utils/binary"
)

// 视频必须是1024*1024

const BlockSize = 1024 * 1024

type Transaction struct {
	CommandId uint32
	Body      io.Reader
	Sum       []byte // md5 sum of body
	Size      uint64 // body size
	Ticket    []byte
	LoginSig  []byte
	Ext       []byte
	Encrypt   bool
}

func (trans *Transaction) encrypt(key []byte) error {
	if !trans.Encrypt {
		return nil
	}
	if len(key) == 0 {
		return errors.New("session key not found. maybe miss some packet?")
	}
	trans.Ext = tea.NewTeaCipher(key).Encrypt(trans.Ext)
	return nil
}

func (trans *Transaction) Build(s *Session, offset uint64, length uint32, md5hash []byte) *highway.ReqDataHighwayHead {
	return &highway.ReqDataHighwayHead{
		MsgBaseHead: &highway.DataHighwayHead{
			Version:    proto.Some[uint32](1),
			Uin:        proto.Some(strconv.Itoa(int(*s.Uin))),
			Command:    proto.Some(_REQ_CMD_DATA),
			Seq:        proto.Some(s.NextSeq()),
			RetryTimes: proto.Some(uint32(0)),
			AppId:      proto.Some[uint32](s.SubAppId),
			DataFlag:   proto.Some[uint32](16),
			CommandId:  proto.Some[uint32](trans.CommandId),
			//LocaleId:  2052,
		},
		MsgSegHead: &highway.SegHead{
			ServiceId:     proto.Some(uint32(0)),
			Filesize:      proto.Some[uint64](trans.Size),
			DataOffset:    proto.Some(offset),
			DataLength:    proto.Some[uint32](length),
			RetCode:       proto.Some(uint32(0)),
			ServiceTicket: trans.Ticket,
			Md5:           md5hash,
			FileMd5:       trans.Sum,
			CacheAddr:     proto.Some(uint32(0)),
			CachePort:     proto.Some(uint32(0)),
		},
		ReqExtendInfo: trans.Ext,
		MsgLoginSigHead: &highway.LoginSigHead{
			LoginSigType: proto.Some[uint32](8),
			LoginSig:     trans.LoginSig,
			AppId:        proto.Some[uint32](s.AppId),
		},
	}
}

func (s *Session) uploadSingle(trans *Transaction) ([]byte, error) {
	pc, err := s.selectConn()
	if err != nil {
		return nil, err
	}
	defer s.putIdleConn(pc)

	reader := binary.NewNetworkReader(pc.conn)
	var rspExt []byte
	offset := 0
	chunk := make([]byte, BlockSize)
	for {
		chunk = chunk[:cap(chunk)]
		rl, err := io.ReadFull(trans.Body, chunk)
		if rl == 0 {
			break
		}
		if errors.Is(err, io.ErrUnexpectedEOF) {
			chunk = chunk[:rl]
		}
		ch := md5.Sum(chunk)
		head, _ := proto.Marshal(trans.Build(s, uint64(offset), uint32(rl), ch[:]))
		offset += rl
		buffers := Frame(head, chunk)
		_, err = buffers.WriteTo(pc.conn)
		if err != nil {
			return nil, errors.Wrap(err, "write conn error")
		}
		rspHead, err := readResponse(reader)
		if err != nil {
			return nil, errors.Wrap(err, "highway upload error")
		}
		if rspHead.ErrorCode.Unwrap() != 0 {
			return nil, fmt.Errorf("upload failed: %d", rspHead.ErrorCode.Unwrap())
		}
		if rspHead.RspExtendInfo != nil {
			rspExt = rspHead.RspExtendInfo
		}
		if rspHead.MsgSegHead != nil && rspHead.MsgSegHead.ServiceTicket != nil {
			trans.Ticket = rspHead.MsgSegHead.ServiceTicket
		}
	}
	return rspExt, nil
}

func (s *Session) Upload(trans *Transaction) ([]byte, error) {
	// encrypt ext data
	if err := trans.encrypt(s.SessionKey); err != nil {
		return nil, err
	}

	const maxThreadCount = 4
	threadCount := int(trans.Size) / (6 * BlockSize) // 1 thread upload 1.5 MB
	if threadCount > maxThreadCount {
		threadCount = maxThreadCount
	}
	if threadCount < 2 {
		// single thread upload
		return s.uploadSingle(trans)
	}

	// pick a address
	// TODO: pick smarter
	pc, err := s.selectConn()
	if err != nil {
		return nil, err
	}
	addr := pc.addr
	s.putIdleConn(pc)

	var (
		rspExt          []byte
		completedThread uint32
		cond            = sync.NewCond(&sync.Mutex{})
		offset          = uint64(0)
		count           = (trans.Size + BlockSize - 1) / BlockSize
		id              = 0
	)
	doUpload := func() error {
		// send signal complete uploading
		defer func() {
			atomic.AddUint32(&completedThread, 1)
			cond.Signal()
		}()

		// todo: get from pool?
		pc, err := s.connect(addr)
		if err != nil {
			return err
		}
		defer s.putIdleConn(pc)

		reader := binary.NewNetworkReader(pc.conn)
		chunk := make([]byte, BlockSize)
		for {
			cond.L.Lock() // lock protect reading
			off := offset
			offset += BlockSize
			id++
			last := uint64(id) == count
			if last { // last
				for atomic.LoadUint32(&completedThread) != uint32(threadCount-1) {
					cond.Wait()
				}
			} else if uint64(id) > count {
				cond.L.Unlock()
				break
			}
			chunk = chunk[:BlockSize]
			n, err := io.ReadFull(trans.Body, chunk)
			cond.L.Unlock()

			if n == 0 {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				chunk = chunk[:n]
			}
			ch := md5.Sum(chunk)
			head, _ := proto.Marshal(trans.Build(s, off, uint32(n), ch[:]))
			buffers := Frame(head, chunk)
			_, err = buffers.WriteTo(pc.conn)
			if err != nil {
				return errors.Wrap(err, "write conn error")
			}
			rspHead, err := readResponse(reader)
			if err != nil {
				return errors.Wrap(err, "highway upload error")
			}
			if rspHead.ErrorCode.Unwrap() != 0 {
				return fmt.Errorf("upload failed: %d", rspHead.ErrorCode.Unwrap())
			}
			if last && rspHead.RspExtendInfo != nil {
				rspExt = rspHead.RspExtendInfo
			}
		}
		return nil
	}

	group := errgroup.Group{}
	for i := 0; i < threadCount; i++ {
		group.Go(doUpload)
	}
	return rspExt, group.Wait()
}
