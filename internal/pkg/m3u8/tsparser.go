package m3u8

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/nicoxiang/geektime-downloader/internal/pkg/crypto"
)

const (
	PacketLength               = 188
	SyncByte              byte = 0x47
	PayloadStartMask      byte = 0x40
	AtfMask               byte = 0x30
	AtfReserve            byte = 0x00
	AtfPayloadOnly        byte = 0x01
	AtfFieldOnly          byte = 0x02
	AtfFiledFollowPayload byte = 0x03
)

type TSParser struct {
	stream *TSStream
}

type TSPesFragment struct {
	packets []*TSPacket
}

type TSStream struct {
	data    []byte
	key     []byte
	packets []*TSPacket
	videos  []*TSPesFragment
	audios  []*TSPesFragment
}

type TSHeader struct {
	syncByte                   byte //8
	transportErrorIndicator    byte //1
	payloadUnitStartIndicator  byte //1
	transportPriority          byte //1
	pid                        int  //13
	transportScramblingControl byte //2
	adaptationFiled            byte //2
	continuityCounter          byte //4
	hasError                   bool
	isPayloadStart             bool
	hasAdaptationFieldField    bool
	hasPayload                 bool
}

type TSPacket struct {
	header                TSHeader
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

func NewTSPacket() *TSPacket {
	return &TSPacket{
		headerLength: 4,
	}
}

func NewTSParser(data []byte, key string) *TSParser {
	hexKey, err := hex.DecodeString(key)
	if err != nil {
		log.Fatalln(err)
	}
	stream := &TSStream{
		data: data,
		key:  hexKey,
	}
	stream.parseTs()
	return &TSParser{
		stream: stream,
	}
}

func (p *TSParser) Decrypt() []byte {
	p.decryptPES(p.stream.data, p.stream.videos, p.stream.key)
	p.decryptPES(p.stream.data, p.stream.audios, p.stream.key)
	return p.stream.data
}

func (p *TSParser) decryptPES(byteBuf []byte, pesFragments []*TSPesFragment, key []byte) {
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
			buffer.Read(byteBuf[payloadStartOffset : payloadStartOffset+payloadLength])
		}
	}
}

func (pes *TSPesFragment) Add(packet *TSPacket) {
	pes.packets = append(pes.packets, packet)
}

func (stream *TSStream) parseTs() {
	byteBuf := bytes.NewReader(stream.data)
	length := byteBuf.Len()
	if length%PacketLength != 0 {
		panic("not a ts package")
	}
	var pes *TSPesFragment
	packNums := length / PacketLength
	for packageNo := 0; packageNo < packNums; packageNo++ {
		buffer := make([]byte, PacketLength)
		byteBuf.Read(buffer)
		packet := stream.parseTSPacket(buffer, packageNo, packageNo*PacketLength)
		switch packet.header.pid {
		// video data
		case 0x100:
			if packet.header.isPayloadStart {
				if nil != pes {
					stream.videos = append(stream.videos, pes)
				}
				pes = new(TSPesFragment)
			}
			pes.Add(packet)
		//audio data
		case 0x101:
			if packet.header.isPayloadStart {
				if nil != pes {
					stream.audios = append(stream.audios, pes)
				}
				pes = new(TSPesFragment)
			}
			pes.Add(packet)
		}
		stream.packets = append(stream.packets, packet)
	}
}

func (stream *TSStream) parseTSPacket(buffer []byte, packNo, offset int) *TSPacket {
	if buffer[0] != SyncByte {
		panic(fmt.Sprintf("Invalid ts package in :%d offset: %d", packNo, offset))
	}
	header := TSHeader{}
	header.syncByte = buffer[0]
	if buffer[1]&0x80 > 0 {
		header.transportErrorIndicator = 1
	}
	if buffer[1]&PayloadStartMask > 0 {
		header.payloadUnitStartIndicator = 1
	}
	if buffer[1]&0x20 > 0 {
		header.transportErrorIndicator = 1
	}
	header.pid = int(buffer[1]&0x1F)<<8 | int(buffer[2]&0xFF)
	header.transportScramblingControl = ((buffer[3] & 0xC0) >> 6) & 0xFF
	header.adaptationFiled = ((buffer[3] & AtfMask) >> 4) & 0xFF
	header.continuityCounter = (buffer[3] & 0x0F) & 0xFF
	header.hasError = header.transportErrorIndicator != 0
	header.isPayloadStart = header.payloadUnitStartIndicator != 0
	header.hasAdaptationFieldField = header.adaptationFiled == AtfFieldOnly || header.adaptationFiled == AtfFiledFollowPayload
	header.hasPayload = header.adaptationFiled == AtfPayloadOnly || header.adaptationFiled == AtfFiledFollowPayload
	packet := NewTSPacket()
	packet.header = header
	packet.packNo = packNo
	packet.startOffset = offset
	if header.hasAdaptationFieldField {
		atfLength := buffer[4] & 0xFF
		packet.headerLength += 1
		packet.atfLength = int(atfLength)
	}
	if header.isPayloadStart {
		packet.pesOffset = packet.startOffset + packet.headerLength + packet.atfLength
		// 9 bytes : 6 bytes for PES header + 3 bytes for PES extension
		packet.pesHeaderLength = int(6 + 3 + buffer[packet.headerLength+packet.atfLength+8]&0xFF)
	}
	packet.payloadRelativeOffset = packet.headerLength + packet.atfLength + packet.pesHeaderLength
	packet.payloadStartOffset = int(packet.startOffset + packet.payloadRelativeOffset)
	packet.payloadLength = PacketLength - packet.payloadRelativeOffset
	//log.Printf("%+v", packet)
	if packet.payloadLength > 0 {
		packet.payload = buffer[packet.payloadRelativeOffset:PacketLength]
	}
	return packet
}
