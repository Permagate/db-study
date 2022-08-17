package ch4

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClockMap(t *testing.T) {
	clock := NewClockMap(3)
	requireClockState(t, clock, 0, 0, 3, nil, nil, []int{})

	{
		// Fill up the clocks
		clock.Set(123, "item1")
		requireClockState(t, clock, 0, 1, 3,
			[]interface{}{"item1"},
			[]int{0},
			[]int{})

		clock.Set(456, "item2")
		requireClockState(t, clock, 0, 2, 3,
			[]interface{}{"item1", "item2"},
			[]int{0, 0},
			[]int{})

		clock.Set(789, "item3")
		requireClockState(t, clock, 0, 3, 3,
			[]interface{}{"item1", "item2", "item3"},
			[]int{0, 0, 0},
			[]int{})
	}

	{
		// Multiple gets
		getVal, getOk := clock.Get(123)
		require.True(t, getOk)
		require.Equal(t, "item1", getVal)

		for i := 0; i < 3; i++ {
			getVal, getOk = clock.Get(789)
			require.True(t, getOk)
			require.Equal(t, "item3", getVal)
		}

		getVal, getOk = clock.Get(123123123)
		require.False(t, getOk)
		require.Nil(t, getVal)

		requireClockState(t, clock, 0, 3, 3,
			[]interface{}{"item1", "item2", "item3"},
			[]int{1, 0, 3},
			[]int{})
	}

	{
		// Set new item and evict an old item
		clock.Set(321, "item4")
		requireClockState(t, clock, 2, 3, 3,
			[]interface{}{"item1", "item4", "item3"},
			[]int{0, 0, 3},
			[]int{})

		clock.Set(654, "item5")
		requireClockState(t, clock, 1, 3, 3,
			[]interface{}{"item5", "item4", "item3"},
			[]int{0, 0, 2},
			[]int{})

		// Promote item with key 321
		getVal, getOk := clock.Get(321)
		require.True(t, getOk)
		require.Equal(t, "item4", getVal)
		requireClockState(t, clock, 1, 3, 3,
			[]interface{}{"item5", "item4", "item3"},
			[]int{0, 1, 2},
			[]int{})

		clock.Set(987, "item6")
		requireClockState(t, clock, 1, 3, 3,
			[]interface{}{"item6", "item4", "item3"},
			[]int{0, 0, 1},
			[]int{})
	}

	{
		// Delete items and then fill it up again
		require.False(t, clock.Del(123123123))

		require.True(t, clock.Del(789))
		requireClockState(t, clock, 1, 2, 3,
			[]interface{}{"item6", "item4", "item3"},
			[]int{0, 0, 1},
			[]int{2})

		require.True(t, clock.Del(987))
		requireClockState(t, clock, 1, 1, 3,
			[]interface{}{"item6", "item4", "item3"},
			[]int{0, 0, 1},
			[]int{2, 0})

		clock.Set(111, "item7")
		requireClockState(t, clock, 1, 2, 3,
			[]interface{}{"item6", "item4", "item7"},
			[]int{0, 0, 0},
			[]int{0})

		clock.Set(222, "item8")
		requireClockState(t, clock, 1, 3, 3,
			[]interface{}{"item8", "item4", "item7"},
			[]int{0, 0, 0},
			[]int{})

		clock.Set(333, "item9")
		requireClockState(t, clock, 2, 3, 3,
			[]interface{}{"item8", "item9", "item7"},
			[]int{0, 0, 0},
			[]int{})
	}
}

func TestClockMapData(t *testing.T) {
	data := NewClockMapData("anything")

	{
		// Get promotes the Data
		require.Equal(t, "anything", data.Get())
		require.Equal(t, 1, data.counter)
	}

	{
		// Replace resets the counter back to 0
		data.Replace("replaced")
		require.Equal(t, 0, data.counter)
	}

	{
		// counter keeps getting incremented with multiple Get calls
		for i := 1; i <= MAX_COUNTER-1; i++ {
			require.Equal(t, "replaced", data.Get())
			require.Equal(t, i, data.counter)
		}
	}

	{
		// Set should also promote the counter instead of resetting to 0 like with Replace
		data.Set("updated")
		require.Equal(t, MAX_COUNTER, data.counter)
	}

	{
		// counter maxed at MAX_COUNTER no matter what
		for i := 0; i <= 100; i++ {
			require.Equal(t, "updated", data.Get())
			require.Equal(t, MAX_COUNTER, data.counter)
		}
	}

	{
		// Data can get demoted until 0
		for i := MAX_COUNTER - 1; i >= 0; i-- {
			require.True(t, data.Demote())
			require.Equal(t, i, data.counter)
		}
	}

	{
		// Data can no longer be demoted with counter at 0
		require.False(t, data.Demote())
		require.Equal(t, 0, data.counter)
	}
}

func requireClockState(t *testing.T, clock *ClockMap, expectedClockHandIdx, expectedLen, expectedCap int, expectedValues []interface{}, expectedCounters, expectedFreeKeys []int) {
	require.Equal(t, expectedClockHandIdx, clock.clockHandIdx)
	require.Equal(t, expectedLen, clock.Len())
	require.Equal(t, expectedCap, clock.Cap())
	for i := 0; i < len(expectedValues); i++ {
		require.Equal(t, expectedValues[i], clock.data[i].data)
		require.Equal(t, expectedCounters[i], clock.data[i].counter)
	}
	require.Equal(t, expectedFreeKeys, clock.freeKeys)
}
