package main

import (
	"testing"
	"time"
)

func TestMakeSnapshotProgressBarFields(t *testing.T) {
	snap := makeSnapshot(2, 250, 100, 0, "repair", 3, time.Now().UTC().Unix(), 0)
	if snap.ProgressPercent != 50 {
		t.Fatalf("expected progress percent 50, got %d", snap.ProgressPercent)
	}
	if snap.XPToNextLevel != 50 {
		t.Fatalf("expected xp_to_next_level 50, got %d", snap.XPToNextLevel)
	}
	if snap.ProgressMode != "repair" {
		t.Fatalf("expected repair mode, got %s", snap.ProgressMode)
	}
}

func TestPunishModeGetsWorseWithInactivity(t *testing.T) {
	now := time.Now().UTC().Unix()
	snapActive := makeSnapshot(3, 350, 0, 0, "punish", 5, now, 0)
	snapInactive := makeSnapshot(3, 350, 0, 0, "punish", 0, now-(72*60*60), 3)

	if snapInactive.UIBurdenScore <= snapActive.UIBurdenScore {
		t.Fatalf("expected inactive burden > active burden, got %d <= %d", snapInactive.UIBurdenScore, snapActive.UIBurdenScore)
	}
	if snapInactive.ForcedActionDelayMS <= snapActive.ForcedActionDelayMS {
		t.Fatalf("expected inactive delay > active delay, got %d <= %d", snapInactive.ForcedActionDelayMS, snapActive.ForcedActionDelayMS)
	}
}

func TestApplyInactivityPenaltyAppliedOncePerDay(t *testing.T) {
	now := time.Now().UTC()
	ps := playerState{
		Snapshot: makeSnapshot(2, 240, 0, 0, "punish", 4, now.Add(-48*time.Hour).Unix(), 2),
	}

	xpBefore := ps.Snapshot.XP
	ps.applyInactivityPenalty(now)
	if ps.Snapshot.XP >= xpBefore {
		t.Fatalf("expected xp decrease after penalty, before=%d after=%d", xpBefore, ps.Snapshot.XP)
	}

	xpAfterFirst := ps.Snapshot.XP
	ps.applyInactivityPenalty(now)
	if ps.Snapshot.XP != xpAfterFirst {
		t.Fatalf("expected no second penalty on same day, got %d vs %d", ps.Snapshot.XP, xpAfterFirst)
	}
}
