package m3u8

// copy from
// https://github.com/SweetInk/lagou-course-downloader/blob/master/src/main/java/online/githuboy/lagou/course/decrypt/alibaba/TSParser.java
// https://github.com/lbbniu/aliyun-m3u8-downloader/blob/main/pkg/parse/aliyun/tsparser.go

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/crypto"
)

const (
	packetLength               = 188
	syncByte              byte = 0x47
	payloadStartMask      byte = 0x40
	atfMask               byte = 0x30
	atfReserve            byte = 0x00
	atfPayloadOnly        byte = 0x01
	atfFieldOnly          byte = 0x02
	atfFiledFollowPayload byte = 0x03
)

// TSParser ...
type TSParser struct {
	stream *tsStream
}

type tsPesFragment struct {
	packets []*tsPacket
}

type tsStream struct {
	data    []byte
	key     []byte
	packets []*tsPacket
	videos  []*tsPesFragment
	audios  []*tsPesFragment
}

type tsHeader struct {
	syncByte                   byte //8
	transportErrorIndicator    byte //1
	payloadUnitStartIndicator  byte //1
	pid                        int  //13
	transportScramblingControl byte //2
	adaptationFiled            byte //2
	continuityCounter          byte //4
	hasError                   bool
	isPayloadStart             bool
	hasAdaptationFieldField    bool
	hasPayload                 bool
}

type tsPacket struct {
	header                tsHeader
	packNo                int
	startOffset           int
	headerLength          int // 4
	atfLength             int
	pesOffset             int
	pesHeaderLength       int
	payloadStartOffset    int
	payloadRelativeOffset int // 0
	payloadLength         int // 0
	payload               []byte
}

func newTSPacket() *tsPacket {
	return &tsPacket{
		headerLength: 4,
	}
}

// NewTSParser ...
func NewTSParser(data []byte, key string) *TSParser {
	hexKey, _ := hex.DecodeString(key)
	stream := &tsStream{
		data: data,
		key:  hexKey,
	}
	stream.parseTS()
	return &TSParser{
		stream: stream,
	}
}

// Decrypt ...
func (p *TSParser) Decrypt() []byte {
	p.decryptPES(p.stream.data, p.stream.videos, p.stream.key)
	p.decryptPES(p.stream.data, p.stream.audios, p.stream.key)
	return p.stream.data
}

func (p *TSParser) decryptPES(byteBuf []byte, pesFragments []*tsPesFragment, key []byte) {
	for _, pes := range pesFragments {
		buffer := &bytes.Buffer{}
		for _, packet := range pes.packets {
			if nil == packet.payload {
				panic("payload is null")
			}
			buffer.Write(packet.payload)
		}
		length := buffer.Len()
		all := buffer.Bytes()
		buffer.Reset()
		if length%16 > 0 {
			newLength := 16 * (length / 16)
			decrypt := crypto.AESDecryptECB(all[:newLength], key)
			buffer.Write(decrypt)
			buffer.Write(all[newLength:])
		} else {
			decrypt := crypto.AESDecryptECB(all, key)
			buffer.Write(decrypt)
		}
		//Rewrite decrypted bytes to byteBuf
		for _, packet := range pes.packets {
			payloadLength := packet.payloadLength
			payloadStartOffset := packet.payloadStartOffset
			_, _ = buffer.Read(byteBuf[payloadStartOffset : payloadStartOffset+payloadLength])
		}
	}
}

func (pes *tsPesFragment) add(packet *tsPacket) {
	pes.packets = append(pes.packets, packet)
}

func (stream *tsStream) parseTS() {
	byteBuf := bytes.NewReader(stream.data)
	length := byteBuf.Len()
	if length%packetLength != 0 {
		panic("not a ts package")
	}
	var pes *tsPesFragment
	packNums := length / packetLength
	for packageNo := 0; packageNo < packNums; packageNo++ {
		buffer := make([]byte, packetLength)
		_, _ = byteBuf.Read(buffer)
		packet := stream.parseTSPacket(buffer, packageNo, packageNo*packetLength)
		switch packet.header.pid {
		// video data
		case 0x100:
			if packet.header.isPayloadStart {
				if nil != pes {
					stream.videos = append(stream.videos, pes)
				}
				pes = new(tsPesFragment)
			}
			pes.add(packet)
		//audio data
		case 0x101:
			if packet.header.isPayloadStart {
				if nil != pes {
					stream.audios = append(stream.audios, pes)
				}
				pes = new(tsPesFragment)
			}
			pes.add(packet)
		}
		stream.packets = append(stream.packets, packet)
	}
}

func (stream *tsStream) parseTSPacket(buffer []byte, packNo, offset int) *tsPacket {
	if buffer[0] != syncByte {
		panic(fmt.Sprintf("Invalid ts package in :%d offset: %d", packNo, offset))
	}
	header := tsHeader{}
	header.syncByte = buffer[0]
	if buffer[1]&0x80 > 0 {
		header.transportErrorIndicator = 1
	}
	if buffer[1]&payloadStartMask > 0 {
		header.payloadUnitStartIndicator = 1
	}
	if buffer[1]&0x20 > 0 {
		header.transportErrorIndicator = 1
	}
	header.pid = int(buffer[1]&0x1F)<<8 | int(buffer[2]&0xFF)
	header.transportScramblingControl = ((buffer[3] & 0xC0) >> 6) & 0xFF
	header.adaptationFiled = ((buffer[3] & atfMask) >> 4) & 0xFF
	header.continuityCounter = (buffer[3] & 0x0F) & 0xFF
	header.hasError = header.transportErrorIndicator != 0
	header.isPayloadStart = header.payloadUnitStartIndicator != 0
	header.hasAdaptationFieldField = header.adaptationFiled == atfFieldOnly || header.adaptationFiled == atfFiledFollowPayload
	header.hasPayload = header.adaptationFiled == atfPayloadOnly || header.adaptationFiled == atfFiledFollowPayload
	packet := newTSPacket()
	packet.header = header
	packet.packNo = packNo
	packet.startOffset = offset
	if header.hasAdaptationFieldField {
		atfLength := buffer[4] & 0xFF
		packet.headerLength++
		packet.atfLength = int(atfLength)
	}
	if header.isPayloadStart {
		packet.pesOffset = packet.startOffset + packet.headerLength + packet.atfLength
		// 9 bytes : 6 bytes for PES header + 3 bytes for PES extension
		packet.pesHeaderLength = int(6 + 3 + buffer[packet.headerLength+packet.atfLength+8]&0xFF)
	}
	packet.payloadRelativeOffset = packet.headerLength + packet.atfLength + packet.pesHeaderLength
	packet.payloadStartOffset = int(packet.startOffset + packet.payloadRelativeOffset)
	packet.payloadLength = packetLength - packet.payloadRelativeOffset
	if packet.payloadLength > 0 {
		packet.payload = buffer[packet.payloadRelativeOffset:packetLength]
	}
	return packet
}
