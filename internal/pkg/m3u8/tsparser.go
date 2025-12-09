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
	atfPayloadOnly        byte = 0x01
	atfFieldOnly          byte = 0x02
	atfFieldFollowPayload byte = 0x03
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
	syncByte                   byte // 8
	transportErrorIndicator    byte // 1
	payloadUnitStartIndicator  byte // 1
	pid                        int  // 13
	transportScramblingControl byte // 2
	adaptationField            byte // 2
	continuityCounter          byte // 4
	hasError                   bool
	isPayloadStart             bool
	hasAdaptationFieldField    bool
	hasPayload                 bool
}

type tsPacket struct {
	header             tsHeader
	packNo             int
	startOffset        int
	headerLength       int // 4
	atfLength          int
	pesOffset          int
	pesHeaderLength    int
	payloadStartOffset int
	payloadLength      int // 0
	payload            []byte
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

// Decrypt decrypt video and audio PES
func (p *TSParser) Decrypt() []byte {
	p.decryptPES(p.stream.data, p.stream.videos, p.stream.key)
	p.decryptPES(p.stream.data, p.stream.audios, p.stream.key)
	return p.stream.data
}

func (p *TSParser) decryptPES(byteBuf []byte, pesFragments []*tsPesFragment, key []byte) {
	for _, pes := range pesFragments {
		buffer := &bytes.Buffer{}
		for _, packet := range pes.packets {
			if len(packet.payload) == 0 {
				continue
			}
			buffer.Write(packet.payload)
		}

		all := buffer.Bytes()
		length := len(all)
		if length == 0 {
			continue
		}

		buffer.Reset()
		decryptLen := (length / 16) * 16
		if decryptLen > 0 {
			decrypted := crypto.AESDecryptECB(all[:decryptLen], key)
			buffer.Write(decrypted)
		}
		if decryptLen < length {
			buffer.Write(all[decryptLen:])
		}

		// Rewrite decrypted bytes to byteBuf
		bufReader := bytes.NewReader(buffer.Bytes())
		for _, packet := range pes.packets {
			payloadLen := len(packet.payload)
			if payloadLen == 0 {
				continue
			}
			n, err := bufReader.Read(byteBuf[packet.payloadStartOffset : packet.payloadStartOffset+payloadLen])
			if err != nil || n != payloadLen {
				panic(fmt.Sprintf("decrypt write back failed, expected %d got %d", payloadLen, n))
			}
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
		panic("TS data length not multiple of 188")
	}

	var pesVideo, pesAudio *tsPesFragment
	numPackets := length / packetLength

	for packNo := 0; packNo < numPackets; packNo++ {
		buffer := make([]byte, packetLength)
		_, _ = byteBuf.Read(buffer)
		packet := stream.parseTSPacket(buffer, packNo, packNo*packetLength)

		switch packet.header.pid {
		case 0x100: // video
			if packet.header.isPayloadStart {
				if pesVideo != nil {
					stream.videos = append(stream.videos, pesVideo)
				}
				pesVideo = new(tsPesFragment)
			}
			if pesVideo != nil {
				pesVideo.add(packet)
			}
		case 0x101: // audio
			if packet.header.isPayloadStart {
				if pesAudio != nil {
					stream.audios = append(stream.audios, pesAudio)
				}
				pesAudio = new(tsPesFragment)
			}
			if pesAudio != nil {
				pesAudio.add(packet)
			}
		}
		stream.packets = append(stream.packets, packet)
	}

	if pesVideo != nil {
		stream.videos = append(stream.videos, pesVideo)
	}
	if pesAudio != nil {
		stream.audios = append(stream.audios, pesAudio)
	}
}

func (stream *tsStream) parseTSPacket(buffer []byte, packNo, offset int) *tsPacket {
	if buffer[0] != syncByte {
		panic(fmt.Sprintf("Invalid ts package at %d offset %d", packNo, offset))
	}

	header := tsHeader{}
	header.syncByte = buffer[0]
	header.transportErrorIndicator = (buffer[1] & 0x80) >> 7
	header.payloadUnitStartIndicator = (buffer[1] & payloadStartMask) >> 6
	header.pid = int(buffer[1]&0x1F)<<8 | int(buffer[2])
	header.transportScramblingControl = (buffer[3] & 0xC0) >> 6
	header.adaptationField = (buffer[3] & atfMask) >> 4
	header.continuityCounter = buffer[3] & 0x0F
	header.hasError = header.transportErrorIndicator != 0
	header.isPayloadStart = header.payloadUnitStartIndicator != 0
	header.hasAdaptationFieldField = header.adaptationField == atfFieldOnly || header.adaptationField == atfFieldFollowPayload
	header.hasPayload = header.adaptationField == atfPayloadOnly || header.adaptationField == atfFieldFollowPayload

	packet := newTSPacket()
	packet.header = header
	packet.packNo = packNo
	packet.startOffset = offset

	packet.headerLength = 4
	if header.hasAdaptationFieldField {
		packet.atfLength = int(buffer[4])
		packet.headerLength += 1 + packet.atfLength
	}

	packet.pesHeaderLength = 0
	if header.isPayloadStart {
		if packet.headerLength+8 < len(buffer) {
			packet.pesHeaderLength = 6 + 3 + int(buffer[packet.headerLength+8])
		}
		packet.pesOffset = packet.startOffset + packet.headerLength
	}

	packet.payloadStartOffset = packet.startOffset + packet.headerLength + packet.pesHeaderLength
	packet.payloadLength = packetLength - packet.headerLength - packet.pesHeaderLength
	if packet.payloadLength < 0 {
		packet.payloadLength = 0
	}

	if packet.payloadLength > 0 {
		packet.payload = buffer[packet.headerLength+packet.pesHeaderLength:]
	}

	return packet
}
