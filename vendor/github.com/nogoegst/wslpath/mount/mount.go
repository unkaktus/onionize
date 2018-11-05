// +build linux

package mount

// FilterFunc is a type defining a callback function
// to filter out unwanted entries. It takes a pointer
// to an Info struct, and returns two booleans:
//  - skip: true if the entry should be skipped
//  - stop: true if parsing should be stopped after the entry
type FilterFunc func(*Info) (skip, stop bool)

// GetMounts retrieves a list of mounts for the current running process,
// with an optional filter applied (use nil for no filter).
func GetMounts(f FilterFunc) ([]*Info, error) {
	return parseMountTable(f)
}
