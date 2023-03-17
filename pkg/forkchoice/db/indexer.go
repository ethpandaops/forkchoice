package db

import (
	"context"
	"database/sql"
	"errors"

	"github.com/ethpandaops/forkchoice/pkg/forkchoice/types"
	perrors "github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Indexer struct {
	db  *gorm.DB
	log logrus.FieldLogger
}

func NewIndexer(log logrus.FieldLogger, config IndexerConfig, dbConn ...*sql.DB) (*Indexer, error) {
	var db *gorm.DB

	var err error

	switch config.DriverName {
	case "postgres":
		conf := postgres.Config{
			DSN:        config.DSN,
			DriverName: "postgres",
		}

		if len(dbConn) > 0 {
			conf.Conn = dbConn[0]
		}

		dialect := postgres.New(conf)

		db, err = gorm.Open(dialect, &gorm.Config{})
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.DSN), &gorm.Config{})
	default:
		return nil, errors.New("invalid driver name: " + config.DriverName)
	}

	if err != nil {
		return nil, err
	}

	db = db.Session(&gorm.Session{FullSaveAssociations: true})

	err = db.AutoMigrate(&Frame{})
	if err != nil {
		return nil, perrors.Wrap(err, "failed to auto migrate frame")
	}

	err = db.AutoMigrate(&FrameLabel{})
	if err != nil {
		return nil, perrors.Wrap(err, "failed to auto migrate frame_label")
	}

	return &Indexer{
		db:  db,
		log: log.WithField("component", "indexer"),
	}, nil
}

func (i *Indexer) AddFrame(ctx context.Context, frame *types.Frame) error {
	var f Frame

	result := i.db.WithContext(ctx).Create(f.FromFrameMetadata(&frame.Metadata))

	return result.Error
}

func (i *Indexer) RemoveFrame(ctx context.Context, id string) error {
	result := i.db.WithContext(ctx).Where("id = ?", id).Delete(&Frame{})

	return result.Error
}

func (i *Indexer) ListFrames(ctx context.Context, filter *FrameFilter) ([]*Frame, error) {
	var frames []*Frame

	query := i.db.WithContext(ctx).Model(&Frame{})

	// Fetch frames that have ALL labels provided.
	if filter.Labels != nil {
		frameIDs, err := i.getFrameIDsWithLabels(ctx, *filter.Labels)
		if err != nil {
			return nil, err
		}

		query = query.Where("id IN (?)", frameIDs)
	}

	query, err := filter.ApplyToQuery(query)
	if err != nil {
		return nil, err
	}

	result := query.Preload("Labels").Order("fetched_at ASC").Find(&frames).Limit(1000)
	if result.Error != nil {
		return nil, result.Error
	}

	return frames, nil
}

func (i *Indexer) ListNodesWithFrames(ctx context.Context, filter *FrameFilter) ([]string, error) {
	var nodes []string

	query := i.db.WithContext(ctx).Model(&Frame{})

	// Fetch frames that have ALL labels provided.
	if filter.Labels != nil {
		frameIDs, err := i.getFrameIDsWithLabels(ctx, *filter.Labels)
		if err != nil {
			return nil, err
		}

		query = query.Where("id IN (?)", frameIDs)
	}

	query, err := filter.ApplyToQuery(query)
	if err != nil {
		return nil, err
	}

	result := query.Preload("Labels").Distinct("node").Find(&nodes).Limit(1000)
	if result.Error != nil {
		return nil, result.Error
	}

	return nodes, nil
}

func (i *Indexer) getFrameIDsWithLabels(ctx context.Context, labels []string) ([]string, error) {
	frameLabels := []*FrameLabel{}

	if err := i.db.Model(&FrameLabel{}).Where("name IN (?)", labels).Find(&frameLabels).Error; err != nil {
		return nil, err
	}

	frameLabelFrameIDs := map[string]int{}
	for _, frameLabel := range frameLabels {
		frameLabelFrameIDs[frameLabel.FrameID]++
	}

	frameIDs := []string{}

	for frameID, count := range frameLabelFrameIDs {
		// Check if the frame has all required labels.
		if len(labels) != count {
			continue
		}

		frameIDs = append(frameIDs, frameID)
	}

	return frameIDs, nil
}
