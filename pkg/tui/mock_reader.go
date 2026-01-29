package tui

import "time"

// MockEventReader is an EventReader for testing.
type MockEventReader struct {
	events []Event
	index  int
}

// Ensure MockEventReader implements EventReader and InterruptibleReader.
var _ EventReader = (*MockEventReader)(nil)
var _ InterruptibleReader = (*MockEventReader)(nil)

// NewMockEventReader creates a MockEventReader with the given events.
// Events are returned in order by successive calls to PollEvent.
func NewMockEventReader(events ...Event) *MockEventReader {
	return &MockEventReader{
		events: events,
		index:  0,
	}
}

// PollEvent returns the next queued event, ignoring the timeout.
// Returns (nil, false) when all events have been consumed.
func (m *MockEventReader) PollEvent(timeout time.Duration) (Event, bool) {
	if m.index >= len(m.events) {
		return nil, false
	}
	ev := m.events[m.index]
	m.index++
	return ev, true
}

// Close is a no-op for the mock reader.
func (m *MockEventReader) Close() error {
	return nil
}

// Reset resets the reader to return events from the beginning.
func (m *MockEventReader) Reset() {
	m.index = 0
}

// AddEvents adds more events to the queue.
func (m *MockEventReader) AddEvents(events ...Event) {
	m.events = append(m.events, events...)
}

// Remaining returns the number of events yet to be returned.
func (m *MockEventReader) Remaining() int {
	return len(m.events) - m.index
}

// EnableInterrupt is a no-op for the mock reader.
func (m *MockEventReader) EnableInterrupt() error {
	return nil
}

// Interrupt is a no-op for the mock reader.
func (m *MockEventReader) Interrupt() error {
	return nil
}
