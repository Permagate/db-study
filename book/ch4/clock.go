package ch4

const (
	// Max counter that each data can have (so getting the same value won't increase the counter to absurdly high number
	MAX_COUNTER = 5
)

// Implementation of a naive non-concurrent safe map with a limited size
// and it is using clock algorithm as eviction policy
type ClockMap struct {
	currentIdx   int
	data         []*ClockMapData // The real data is stored in a slice
	keyToDataMap map[int]int     // Map key to the index in slice that holds the real data
	freeKeys     []int           // freeKeys stores list of keys that are free to use for storage, mainly filled after an entry deletion
}

func NewClockMap(size int) *ClockMap {
	return &ClockMap{
		currentIdx:   0,
		data:         make([]*ClockMapData, 0, size),
		keyToDataMap: map[int]int{},
		freeKeys:     make([]int, 0),
	}
}

func (cm *ClockMap) Get(key int) (interface{}, bool) {
	if idx, ok := cm.keyToDataMap[key]; ok {
		return cm.data[idx].Get(), true
	}
	return nil, false
}

func (cm *ClockMap) Set(key int, value interface{}) {
	if cm.Len() < cm.Cap() {
		// clock map still not full yet
		cm.data = append(cm.data, NewClockMapData(value))
		cm.keyToDataMap[key] = cm.Len() - 1

	} else if len(cm.freeKeys) > 0 {
		// clock map full, but has free keys
		freeKey := cm.freeKeys[0]
		cm.data[freeKey].Replace(value)
		cm.keyToDataMap[key] = freeKey
		cm.freeKeys = cm.freeKeys[1:]

	} else {
		// clock map full, evict according to clock rule
		evictedIdx := cm.nextEvictable()
		cm.data[evictedIdx].Replace(value)
		cm.keyToDataMap[key] = evictedIdx
	}
}

func (cm *ClockMap) Del(key int) bool {
	if idx, ok := cm.keyToDataMap[key]; ok {
		cm.freeKeys = append(cm.freeKeys, idx)
		// No need to evict the underlying data, it will be lazily replaced when a new entry coming in
		return true
	}
	return false
}

func (cm *ClockMap) Cap() int {
	return cap(cm.data)
}

func (cm *ClockMap) Len() int {
	return len(cm.data)
}

func (cm *ClockMap) nextEvictable() (evictableIdx int) {
	for cm.data[cm.currentIdx].Demote() {
		cm.currentIdx++
	}
	evictableIdx = cm.currentIdx
	cm.currentIdx += 1
	return
}

type ClockMapData struct {
	data    interface{}
	counter int // Incremented on every get (up until maxCounter), decremented on every round robin
}

func NewClockMapData(data interface{}) *ClockMapData {
	return &ClockMapData{data, 0}
}

func (cmd *ClockMapData) Get() interface{} {
	cmd.Promote()
	return cmd.data
}

func (cmd *ClockMapData) Set(newData interface{}) {
	cmd.Promote()
	cmd.data = newData
}

func (cmd *ClockMapData) Replace(newData interface{}) {
	cmd.counter = 0
	cmd.data = newData
}

func (cmd *ClockMapData) Promote() {
	if cmd.counter < MAX_COUNTER {
		cmd.counter++
	}
}

func (cmd *ClockMapData) Demote() bool {
	if cmd.counter == 0 {
		return false
	}
	cmd.counter -= 1
	return true
}
