package stage

import (
	"testing"
)

func TestStageManager_GetOrCreate(t *testing.T) {
	sm := NewStageManager()

	// First call should create a new stage
	stage1 := sm.GetOrCreate(10000)
	if stage1 == nil {
		t.Fatal("GetOrCreate should return a stage")
	}
	if stage1.MapID() != 10000 {
		t.Errorf("Expected map ID 10000, got %d", stage1.MapID())
	}

	// Second call should return the same stage
	stage2 := sm.GetOrCreate(10000)
	if stage1 != stage2 {
		t.Error("GetOrCreate should return the same stage for the same map ID")
	}

	// Different map should create different stage
	stage3 := sm.GetOrCreate(20000)
	if stage3 == stage1 {
		t.Error("Different map IDs should create different stages")
	}
}

func TestStageManager_Get(t *testing.T) {
	sm := NewStageManager()

	// Get non-existent stage should return nil
	stage := sm.Get(10000)
	if stage != nil {
		t.Error("Get should return nil for non-existent stage")
	}

	// Create stage then get it
	sm.GetOrCreate(10000)
	stage = sm.Get(10000)
	if stage == nil {
		t.Error("Get should return created stage")
	}
}

func TestStageManager_Remove(t *testing.T) {
	sm := NewStageManager()

	// Create and then remove
	sm.GetOrCreate(10000)
	sm.Remove(10000)

	stage := sm.Get(10000)
	if stage != nil {
		t.Error("Stage should be removed")
	}
}

func TestStageManager_GetAll(t *testing.T) {
	sm := NewStageManager()

	// Empty manager
	stages := sm.GetAll()
	if len(stages) != 0 {
		t.Errorf("Expected 0 stages, got %d", len(stages))
	}

	// Add some stages
	sm.GetOrCreate(10000)
	sm.GetOrCreate(20000)
	sm.GetOrCreate(30000)

	stages = sm.GetAll()
	if len(stages) != 3 {
		t.Errorf("Expected 3 stages, got %d", len(stages))
	}
}

func TestStageManager_Count(t *testing.T) {
	sm := NewStageManager()

	if sm.Count() != 0 {
		t.Errorf("Expected count 0, got %d", sm.Count())
	}

	sm.GetOrCreate(10000)
	if sm.Count() != 1 {
		t.Errorf("Expected count 1, got %d", sm.Count())
	}

	sm.GetOrCreate(20000)
	if sm.Count() != 2 {
		t.Errorf("Expected count 2, got %d", sm.Count())
	}
}

func TestStageManager_TotalUsers(t *testing.T) {
	sm := NewStageManager()

	// No users initially
	if sm.TotalUsers() != 0 {
		t.Errorf("Expected 0 users, got %d", sm.TotalUsers())
	}
}

func TestStageManager_CleanupEmpty(t *testing.T) {
	sm := NewStageManager()

	// Create some stages (they're empty by default)
	sm.GetOrCreate(10000)
	sm.GetOrCreate(20000)

	// Cleanup should remove all empty stages
	removed := sm.CleanupEmpty()
	if removed != 2 {
		t.Errorf("Expected 2 removed, got %d", removed)
	}

	if sm.Count() != 0 {
		t.Errorf("Expected 0 stages after cleanup, got %d", sm.Count())
	}
}

