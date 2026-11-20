package xs

// EntrySymbol exposes the entry table's symbol name to the tests.
var EntrySymbol = entrySymbol

// Unregister removes a module Register added, so one test's linked module
// does not reach the instances of the others.
func Unregister(module string) {
	linkedMu.Lock()
	delete(linked, module)
	linkedMu.Unlock()
}
