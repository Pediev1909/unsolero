package mediaobject

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"regexp"
	"strings"

	catalog "rigmark/internal/modules/catalog/domain"
)

const MaximumImageBytes = 5 * 1024 * 1024

var (
	objectNamePattern = regexp.MustCompile(`^([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})_([a-f0-9]{64})(\.(jpg|png|webp))$`)
	productIDPattern  = regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)
)

type Description struct {
	Name        string
	ObjectKey   string
	ProductID   catalog.ProductID
	Digest      string
	Extension   string
	ContentType string
}

func Describe(productID catalog.ProductID, data []byte, extension string) (Description, error) {
	if !productIDPattern.MatchString(string(productID)) || !ValidBytes(data, extension) {
		return Description{}, errors.New("invalid image object")
	}
	digestBytes := sha256.Sum256(data)
	digest := hex.EncodeToString(digestBytes[:])
	name := string(productID) + "_" + digest + extension
	return Description{
		Name: name, ObjectKey: "products/" + string(productID) + "/" + digest + extension,
		ProductID: productID, Digest: digest, Extension: extension, ContentType: ContentType(extension),
	}, nil
}

func Parse(name string) (Description, error) {
	matches := objectNamePattern.FindStringSubmatch(name)
	if len(matches) != 5 || strings.ContainsAny(name, `/\\`) {
		return Description{}, errors.New("invalid image object name")
	}
	productID := catalog.ProductID(matches[1])
	extension := matches[3]
	return Description{
		Name: name, ObjectKey: "products/" + string(productID) + "/" + matches[2] + extension,
		ProductID: productID, Digest: matches[2], Extension: extension, ContentType: ContentType(extension),
	}, nil
}

func ParseObjectKey(key string) (Description, error) {
	if !strings.HasPrefix(key, "products/") || strings.ContainsAny(key, `\\`) {
		return Description{}, errors.New("invalid image object key")
	}
	parts := strings.Split(key, "/")
	if len(parts) != 3 || !productIDPattern.MatchString(parts[1]) {
		return Description{}, errors.New("invalid image object key")
	}
	name := parts[1] + "_" + parts[2]
	description, err := Parse(name)
	if err != nil || description.ObjectKey != key {
		return Description{}, errors.New("invalid image object key")
	}
	return description, nil
}

func BelongsTo(productID catalog.ProductID, name string) bool {
	description, err := Parse(name)
	return err == nil && description.ProductID == productID
}

func ContentType(extension string) string {
	return map[string]string{".jpg": "image/jpeg", ".png": "image/png", ".webp": "image/webp"}[extension]
}

func ValidBytes(data []byte, extension string) bool {
	if len(data) == 0 || len(data) > MaximumImageBytes {
		return false
	}
	switch extension {
	case ".jpg":
		return len(data) >= 5 && bytes.Equal(data[:3], []byte{0xff, 0xd8, 0xff}) && bytes.Equal(data[len(data)-2:], []byte{0xff, 0xd9})
	case ".png":
		return len(data) >= 16 && bytes.Equal(data[:8], []byte("\x89PNG\r\n\x1a\n"))
	case ".webp":
		return len(data) >= 12 && bytes.Equal(data[:4], []byte("RIFF")) &&
			bytes.Equal(data[8:12], []byte("WEBP")) && int(binary.LittleEndian.Uint32(data[4:8]))+8 == len(data)
	default:
		return false
	}
}
