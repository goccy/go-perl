package gperl

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"os"
	"strings"
)

// privatizeELF makes every symbol an ELF relocatable object defines,
// except keep, private to the object: each gets a name no other object
// has (its own name plus suffix), so objects that define the same names
// link into one binary. References inside the object go by symbol index
// and are unaffected; an undefined symbol (what the object needs from libc
// and friends) keeps its name, unless imports gives it another, or it is
// one of the SDK's references to the C library function a wrapper stands in
// for (libcRealPrefix + the function's name), which becomes that name, bound
// global so that a static link takes the function from the C library. It returns
// the names of the renamed symbols of default visibility: what the module's
// shared object exports.
//
// The file is edited in place: the renamed symbols point into a new string
// table appended to the file, and the string table's section header is
// redirected to it. Nothing else moves.
func privatizeELF(path, keep, suffix string, imports map[string]string) (exports []string, err error) {
	// GOPERL_LIBC in the SDK's perl.h: declared weak, so a loadable module,
	// which has no such symbol, loads and asks the dynamic linker instead.
	const libcRealPrefix = "goperl_libc_real_"
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, f.Close()) }()
	ef, err := elf.NewFile(f)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	if ef.Class != elf.ELFCLASS64 || ef.Type != elf.ET_REL {
		return nil, fmt.Errorf("%s: want a 64-bit relocatable ELF object, got %s %s", path, ef.Class, ef.Type)
	}
	var symtab *elf.Section
	for _, s := range ef.Sections {
		if s.Type == elf.SHT_SYMTAB {
			symtab = s
		}
	}
	if symtab == nil {
		return nil, fmt.Errorf("%s: no symbol table", path)
	}
	if int(symtab.Link) >= len(ef.Sections) {
		return nil, fmt.Errorf("%s: symbol table names a string table that does not exist", path)
	}
	strtab := ef.Sections[symtab.Link]
	names, err := strtab.Data()
	if err != nil {
		return nil, err
	}
	syms, err := symtab.Data()
	if err != nil {
		return nil, err
	}
	// Elf64_Sym: st_name u32, st_info u8, st_other u8, st_shndx u16,
	// st_value u64, st_size u64.
	const symSize = 24
	order := ef.ByteOrder
	renamed := 0
	for off := 0; off+symSize <= len(syms); off += symSize {
		nameOff := order.Uint32(syms[off:])
		bind := elf.SymBind(syms[off+4] >> 4)
		shndx := elf.SectionIndex(order.Uint16(syms[off+6:]))
		if (bind != elf.STB_GLOBAL && bind != elf.STB_WEAK) || int(nameOff) >= len(names) {
			continue
		}
		end := bytes.IndexByte(names[nameOff:], 0)
		if end < 0 {
			return nil, fmt.Errorf("%s: unterminated symbol name", path)
		}
		name := string(names[nameOff : int(nameOff)+end])
		newName := imports[name]
		if libc, ok := strings.CutPrefix(name, libcRealPrefix); ok && shndx == elf.SHN_UNDEF {
			newName = libc
			syms[off+4] = byte(elf.STB_GLOBAL)<<4 | syms[off+4]&0xf
		}
		if shndx != elf.SHN_UNDEF {
			if name == "" || name == keep {
				continue
			}
			newName = name + suffix
			if elf.SymVis(syms[off+5]&3) == elf.STV_DEFAULT {
				exports = append(exports, name)
			}
		}
		if newName == "" {
			continue
		}
		order.PutUint32(syms[off:], uint32(len(names)))
		names = append(append(names, newName...), 0)
		renamed++
	}
	if renamed == 0 {
		return exports, nil
	}
	if _, err := f.WriteAt(syms, int64(symtab.Offset)); err != nil {
		return nil, err
	}
	end, err := f.Seek(0, 2)
	if err != nil {
		return nil, err
	}
	if _, err := f.WriteAt(names, end); err != nil {
		return nil, err
	}
	// Redirect the string table's section header: Elf64_Shdr has sh_offset
	// at byte 24 and sh_size at byte 32.
	var ehdr [64]byte
	if _, err := f.ReadAt(ehdr[:], 0); err != nil {
		return nil, err
	}
	shoff := int64(order.Uint64(ehdr[0x28:]))
	shentsize := int64(order.Uint16(ehdr[0x3a:]))
	hdr := shoff + int64(symtab.Link)*shentsize
	var buf [16]byte
	order.PutUint64(buf[0:], uint64(end))
	order.PutUint64(buf[8:], uint64(len(names)))
	_, err = f.WriteAt(buf[:], hdr+24)
	return exports, err
}
