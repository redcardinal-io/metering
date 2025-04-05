package pagination

import (
	"encoding/base64"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestItem implements the Item interface for testing
type TestItem struct {
	ID        string
	CreatedAt time.Time
	Data      string
}

// Cursor implements the Item interface
func (i TestItem) Cursor() Cursor {
	return NewCursor(i.CreatedAt, i.ID)
}

// createTestItems generates n test items with descending timestamps
func createTestItems(n int) []TestItem {
	items := make([]TestItem, n)
	now := time.Now().UTC()

	for i := range items {
		items[i] = TestItem{
			ID:        fmt.Sprintf("item-%d", i+1),
			CreatedAt: now.Add(time.Duration(-i) * time.Hour),
			Data:      fmt.Sprintf("Data for item %d", i+1),
		}
	}

	return items
}

func TestNewCursor(t *testing.T) {
	t.Run("Creates cursor with UTC time", func(t *testing.T) {
		// Use a non-UTC time
		localTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.Local)
		id := "test-id"

		cursor := NewCursor(localTime, id)

		// Verify time is converted to UTC
		assert.Equal(t, localTime.UTC(), cursor.Time)
		assert.Equal(t, id, cursor.ID)
	})
}

func TestCursorValidate(t *testing.T) {
	t.Run("Valid cursor", func(t *testing.T) {
		cursor := Cursor{
			Time: time.Now(),
			ID:   "valid-id",
		}

		err := cursor.Validate()
		assert.NoError(t, err)
	})

	t.Run("Zero time", func(t *testing.T) {
		cursor := Cursor{
			Time: time.Time{},
			ID:   "id",
		}

		err := cursor.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "time is zero")
	})

	t.Run("Empty ID", func(t *testing.T) {
		cursor := Cursor{
			Time: time.Now(),
			ID:   "",
		}

		err := cursor.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id is empty")
	})

	t.Run("Multiple errors", func(t *testing.T) {
		cursor := Cursor{
			Time: time.Time{},
			ID:   "",
		}

		err := cursor.Validate()
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "time is zero")
		assert.Contains(t, err.Error(), "id is empty")
	})
}

func TestCursorEncodeDecode(t *testing.T) {
	t.Run("Encode and decode", func(t *testing.T) {
		originalCursor := Cursor{
			Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			ID:   "test-id",
		}

		encoded := originalCursor.Encode()

		decoded, err := DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, originalCursor.Time, decoded.Time)
		assert.Equal(t, originalCursor.ID, decoded.ID)
	})

	t.Run("Encode with ID containing delimiter", func(t *testing.T) {
		originalCursor := Cursor{
			Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			ID:   "test" + cursorDelimiter + "id",
		}

		encoded := originalCursor.Encode()

		decoded, err := DecodeCursor(encoded)
		require.NoError(t, err)
		assert.Equal(t, originalCursor.Time, decoded.Time)
		assert.Equal(t, originalCursor.ID, decoded.ID)
	})
}

func TestDecodeCursorErrors(t *testing.T) {
	t.Run("Empty string", func(t *testing.T) {
		_, err := DecodeCursor("")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "text is empty")
	})

	t.Run("Invalid base64", func(t *testing.T) {
		_, err := DecodeCursor("!invalid-base64!")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "decode cursor")
	})

	t.Run("No delimiter", func(t *testing.T) {
		encoded := base64.StdEncoding.EncodeToString([]byte("2023-01-01T12:00:00Z"))
		_, err := DecodeCursor(encoded)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "no delimiter found")
	})

	t.Run("Invalid time format", func(t *testing.T) {
		invalidTime := "invalid-time" + cursorDelimiter + "id"
		encoded := base64.StdEncoding.EncodeToString([]byte(invalidTime))
		_, err := DecodeCursor(encoded)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "parse cursor timestamp")
	})
}

func TestNewResult(t *testing.T) {
	t.Run("With items", func(t *testing.T) {
		items := []TestItem{
			{
				ID:        "1",
				CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
				Data:      "Item 1",
			},
			{
				ID:        "2",
				CreatedAt: time.Date(2023, 1, 2, 12, 0, 0, 0, time.UTC),
				Data:      "Item 2",
			},
			{
				ID:        "3",
				CreatedAt: time.Date(2023, 1, 3, 12, 0, 0, 0, time.UTC),
				Data:      "Item 3",
			},
		}

		result := NewResult(items)

		assert.Equal(t, items, result.Items, "Items should be stored without modification")

		require.NotNil(t, result.NextCursor, "NextCursor should not be nil")

		cursor, err := DecodeCursor(*result.NextCursor)
		require.NoError(t, err, "Cursor should decode without error")
		assert.Equal(t, items[2].CreatedAt.UTC(), cursor.Time, "Cursor time should match last item")
		assert.Equal(t, items[2].ID, cursor.ID, "Cursor ID should match last item")
	})

	t.Run("With empty items", func(t *testing.T) {
		result := NewResult([]TestItem{})

		assert.Empty(t, result.Items, "Items should be empty")

		assert.Nil(t, result.NextCursor, "NextCursor should be nil for empty results")
	})

	t.Run("With single item", func(t *testing.T) {
		item := TestItem{
			ID:        "single",
			CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			Data:      "Single item",
		}

		result := NewResult([]TestItem{item})

		assert.Equal(t, []TestItem{item}, result.Items, "Item should be stored without modification")

		require.NotNil(t, result.NextCursor, "NextCursor should not be nil")

		cursor, err := DecodeCursor(*result.NextCursor)
		require.NoError(t, err, "Cursor should decode without error")
		assert.Equal(t, item.CreatedAt.UTC(), cursor.Time, "Cursor time should match the item")
		assert.Equal(t, item.ID, cursor.ID, "Cursor ID should match the item")
	})
}

func TestEdgeCases(t *testing.T) {
	t.Run("Items with same timestamp", func(t *testing.T) {
		sameTime := time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC)
		items := []TestItem{
			{ID: "a", CreatedAt: sameTime, Data: "Item A"},
			{ID: "b", CreatedAt: sameTime, Data: "Item B"},
			{ID: "c", CreatedAt: sameTime, Data: "Item C"},
		}

		result := NewResult(items)

		assert.NotNil(t, result.NextCursor, "Should generate cursor with identical timestamps")

		cursor, err := DecodeCursor(*result.NextCursor)
		assert.NoError(t, err, "Should decode cursor without error")

		assert.Equal(t, sameTime, cursor.Time)
		assert.Equal(t, "c", cursor.ID)
	})

	t.Run("Items with different time zones", func(t *testing.T) {
		nyZone, _ := time.LoadLocation("America/New_York")
		tokyoZone, _ := time.LoadLocation("Asia/Tokyo")

		items := []TestItem{
			{ID: "ny", CreatedAt: time.Date(2023, 1, 1, 12, 0, 0, 0, nyZone), Data: "NY Item"},
			{ID: "utc", CreatedAt: time.Date(2023, 1, 1, 17, 0, 0, 0, time.UTC), Data: "UTC Item"},
			{ID: "tokyo", CreatedAt: time.Date(2023, 1, 2, 2, 0, 0, 0, tokyoZone), Data: "Tokyo Item"},
		}

		result := NewResult(items)

		cursor, err := DecodeCursor(*result.NextCursor)
		assert.NoError(t, err)

		assert.Equal(t, items[2].CreatedAt.UTC(), cursor.Time)
		assert.Equal(t, "tokyo", cursor.ID)
	})
}

func TestCursorMarshalUnmarshal(t *testing.T) {
	t.Run("Marshal and unmarshal", func(t *testing.T) {
		originalCursor := Cursor{
			Time: time.Date(2023, 1, 1, 12, 0, 0, 0, time.UTC),
			ID:   "test-id",
		}

		bytes, err := originalCursor.MarshalText()
		require.NoError(t, err)

		var decodedCursor Cursor
		err = decodedCursor.UnmarshalText(bytes)
		require.NoError(t, err)

		assert.Equal(t, originalCursor.Time, decodedCursor.Time)
		assert.Equal(t, originalCursor.ID, decodedCursor.ID)
	})
}

func BenchmarkPagination(b *testing.B) {
	items := createTestItems(100)

	b.Run("NewResult", func(b *testing.B) {
		b.ResetTimer()
		for b.Loop() {
			_ = NewResult(items)
		}
	})

	b.Run("Cursor encoding", func(b *testing.B) {
		cursor := items[0].Cursor()
		b.ResetTimer()
		for b.Loop() {
			_ = cursor.Encode()
		}
	})

	b.Run("Cursor decoding", func(b *testing.B) {
		encoded := items[0].Cursor().Encode()
		b.ResetTimer()
		for b.Loop() {
			_, _ = DecodeCursor(encoded)
		}
	})
}
