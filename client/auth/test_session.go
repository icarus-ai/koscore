package auth

import (
	"bytes"
	"fmt"
)

type test_session_t struct {
	byt *bytes.Buffer
}

func (m *test_session_t) fn_hex(tag string, data []byte) {
	if len(data) == 0 {
		m.byt.WriteString(fmt.Sprintf("\n  %s: no date", tag))
	} else {
		m.byt.WriteString(fmt.Sprintf("\n  %s: %02X", tag, data))
	}
}

func (m *test_session_t) fn_str(tag, data string) {
	if len(data) == 0 {
		m.byt.WriteString(fmt.Sprintf("\n  %s", tag))
	} else {
		m.byt.WriteString(fmt.Sprintf("\n  %s: %s", tag, data))
	}
}

func (m *Session) ToStr() string {
	p := &test_session_t{byt: new(bytes.Buffer)}

	p.byt.WriteString("\nSig: WLoginSigs")
	p.fn_hex("A1", m.Sig.A1)
	p.fn_hex("A1Key", m.Sig.A1Key)
	p.fn_hex("A2", m.Sig.A2)
	p.fn_hex("A2Key", m.Sig.A2Key)
	p.fn_hex("D2", m.Sig.D2)
	p.fn_hex("D2Key", m.Sig.D2Key)
	p.fn_hex("NoPicSig", m.Sig.NoPicSig)
	p.fn_hex("TgtgtKey", m.Sig.TgtgtKey)
	p.fn_hex("Ksid", m.Sig.Ksid)
	p.fn_hex("SuperKey", m.Sig.SuperKey)
	p.fn_hex("St", m.Sig.St)
	p.fn_hex("StKey", m.Sig.StKey)
	p.fn_hex("StWeb", m.Sig.StWeb)
	p.fn_hex("WtSessionTicket", m.Sig.WtSessionTicket)
	p.fn_hex("WtSessionTicketKey", m.Sig.WtSessionTicketKey)
	p.fn_hex("RandomKey", m.Sig.RandomKey)
	p.fn_hex("SKey", m.Sig.SKey)
	p.byt.WriteString("\n  PsKey")
	if len(m.Sig.PsKey) == 0 {
		p.byt.WriteString(": no date")
	} else {
		for k, v := range m.Sig.PsKey {
			p.byt.WriteString(fmt.Sprintf("\n    %s : %s", k, v))
		}
	}

	p.byt.WriteString("\nState")
	p.fn_hex("Tlv104", m.State.Tlv104)
	p.fn_hex("Tlv547", m.State.Tlv547)
	p.fn_hex("Tlv174", m.State.Tlv174)
	p.fn_str("Cookie", m.State.Cookie)
	p.fn_hex("QrSig", m.State.QrSig)
	p.byt.WriteString("\n  KeyExchange")
	if m.State.KeyExchange == nil {
		p.byt.WriteString(": no date")
	} else {
		p.fn_hex("  SessionKey", m.State.KeyExchange.SessionKey)
		p.fn_hex("  SessionTicket", m.State.KeyExchange.SessionTicket)
	}

	return p.byt.String()[1:]
}
