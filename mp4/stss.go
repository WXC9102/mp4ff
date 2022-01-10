package mp4

import (
	"encoding/binary"
	"io"

	"github.com/edgeware/mp4ff/bits"
)

// StssBox - Sync Sample Box (stss - optional)
//
// Contained in : Sample Table box (stbl)
//
// This lists all sync samples (key frames for video tracks) in the data. If absent, all samples are sync samples.
type StssBox struct {
	Version      byte
	Flags        uint32
	SampleNumber []uint32
}

// DecodeStss - box-specific decode
func DecodeStss(hdr boxHeader, startPos uint64, r io.Reader) (Box, error) {
	data, err := readBoxBody(r, hdr)
	if err != nil {
		return nil, err
	}
	versionAndFlags := binary.BigEndian.Uint32(data[0:4])
	b := &StssBox{
		Version:      byte(versionAndFlags >> 24),
		Flags:        versionAndFlags & flagsMask,
		SampleNumber: []uint32{},
	}
	ec := binary.BigEndian.Uint32(data[4:8])
	for i := 0; i < int(ec); i++ {
		sample := binary.BigEndian.Uint32(data[(8 + 4*i):(12 + 4*i)])
		b.SampleNumber = append(b.SampleNumber, sample)
	}
	return b, nil
}

// EntryCount - number of sync samples
func (b *StssBox) EntryCount() uint32 {
	return uint32(len(b.SampleNumber))
}

// Type - box-specific type
func (b *StssBox) Type() string {
	return "stss"
}

// Size - box-specific size
func (b *StssBox) Size() uint64 {
	return uint64(boxHeaderSize + 8 + len(b.SampleNumber)*4)
}

// IsSyncSample - check if sample (one-based) sampleNr is a sync sample
func (b *StssBox) IsSyncSample(sampleNr uint32) (isSync bool) {
	// Based on a binary search algorithm from the Go standard library code.
	// i will be the lowest index such that b.SampleNumber[i] >= sampleNr
	// or len(b.SampleNumber) if not possible.
	nrSamples := len(b.SampleNumber)
	i, j := 0, nrSamples
	for i < j {
		h := (i + j) >> 1
		// i ≤ h < j
		if b.SampleNumber[h] < sampleNr {
			i = h + 1
		} else {
			j = h
		}
	}
	return i < nrSamples && b.SampleNumber[i] == sampleNr
}

// Encode - write box to w
func (b *StssBox) Encode(w io.Writer) error {
	sw := bits.NewSliceWriterWithSize(int(b.Size()))
	err := b.EncodeSW(sw)
	if err != nil {
		return err
	}
	_, err = w.Write(sw.Bytes())
	return err
}

// EncodeSW - box-specific encode to slicewriter
func (b *StssBox) EncodeSW(sw bits.SliceWriter) error {
	err := EncodeHeaderSW(b, sw)
	if err != nil {
		return err
	}
	versionAndFlags := (uint32(b.Version) << 24) + b.Flags
	sw.WriteUint32(versionAndFlags)
	sw.WriteUint32(uint32(len(b.SampleNumber)))
	for i := range b.SampleNumber {
		sw.WriteUint32(b.SampleNumber[i])
	}
	return sw.AccError()
}

// Info - write box-specific information
func (b *StssBox) Info(w io.Writer, specificBoxLevels, indent, indentStep string) error {
	bd := newInfoDumper(w, indent, b, int(b.Version), b.Flags)
	level := getInfoLevel(b, specificBoxLevels)
	if level >= 1 {
		for i := range b.SampleNumber {
			bd.write(" - syncSample[%d]: sampleNumber=%d", i+1, b.SampleNumber[i])
		}
	}
	return bd.err
}
