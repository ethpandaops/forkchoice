package types

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io/ioutil"
	"time"

	v1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec/phase0"
)

type FrameMetadata struct {
	// Node is the node that provided the frame.
	// In the case of a beacon node, this is the beacon node's ID as specified in the config for this service.
	// In the case of Xatu, this is the Xatu Sentry ID.
	Node string `json:"node"`
	// ID is the ID of the frame.
	ID string `json:"id"`
	// FetchedAt is the time the frame was fetched.
	FetchedAt time.Time `json:"fetched_at"`
	// WallClockSlot is the wall clock slot at the time the frame was fetched.
	WallClockSlot phase0.Slot `json:"wall_clock_slot"`
	// WallClockEpoch is the wall clock epoch at the time the frame was fetched.
	WallClockEpoch phase0.Epoch `json:"wall_clock_epoch"`
}

func (f *FrameMetadata) Validate() error {
	if f.Node == "" {
		return errors.New("invalid node")
	}

	if f.ID == "" {
		return errors.New("invalid ID")
	}

	if f.FetchedAt.IsZero() {
		return errors.New("invalid fetched_at")
	}

	if f.WallClockSlot == 0 {
		return errors.New("invalid wall clock slot")
	}

	if f.WallClockEpoch == 0 {
		return errors.New("invalid wall clock epoch")
	}

	return nil
}

// Frame holds a fork choice dump with a timestamp.
type Frame struct {
	// Data is the fork choice dump.
	Data *v1.ForkChoice `json:"data"`
	// Metadata is the metadata of the frame.
	Metadata FrameMetadata `json:"metadata"`
}

// FrameFilter is a filter for frames.
type FrameFilter struct {
	// WallClockSlot is the wall clock slot to filter on.
	WallClockSlot *phase0.Slot `json:"wall_clock_slot"`
	// WallClockEpoch is the wall clock epoch to filter on.
	WallClockEpoch *phase0.Epoch `json:"wall_clock_epoch"`
}

func (f *Frame) AsGzipJSON() ([]byte, error) {
	// Convert to JSON
	asJSON, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}

	// Gzip before writing to disk
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err = gz.Write(asJSON)
	if err != nil {
		return nil, err
	}

	if err := gz.Close(); err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (f *Frame) FromGzipJSON(data []byte) error {
	// Unzip
	r, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer r.Close()

	// Read the uncompressed data
	uncompressed, err := ioutil.ReadAll(r)
	if err != nil {
		return err
	}

	// Convert from JSON
	var returnFile Frame

	err = json.Unmarshal(uncompressed, &returnFile)
	if err != nil {
		return err
	}

	f.Data = returnFile.Data
	f.Metadata = returnFile.Metadata

	return nil
}
