package xs

import (
	"fmt"
	"strings"
	"sync"
	"unsafe"
)

// entryNames lists, in slot order, what a module's entry table
// (sdk/include/goperl_entry.c) holds: the module's boot function, the SDK
// entry points the loader resolves by name, and the C library functions
// the loader calls.
var entryNames = []string{
	"", // boot_<Module>
	"__goperl_xs_init",
	"__goperl_xs_layout",
	"__goperl_pp_invoke",
	"__goperl_shared_clone",
	"__goperl_shared_release",
	"__goperl_svt_invoke",
	"__goperl_xs_invoke",
	"__goperl_xs_epilogue",
	"__goperl_keyword_invoke",
	"__goperl_perlio_invoke",
	"__goperl_dtor_invoke",
	"__goperl_host_invoke",
	"strlen",
	"memcpy",
	"malloc",
	"free",
}

// entrySlots is the number of slots in an entry table.
const entrySlots = 17

// entrySlotOf maps a name of entryNames to its slot.
var entrySlotOf = func() map[string]int {
	slots := make(map[string]int, len(entryNames))
	for i, name := range entryNames {
		slots[name] = i
	}
	return slots
}()

// mangledModule spells a Perl package the way its C symbols do
// ("Demo::XS" -> "Demo__XS").
func mangledModule(module string) string {
	return strings.ReplaceAll(module, "::", "__")
}

// entrySymbol names a module's entry table.
func entrySymbol(module string) string { return "goperl_entry_" + mangledModule(module) }

// entryResolver resolves the loader's symbols from a module's entry table.
func entryResolver(module string, entry unsafe.Pointer) func(string) (uintptr, error) {
	slots := (*[entrySlots]uintptr)(entry)
	boot := "boot_" + mangledModule(module)
	return func(name string) (uintptr, error) {
		slot, ok := entrySlotOf[name]
		if name == boot {
			slot, ok = 0, true
		}
		if !ok || name == "" || slots[slot] == 0 {
			return 0, fmt.Errorf("%s: no %s in the entry table", module, name)
		}
		return slots[slot], nil
	}
}

var (
	linkedMu sync.Mutex
	linked   = map[string]unsafe.Pointer{}
)

// Register makes a native XS module linked into the binary available to
// every Perl instance: module is the Perl package the .xs declared (e.g.
// "Cpanel::JSON::XS") and entry is the address of its entry table. It is
// called from the init function of the Go package that links the module's
// object, before any instance exists.
func Register(module string, entry unsafe.Pointer) {
	linkedMu.Lock()
	linked[module] = entry
	linkedMu.Unlock()
}

// linkedModules snapshots the registered modules.
func linkedModules() map[string]unsafe.Pointer {
	linkedMu.Lock()
	defer linkedMu.Unlock()
	out := make(map[string]unsafe.Pointer, len(linked))
	for module, entry := range linked {
		out[module] = entry
	}
	return out
}
