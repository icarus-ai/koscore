package client

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"

	"github.com/RomiChan/protobuf/proto"
	"github.com/kernel-ai/koscore/client/packets/pb/v2/service/highway"
	"github.com/kernel-ai/koscore/utils"
	"github.com/kernel-ai/koscore/utils/binary"
	"github.com/kernel-ai/koscore/utils/crypto"

	hw "github.com/kernel-ai/koscore/client/internal/highway"
	pkt_hw "github.com/kernel-ai/koscore/client/packets/highway"
)

//var m_hw_session *hw.Session
/*
func (c *QQClient) initHighwayServers() {
	c.hw_session.Uin      = &c.transport.Sig.Uin
	c.hw_session.AppId    = uint32(c.transport.Version.AppId)
	c.hw_session.SubAppId = uint32(c.transport.Version.SubAppId)
}
*/

func (m *QQClient) ensureHighwayServers() error {
	if m.hw_session.SsoAddr == nil || m.hw_session.SigSession == nil || m.hw_session.SessionKey == nil {
		pkt, err := m.sendOidbPacketAndWait(pkt_hw.BuildHighWayURLReq(m.session.Sig.A2))
		if err != nil {
			return fmt.Errorf("get highway server: %w", err)
		}
		rsp, err := pkt_hw.ParseHighWayURLReq(pkt.Data)
		if err != nil {
			return fmt.Errorf("parse highway server: %w", err)
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

func (m *QQClient) highwayUpload(commonId int, r io.Reader, fileSize uint64, md5 []byte, extendInfo []byte) error {
	// 能close的io就close
	defer utils.CloseIO(r)
	err := m.ensureHighwayServers()
	if err != nil {
		return err
	}
	trans := &hw.Transaction{
		CommandId: uint32(commonId),
		Body:      r,
		Sum:       md5,
		Size:      fileSize,
		Ticket:    m.hw_session.SigSession,
		LoginSig:  m.session.Sig.A2,
		Ext:       extendInfo,
	}
	if _, err = m.hw_session.Upload(trans); err == nil {
		return nil
	}
	// fallback to http upload
	servers := m.hw_session.SsoAddr
	saddr := servers[rand.Intn(len(servers))]
	server := fmt.Sprintf("http://%s:%d/cgi-bin/httpconn?htcmd=0x6FF0087&uin=%d", binary.UInt32ToIPV4Address(saddr.IP), saddr.Port, m.Uin())
	buffer := make([]byte, hw.BlockSize)
	for offset := uint64(0); offset < fileSize; offset += hw.BlockSize {
		if hw.BlockSize > fileSize-offset {
			buffer = buffer[:fileSize-offset]
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
		return fmt.Errorf("send highway packet: %w", err)
	}
	defer payload.Close()
	rsphead, _, err := parseHighwayPacket(payload)
	if err != nil {
		return fmt.Errorf("parse highway packet: %w", err)
	}
	//m.LOGD("Highway Block Result: %d | %d | %x | %v", rsphead.ErrorCode, rsphead.MsgSegHead.RetCode.Unwrap(), rsphead.BytesRspExtendInfo, rspbody)
	if rsphead.ErrorCode.Unwrap() != 0 {
		return errors.New("highway error code: " + strconv.Itoa(int(rsphead.ErrorCode.Unwrap())))
	}
	return nil
}

func parseHighwayPacket(data io.Reader) (head *highway.RespDataHighwayHead, body *binary.Reader, e error) {
	body = binary.ParseReader(data)
	if body.ReadBytesNoCopy(1)[0] != 0x28 {
		return nil, nil, errors.New("invalid highway packet")
	}
	size := body.ReadU32() // head length
	_ = body.ReadU32()     // body len
	head = &highway.RespDataHighwayHead{}
	d := body.ReadBytesNoCopy(int(int64(size)))
	if e = proto.Unmarshal(d, head); e != nil {
		return nil, nil, e
	}
	if body.ReadBytesNoCopy(1)[0] != 0x29 {
		return nil, nil, errors.New("invalid highway head")
	}
	return
}

func sendHighwayPacket(packet *highway.ReqDataHighwayHead, block []byte, serverURL string, end bool) (io.ReadCloser, error) {
	data, err := proto.Marshal(packet)
	if err != nil {
		return nil, err
	}
	buf := hw.Frame(data, block)
	if data, err = io.ReadAll(&buf); err != nil {
		return nil, err
	}
	return postHighwayContent(bytes.NewReader(data), serverURL, end)
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

func postHighwayContent(content io.Reader, serverURL string, end bool) (io.ReadCloser, error) {
	// Parse server URL
	server, err := url.Parse(serverURL)
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
