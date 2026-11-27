package xs

// EntrySymbol exposes the entry table's symbol name to the tests.
var EntrySymbol = entrySymbol

// LibcSymbol exposes how the runtime finds a C library function.
var LibcSymbol = libcSymbol

// Unregister removes a module Register added, so one test's linked module
// does not reach the instances of the others.
func Unregister(module string) {
	linkedMu.Lock()
	delete(linked, module)
	linkedMu.Unlock()
}
