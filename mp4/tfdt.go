package mp4

import (
	"io"

	"github.com/edgeware/mp4ff/bits"
)

// TfdtBox - Track Fragment Decode Time (tfdt)
//
// Contained in : Track Fragment box (traf)
type TfdtBox struct {
	Version             byte
	Flags               uint32
	BaseMediaDecodeTime uint64
}

// DecodeTfdt - box-specific decode
func DecodeTfdt(hdr boxHeader, startPos uint64, r io.Reader) (Box, error) {
	data, err := readBoxBody(r, hdr)
	if err != nil {
		return nil, err
	}
	s := NewSliceReader(data)
	versionAndFlags := s.ReadUint32()
	version := byte(versionAndFlags >> 24)
	var baseMediaDecodeTime uint64
	if version == 0 {
		baseMediaDecodeTime = uint64(s.ReadUint32())
	} else {
		baseMediaDecodeTime = s.ReadUint64()
	}

	b := &TfdtBox{
		Version:             version,
		Flags:               versionAndFlags & flagsMask,
		BaseMediaDecodeTime: baseMediaDecodeTime,
	}
	return b, nil
}

// CreateTfdt - Create a new TfdtBox with baseMediaDecodeTime
func CreateTfdt(baseMediaDecodeTime uint64) *TfdtBox {
	var version byte = 0
	if baseMediaDecodeTime >= 4294967296 {
		version = 1
	}
	return &TfdtBox{
		Version:             version,
		Flags:               0,
		BaseMediaDecodeTime: baseMediaDecodeTime,
	}
}

// SetBaseMediaDecodeTime - Set time of TfdtBox
func (t *TfdtBox) SetBaseMediaDecodeTime(bTime uint64) {
	if bTime >= 4294967296 {
		t.Version = 1
	} else {
		t.Version = 0
	}
	t.BaseMediaDecodeTime = bTime
}

// Type - return box type
func (t *TfdtBox) Type() string {
	return "tfdt"
}

// Size - return calculated size
func (t *TfdtBox) Size() uint64 {
	return uint64(boxHeaderSize + 8 + 4*int(t.Version))
}

// Encode - write box to w
func (t *TfdtBox) Encode(w io.Writer) error {
	sw := bits.NewSliceWriterWithSize(int(t.Size()))
	err := t.EncodeSW(sw)
	if err != nil {
		return err
	}
	_, err = w.Write(sw.Bytes())
	return err
}

// EncodeSW - box-specific encode to slicewriter
func (t *TfdtBox) EncodeSW(sw bits.SliceWriter) error {
	err := EncodeHeaderSW(t, sw)
	if err != nil {
		return err
	}
	versionAndFlags := (uint32(t.Version) << 24) + t.Flags
	sw.WriteUint32(versionAndFlags)
	if t.Version == 0 {
		sw.WriteUint32(uint32(t.BaseMediaDecodeTime))
	} else {
		sw.WriteUint64(t.BaseMediaDecodeTime)
	}
	return sw.AccError()
}

// Info - write box info to w
func (t *TfdtBox) Info(w io.Writer, specificBoxLevels, indent, indentStep string) (err error) {
	bd := newInfoDumper(w, indent, t, int(t.Version), t.Flags)
	bd.write(" - baseMediaDecodeTime: %d", t.BaseMediaDecodeTime)
	return bd.err
}
