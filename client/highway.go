package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"

	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/highway"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/crypto"
	"github.com/kernel-ai/koscore/utils/exception"
	"github.com/kernel-ai/koscore/utils/proto"

	hw "github.com/kernel-ai/koscore/client/internal/highway"
	pkt_hw "github.com/kernel-ai/koscore/client/packets/highway"
)

func (m *QQClient) initHighwayServers() {
	m.hw_session.Uin = &m.session.Info.Uin
	m.hw_session.AppId = m.version.AppId
	m.hw_session.SubAppId = m.version.SubAppId
}

func (m *QQClient) ensureHighwayServers() error {
	if m.hw_session.SsoAddr == nil || m.hw_session.SigSession == nil || m.hw_session.SessionKey == nil {
		pkt, err := m.sendOidbPacketAndWait(pkt_hw.BuildHighWayURLReq(m.session.Sig.A2))
		if err != nil {
			return exception.NewFormat("get highway server: %w", err)
		}
		rsp, err := pkt_hw.ParseHighWayURLReq(pkt.Data)
		if err != nil {
			return exception.NewFormat("parse highway server: %w", err)
		}

		m.hw_session.SigSession = rsp.RspBody.SigSession
		m.hw_session.SessionKey = rsp.RspBody.SessionKey

		for _, info := range rsp.RspBody.Addrs {
			if info.ServiceType.Unwrap() == 1 {
				for _, addr := range info.Addrs {
					m.LOGD("add highway server %s:%d", binary.UInt32ToIPV4Address(addr.Ip.Unwrap()), addr.Port)
					m.hw_session.AppendAddr(addr.Ip.Unwrap(), addr.Port.Unwrap())
				}
			}
		}
	}
	if m.hw_session.SsoAddr == nil || m.hw_session.SigSession == nil || m.hw_session.SessionKey == nil {
		return errors.New("empty highway servers")
	}
	return nil
}

func (m *QQClient) highwayUpload(common_id uint32, r io.Reader, file_size uint64, md5 []byte, extend_info []byte) error {
	// 能close的io就close
	defer utils.CloseIO(r)
	err := m.ensureHighwayServers()
	if err != nil {
		return err
	}
	trans := &hw.Transaction{
		CommandId: common_id,
		Body:      r,
		Sum:       md5,
		Size:      file_size,
		Ticket:    m.hw_session.SigSession,
		LoginSig:  m.session.Sig.A2,
		Ext:       extend_info,
	}
	if _, err = m.hw_session.Upload(trans); err == nil {
		return nil
	}
	// fallback to http upload
	servers := m.hw_session.SsoAddr
	saddr := servers[rand.Intn(len(servers))]
	server := fmt.Sprintf("http://%s:%d/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%d", binary.UInt32ToIPV4Address(saddr.IP), saddr.Port, m.session.Info.Uin)
	buffer := make([]byte, hw.BlockSize)
	for offset := uint64(0); offset < file_size; offset += hw.BlockSize {
		if hw.BlockSize > file_size-offset {
			buffer = buffer[:file_size-offset]
		}
		if _, err = io.ReadFull(r, buffer); err != nil {
			return err
		}
		if err = m.highwayUploadBlock(trans, server, offset, crypto.MD5Digest(buffer), buffer); err != nil {
			return err
		}
	}
	return nil
}

func (m *QQClient) highwayUploadBlock(trans *hw.Transaction, server string, offset uint64, blkmd5 []byte, blk []byte) error {
	blksz := uint64(len(blk))
	isEnd := offset+blksz == trans.Size
	payload, err := sendHighwayPacket(trans.Build(&m.hw_session, offset, uint32(blksz), blkmd5), blk, server, isEnd)
	if err != nil {
		return exception.NewFormat("send highway packet: %w", err)
	}
	defer payload.Close()
	if _, _, err = parseHighwayPacket(payload); err != nil {
		return err
	}
	//m.LOGD("highway block result: %d | %d | %x | %v", rsphead.ErrorCode, rsphead.MsgSegHead.RetCode.Unwrap(), rsphead.BytesRspExtendInfo, rspbody)
	return nil
}

func parseHighwayPacket(data io.Reader) (head *highway.RespDataHighwayHead, body *binary.Reader, e error) {
	body = binary.ParseReader(data)
	if body.ReadBytesNoCopy(1)[0] != 0x28 {
		return nil, nil, exception.New("parse highway packet: invalid highway packet")
	}
	size := body.ReadU32() // head length
	_ = body.ReadU32()     // body len
	d := body.ReadBytesNoCopy(int(int64(size)))
	if head, e = proto.Unmarshal[highway.RespDataHighwayHead](d); e != nil {
		return nil, nil, exception.NewUnmarshalProtoException(e, "parse highway packet")
	}
	if body.ReadBytesNoCopy(1)[0] != 0x29 {
		return nil, nil, exception.New("parse highway packet: invalid highway head")
	}
	if head.ErrorCode.Unwrap() != 0 {
		e = exception.NewOperationExceptionCode(head.ErrorCode.Unwrap(), "parse highway packet: code")
	}
	return
}

func sendHighwayPacket(packet *highway.ReqDataHighwayHead, block []byte, server_url string, end bool) (io.ReadCloser, error) {
	data, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}
	buf := hw.Frame(data, block)
	if data, err = io.ReadAll(&buf); err != nil {
		return nil, err
	}
	return postHighwayContent(bytes.NewReader(data), server_url, end)
	/*
	   return postHighwayContent(

	   	binary.NewBuilder(nil).
	   		WriteBytes([]byte{0x28}).
	   		WriteU32(uint32(len(marshal))).
	   		WriteU32(uint32(len(block))).
	   		WriteBytes(marshal).
	   		WriteBytes(block).
	   		WriteBytes([]byte{0x29}).
	   	ToReader(), serverURL, end)
	*/
}

func postHighwayContent(content io.Reader, server_url string, end bool) (io.ReadCloser, error) {
	// Parse server URL
	server, err := url.Parse(server_url)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", server.String(), content)
	if err != nil {
		return nil, err
	}

	// Set headers
	if end {
		req.Header.Set("Connection", "close")
	} else {
		req.Header.Set("Connection", "keep-alive")
	}

	// Send request
	rsp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return rsp.Body, nil
}
