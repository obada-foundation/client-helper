// Copyright 2018 The LevelDB-Go and Pebble Authors. All rights reserved. Use
// of this source code is governed by a BSD-style license that can be found in
// the LICENSE file.

package pebble

import (
	"fmt"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/pebble/internal/base"
	"github.com/cockroachdb/pebble/internal/humanize"
	"github.com/cockroachdb/pebble/internal/invariants"
	"github.com/cockroachdb/pebble/internal/manifest"
	"github.com/cockroachdb/pebble/vfs"
	"github.com/cockroachdb/redact"
)

// TableInfo exports the manifest.TableInfo type.
type TableInfo = manifest.TableInfo

func tablesTotalSize(tables []TableInfo) uint64 {
	var size uint64
	for i := range tables {
		size += tables[i].Size
	}
	return size
}

func formatFileNums(tables []TableInfo) string {
	var buf strings.Builder
	for i := range tables {
		if i > 0 {
			buf.WriteString(" ")
		}
		buf.WriteString(tables[i].FileNum.String())
	}
	return buf.String()
}

// LevelInfo contains info pertaining to a particular level.
type LevelInfo struct {
	Level  int
	Tables []TableInfo
	Score  float64
}

func (i LevelInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i LevelInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	w.Printf("L%d [%s] (%s) Score=%.2f",
		redact.Safe(i.Level),
		redact.Safe(formatFileNums(i.Tables)),
		redact.Safe(humanize.Bytes.Uint64(tablesTotalSize(i.Tables))),
		redact.Safe(i.Score))
}

// CompactionInfo contains the info for a compaction event.
type CompactionInfo struct {
	// JobID is the ID of the compaction job.
	JobID int
	// Reason is the reason for the compaction.
	Reason string
	// Input contains the input tables for the compaction organized by level.
	Input []LevelInfo
	// Output contains the output tables generated by the compaction. The output
	// tables are empty for the compaction begin event.
	Output LevelInfo
	// Duration is the time spent compacting, including reading and writing
	// sstables.
	Duration time.Duration
	// TotalDuration is the total wall-time duration of the compaction,
	// including applying the compaction to the database. TotalDuration is
	// always ≥ Duration.
	TotalDuration time.Duration
	Done          bool
	Err           error

	SingleLevelOverlappingRatio float64
	MultiLevelOverlappingRatio  float64

	// Annotations specifies additional info to appear in a compaction's event log line
	Annotations compactionAnnotations
}

type compactionAnnotations []string

// SafeFormat implements redact.SafeFormatter.
func (ca compactionAnnotations) SafeFormat(w redact.SafePrinter, _ rune) {
	if len(ca) == 0 {
		return
	}
	for i := range ca {
		if i != 0 {
			w.Print(" ")
		}
		w.Printf("%s", redact.SafeString(ca[i]))
	}
}

func (i CompactionInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i CompactionInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] compaction(%s) to L%d error: %s",
			redact.Safe(i.JobID), redact.SafeString(i.Reason), redact.Safe(i.Output.Level), i.Err)
		return
	}

	if !i.Done {
		w.Printf("[JOB %d] compacting(%s) ",
			redact.Safe(i.JobID),
			redact.SafeString(i.Reason))
		w.Printf("%s", i.Annotations)
		w.Printf("%s; ", levelInfos(i.Input))
		w.Printf("OverlappingRatio: Single %.2f, Multi %.2f", i.SingleLevelOverlappingRatio, i.MultiLevelOverlappingRatio)
		return
	}
	outputSize := tablesTotalSize(i.Output.Tables)
	w.Printf("[JOB %d] compacted(%s) ", redact.Safe(i.JobID), redact.SafeString(i.Reason))
	w.Printf("%s", i.Annotations)
	w.Print(levelInfos(i.Input))
	w.Printf(" -> L%d [%s] (%s), in %.1fs (%.1fs total), output rate %s/s",
		redact.Safe(i.Output.Level),
		redact.Safe(formatFileNums(i.Output.Tables)),
		redact.Safe(humanize.Bytes.Uint64(outputSize)),
		redact.Safe(i.Duration.Seconds()),
		redact.Safe(i.TotalDuration.Seconds()),
		redact.Safe(humanize.Bytes.Uint64(uint64(float64(outputSize)/i.Duration.Seconds()))))
}

type levelInfos []LevelInfo

func (i levelInfos) SafeFormat(w redact.SafePrinter, _ rune) {
	for j, levelInfo := range i {
		if j > 0 {
			w.Printf(" + ")
		}
		w.Print(levelInfo)
	}
}

// DiskSlowInfo contains the info for a disk slowness event when writing to a
// file.
type DiskSlowInfo = vfs.DiskSlowInfo

// FlushInfo contains the info for a flush event.
type FlushInfo struct {
	// JobID is the ID of the flush job.
	JobID int
	// Reason is the reason for the flush.
	Reason string
	// Input contains the count of input memtables that were flushed.
	Input int
	// InputBytes contains the total in-memory size of the memtable(s) that were
	// flushed. This size includes skiplist indexing data structures.
	InputBytes uint64
	// Output contains the ouptut table generated by the flush. The output info
	// is empty for the flush begin event.
	Output []TableInfo
	// Duration is the time spent flushing. This duration includes writing and
	// syncing all of the flushed keys to sstables.
	Duration time.Duration
	// TotalDuration is the total wall-time duration of the flush, including
	// applying the flush to the database. TotalDuration is always ≥ Duration.
	TotalDuration time.Duration
	// Ingest is set to true if the flush is handling tables that were added to
	// the flushable queue via an ingestion operation.
	Ingest bool
	// IngestLevels are the output levels for each ingested table in the flush.
	// This field is only populated when Ingest is true.
	IngestLevels []int
	Done         bool
	Err          error
}

func (i FlushInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i FlushInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] flush error: %s", redact.Safe(i.JobID), i.Err)
		return
	}

	plural := redact.SafeString("s")
	if i.Input == 1 {
		plural = ""
	}
	if !i.Done {
		w.Printf("[JOB %d] ", redact.Safe(i.JobID))
		if !i.Ingest {
			w.Printf("flushing %d memtable", redact.Safe(i.Input))
			w.SafeString(plural)
			w.Printf(" (%s) to L0", redact.Safe(humanize.Bytes.Uint64(i.InputBytes)))
		} else {
			w.Printf("flushing %d ingested table%s", redact.Safe(i.Input), plural)
		}
		return
	}

	outputSize := tablesTotalSize(i.Output)
	if !i.Ingest {
		if invariants.Enabled && len(i.IngestLevels) > 0 {
			panic(errors.AssertionFailedf("pebble: expected len(IngestedLevels) == 0"))
		}
		w.Printf("[JOB %d] flushed %d memtable%s (%s) to L0 [%s] (%s), in %.1fs (%.1fs total), output rate %s/s",
			redact.Safe(i.JobID), redact.Safe(i.Input), plural,
			redact.Safe(humanize.Bytes.Uint64(i.InputBytes)),
			redact.Safe(formatFileNums(i.Output)),
			redact.Safe(humanize.Bytes.Uint64(outputSize)),
			redact.Safe(i.Duration.Seconds()),
			redact.Safe(i.TotalDuration.Seconds()),
			redact.Safe(humanize.Bytes.Uint64(uint64(float64(outputSize)/i.Duration.Seconds()))))
	} else {
		if invariants.Enabled && len(i.IngestLevels) == 0 {
			panic(errors.AssertionFailedf("pebble: expected len(IngestedLevels) > 0"))
		}
		w.Printf("[JOB %d] flushed %d ingested flushable%s",
			redact.Safe(i.JobID), redact.Safe(len(i.Output)), plural)
		for j, level := range i.IngestLevels {
			file := i.Output[j]
			if j > 0 {
				w.Printf(" +")
			}
			w.Printf(" L%d:%s (%s)", level, file.FileNum, humanize.Bytes.Uint64(file.Size))
		}
		w.Printf(" in %.1fs (%.1fs total), output rate %s/s",
			redact.Safe(i.Duration.Seconds()),
			redact.Safe(i.TotalDuration.Seconds()),
			redact.Safe(humanize.Bytes.Uint64(uint64(float64(outputSize)/i.Duration.Seconds()))))
	}
}

// ManifestCreateInfo contains info about a manifest creation event.
type ManifestCreateInfo struct {
	// JobID is the ID of the job the caused the manifest to be created.
	JobID int
	Path  string
	// The file number of the new Manifest.
	FileNum base.DiskFileNum
	Err     error
}

func (i ManifestCreateInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i ManifestCreateInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] MANIFEST create error: %s", redact.Safe(i.JobID), i.Err)
		return
	}
	w.Printf("[JOB %d] MANIFEST created %s", redact.Safe(i.JobID), i.FileNum)
}

// ManifestDeleteInfo contains the info for a Manifest deletion event.
type ManifestDeleteInfo struct {
	// JobID is the ID of the job the caused the Manifest to be deleted.
	JobID   int
	Path    string
	FileNum FileNum
	Err     error
}

func (i ManifestDeleteInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i ManifestDeleteInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] MANIFEST delete error: %s", redact.Safe(i.JobID), i.Err)
		return
	}
	w.Printf("[JOB %d] MANIFEST deleted %s", redact.Safe(i.JobID), i.FileNum)
}

// TableCreateInfo contains the info for a table creation event.
type TableCreateInfo struct {
	JobID int
	// Reason is the reason for the table creation: "compacting", "flushing", or
	// "ingesting".
	Reason  string
	Path    string
	FileNum FileNum
}

func (i TableCreateInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i TableCreateInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	w.Printf("[JOB %d] %s: sstable created %s",
		redact.Safe(i.JobID), redact.Safe(i.Reason), i.FileNum)
}

// TableDeleteInfo contains the info for a table deletion event.
type TableDeleteInfo struct {
	JobID   int
	Path    string
	FileNum FileNum
	Err     error
}

func (i TableDeleteInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i TableDeleteInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] sstable delete error %s: %s",
			redact.Safe(i.JobID), i.FileNum, i.Err)
		return
	}
	w.Printf("[JOB %d] sstable deleted %s", redact.Safe(i.JobID), i.FileNum)
}

// TableIngestInfo contains the info for a table ingestion event.
type TableIngestInfo struct {
	// JobID is the ID of the job the caused the table to be ingested.
	JobID  int
	Tables []struct {
		TableInfo
		Level int
	}
	// GlobalSeqNum is the sequence number that was assigned to all entries in
	// the ingested table.
	GlobalSeqNum uint64
	// flushable indicates whether the ingested sstable was treated as a
	// flushable.
	flushable bool
	Err       error
}

func (i TableIngestInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i TableIngestInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] ingest error: %s", redact.Safe(i.JobID), i.Err)
		return
	}

	if i.flushable {
		w.Printf("[JOB %d] ingested as flushable", redact.Safe(i.JobID))
	} else {
		w.Printf("[JOB %d] ingested", redact.Safe(i.JobID))
	}

	for j := range i.Tables {
		t := &i.Tables[j]
		if j > 0 {
			w.Printf(",")
		}
		levelStr := ""
		if !i.flushable {
			levelStr = fmt.Sprintf("L%d:", t.Level)
		}
		w.Printf(" %s%s (%s)", redact.Safe(levelStr), t.FileNum,
			redact.Safe(humanize.Bytes.Uint64(t.Size)))
	}
}

// TableStatsInfo contains the info for a table stats loaded event.
type TableStatsInfo struct {
	// JobID is the ID of the job that finished loading the initial tables'
	// stats.
	JobID int
}

func (i TableStatsInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i TableStatsInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	w.Printf("[JOB %d] all initial table stats loaded", redact.Safe(i.JobID))
}

// TableValidatedInfo contains information on the result of a validation run
// on an sstable.
type TableValidatedInfo struct {
	JobID int
	Meta  *fileMetadata
}

func (i TableValidatedInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i TableValidatedInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	w.Printf("[JOB %d] validated table: %s", redact.Safe(i.JobID), i.Meta)
}

// WALCreateInfo contains info about a WAL creation event.
type WALCreateInfo struct {
	// JobID is the ID of the job the caused the WAL to be created.
	JobID int
	Path  string
	// The file number of the new WAL.
	FileNum base.DiskFileNum
	// The file number of a previous WAL which was recycled to create this
	// one. Zero if recycling did not take place.
	RecycledFileNum FileNum
	Err             error
}

func (i WALCreateInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i WALCreateInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] WAL create error: %s", redact.Safe(i.JobID), i.Err)
		return
	}

	if i.RecycledFileNum == 0 {
		w.Printf("[JOB %d] WAL created %s", redact.Safe(i.JobID), i.FileNum)
		return
	}

	w.Printf("[JOB %d] WAL created %s (recycled %s)",
		redact.Safe(i.JobID), i.FileNum, i.RecycledFileNum)
}

// WALDeleteInfo contains the info for a WAL deletion event.
type WALDeleteInfo struct {
	// JobID is the ID of the job the caused the WAL to be deleted.
	JobID   int
	Path    string
	FileNum FileNum
	Err     error
}

func (i WALDeleteInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i WALDeleteInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	if i.Err != nil {
		w.Printf("[JOB %d] WAL delete error: %s", redact.Safe(i.JobID), i.Err)
		return
	}
	w.Printf("[JOB %d] WAL deleted %s", redact.Safe(i.JobID), i.FileNum)
}

// WriteStallBeginInfo contains the info for a write stall begin event.
type WriteStallBeginInfo struct {
	Reason string
}

func (i WriteStallBeginInfo) String() string {
	return redact.StringWithoutMarkers(i)
}

// SafeFormat implements redact.SafeFormatter.
func (i WriteStallBeginInfo) SafeFormat(w redact.SafePrinter, _ rune) {
	w.Printf("write stall beginning: %s", redact.Safe(i.Reason))
}

// EventListener contains a set of functions that will be invoked when various
// significant DB events occur. Note that the functions should not run for an
// excessive amount of time as they are invoked synchronously by the DB and may
// block continued DB work. For a similar reason it is advisable to not perform
// any synchronous calls back into the DB.
type EventListener struct {
	// BackgroundError is invoked whenever an error occurs during a background
	// operation such as flush or compaction.
	BackgroundError func(error)

	// CompactionBegin is invoked after the inputs to a compaction have been
	// determined, but before the compaction has produced any output.
	CompactionBegin func(CompactionInfo)

	// CompactionEnd is invoked after a compaction has completed and the result
	// has been installed.
	CompactionEnd func(CompactionInfo)

	// DiskSlow is invoked after a disk write operation on a file created with a
	// disk health checking vfs.FS (see vfs.DefaultWithDiskHealthChecks) is
	// observed to exceed the specified disk slowness threshold duration. DiskSlow
	// is called on a goroutine that is monitoring slowness/stuckness. The callee
	// MUST return without doing any IO, or blocking on anything (like a mutex)
	// that is waiting on IO. This is imperative in order to reliably monitor for
	// slowness, since if this goroutine gets stuck, the monitoring will stop
	// working.
	DiskSlow func(DiskSlowInfo)

	// FlushBegin is invoked after the inputs to a flush have been determined,
	// but before the flush has produced any output.
	FlushBegin func(FlushInfo)

	// FlushEnd is invoked after a flush has complated and the result has been
	// installed.
	FlushEnd func(FlushInfo)

	// FormatUpgrade is invoked after the database's FormatMajorVersion
	// is upgraded.
	FormatUpgrade func(FormatMajorVersion)

	// ManifestCreated is invoked after a manifest has been created.
	ManifestCreated func(ManifestCreateInfo)

	// ManifestDeleted is invoked after a manifest has been deleted.
	ManifestDeleted func(ManifestDeleteInfo)

	// TableCreated is invoked when a table has been created.
	TableCreated func(TableCreateInfo)

	// TableDeleted is invoked after a table has been deleted.
	TableDeleted func(TableDeleteInfo)

	// TableIngested is invoked after an externally created table has been
	// ingested via a call to DB.Ingest().
	TableIngested func(TableIngestInfo)

	// TableStatsLoaded is invoked at most once, when the table stats
	// collector has loaded statistics for all tables that existed at Open.
	TableStatsLoaded func(TableStatsInfo)

	// TableValidated is invoked after validation runs on an sstable.
	TableValidated func(TableValidatedInfo)

	// WALCreated is invoked after a WAL has been created.
	WALCreated func(WALCreateInfo)

	// WALDeleted is invoked after a WAL has been deleted.
	WALDeleted func(WALDeleteInfo)

	// WriteStallBegin is invoked when writes are intentionally delayed.
	WriteStallBegin func(WriteStallBeginInfo)

	// WriteStallEnd is invoked when delayed writes are released.
	WriteStallEnd func()
}

// EnsureDefaults ensures that background error events are logged to the
// specified logger if a handler for those events hasn't been otherwise
// specified. Ensure all handlers are non-nil so that we don't have to check
// for nil-ness before invoking.
func (l *EventListener) EnsureDefaults(logger Logger) {
	if l.BackgroundError == nil {
		if logger != nil {
			l.BackgroundError = func(err error) {
				logger.Errorf("background error: %s", err)
			}
		} else {
			l.BackgroundError = func(error) {}
		}
	}
	if l.CompactionBegin == nil {
		l.CompactionBegin = func(info CompactionInfo) {}
	}
	if l.CompactionEnd == nil {
		l.CompactionEnd = func(info CompactionInfo) {}
	}
	if l.DiskSlow == nil {
		l.DiskSlow = func(info DiskSlowInfo) {}
	}
	if l.FlushBegin == nil {
		l.FlushBegin = func(info FlushInfo) {}
	}
	if l.FlushEnd == nil {
		l.FlushEnd = func(info FlushInfo) {}
	}
	if l.FormatUpgrade == nil {
		l.FormatUpgrade = func(v FormatMajorVersion) {}
	}
	if l.ManifestCreated == nil {
		l.ManifestCreated = func(info ManifestCreateInfo) {}
	}
	if l.ManifestDeleted == nil {
		l.ManifestDeleted = func(info ManifestDeleteInfo) {}
	}
	if l.TableCreated == nil {
		l.TableCreated = func(info TableCreateInfo) {}
	}
	if l.TableDeleted == nil {
		l.TableDeleted = func(info TableDeleteInfo) {}
	}
	if l.TableIngested == nil {
		l.TableIngested = func(info TableIngestInfo) {}
	}
	if l.TableStatsLoaded == nil {
		l.TableStatsLoaded = func(info TableStatsInfo) {}
	}
	if l.TableValidated == nil {
		l.TableValidated = func(validated TableValidatedInfo) {}
	}
	if l.WALCreated == nil {
		l.WALCreated = func(info WALCreateInfo) {}
	}
	if l.WALDeleted == nil {
		l.WALDeleted = func(info WALDeleteInfo) {}
	}
	if l.WriteStallBegin == nil {
		l.WriteStallBegin = func(info WriteStallBeginInfo) {}
	}
	if l.WriteStallEnd == nil {
		l.WriteStallEnd = func() {}
	}
}

// MakeLoggingEventListener creates an EventListener that logs all events to the
// specified logger.
func MakeLoggingEventListener(logger Logger) EventListener {
	if logger == nil {
		logger = DefaultLogger
	}

	return EventListener{
		BackgroundError: func(err error) {
			logger.Errorf("background error: %s", err)
		},
		CompactionBegin: func(info CompactionInfo) {
			logger.Infof("%s", info)
		},
		CompactionEnd: func(info CompactionInfo) {
			logger.Infof("%s", info)
		},
		DiskSlow: func(info DiskSlowInfo) {
			logger.Infof("%s", info)
		},
		FlushBegin: func(info FlushInfo) {
			logger.Infof("%s", info)
		},
		FlushEnd: func(info FlushInfo) {
			logger.Infof("%s", info)
		},
		FormatUpgrade: func(v FormatMajorVersion) {
			logger.Infof("upgraded to format version: %s", v)
		},
		ManifestCreated: func(info ManifestCreateInfo) {
			logger.Infof("%s", info)
		},
		ManifestDeleted: func(info ManifestDeleteInfo) {
			logger.Infof("%s", info)
		},
		TableCreated: func(info TableCreateInfo) {
			logger.Infof("%s", info)
		},
		TableDeleted: func(info TableDeleteInfo) {
			logger.Infof("%s", info)
		},
		TableIngested: func(info TableIngestInfo) {
			logger.Infof("%s", info)
		},
		TableStatsLoaded: func(info TableStatsInfo) {
			logger.Infof("%s", info)
		},
		TableValidated: func(info TableValidatedInfo) {
			logger.Infof("%s", info)
		},
		WALCreated: func(info WALCreateInfo) {
			logger.Infof("%s", info)
		},
		WALDeleted: func(info WALDeleteInfo) {
			logger.Infof("%s", info)
		},
		WriteStallBegin: func(info WriteStallBeginInfo) {
			logger.Infof("%s", info)
		},
		WriteStallEnd: func() {
			logger.Infof("write stall ending")
		},
	}
}

// TeeEventListener wraps two EventListeners, forwarding all events to both.
func TeeEventListener(a, b EventListener) EventListener {
	a.EnsureDefaults(nil)
	b.EnsureDefaults(nil)
	return EventListener{
		BackgroundError: func(err error) {
			a.BackgroundError(err)
			b.BackgroundError(err)
		},
		CompactionBegin: func(info CompactionInfo) {
			a.CompactionBegin(info)
			b.CompactionBegin(info)
		},
		CompactionEnd: func(info CompactionInfo) {
			a.CompactionEnd(info)
			b.CompactionEnd(info)
		},
		DiskSlow: func(info DiskSlowInfo) {
			a.DiskSlow(info)
			b.DiskSlow(info)
		},
		FlushBegin: func(info FlushInfo) {
			a.FlushBegin(info)
			b.FlushBegin(info)
		},
		FlushEnd: func(info FlushInfo) {
			a.FlushEnd(info)
			b.FlushEnd(info)
		},
		FormatUpgrade: func(v FormatMajorVersion) {
			a.FormatUpgrade(v)
			b.FormatUpgrade(v)
		},
		ManifestCreated: func(info ManifestCreateInfo) {
			a.ManifestCreated(info)
			b.ManifestCreated(info)
		},
		ManifestDeleted: func(info ManifestDeleteInfo) {
			a.ManifestDeleted(info)
			b.ManifestDeleted(info)
		},
		TableCreated: func(info TableCreateInfo) {
			a.TableCreated(info)
			b.TableCreated(info)
		},
		TableDeleted: func(info TableDeleteInfo) {
			a.TableDeleted(info)
			b.TableDeleted(info)
		},
		TableIngested: func(info TableIngestInfo) {
			a.TableIngested(info)
			b.TableIngested(info)
		},
		TableStatsLoaded: func(info TableStatsInfo) {
			a.TableStatsLoaded(info)
			b.TableStatsLoaded(info)
		},
		TableValidated: func(info TableValidatedInfo) {
			a.TableValidated(info)
			b.TableValidated(info)
		},
		WALCreated: func(info WALCreateInfo) {
			a.WALCreated(info)
			b.WALCreated(info)
		},
		WALDeleted: func(info WALDeleteInfo) {
			a.WALDeleted(info)
			b.WALDeleted(info)
		},
		WriteStallBegin: func(info WriteStallBeginInfo) {
			a.WriteStallBegin(info)
			b.WriteStallBegin(info)
		},
		WriteStallEnd: func() {
			a.WriteStallEnd()
			b.WriteStallEnd()
		},
	}
}
