package structs

import (
	"io"
	"time"
)

type Downloadable interface {
	Download() (io.ReadCloser, error)
}

type MediaType int

const (
	Image MediaType = iota
	Video
	Audio
	Other
)

func (m MediaType) String() string {
	switch m {
	case Image:
		return "image"
	case Video:
		return "video"
	case Audio:
		return "audio"
	default:
		return "other"
	}
}

func ParseMediaType(s string) (MediaType, error) {
	switch s {
	case "image":
		return Image, nil
	case "video":
		return Video, nil
	case "audio":
		return Audio, nil
	default:
		return Other, nil
	}
}

type MediaFile struct {
	Id       string        `bson:"id" json:"id"`
	Type     MediaType     `bson:"type" json:"type"`
	MediaId  string        `bson:"media_id" json:"media_id"`
	Metadata *FileMetadata `bson:"metadata" json:"metadata"`
}

type FileMetadata struct {
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	CreatedBy string    `bson:"createdBy" json:"createdBy"`
	Size      int64     `bson:"size" json:"size"`
	RoomIds   []string  `bson:"roomIds" json:"roomIds"`
}
