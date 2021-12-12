package smoothstreaming

import (
	"bytes"
	"fmt"

	"github.com/go-webdl/media-codec/avc"
	"github.com/go-webdl/media-codec/hevc"
	"github.com/go-webdl/mp4"

	"github.com/google/uuid"
	"golang.org/x/text/language"
)

type MoovProcessor struct {
	TrackID            uint32
	Codec              mp4.FourCC
	Width              uint32
	Height             uint32
	Duration           uint64
	Timescale          uint64
	Language           language.Base
	CodecPrivateData   []byte
	StreamType         StreamType
	StreamName         string
	Protected          bool
	KID                [16]byte
	SystemID           uuid.UUID
	ProtectionInitData []byte
}

func (p MoovProcessor) CreateFtypMp4Box() (ftyp mp4.Box, err error) {
	ftyp = &mp4.FileTypeBox{
		MajorBrand:   mp4.Iso6FourCC,
		MinorVersion: 1,
		CompatibleBrands: []mp4.FourCC{
			mp4.IsomFourCC,
			mp4.Iso6FourCC,
			mp4.MsdhFourCC,
		},
	}
	ftyp.Mp4BoxUpdate()
	return
}

func (p MoovProcessor) CreateMoovMp4Box() (moov mp4.Box, err error) {
	mvhd, err := p.CreateMvhdMp4Box()
	if err != nil {
		return
	}

	trak, err := p.CreateTrakMp4Box()
	if err != nil {
		return
	}

	mvex, err := p.CreateMvexMp4Box()
	if err != nil {
		return
	}

	children := []mp4.Box{mvhd, trak, mvex}

	if p.Protected {
		var pssh mp4.Box
		if pssh, err = p.CreatePsshMp4Box(); err != nil {
			return
		}
		children = append(children, pssh)
	}

	moov = &mp4.MovieBox{}
	if err = moov.Mp4BoxReplaceChildren(children); err != nil {
		return
	}
	moov.Mp4BoxUpdate()
	return
}

func (p MoovProcessor) CreateMvhdMp4Box() (mvhd mp4.Box, err error) {
	mvhd = &mp4.MovieHeaderBox{
		FullHeader: mp4.FullHeader{Version: 1}, // in order to have 64bits duration value
		Timescale:  uint32(p.Timescale),
		Duration:   p.Duration * p.Timescale,
		Rate:       0x00010000, // typically 1.0
		Volume:     0x0100,     // typically, full volume
		Matrix: [9]int32{ // Unity matrix
			0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000,
		},
		NextTrackID: p.TrackID + 1,
	}
	return
}

func (p MoovProcessor) CreatePsshMp4Box() (pssh mp4.Box, err error) {
	pssh = &mp4.ProtectionSystemSpecificHeaderBox{
		SystemID: p.SystemID,
		Data:     p.ProtectionInitData,
	}
	return
}

func (p MoovProcessor) CreateMvexMp4Box() (mvex mp4.Box, err error) {
	trex := &mp4.TrackExtendsBox{
		TrackID:                      p.TrackID,
		DefaultSampleDescrptionIndex: 1,
	}
	mvex = &mp4.MovieExtendsBox{}
	if err = mvex.Mp4BoxReplaceChildren([]mp4.Box{trex}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateTrakMp4Box() (trak mp4.Box, err error) {
	tkhd := &mp4.TrackHeaderBox{
		TrackID:  p.TrackID,
		Duration: p.Duration * p.Timescale,
		Volume:   0x0100,
		Matrix: [9]int32{ // Unity matrix
			0x00010000, 0, 0, 0, 0x00010000, 0, 0, 0, 0x40000000,
		},
		Width:  p.Width,
		Height: p.Height,
	}
	tkhd.Version = 1
	tkhd.Mp4BoxSetFlags(mp4.FLAG_TKHD_TRACK_ENABLED | mp4.FLAG_TKHD_TRACK_IN_MOVIE | mp4.FLAG_TKHD_TRACK_IN_PREVIEW)

	mdia, err := p.CreateMdiaMp4Box()
	if err != nil {
		return
	}

	trak = &mp4.TrackBox{}
	if err = trak.Mp4BoxReplaceChildren([]mp4.Box{tkhd, mdia}); err != nil {
		return
	}

	return
}

func (p MoovProcessor) CreateMdiaMp4Box() (mdia mp4.Box, err error) {
	mdhd := &mp4.MediaHeaderBox{
		Timescale: uint32(p.Timescale),
		Duration:  p.Duration * p.Timescale,
		Language:  p.Language,
	}
	mdhd.Version = 1

	hdlr := &mp4.HandlerBox{
		HandlerType: mp4.VideFourCC,
		Name:        mp4.NullTerminatedString(p.StreamName),
	}
	switch p.StreamType {
	case VideoStream:
		hdlr.HandlerType = mp4.VideFourCC
	case AudioStream:
		hdlr.HandlerType = mp4.SounFourCC
	default:
		hdlr.HandlerType = mp4.MetaFourCC
	}

	minf, err := p.CreateMinfMp4Box()
	if err != nil {
		return
	}

	mdia = &mp4.MediaBox{}
	if err = mdia.Mp4BoxReplaceChildren([]mp4.Box{mdhd, hdlr, minf}); err != nil {
		return
	}

	return
}

func (p MoovProcessor) CreateMinfMp4Box() (minf mp4.Box, err error) {
	mhd, err := p.CreateMhdMp4Box()
	if err != nil {
		return
	}

	dinf, err := p.CreateDinfMp4Box()
	if err != nil {
		return
	}

	stbl, err := p.CreateStblMp4Box()
	if err != nil {
		return
	}

	childred := []mp4.Box{dinf, stbl}
	if mhd != nil {
		childred = append([]mp4.Box{mhd}, childred...)
	}

	minf = &mp4.MediaInformationBox{}
	if err = minf.Mp4BoxReplaceChildren(childred); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateStblMp4Box() (stbl mp4.Box, err error) {
	stsd, err := p.CreateStsdMp4Box()
	if err != nil {
		return
	}

	stbl = &mp4.SampleTableBox{}
	if err = stbl.Mp4BoxReplaceChildren([]mp4.Box{
		stsd,
		&mp4.TimeToSampleBox{},
		&mp4.SampleToChunkBox{},
		&mp4.ChunkOffsetBox{},
		&mp4.SampleSizeBox{},
	}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateStsdMp4Box() (stsd mp4.Box, err error) {
	sampleEntry, err := p.CreateSampleEntryMp4Box()
	if err != nil {
		return
	}

	stsd = &mp4.SampleDescriptionBox{}
	if err = stsd.Mp4BoxReplaceChildren([]mp4.Box{sampleEntry}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateSampleEntryMp4Box() (sampleEntry mp4.Box, err error) {
	switch p.Codec {
	case mp4.Avc1FourCC:
		sampleEntry, err = p.CreateAvc1Mp4Box()
	case mp4.Hvc1FourCC, mp4.Hev1FourCC:
		sampleEntry, err = p.CreateHvc1Mp4Box()
	default:
		err = fmt.Errorf("codec %s not supported: %w", p.Codec, ErrUnknownCodec)
	}
	return
}

func (p MoovProcessor) CreateHvc1Mp4Box() (hvc1 mp4.Box, err error) {
	hvc1 = &mp4.VisualSampleEntryBox{
		SampleEntry: mp4.SampleEntry{
			Header:             mp4.Header{Type: mp4.BoxType(p.Codec)},
			DataReferenceIndex: 1,
		},
		Width:           uint16(p.Width),
		Height:          uint16(p.Height),
		HorizResolution: 72, // 72 dpi
		VertResolution:  72, // 72 dpi,
		FrameCount:      1,
		CompressorName:  "HEVC Coding",
		Depth:           0x0018, // 0x0018 – images are in colour with no alpha.
	}
	hvcC, err := p.CreateHvcCMp4Box()
	if err != nil {
		return
	}
	children := []mp4.Box{hvcC}
	if p.Protected {
		hvc1.Mp4BoxSetType(mp4.EncvBoxType)

		var sinf mp4.Box
		if sinf, err = p.CreateSinfMp4Box(); err != nil {
			return
		}

		children = append(children, sinf)
	}
	if err = hvc1.Mp4BoxReplaceChildren(children); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateAvc1Mp4Box() (avc1 mp4.Box, err error) {
	avc1 = &mp4.VisualSampleEntryBox{
		SampleEntry: mp4.SampleEntry{
			Header:             mp4.Header{Type: mp4.BoxType(mp4.Avc1FourCC)},
			DataReferenceIndex: 1,
		},
		Width:           uint16(p.Width),
		Height:          uint16(p.Height),
		HorizResolution: 72, // 72 dpi
		VertResolution:  72, // 72 dpi,
		FrameCount:      1,
		CompressorName:  "AVC Coding",
		Depth:           0x0018, // 0x0018 – images are in colour with no alpha.
	}
	avcC, err := p.CreateAvcCMp4Box()
	if err != nil {
		return
	}
	children := []mp4.Box{avcC}
	if p.Protected {
		avc1.Mp4BoxSetType(mp4.EncvBoxType)

		var sinf mp4.Box
		if sinf, err = p.CreateSinfMp4Box(); err != nil {
			return
		}

		children = append(children, sinf)
	}
	if err = avc1.Mp4BoxReplaceChildren(children); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateSinfMp4Box() (sinf mp4.Box, err error) {
	sinf = &mp4.ProtectionSchemeInfoBox{}
	frmt := &mp4.OriginalFormatBox{
		DataFormat: p.Codec,
	}
	schm := &mp4.SchemeTypeBox{
		SchemeType:    mp4.CencFourCC, // 'cenc' => common encryption
		SchemeVersion: 0x00010000,     // version set to 0x00010000 (Major version 1, Minor version 0)
	}
	schi, err := p.CreateSchiMp4Box()
	if err != nil {
		return
	}
	if err = sinf.Mp4BoxReplaceChildren([]mp4.Box{frmt, schm, schi}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateSchiMp4Box() (schi mp4.Box, err error) {
	tenc := &mp4.TrackEncryptionBox{
		DefaultIsProtected:     1,
		DefaultPerSampleIVSize: 8,
		DefaultKID:             p.KID,
	}
	schi = &mp4.SchemeInformationBox{}
	if err = schi.Mp4BoxReplaceChildren([]mp4.Box{tenc}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateAvcCMp4Box() (avcC mp4.Box, err error) {
	nalus := bytes.Split(p.CodecPrivateData, []byte{0, 0, 0, 1})
	if len(nalus) < 1 {
		err = fmt.Errorf("invalid CodecPrivateData for avcC: %w", ErrInvalidParam)
		return
	}
	var sps []avc.AVCSequenceParameterSet
	var pps []avc.AVCPictureParameterSet
	for _, nalu := range nalus[1:] {
		naluType := avc.GetNaluType(nalu[0])
		switch naluType {
		case avc.NALU_SPS:
			sps = append(sps, avc.AVCSequenceParameterSet{NALUnit: nalu})
		case avc.NALU_PPS:
			pps = append(pps, avc.AVCPictureParameterSet{NALUnit: nalu})
		}
	}
	var avcProfile, avcProfileCompatibility, avcLevel uint8
	if len(sps) > 0 {
		avcProfile = sps[0].NALUnit[1]
		avcProfileCompatibility = sps[0].NALUnit[2]
		avcLevel = sps[0].NALUnit[3]
	}
	avcC = &mp4.AVCConfigurationBox{
		AVCConfig: avc.AVCDecoderConfigurationRecord{
			ConfigurationVersion:  1,
			AVCProfileIndication:  avcProfile,
			ProfileCompatibility:  avcProfileCompatibility,
			AVCLevelIndication:    avcLevel,
			LengthSizeMinusOne:    3,
			SequenceParameterSets: sps,
			PictureParameterSets:  pps,
		},
	}
	return
}

func (p MoovProcessor) CreateHvcCMp4Box() (hvcC mp4.Box, err error) {
	nalus := bytes.Split(p.CodecPrivateData, []byte{0, 0, 0, 1})
	if len(nalus) < 1 {
		err = fmt.Errorf("invalid CodecPrivateData for hvcC: %w", ErrInvalidParam)
		return
	}
	var vpsNalus, spsNalus, ppsNalus [][]byte
	for _, nalu := range nalus[1:] {
		naluType := hevc.GetNaluType(nalu[0])
		switch naluType {
		case hevc.NALU_VPS:
			vpsNalus = append(vpsNalus, nalu)
		case hevc.NALU_SPS:
			spsNalus = append(spsNalus, nalu)
		case hevc.NALU_PPS:
			ppsNalus = append(ppsNalus, nalu)
		}
	}
	if len(spsNalus) == 0 {
		err = fmt.Errorf("cannot find hevc sps nalu")
		return
	}
	conf, err := hevc.CreateHEVCDecoderConfigurationRecord(vpsNalus, spsNalus, ppsNalus, true, true, true)
	if err != nil {
		return
	}
	hvcC = &mp4.HEVCConfigurationBox{
		HEVCConfig: conf,
	}
	return
}

func (p MoovProcessor) CreateDinfMp4Box() (dinf mp4.Box, err error) {
	dref, err := p.CreateDrefMp4Box()
	if err != nil {
		return
	}
	dinf = &mp4.DataInformationBox{}
	if err = dinf.Mp4BoxReplaceChildren([]mp4.Box{dref}); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateDrefMp4Box() (dref mp4.Box, err error) {
	url := &mp4.DataEntryBox{}
	url.Mp4BoxSetFlags(mp4.FLAG_DREF_SAME_FILE)
	dref = &mp4.DataReferenceBox{}
	if err = dref.Mp4BoxAppend(url); err != nil {
		return
	}
	return
}

func (p MoovProcessor) CreateMhdMp4Box() (mhd mp4.Box, err error) {
	switch p.StreamType {
	case VideoStream:
		mhd = &mp4.VideoMediaHeaderBox{}
	case AudioStream:
		mhd = &mp4.SoundMediaHeaderBox{}
	}
	return
}

func (p MoovProcessor) CreateInitMp4Box() (ftyp, moov mp4.Box, err error) {
	if ftyp, err = p.CreateFtypMp4Box(); err != nil {
		return
	}
	if moov, err = p.CreateMoovMp4Box(); err != nil {
		return
	}
	return
}
