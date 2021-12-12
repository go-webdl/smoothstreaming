// [MS-SSTR]: Smooth Streaming Protocol
// https://docs.microsoft.com/en-us/openspecs/windows_protocols/ms-sstr/8383f27f-7efe-4c60-832a-387274457251
package smoothstreaming

import (
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/go-webdl/encodetype"

	"github.com/google/uuid"
)

// The SmoothStreamingMedia field and related fields encapsulate metadata that
// is required to play the presentation.
//
// An XML element that encapsulates all metadata that is required by the client
// to play back the presentation.
//
// Attributes can appear in any order. However, the following fields are
// required and MUST be present in SmoothStreamingMediaAttributes:
// MajorVersionAttribute, MinorVersionAttribute, and DurationAttribute.
type SmoothStreamingMedia struct {
	// The major version of the Manifest Response message. MUST be set to 2.
	MajorVersion uint `xml:",attr"`

	// The minor version of the Manifest Response message. MUST be set to 0 or
	// 2.
	MinorVersion uint `xml:",attr"`

	// The duration of the presentation, specified as the number of time
	// increments indicated by the value of the TimeScale field.
	Duration uint64 `xml:",attr"`

	// The timescale of the Duration attribute, specified as the number of
	// increments in 1 second. The default value is 10000000.
	TimeScale *uint64 `xml:",attr"`

	// Specifies the presentation type. If this field contains a TRUE value, it
	// specifies that the presentation is a live presentation. Otherwise, the
	// presentation is an on-demand presentation.
	IsLive *bool `xml:",attr"`

	// Specifies the size of the server buffer, as an integer number of
	// fragments. This field MUST be omitted for on-demand presentations.
	LookaheadCount *uint32 `xml:",attr"`

	// The length of the DVR window, specified as the number of time increments
	// indicated by the value of the TimeScale field. If this field is omitted
	// for a live presentation or set to 0, the DVR window is effectively
	// infinite. This field MUST be omitted for on-demand presentations.
	DVRWindowLength *uint64 `xml:",attr"`

	// The StreamElement field and related fields encapsulate metadata that is
	// required to play a specific stream in the presentation.
	Streams []*StreamIndex `xml:"StreamIndex"`

	// The ProtectionElement field and related fields encapsulate metadata that
	// is required to play back protected content.
	Protection *Protection
}

// The StreamElement field and related fields encapsulate metadata that is
// required to play a specific stream in the presentation.
//
// An XML element that encapsulates all metadata that is required by the client
// to play back a stream.
//
// Attributes can appear in any order. However, the following field is required
// and MUST be present in StreamAttributes: TypeAttribute. The following
// additional fields are required and MUST be present in StreamAttributes unless
// an Embedded Track is used in the StreamContent field:
// NumberOfFragmentsAttribute, NumberOfTracksAttribute, and UrlAttribute.
type StreamIndex struct {
	// The type of the stream: video, audio, or text. If the specified type is
	// text, the following field is required and MUST appear in
	// StreamAttributes: SubtypeAttribute. Unless the specified type is video,
	// the following fields MUST NOT appear in StreamAttributes:
	// StreamMaxWidthAttribute, StreamMaxHeightAttribute, DisplayWidthAttribute,
	// and DisplayHeightAttribute.
	Type StreamType `xml:",attr"`

	// A four-character code that identifies the intended use category for each
	// sample in a text track. However, the FourCC field, specified in section
	// 2.2.2.5, is used to identify the media format for each sample. The
	// following range of values is reserved, with the following semantic
	// meanings:
	//
	// * "SCMD": Triggers for actions by the higher-layer implementation on the
	// client.
	//
	// * "CHAP": Chapter markers.
	//
	// * "SUBT": Subtitles that are used for foreign-language audio.
	//
	// * "CAPT": Closed captions for people who are deaf.
	//
	// * "DESC": Media descriptions for people who are deaf.
	//
	// * "CTRL": Events that control the application business logic.
	//
	// * "DATA": Application data that does not fall into any of the previous
	// categories.
	Subtype *string `xml:",attr"`

	// The timescale for duration and time values in this stream, specified as
	// the number of increments in 1 second.
	TimeScale *uint64 `xml:",attr"`

	// The name of the stream.
	Name *string `xml:",attr"`

	// The number of fragments that are available for this stream.
	NumberOfFragments *uint32 `xml:"Chunks,attr"`

	// The number of tracks that are available for this stream.
	NumberOfTracks *uint32 `xml:"QualityLevels,attr"`

	// A pattern that is used by the client to generate Fragment Request
	// messages.
	URL *string `xml:"Url,attr"`

	// The maximum width of a video sample, in pixels.
	MaxWidth *uint32 `xml:",attr"`

	// The maximum height of a video sample, in pixels.
	MaxHeight *uint32 `xml:",attr"`

	// The suggested display width of a video sample, in pixels.
	DisplayWidth *uint32 `xml:",attr"`

	// The suggested display height of a video sample, in pixels.
	DisplayHeight *uint32 `xml:",attr"`

	// Specifies the non-sparse stream that is used to transmit timing
	// information for this stream. If the ParentStream field is present, it
	// indicates that the stream that is described by the containing
	// StreamElement field is a sparse stream. If present, the value of this
	// field MUST match the value of the Name field for a non-sparse stream in
	// the presentation.
	ParentStreamIndex *string `xml:",attr"`

	// Specifies whether sample data for this stream appears directly in the
	// manifest as part of the ManifestOutputSample field, specified in section
	// 2.2.2.6.1, if this field contains a TRUE value. Otherwise, the
	// ManifestOutputSample field for fragments that are part of this stream
	// MUST be omitted.
	ManifestOutput bool `xml:",attr"`

	// Metadata describing available tracks.
	Tracks []*Track `xml:"QualityLevel"`

	// Metadata describing available fragments.
	Fragments []*StreamFragment `xml:"c"`
}

// The TrackElement field and related fields encapsulate metadata that is
// required to play a specific track in the stream.
//
// An XML element that encapsulates all metadata that is required by the client
// to play a track.
//
// Attributes can appear in any order. However, the following fields are
// required and MUST be present in TrackAttributes: IndexAttribute and
// BitrateAttribute. If the track is contained in a stream whose Type is video,
// the following additional fields are also required and MUST be present in
// TrackAttributes: MaxWidthAttribute, MaxHeightAttribute, and
// CodecPrivateDataAttribute. If the track is contained in a stream whose Type
// is audio, the following additional fields are also required and MUST be
// present in TrackAttributes: MaxWidthAttribute, MaxHeightAttribute,
// CodecPrivateDataAttribute, SamplingRateAttribute, ChannelsAttribute,
// BitsPerSampleAttribute, PacketSizeAttribute, AudioTagAttribute, and
// FourCCAttribute.
type Track struct {
	// An ordinal that identifies the track and MUST be unique for each track in
	// the stream. Index SHOULD start at 0 and increment by 1 for each
	// subsequent track in the stream.
	Index uint32 `xml:",attr"`

	// The average bandwidth that is consumed by the track, in bits per second
	// (bps). The value 0 MAY be used for tracks whose bit rate is negligible
	// relative to other tracks in the presentation.
	Bitrate uint32 `xml:",attr"`

	// The maximum width of a video sample, in pixels.
	MaxWidth *uint32 `xml:",attr"`

	// The maximum height of a video sample, in pixels.
	MaxHeight *uint32 `xml:",attr"`

	// The Sampling Rate of an audio track, as defined in [ISO/IEC-14496-12].
	SamplingRate *uint32 `xml:",attr"`

	// The Channel Count of an audio track, as defined in [ISO/IEC-14496-12].
	Channels *uint16 `xml:",attr"`

	// A numeric code that identifies which media format and variant of the
	// media format is used for each sample in an audio track. The following
	// range of values is reserved with the following semantic meanings:
	//
	// * "1": The sample media format is Linear 8 or 16-bit pulse code
	// modulation.
	//
	// * "353": Microsoft Windows Media Audio v7, v8 and v9.x Standard (WMA
	// Standard)
	//
	// * "353": Microsoft Windows Media Audio v9.x and v10 Professional (WMA
	// Professional).
	//
	// * "85": International Organization for Standardization (ISO) MPEG-1 Layer
	// III (MP3).
	//
	// * "255": ISO Advanced Audio Coding (AAC).
	//
	// * "65534": Vendor-extensible format. If specified, the CodecPrivateData
	// field SHOULD contain a hexadecimal-encoded version of the
	// WAVE_FORMAT_EXTENSIBLE structure [WFEX].
	AudioTag *uint32 `xml:",attr"`

	// The sample size of an audio track, as defined in [ISO/IEC-14496-12].
	BitsPerSample *uint16 `xml:",attr"`

	// The size of each audio packet, in bytes.
	PacketSize *uint32 `xml:",attr"`

	// A four-character code that identifies which media format is used for each
	// sample. The following range of values is reserved with the following
	// semantic meanings:
	//
	// * "H264": Video samples for this track use Advanced Video Coding, as
	// described in [ISO/IEC-14496-15].
	//
	// * "WVC1": Video samples for this track use VC-1, as described in [VC-1].
	//
	// * "AACL": Audio samples for this track use AAC (Low Complexity), as
	// specified in [ISO/IEC-14496-3].
	//
	// * "WMAP": Audio samples for this track use WMA Professional.
	//
	// * A vendor extension value containing a registered with MPEG4-RA, as
	// specified in [ISO/IEC-14496-12].
	FourCC *string `xml:",attr"`

	// Data that specifies parameters that are specific to the media format and
	// common to all samples in the track, represented as a string of
	// hexadecimal-coded bytes. The format and semantic meaning of byte sequence
	// varies with the value of the FourCC field as follows:
	//
	// * The FourCC field equals "H264": The CodecPrivateData field contains a
	// hexadecimal-coded string representation of the following byte sequence,
	// specified in ABNF [RFC5234]:
	//
	//     %x00 %x00 %x00 %x01 SPSField %x00 %x00 %x00 %x01 PPSField
	//
	// * SPSField contains the Sequence Parameter Set (SPS).
	//
	// * PPSField contains the Picture Parameter Set (PPS).
	//
	// * The FourCC field equals "WVC1": The CodecPrivateData field contains a
	// hexadecimal-coded string representation of the VIDEOINFOHEADER structure,
	// specified in [MSDN-VIH].
	//
	// * The FourCC field equals "AACL": The CodecPrivateData field SHOULD be
	// empty.
	//
	// * The FourCC field equals "WMAP": The CodecPrivateData field contains the
	// WAVEFORMATEX structure, specified in [WFEX], if the AudioTag field equals
	// "65534". Otherwise, it SHOULD be empty.
	//
	// * The FourCC field is a vendor extension value: The format of the
	// CodecPrivateData field is also vendor-extensible. Registration of the
	// FourCC field value with MPEG4-RA, as specified in [ISO/IEC-14496-12], can
	// be used to avoid collision between extensions.
	CodecPrivateData encodetype.HexBytes `xml:",attr"`

	// The number of bytes that specifies the length of each Network Abstraction
	// Layer (NAL) unit. This field SHOULD be omitted unless the value of the
	// FourCC field is "H264". The default value is 4.
	NALUnitLengthField *uint16 `xml:",attr"`

	// Specify metadata that disambiguates tracks in a stream.
	CustomAttributes *CustomAttributes
}

// The StreamFragmentElement field and related fields are used to specify
// metadata for one set of related fragments in a stream. The order of repeated
// StreamFragmentElement fields in a containing StreamElement field is
// significant for the correct function of the Smooth Streaming Transport
// Protocol. To this end, the following elements make use of the terms
// "preceding" and "subsequent" StreamFragmentElement in reference to the order
// of these fields.
//
// An XML element that encapsulates metadata for a set of related fragments.
// Attributes can appear in any order. However, either one or both of the
// following fields are required and MUST be present in
// StreamFragmentAttributes: FragmentDuration and FragmentTime. Additionally, a
// contiguous sequence of fragments MUST be represented using one of the
// following schemes. A sequence of fragments is termed contiguous if, with the
// exception of the first fragment, the StreamFragmentElement's FragmentTime
// field of any fragment in the sequence is equal to the sum of the
// StreamFragmentElement's FragmentTime field and the FragmentDuration field of
// the preceding fragment.
//
// Attributes can appear in any order. However, either one or both of the
// following fields are required and MUST be present in
// StreamFragmentAttributes: FragmentDuration and FragmentTime. Additionally, a
// contiguous sequence of fragments MUST be represented using one of the
// following schemes. A sequence of fragments is termed contiguous if, with the
// exception of the first fragment, the StreamFragmentElement's FragmentTime
// field of any fragment in the sequence is equal to the sum of the
// StreamFragmentElement's FragmentTime field and the FragmentDuration field of
// the preceding fragment.
//
// § Start-time coding – Each fragment in the sequence has an explicit value for
// the StreamFragmentElement's FragmentTime field and an implicit value for the
// StreamFragmentElement's FragmentDuration field, except the last fragment, for
// which the value of the StreamFragmentElement's FragmentDuration field is
// explicit.
//
// § Duration coding – Each fragment in the sequence has an explicit value for
// the StreamFragmentElement's FragmentDuration field and an implicit value for
// the StreamFragmentElement's FragmentTime field, except the first fragment,
// whose start-time is explicit unless the implicit value of zero is desired.
type StreamFragment struct {
	// The ordinal of the StreamFragmentElement field in the stream. If
	// FragmentNumber is specified, its value MUST monotonically increase with
	// the value of the FragmentTime field.
	Number *uint32 `xml:"n,attr"`

	// The duration of the fragment, specified as a number of increments defined
	// by the implicit or explicit value of the containing StreamElement's
	// StreamTimeScale field. If the FragmentDuration field is omitted, its
	// implicit value MUST be computed by the client by subtracting the value of
	// the preceding StreamFragmentElement's FragmentTime field from the value
	// of this StreamFragmentElement's FragmentTime field. If no preceding
	// StreamFragmentElement exists, the implicit value of the FragmentDuration
	// field MUST be computed by the client by subtracting the value of this
	// StreamFragmentElement FragmentTime field from the subsequent
	// StreamFragmentElement's FragmentTime field.
	//
	// If no preceding or subsequent StreamFragmentElement field exists, the
	// implicit value of the FragmentDuration field is the value of the
	// SmoothStreamingMedia's Duration field.
	Duration *uint64 `xml:"d,attr"`

	// The time of the fragment, specified as a number of increments defined by
	// the implicit or explicit value of the containing StreamElement's
	// StreamTimeScale field. If the FragmentTime field is omitted, its implicit
	// value MUST be computed by the client by adding the value of the preceding
	// StreamFragmentElement's FragmentTime field to the value of the preceding
	// StreamFragmentElement's FragmentDuration field. If no preceding
	// StreamFragmentElement exists, the implicit value of the FragmentTime
	// field is 0.
	Time *uint64 `xml:"t,attr"`

	// The repeat count of the fragment, specified as the number of contiguous
	// fragments with the same duration defined by the StreamFragmentElement's
	// FragmentTime field. This value is one-based. (A value of 2 means two
	// fragments in the contiguous series). The SmoothStreamingMedia's
	// MajorVersion and MinorVersion fields MUST both be set to 2.
	Repeat *uint64 `xml:"r,attr"`

	// The TrackFragmentElement field and related fields are used to specify
	// metadata pertaining to a fragment for a specific track, rather than all
	// versions of a fragment for a stream.
	TrackFragments []*TrackFragment `xml:"f"`
}

// An XML element that encapsulates informative track-specific metadata for a
// specific fragment. Attributes can appear in any order. However, the following
// field is required and MUST be present in TrackFragmentAttributes:
// TrackFragmentIndexAttribute.
type TrackFragment struct {
	// An ordinal that MUST match the value of the Index field for the track to
	// which this TrackFragment field pertains.
	Index uint32 `xml:"i,attr"`

	// A string that contains the base64-encoded representation of the raw bytes
	// of the sample data for this fragment. This field MUST be omitted unless
	// the ManifestOutput field for the corresponding stream contains a TRUE
	// value.
	ManifestOutputSample encodetype.Base64Bytes `xml:",chardata"`
}

// The CustomAttributesElement field and related fields are used to specify
// metadata that disambiguates tracks in a stream.
type CustomAttributes struct {
	// Metadata that is expressed as key/value pairs that disambiguate tracks.
	Attributes []*Attribute `xml:"Attribute"`
}

// Metadata that is expressed as key/value pairs that disambiguate tracks.
type Attribute struct {
	// The name of a custom attribute for a track.
	Name string `xml:",attr"`

	// The value of a custom attribute for a track.
	Value string `xml:",attr"`
}

// An XML element that encapsulates metadata that is required by the client to
// play back protected content.
type Protection struct {
	ProtectionHeaders []*ProtectionHeader `xml:"ProtectionHeader"`
}

// An XML element that encapsulates content-protection metadata for a specific
// content-protection system.
type ProtectionHeader struct {
	// A UUID that uniquely identifies the Content Protection System to which
	// this ProtectionElement field pertains.
	SystemID uuid.UUID `xml:",attr"`

	// Opaque data that the Content Protection System that is identified in the
	// SystemID field can use to enable playback for authorized users, encoded
	// using base64 encoding [RFC3548].
	Content string `xml:",chardata"`
}

type StreamType string

const (
	VideoStream StreamType = "video"
	AudioStream StreamType = "audio"
	TextStream  StreamType = "text"
)

func ChunkURL(baseURL *url.URL, stream *StreamIndex, level *Track, startTime uint64) *url.URL {
	u := *baseURL
	c := *stream.URL
	bitrateStr := strconv.FormatUint(uint64(level.Bitrate), 10)
	starttimeStr := strconv.FormatUint(startTime, 10)
	c = strings.ReplaceAll(c, "{bitrate}", bitrateStr)
	c = strings.ReplaceAll(c, "{Bitrate}", bitrateStr)
	c = strings.ReplaceAll(c, "{start time}", starttimeStr)
	c = strings.ReplaceAll(c, "{start_time}", starttimeStr)
	u.Path = path.Join(path.Dir(u.Path), c)
	return &u
}
