/* Flag and type constants of the embedded perl, for the names perl.h
 * does not define itself. Generated from the interpreter's own headers
 * (sv.h, mg_vtable.h, cop.h, hv.h, utf8.h, warnings.h, ...): the values
 * cross to the interpreter unchanged, so they must be its own. */
#ifndef GOPERL_XS_SDK_CONSTS_H
#define GOPERL_XS_SDK_CONSTS_H
#ifndef CVf_ANONCONST
#define CVf_ANONCONST 0x20000
#endif
#ifndef CVf_AUTOLOAD
#define CVf_AUTOLOAD 0x2000
#endif
#ifndef CVf_CLONE
#define CVf_CLONE 0x0020
#endif
#ifndef CVf_CLONED
#define CVf_CLONED 0x0040
#endif
#ifndef CVf_CONST
#define CVf_CONST 0x0004
#endif
#ifndef CVf_CVGV_RC
#define CVf_CVGV_RC 0x0400
#endif
#ifndef CVf_DYNFILE
#define CVf_DYNFILE 0x1000
#endif
#ifndef CVf_EVAL_COMPILED
#define CVf_EVAL_COMPILED 0x400000
#endif
#ifndef CVf_HASEVAL
#define CVf_HASEVAL 0x4000
#endif
#ifndef CVf_ISXSUB
#define CVf_ISXSUB 0x0008
#endif
#ifndef CVf_IsMETHOD
#define CVf_IsMETHOD 0x100000
#endif
#ifndef CVf_LEXICAL
#define CVf_LEXICAL 0x10000
#endif
#ifndef CVf_LVALUE
#define CVf_LVALUE 0x0002
#endif
#ifndef CVf_NAMED
#define CVf_NAMED 0x8000
#endif
#ifndef CVf_NODEBUG
#define CVf_NODEBUG 0x0200
#endif
#ifndef CVf_NOWARN_AMBIGUOUS
#define CVf_NOWARN_AMBIGUOUS 0x0001
#endif
#ifndef CVf_REFCOUNTED_ANYSV
#define CVf_REFCOUNTED_ANYSV 0x80000
#endif
#ifndef CVf_SIGNATURE
#define CVf_SIGNATURE 0x40000
#endif
#ifndef CVf_SLABBED
#define CVf_SLABBED 0x0800
#endif
#ifndef CVf_UNIQUE
#define CVf_UNIQUE 0x0100
#endif
#ifndef CVf_WEAKOUTSIDE
#define CVf_WEAKOUTSIDE 0x0010
#endif
#ifndef CVf_XS_RCSTACK
#define CVf_XS_RCSTACK 0x200000
#endif
#ifndef CXp_EVALBLOCK
#define CXp_EVALBLOCK 0x40
#endif
#ifndef CXp_FINALLY
#define CXp_FINALLY 0x20
#endif
#ifndef CXp_FOR_DEF
#define CXp_FOR_DEF 0x10
#endif
#ifndef CXp_FOR_GV
#define CXp_FOR_GV 0x40
#endif
#ifndef CXp_FOR_LVREF
#define CXp_FOR_LVREF 0x20
#endif
#ifndef CXp_FOR_PAD
#define CXp_FOR_PAD 0x80
#endif
#ifndef CXp_HASARGS
#define CXp_HASARGS 0x20
#endif
#ifndef CXp_MULTICALL
#define CXp_MULTICALL 0x10
#endif
#ifndef CXp_ONCE
#define CXp_ONCE 0x10
#endif
#ifndef CXp_REAL
#define CXp_REAL 0x20
#endif
#ifndef CXp_SUB_RE
#define CXp_SUB_RE 0x40
#endif
#ifndef CXp_SUB_RE_FAKE
#define CXp_SUB_RE_FAKE 0x80
#endif
#ifndef CXp_TRY
#define CXp_TRY 0x80
#endif
#ifndef EVAL_INEVAL
#define EVAL_INEVAL 1
#endif
#ifndef EVAL_INREQUIRE
#define EVAL_INREQUIRE 8
#endif
#ifndef EVAL_KEEPERR
#define EVAL_KEEPERR 4
#endif
#ifndef EVAL_NULL
#define EVAL_NULL 0
#endif
#ifndef EVAL_RE_REPARSING
#define EVAL_RE_REPARSING 0x10
#endif
#ifndef EVAL_WARNONLY
#define EVAL_WARNONLY 2
#endif
#ifndef GV_ADDMG
#define GV_ADDMG 0x400
#endif
#ifndef GV_AUTOLOAD
#define GV_AUTOLOAD 0x100
#endif
#ifndef GV_AUTOLOAD_ISMETHOD
#define GV_AUTOLOAD_ISMETHOD 1
#endif
#ifndef GV_CACHE_ONLY
#define GV_CACHE_ONLY 0x1000
#endif
#ifndef GV_CROAK
#define GV_CROAK 0x200
#endif
#ifndef GV_NOADD_NOINIT
#define GV_NOADD_NOINIT 0x20
#endif
#ifndef GV_NOEXPAND
#define GV_NOEXPAND 0x40
#endif
#ifndef GV_NOINIT
#define GV_NOINIT 0x10
#endif
#ifndef GV_NOTQUAL
#define GV_NOTQUAL 0x80
#endif
#ifndef GV_NOUNIVERSAL
#define GV_NOUNIVERSAL 0x2000
#endif
#ifndef GV_NO_SVGMAGIC
#define GV_NO_SVGMAGIC 0x800
#endif
#ifndef GV_SUPER
#define GV_SUPER 0x1000
#endif
#ifndef GVf_ASSUMECV
#define GVf_ASSUMECV 0x04
#endif
#ifndef GVf_IMPORTED
#define GVf_IMPORTED 0xF0
#endif
#ifndef GVf_IMPORTED_AV
#define GVf_IMPORTED_AV 0x20
#endif
#ifndef GVf_IMPORTED_CV
#define GVf_IMPORTED_CV 0x80
#endif
#ifndef GVf_IMPORTED_HV
#define GVf_IMPORTED_HV 0x40
#endif
#ifndef GVf_IMPORTED_SV
#define GVf_IMPORTED_SV 0x10
#endif
#ifndef GVf_INTRO
#define GVf_INTRO 0x01
#endif
#ifndef GVf_MULTI
#define GVf_MULTI 0x02
#endif
#ifndef GVf_ONCE_FATAL
#define GVf_ONCE_FATAL 0x100
#endif
#ifndef GVf_RESERVED
#define GVf_RESERVED 0x08
#endif
#ifndef G_FAKINGEVAL
#define G_FAKINGEVAL 0x100
#endif
#ifndef G_METHOD_NAMED
#define G_METHOD_NAMED 0x1000
#endif
#ifndef G_NODEBUG
#define G_NODEBUG 0x40
#endif
#ifndef G_RETHROW
#define G_RETHROW 0x2000
#endif
#ifndef G_RE_REPARSING
#define G_RE_REPARSING 0x800
#endif
#ifndef G_UNDEF_FILL
#define G_UNDEF_FILL 0x200
#endif
#ifndef G_USEHINTS
#define G_USEHINTS 0x4000
#endif
#ifndef G_WARN_ALL_OFF
#define G_WARN_ALL_OFF 4
#endif
#ifndef G_WARN_ALL_ON
#define G_WARN_ALL_ON 2
#endif
#ifndef G_WARN_OFF
#define G_WARN_OFF 0
#endif
#ifndef G_WARN_ONCE
#define G_WARN_ONCE 8
#endif
#ifndef G_WRITING_TO_STDERR
#define G_WRITING_TO_STDERR 0x400
#endif
#ifndef HINT_ASCII_ENCODING
#define HINT_ASCII_ENCODING 0x00000010
#endif
#ifndef HINT_EXPLICIT_STRICT_REFS
#define HINT_EXPLICIT_STRICT_REFS 0x00000020
#endif
#ifndef HINT_EXPLICIT_STRICT_SUBS
#define HINT_EXPLICIT_STRICT_SUBS 0x00000040
#endif
#ifndef HINT_EXPLICIT_STRICT_VARS
#define HINT_EXPLICIT_STRICT_VARS 0x00000080
#endif
#ifndef HINT_FEATURE_MASK
#define HINT_FEATURE_MASK 0x3c000000
#endif
#ifndef HINT_FILETEST_ACCESS
#define HINT_FILETEST_ACCESS 0x00400000
#endif
#ifndef HINT_INTEGER
#define HINT_INTEGER 0x00000001
#endif
#ifndef HINT_LEXICAL_IO_IN
#define HINT_LEXICAL_IO_IN 0x00040000
#endif
#ifndef HINT_LEXICAL_IO_OUT
#define HINT_LEXICAL_IO_OUT 0x00080000
#endif
#ifndef HINT_LOCALE
#define HINT_LOCALE 0x00000004
#endif
#ifndef HINT_LOCALIZE_HH
#define HINT_LOCALIZE_HH 0x00020000
#endif
#ifndef HINT_NEW_BINARY
#define HINT_NEW_BINARY 0x00004000
#endif
#ifndef HINT_NEW_FLOAT
#define HINT_NEW_FLOAT 0x00002000
#endif
#ifndef HINT_NEW_INTEGER
#define HINT_NEW_INTEGER 0x00001000
#endif
#ifndef HINT_NEW_RE
#define HINT_NEW_RE 0x00010000
#endif
#ifndef HINT_NEW_STRING
#define HINT_NEW_STRING 0x00008000
#endif
#ifndef HINT_NO_AMAGIC
#define HINT_NO_AMAGIC 0x01000000
#endif
#ifndef HINT_RE_EVAL
#define HINT_RE_EVAL 0x00200000
#endif
#ifndef HINT_RE_FLAGS
#define HINT_RE_FLAGS 0x02000000
#endif
#ifndef HINT_RE_TAINT
#define HINT_RE_TAINT 0x00100000
#endif
#ifndef HINT_STRICT_SUBS
#define HINT_STRICT_SUBS 0x00000200
#endif
#ifndef HINT_STRICT_VARS
#define HINT_STRICT_VARS 0x00000400
#endif
#ifndef HINT_UNI_8_BIT
#define HINT_UNI_8_BIT 0x00000800
#endif
#ifndef HV_DISABLE_UVAR_XKEY
#define HV_DISABLE_UVAR_XKEY 0x01
#endif
#ifndef HV_FETCH_EMPTY_HE
#define HV_FETCH_EMPTY_HE 0x80
#endif
#ifndef HV_ITERNEXT_WANTPLACEHOLDERS
#define HV_ITERNEXT_WANTPLACEHOLDERS 0x01
#endif
#ifndef HV_NAME_SETALL
#define HV_NAME_SETALL 0x02
#endif
#ifndef HVhek_FREEKEY
#define HVhek_FREEKEY 0x100
#endif
#ifndef HVhek_KEYCANONICAL
#define HVhek_KEYCANONICAL 0x400
#endif
#ifndef HVhek_NOTSHARED
#define HVhek_NOTSHARED 0x04
#endif
#ifndef HVhek_PLACEHOLD
#define HVhek_PLACEHOLD 0x200
#endif
#ifndef HVhek_WASUTF8
#define HVhek_WASUTF8 0x02
#endif
#ifndef IOf_ARGV
#define IOf_ARGV 1
#endif
#ifndef IOf_DIDTOP
#define IOf_DIDTOP 8
#endif
#ifndef IOf_FAKE_DIRP
#define IOf_FAKE_DIRP 64
#endif
#ifndef IOf_FLUSH
#define IOf_FLUSH 4
#endif
#ifndef IOf_NOLINE
#define IOf_NOLINE 32
#endif
#ifndef IOf_START
#define IOf_START 2
#endif
#ifndef IOf_UNTAINT
#define IOf_UNTAINT 16
#endif
#ifndef IS_NUMBER_TRAILING
#define IS_NUMBER_TRAILING 0x40
#endif
#ifndef IoTYPE_APPEND
#define IoTYPE_APPEND 'a'
#endif
#ifndef IoTYPE_CLOSED
#define IoTYPE_CLOSED ' '
#endif
#ifndef IoTYPE_IMPLICIT
#define IoTYPE_IMPLICIT 'I'
#endif
#ifndef IoTYPE_NUMERIC
#define IoTYPE_NUMERIC '#'
#endif
#ifndef IoTYPE_PIPE
#define IoTYPE_PIPE '|'
#endif
#ifndef IoTYPE_RDONLY
#define IoTYPE_RDONLY '<'
#endif
#ifndef IoTYPE_RDWR
#define IoTYPE_RDWR '+'
#endif
#ifndef IoTYPE_SOCKET
#define IoTYPE_SOCKET 's'
#endif
#ifndef IoTYPE_STD
#define IoTYPE_STD '-'
#endif
#ifndef IoTYPE_WRONLY
#define IoTYPE_WRONLY '>'
#endif
#ifndef MGf_BYTES
#define MGf_BYTES 0x40
#endif
#ifndef MGf_GSKIP
#define MGf_GSKIP 4
#endif
#ifndef MGf_MINMATCH
#define MGf_MINMATCH 1
#endif
#ifndef MGf_REQUIRE_GV
#define MGf_REQUIRE_GV 1
#endif
#ifndef MGf_TAINTEDDIR
#define MGf_TAINTEDDIR 1
#endif
#ifndef OPf_FOLDED
#define OPf_FOLDED (1<<16)
#endif
#ifndef PADNAMEf_FIELD
#define PADNAMEf_FIELD 0x20
#endif
#ifndef PADNAMEf_LVALUE
#define PADNAMEf_LVALUE 0x04
#endif
#ifndef PADNAMEf_OUR
#define PADNAMEf_OUR 0x10
#endif
#ifndef PADNAMEf_OUTER
#define PADNAMEf_OUTER 0x01
#endif
#ifndef PADNAMEf_STATE
#define PADNAMEf_STATE 0x02
#endif
#ifndef PADNAMEf_TYPED
#define PADNAMEf_TYPED 0x08
#endif
#ifndef PERLSI_CONSTRUCTOR
#define PERLSI_CONSTRUCTOR 13
#endif
#ifndef PERLSI_DIEHOOK
#define PERLSI_DIEHOOK 8
#endif
#ifndef PERLSI_MULTICALL
#define PERLSI_MULTICALL 10
#endif
#ifndef PERLSI_REGCOMP
#define PERLSI_REGCOMP 11
#endif
#ifndef PERLSI_REQUIRE
#define PERLSI_REQUIRE 9
#endif
#ifndef PERLSI_SMARTMATCH
#define PERLSI_SMARTMATCH 12
#endif
#ifndef PERLSI_WARNHOOK
#define PERLSI_WARNHOOK 7
#endif
#ifndef PERL_EXIT_ABORT
#define PERL_EXIT_ABORT 0x08
#endif
#ifndef PERL_EXIT_EXPECTED
#define PERL_EXIT_EXPECTED 0x01
#endif
#ifndef PERL_EXIT_WARN
#define PERL_EXIT_WARN 0x04
#endif
#ifndef PERL_LOADMOD_DENY
#define PERL_LOADMOD_DENY 0x1
#endif
#ifndef PERL_LOADMOD_IMPORT_OPS
#define PERL_LOADMOD_IMPORT_OPS 0x4
#endif
#ifndef PERL_LOADMOD_NOIMPORT
#define PERL_LOADMOD_NOIMPORT 0x2
#endif
#ifndef PERL_MAGIC_READONLY_ACCEPTABLE
#define PERL_MAGIC_READONLY_ACCEPTABLE 0x40
#endif
#ifndef PERL_MAGIC_UTF8_CACHESIZE
#define PERL_MAGIC_UTF8_CACHESIZE 2
#endif
#ifndef PERL_MAGIC_VALUE_MAGIC
#define PERL_MAGIC_VALUE_MAGIC 0x80
#endif
#ifndef PERL_MAGIC_VTABLE_MASK
#define PERL_MAGIC_VTABLE_MASK 0x3F
#endif
#ifndef PERL_MAGIC_arylen
#define PERL_MAGIC_arylen '#'
#endif
#ifndef PERL_MAGIC_arylen_p
#define PERL_MAGIC_arylen_p '@'
#endif
#ifndef PERL_MAGIC_bm
#define PERL_MAGIC_bm 'B'
#endif
#ifndef PERL_MAGIC_checkcall
#define PERL_MAGIC_checkcall ']'
#endif
#ifndef PERL_MAGIC_collxfrm
#define PERL_MAGIC_collxfrm 'o'
#endif
#ifndef PERL_MAGIC_dbfile
#define PERL_MAGIC_dbfile 'L'
#endif
#ifndef PERL_MAGIC_dbline
#define PERL_MAGIC_dbline 'l'
#endif
#ifndef PERL_MAGIC_debugvar
#define PERL_MAGIC_debugvar '*'
#endif
#ifndef PERL_MAGIC_defelem
#define PERL_MAGIC_defelem 'y'
#endif
#ifndef PERL_MAGIC_destruct
#define PERL_MAGIC_destruct 'X'
#endif
#ifndef PERL_MAGIC_env
#define PERL_MAGIC_env 'E'
#endif
#ifndef PERL_MAGIC_envelem
#define PERL_MAGIC_envelem 'e'
#endif
#ifndef PERL_MAGIC_fm
#define PERL_MAGIC_fm 'f'
#endif
#ifndef PERL_MAGIC_hints
#define PERL_MAGIC_hints 'H'
#endif
#ifndef PERL_MAGIC_hintselem
#define PERL_MAGIC_hintselem 'h'
#endif
#ifndef PERL_MAGIC_hook
#define PERL_MAGIC_hook 'Z'
#endif
#ifndef PERL_MAGIC_hookelem
#define PERL_MAGIC_hookelem 'z'
#endif
#ifndef PERL_MAGIC_isa
#define PERL_MAGIC_isa 'I'
#endif
#ifndef PERL_MAGIC_isaelem
#define PERL_MAGIC_isaelem 'i'
#endif
#ifndef PERL_MAGIC_lvref
#define PERL_MAGIC_lvref '\\'
#endif
#ifndef PERL_MAGIC_nkeys
#define PERL_MAGIC_nkeys 'k'
#endif
#ifndef PERL_MAGIC_nonelem
#define PERL_MAGIC_nonelem 'Y'
#endif
#ifndef PERL_MAGIC_pos
#define PERL_MAGIC_pos '.'
#endif
#ifndef PERL_MAGIC_qr
#define PERL_MAGIC_qr 'r'
#endif
#ifndef PERL_MAGIC_regdata
#define PERL_MAGIC_regdata 'D'
#endif
#ifndef PERL_MAGIC_regdatum
#define PERL_MAGIC_regdatum 'd'
#endif
#ifndef PERL_MAGIC_regex_global
#define PERL_MAGIC_regex_global 'g'
#endif
#ifndef PERL_MAGIC_rhash
#define PERL_MAGIC_rhash '%'
#endif
#ifndef PERL_MAGIC_shared
#define PERL_MAGIC_shared 'N'
#endif
#ifndef PERL_MAGIC_shared_scalar
#define PERL_MAGIC_shared_scalar 'n'
#endif
#ifndef PERL_MAGIC_sig
#define PERL_MAGIC_sig 'S'
#endif
#ifndef PERL_MAGIC_sigelem
#define PERL_MAGIC_sigelem 's'
#endif
#ifndef PERL_MAGIC_substr
#define PERL_MAGIC_substr 'x'
#endif
#ifndef PERL_MAGIC_sv
#define PERL_MAGIC_sv '\0'
#endif
#ifndef PERL_MAGIC_symtab
#define PERL_MAGIC_symtab ':'
#endif
#ifndef PERL_MAGIC_tiedelem
#define PERL_MAGIC_tiedelem 'p'
#endif
#ifndef PERL_MAGIC_tiedscalar
#define PERL_MAGIC_tiedscalar 'q'
#endif
#ifndef PERL_MAGIC_utf8
#define PERL_MAGIC_utf8 'w'
#endif
#ifndef PERL_MAGIC_uvar
#define PERL_MAGIC_uvar 'U'
#endif
#ifndef PERL_MAGIC_uvar_elem
#define PERL_MAGIC_uvar_elem 'u'
#endif
#ifndef PERL_MAGIC_vec
#define PERL_MAGIC_vec 'v'
#endif
#ifndef PERL_MAGIC_vstring
#define PERL_MAGIC_vstring 'V'
#endif
#ifndef PERL_PV_ESCAPE_ALL
#define PERL_PV_ESCAPE_ALL 0x001000
#endif
#ifndef PERL_PV_ESCAPE_DWIM
#define PERL_PV_ESCAPE_DWIM 0x010000
#endif
#ifndef PERL_PV_ESCAPE_DWIM_ALL_HEX
#define PERL_PV_ESCAPE_DWIM_ALL_HEX 0x020000
#endif
#ifndef PERL_PV_ESCAPE_FIRSTCHAR
#define PERL_PV_ESCAPE_FIRSTCHAR 0x000800
#endif
#ifndef PERL_PV_ESCAPE_NOBACKSLASH
#define PERL_PV_ESCAPE_NOBACKSLASH 0x002000
#endif
#ifndef PERL_PV_ESCAPE_NOCLEAR
#define PERL_PV_ESCAPE_NOCLEAR 0x004000
#endif
#ifndef PERL_PV_ESCAPE_NONASCII
#define PERL_PV_ESCAPE_NONASCII 0x000400
#endif
#ifndef PERL_PV_ESCAPE_NON_WC
#define PERL_PV_ESCAPE_NON_WC 0x040000
#endif
#ifndef PERL_PV_ESCAPE_QUOTE
#define PERL_PV_ESCAPE_QUOTE 0x000001
#endif
#ifndef PERL_PV_ESCAPE_RE
#define PERL_PV_ESCAPE_RE 0x008000
#endif
#ifndef PERL_PV_ESCAPE_TRUNC_MIDDLE
#define PERL_PV_ESCAPE_TRUNC_MIDDLE 0x080000
#endif
#ifndef PERL_PV_ESCAPE_UNI
#define PERL_PV_ESCAPE_UNI 0x000100
#endif
#ifndef PERL_PV_ESCAPE_UNI_DETECT
#define PERL_PV_ESCAPE_UNI_DETECT 0x000200
#endif
#ifndef PERL_PV_PRETTY_ELLIPSES
#define PERL_PV_PRETTY_ELLIPSES 0x000002
#endif
#ifndef PERL_PV_PRETTY_EXACTSIZE
#define PERL_PV_PRETTY_EXACTSIZE 0x000008
#endif
#ifndef PERL_PV_PRETTY_LTGT
#define PERL_PV_PRETTY_LTGT 0x000004
#endif
#ifndef PERL_SCAN_ALLOW_MEDIAL_UNDERSCORES_ONLY
#define PERL_SCAN_ALLOW_MEDIAL_UNDERSCORES_ONLY 0x100
#endif
#ifndef PERL_SCAN_ALLOW_UNDERSCORES
#define PERL_SCAN_ALLOW_UNDERSCORES 0x01
#endif
#ifndef PERL_SCAN_DISALLOW_PREFIX
#define PERL_SCAN_DISALLOW_PREFIX 0x02
#endif
#ifndef PERL_SCAN_DISCARD_INSTEAD_OF_OVERFLOW
#define PERL_SCAN_DISCARD_INSTEAD_OF_OVERFLOW 0x200
#endif
#ifndef PERL_SCAN_GREATER_THAN_UV_MAX
#define PERL_SCAN_GREATER_THAN_UV_MAX 0x04
#endif
#ifndef PERL_SCAN_NOTIFY_ILLDIGIT
#define PERL_SCAN_NOTIFY_ILLDIGIT 0x40
#endif
#ifndef PERL_SCAN_SILENT_ILLDIGIT
#define PERL_SCAN_SILENT_ILLDIGIT 0x08
#endif
#ifndef PERL_SCAN_SILENT_NON_PORTABLE
#define PERL_SCAN_SILENT_NON_PORTABLE 0x20
#endif
#ifndef PERL_SCAN_SILENT_OVERFLOW
#define PERL_SCAN_SILENT_OVERFLOW 0x80
#endif
#ifndef PERL_SCAN_TRAILING
#define PERL_SCAN_TRAILING 0x10
#endif
#ifndef PERL_UNICODE_ALL_FLAGS
#define PERL_UNICODE_ALL_FLAGS 0x01ff
#endif
#ifndef PERL_UNICODE_ARGV
#define PERL_UNICODE_ARGV 'A'
#endif
#ifndef PERL_UNICODE_ARGV_FLAG
#define PERL_UNICODE_ARGV_FLAG 0x0020
#endif
#ifndef PERL_UNICODE_IN
#define PERL_UNICODE_IN 'i'
#endif
#ifndef PERL_UNICODE_INOUT
#define PERL_UNICODE_INOUT 'D'
#endif
#ifndef PERL_UNICODE_IN_FLAG
#define PERL_UNICODE_IN_FLAG 0x0008
#endif
#ifndef PERL_UNICODE_LOCALE
#define PERL_UNICODE_LOCALE 'L'
#endif
#ifndef PERL_UNICODE_LOCALE_FLAG
#define PERL_UNICODE_LOCALE_FLAG 0x0040
#endif
#ifndef PERL_UNICODE_MAX
#define PERL_UNICODE_MAX 0x10FFFF
#endif
#ifndef PERL_UNICODE_OUT
#define PERL_UNICODE_OUT 'o'
#endif
#ifndef PERL_UNICODE_OUT_FLAG
#define PERL_UNICODE_OUT_FLAG 0x0010
#endif
#ifndef PERL_UNICODE_STD
#define PERL_UNICODE_STD 'S'
#endif
#ifndef PERL_UNICODE_STDERR
#define PERL_UNICODE_STDERR 'E'
#endif
#ifndef PERL_UNICODE_STDERR_FLAG
#define PERL_UNICODE_STDERR_FLAG 0x0004
#endif
#ifndef PERL_UNICODE_STDIN
#define PERL_UNICODE_STDIN 'I'
#endif
#ifndef PERL_UNICODE_STDIN_FLAG
#define PERL_UNICODE_STDIN_FLAG 0x0001
#endif
#ifndef PERL_UNICODE_STDOUT
#define PERL_UNICODE_STDOUT 'O'
#endif
#ifndef PERL_UNICODE_STDOUT_FLAG
#define PERL_UNICODE_STDOUT_FLAG 0x0002
#endif
#ifndef PERL_UNICODE_UTF8CACHEASSERT
#define PERL_UNICODE_UTF8CACHEASSERT 'a'
#endif
#ifndef PERL_UNICODE_UTF8CACHEASSERT_FLAG
#define PERL_UNICODE_UTF8CACHEASSERT_FLAG 0x0100
#endif
#ifndef PERL_UNICODE_WIDESYSCALLS
#define PERL_UNICODE_WIDESYSCALLS 'W'
#endif
#ifndef PERL_UNICODE_WIDESYSCALLS_FLAG
#define PERL_UNICODE_WIDESYSCALLS_FLAG 0x0080
#endif
#ifndef RXf_PMf_STD_PMMOD_SHIFT
#define RXf_PMf_STD_PMMOD_SHIFT 0
#endif
#ifndef SV_CATBYTES
#define SV_CATBYTES (1 << 14)
#endif
#ifndef SV_CATUTF8
#define SV_CATUTF8 (1 << 15)
#endif
#ifndef SV_CONSTS_COUNT
#define SV_CONSTS_COUNT 35
#endif
#ifndef SV_CONST_BINMODE
#define SV_CONST_BINMODE 28
#endif
#ifndef SV_CONST_CLEAR
#define SV_CONST_CLEAR 32
#endif
#ifndef SV_CONST_CLOSE
#define SV_CONST_CLOSE 30
#endif
#ifndef SV_CONST_DELETE
#define SV_CONST_DELETE 31
#endif
#ifndef SV_CONST_DESTROY
#define SV_CONST_DESTROY 34
#endif
#ifndef SV_CONST_EOF
#define SV_CONST_EOF 27
#endif
#ifndef SV_CONST_EXISTS
#define SV_CONST_EXISTS 8
#endif
#ifndef SV_CONST_EXTEND
#define SV_CONST_EXTEND 14
#endif
#ifndef SV_CONST_FETCH
#define SV_CONST_FETCH 4
#endif
#ifndef SV_CONST_FETCHSIZE
#define SV_CONST_FETCHSIZE 5
#endif
#ifndef SV_CONST_FILENO
#define SV_CONST_FILENO 29
#endif
#ifndef SV_CONST_FIRSTKEY
#define SV_CONST_FIRSTKEY 15
#endif
#ifndef SV_CONST_GETC
#define SV_CONST_GETC 24
#endif
#ifndef SV_CONST_NEXTKEY
#define SV_CONST_NEXTKEY 16
#endif
#ifndef SV_CONST_OPEN
#define SV_CONST_OPEN 18
#endif
#ifndef SV_CONST_POP
#define SV_CONST_POP 10
#endif
#ifndef SV_CONST_PRINT
#define SV_CONST_PRINT 20
#endif
#ifndef SV_CONST_PRINTF
#define SV_CONST_PRINTF 21
#endif
#ifndef SV_CONST_PUSH
#define SV_CONST_PUSH 9
#endif
#ifndef SV_CONST_READ
#define SV_CONST_READ 22
#endif
#ifndef SV_CONST_READLINE
#define SV_CONST_READLINE 23
#endif
#ifndef SV_CONST_RETURN
#define SV_CONST_RETURN (1 <<  5)
#endif
#ifndef SV_CONST_SCALAR
#define SV_CONST_SCALAR 17
#endif
#ifndef SV_CONST_SEEK
#define SV_CONST_SEEK 25
#endif
#ifndef SV_CONST_SHIFT
#define SV_CONST_SHIFT 11
#endif
#ifndef SV_CONST_SPLICE
#define SV_CONST_SPLICE 13
#endif
#ifndef SV_CONST_STORE
#define SV_CONST_STORE 6
#endif
#ifndef SV_CONST_STORESIZE
#define SV_CONST_STORESIZE 7
#endif
#ifndef SV_CONST_TELL
#define SV_CONST_TELL 26
#endif
#ifndef SV_CONST_TIEARRAY
#define SV_CONST_TIEARRAY 1
#endif
#ifndef SV_CONST_TIEHANDLE
#define SV_CONST_TIEHANDLE 3
#endif
#ifndef SV_CONST_TIEHASH
#define SV_CONST_TIEHASH 2
#endif
#ifndef SV_CONST_TIESCALAR
#define SV_CONST_TIESCALAR 0
#endif
#ifndef SV_CONST_UNSHIFT
#define SV_CONST_UNSHIFT 12
#endif
#ifndef SV_CONST_UNTIE
#define SV_CONST_UNTIE 33
#endif
#ifndef SV_CONST_WRITE
#define SV_CONST_WRITE 19
#endif
#ifndef SV_COW_DROP_PV
#define SV_COW_DROP_PV (1 <<  2)
#endif
#ifndef SV_COW_OTHER_PVS
#define SV_COW_OTHER_PVS (1 << 10)
#endif
#ifndef SV_COW_SHARED_HASH_KEYS
#define SV_COW_SHARED_HASH_KEYS (1 <<  9)
#endif
#ifndef SV_DO_COW_SVSETSV
#define SV_DO_COW_SVSETSV 0
#endif
#ifndef SV_FORCE_OVERLOAD
#define SV_FORCE_OVERLOAD (1 << 16)
#endif
#ifndef SV_FORCE_UTF8_UPGRADE
#define SV_FORCE_UTF8_UPGRADE (1 << 12)
#endif
#ifndef SV_GMAGIC
#define SV_GMAGIC (1 <<  1)
#endif
#ifndef SV_HAS_TRAILING_NUL
#define SV_HAS_TRAILING_NUL (1 <<  8)
#endif
#ifndef SV_IMMEDIATE_UNREF
#define SV_IMMEDIATE_UNREF (1 <<  0)
#endif
#ifndef SV_MUTABLE_RETURN
#define SV_MUTABLE_RETURN (1 <<  6)
#endif
#ifndef SV_NOSTEAL
#define SV_NOSTEAL (1 <<  4)
#endif
#ifndef SV_SMAGIC
#define SV_SMAGIC (1 <<  7)
#endif
#ifndef SV_UNDEF_RETURNS_NULL
#define SV_UNDEF_RETURNS_NULL (1 << 11)
#endif
#ifndef SV_UTF8_NO_ENCODING
#define SV_UTF8_NO_ENCODING 0
#endif
#ifndef SVf_AMAGIC
#define SVf_AMAGIC 0x10000000
#endif
#ifndef SVf_BREAK
#define SVf_BREAK 0x04000000
#endif
#ifndef SVf_FAKE
#define SVf_FAKE 0x01000000
#endif
#ifndef SVf_IOK
#define SVf_IOK 0x00000100
#endif
#ifndef SVf_IVisUV
#define SVf_IVisUV 0x80000000
#endif
#ifndef SVf_IsCOW
#define SVf_IsCOW 0x10000000
#endif
#ifndef SVf_NOK
#define SVf_NOK 0x00000200
#endif
#ifndef SVf_OOK
#define SVf_OOK 0x02000000
#endif
#ifndef SVf_POK
#define SVf_POK 0x00000400
#endif
#ifndef SVf_PROTECT
#define SVf_PROTECT 0x00010000
#endif
#ifndef SVf_READONLY
#define SVf_READONLY 0x08000000
#endif
#ifndef SVf_ROK
#define SVf_ROK 0x00000800
#endif
#ifndef SVf_UTF8
#define SVf_UTF8 0x20000000
#endif
#ifndef SVp_IOK
#define SVp_IOK 0x00001000
#endif
#ifndef SVp_NOK
#define SVp_NOK 0x00002000
#endif
#ifndef SVp_POK
#define SVp_POK 0x00004000
#endif
#ifndef SVp_SCREAM
#define SVp_SCREAM 0x00008000
#endif
#ifndef SVpav_REAL
#define SVpav_REAL 0x40000000
#endif
#ifndef SVpav_REIFY
#define SVpav_REIFY 0x80000000
#endif
#ifndef SVphv_HASKFLAGS
#define SVphv_HASKFLAGS 0x80000000
#endif
#ifndef SVphv_LAZYDEL
#define SVphv_LAZYDEL 0x40000000
#endif
#ifndef SVphv_SHAREKEYS
#define SVphv_SHAREKEYS 0x20000000
#endif
#ifndef SVprv_WEAKREF
#define SVprv_WEAKREF 0x80000000
#endif
#ifndef SVs_GMG
#define SVs_GMG 0x00200000
#endif
#ifndef SVs_OBJECT
#define SVs_OBJECT 0x00100000
#endif
#ifndef SVs_PADSTALE
#define SVs_PADSTALE 0x00040000
#endif
#ifndef SVs_PADTMP
#define SVs_PADTMP 0x00020000
#endif
#ifndef SVs_RMG
#define SVs_RMG 0x00800000
#endif
#ifndef SVs_SMG
#define SVs_SMG 0x00400000
#endif
#ifndef SVs_TEMP
#define SVs_TEMP 0x00080000
#endif
#ifndef SVt_MASK
#define SVt_MASK 0x1f
#endif
#ifndef UNICODE_ALLOW_ABOVE_IV_MAX
#define UNICODE_ALLOW_ABOVE_IV_MAX 0x0100
#endif
#ifndef UNICODE_ALLOW_ANY
#define UNICODE_ALLOW_ANY 0
#endif
#ifndef UNICODE_ALLOW_SUPER
#define UNICODE_ALLOW_SUPER 0
#endif
#ifndef UNICODE_ALLOW_SURROGATE
#define UNICODE_ALLOW_SURROGATE 0
#endif
#ifndef UNICODE_BYTE_ORDER_MARK
#define UNICODE_BYTE_ORDER_MARK 0xFEFF
#endif
#ifndef UNICODE_DISALLOW_NONCHAR
#define UNICODE_DISALLOW_NONCHAR 0x0020
#endif
#ifndef UNICODE_DISALLOW_PERL_EXTENDED
#define UNICODE_DISALLOW_PERL_EXTENDED 0x0080
#endif
#ifndef UNICODE_DISALLOW_SUPER
#define UNICODE_DISALLOW_SUPER 0x0040
#endif
#ifndef UNICODE_DISALLOW_SURROGATE
#define UNICODE_DISALLOW_SURROGATE 0x0010
#endif
#ifndef UNICODE_GREEK_CAPITAL_LETTER_SIGMA
#define UNICODE_GREEK_CAPITAL_LETTER_SIGMA 0x03A3
#endif
#ifndef UNICODE_GREEK_SMALL_LETTER_FINAL_SIGMA
#define UNICODE_GREEK_SMALL_LETTER_FINAL_SIGMA 0x03C2
#endif
#ifndef UNICODE_GREEK_SMALL_LETTER_SIGMA
#define UNICODE_GREEK_SMALL_LETTER_SIGMA 0x03C3
#endif
#ifndef UNICODE_PAT_MOD
#define UNICODE_PAT_MOD 'u'
#endif
#ifndef UNICODE_REPLACEMENT
#define UNICODE_REPLACEMENT 0xFFFD
#endif
#ifndef UNICODE_SURROGATE_FIRST
#define UNICODE_SURROGATE_FIRST 0xD800
#endif
#ifndef UNICODE_SURROGATE_LAST
#define UNICODE_SURROGATE_LAST 0xDFFF
#endif
#ifndef UNICODE_WARN_NONCHAR
#define UNICODE_WARN_NONCHAR 0x0002
#endif
#ifndef UNICODE_WARN_PERL_EXTENDED
#define UNICODE_WARN_PERL_EXTENDED 0x0008
#endif
#ifndef UNICODE_WARN_SUPER
#define UNICODE_WARN_SUPER 0x0004
#endif
#ifndef UNICODE_WARN_SURROGATE
#define UNICODE_WARN_SURROGATE 0x0001
#endif
#ifndef UTF8_ALLOW_ANYUV
#define UTF8_ALLOW_ANYUV 0
#endif
#ifndef UTF8_ALLOW_FE_FF
#define UTF8_ALLOW_FE_FF 0
#endif
#ifndef UTF8_ALLOW_FFFF
#define UTF8_ALLOW_FFFF 0
#endif
#ifndef UTF8_ALLOW_SURROGATE
#define UTF8_ALLOW_SURROGATE 0
#endif
#ifndef UTF8_CHECK_ONLY_BIT_POS_
#define UTF8_CHECK_ONLY_BIT_POS_ 15
#endif
#ifndef UTF8_DIE_IF_MALFORMED_BIT_POS_
#define UTF8_DIE_IF_MALFORMED_BIT_POS_ 16
#endif
#ifndef UTF8_FORCE_WARN_IF_MALFORMED_BIT_POS_
#define UTF8_FORCE_WARN_IF_MALFORMED_BIT_POS_ 17
#endif
#ifndef UTF8_GOT_CONTINUATION_BIT_POS_
#define UTF8_GOT_CONTINUATION_BIT_POS_ 2
#endif
#ifndef UTF8_GOT_EMPTY_BIT_POS_
#define UTF8_GOT_EMPTY_BIT_POS_ 1
#endif
#ifndef UTF8_GOT_LONG_BIT_POS_
#define UTF8_GOT_LONG_BIT_POS_ 13
#endif
#ifndef UTF8_GOT_LONG_WITH_VALUE_BIT_POS_
#define UTF8_GOT_LONG_WITH_VALUE_BIT_POS_ 14
#endif
#ifndef UTF8_GOT_NONCHAR_BIT_POS_
#define UTF8_GOT_NONCHAR_BIT_POS_ 11
#endif
#ifndef UTF8_GOT_NON_CONTINUATION_BIT_POS_
#define UTF8_GOT_NON_CONTINUATION_BIT_POS_ 4
#endif
#ifndef UTF8_GOT_OVERFLOW_BIT_POS_
#define UTF8_GOT_OVERFLOW_BIT_POS_ 0
#endif
#ifndef UTF8_GOT_PERL_EXTENDED_BIT_POS_
#define UTF8_GOT_PERL_EXTENDED_BIT_POS_ 7
#endif
#ifndef UTF8_GOT_SHORT_BIT_POS_
#define UTF8_GOT_SHORT_BIT_POS_ 3
#endif
#ifndef UTF8_GOT_SUPER_BIT_POS_
#define UTF8_GOT_SUPER_BIT_POS_ 9
#endif
#ifndef UTF8_GOT_SURROGATE_BIT_POS_
#define UTF8_GOT_SURROGATE_BIT_POS_ 5
#endif
#ifndef UTF8_MAX_FOLD_CHAR_EXPAND
#define UTF8_MAX_FOLD_CHAR_EXPAND 3
#endif
#ifndef UTF8_NO_CONFIDENCE_IN_CURLEN_BIT_POS_
#define UTF8_NO_CONFIDENCE_IN_CURLEN_BIT_POS_ 18
#endif
#ifndef UTF8_WARN_NONCHAR_BIT_POS_
#define UTF8_WARN_NONCHAR_BIT_POS_ 12
#endif
#ifndef UTF8_WARN_PERL_EXTENDED_BIT_POS_
#define UTF8_WARN_PERL_EXTENDED_BIT_POS_ 8
#endif
#ifndef UTF8_WARN_SUPER_BIT_POS_
#define UTF8_WARN_SUPER_BIT_POS_ 10
#endif
#ifndef UTF8_WARN_SURROGATE_BIT_POS_
#define UTF8_WARN_SURROGATE_BIT_POS_ 6
#endif
#ifndef WARN_ALL
#define WARN_ALL 0
#endif
#ifndef WARN_AMBIGUOUS
#define WARN_AMBIGUOUS 29
#endif
#ifndef WARN_BAREWORD
#define WARN_BAREWORD 30
#endif
#ifndef WARN_CLOSED
#define WARN_CLOSED 6
#endif
#ifndef WARN_CLOSURE
#define WARN_CLOSURE 1
#endif
#ifndef WARN_DEBUGGING
#define WARN_DEBUGGING 22
#endif
#ifndef WARN_DEPRECATED
#define WARN_DEPRECATED 2
#endif
#ifndef WARN_DEPRECATED__DELIMITER_WILL_BE_PAIRED
#define WARN_DEPRECATED__DELIMITER_WILL_BE_PAIRED 73
#endif
#ifndef WARN_DEPRECATED__DOT_IN_INC
#define WARN_DEPRECATED__DOT_IN_INC 61
#endif
#ifndef WARN_DEPRECATED__SUBSEQUENT_USE_VERSION
#define WARN_DEPRECATED__SUBSEQUENT_USE_VERSION 75
#endif
#ifndef WARN_DEPRECATED__UNICODE_PROPERTY_NAME
#define WARN_DEPRECATED__UNICODE_PROPERTY_NAME 48
#endif
#ifndef WARN_DEPRECATED__VERSION_DOWNGRADE
#define WARN_DEPRECATED__VERSION_DOWNGRADE 72
#endif
#ifndef WARN_DIGIT
#define WARN_DIGIT 31
#endif
#ifndef WARN_EXEC
#define WARN_EXEC 7
#endif
#ifndef WARN_EXITING
#define WARN_EXITING 3
#endif
#ifndef WARN_EXPERIMENTAL
#define WARN_EXPERIMENTAL 52
#endif
#ifndef WARN_EXPERIMENTAL__ARGS_ARRAY_WITH_SIGNATURES
#define WARN_EXPERIMENTAL__ARGS_ARRAY_WITH_SIGNATURES 67
#endif
#ifndef WARN_EXPERIMENTAL__BUILTIN
#define WARN_EXPERIMENTAL__BUILTIN 68
#endif
#ifndef WARN_EXPERIMENTAL__CLASS
#define WARN_EXPERIMENTAL__CLASS 74
#endif
#ifndef WARN_EXPERIMENTAL__DECLARED_REFS
#define WARN_EXPERIMENTAL__DECLARED_REFS 60
#endif
#ifndef WARN_EXPERIMENTAL__DEFER
#define WARN_EXPERIMENTAL__DEFER 69
#endif
#ifndef WARN_EXPERIMENTAL__ENHANCED_XX
#define WARN_EXPERIMENTAL__ENHANCED_XX 78
#endif
#ifndef WARN_EXPERIMENTAL__EXTRA_PAIRED_DELIMITERS
#define WARN_EXPERIMENTAL__EXTRA_PAIRED_DELIMITERS 70
#endif
#ifndef WARN_EXPERIMENTAL__KEYWORD_ALL
#define WARN_EXPERIMENTAL__KEYWORD_ALL 76
#endif
#ifndef WARN_EXPERIMENTAL__KEYWORD_ANY
#define WARN_EXPERIMENTAL__KEYWORD_ANY 77
#endif
#ifndef WARN_EXPERIMENTAL__PRIVATE_USE
#define WARN_EXPERIMENTAL__PRIVATE_USE 63
#endif
#ifndef WARN_EXPERIMENTAL__REFALIASING
#define WARN_EXPERIMENTAL__REFALIASING 56
#endif
#ifndef WARN_EXPERIMENTAL__REGEX_SETS
#define WARN_EXPERIMENTAL__REGEX_SETS 53
#endif
#ifndef WARN_EXPERIMENTAL__RE_STRICT
#define WARN_EXPERIMENTAL__RE_STRICT 55
#endif
#ifndef WARN_EXPERIMENTAL__SIGNATURE_NAMED_PARAMETERS
#define WARN_EXPERIMENTAL__SIGNATURE_NAMED_PARAMETERS 79
#endif
#ifndef WARN_EXPERIMENTAL__TRY
#define WARN_EXPERIMENTAL__TRY 66
#endif
#ifndef WARN_EXPERIMENTAL__UNIPROP_WILDCARDS
#define WARN_EXPERIMENTAL__UNIPROP_WILDCARDS 64
#endif
#ifndef WARN_EXPERIMENTAL__VLB
#define WARN_EXPERIMENTAL__VLB 65
#endif
#ifndef WARN_GLOB
#define WARN_GLOB 4
#endif
#ifndef WARN_ILLEGALPROTO
#define WARN_ILLEGALPROTO 47
#endif
#ifndef WARN_IMPRECISION
#define WARN_IMPRECISION 46
#endif
#ifndef WARN_INPLACE
#define WARN_INPLACE 23
#endif
#ifndef WARN_INTERNAL
#define WARN_INTERNAL 24
#endif
#ifndef WARN_IO
#define WARN_IO 5
#endif
#ifndef WARN_LAYER
#define WARN_LAYER 8
#endif
#ifndef WARN_LOCALE
#define WARN_LOCALE 57
#endif
#ifndef WARN_MALLOC
#define WARN_MALLOC 25
#endif
#ifndef WARN_MISC
#define WARN_MISC 12
#endif
#ifndef WARN_MISSING
#define WARN_MISSING 58
#endif
#ifndef WARN_MISSING_IMPORT
#define WARN_MISSING_IMPORT 80
#endif
#ifndef WARN_NEWLINE
#define WARN_NEWLINE 9
#endif
#ifndef WARN_NONCHAR
#define WARN_NONCHAR 50
#endif
#ifndef WARN_NON_UNICODE
#define WARN_NON_UNICODE 49
#endif
#ifndef WARN_NUMERIC
#define WARN_NUMERIC 13
#endif
#ifndef WARN_ONCE
#define WARN_ONCE 14
#endif
#ifndef WARN_OVERFLOW
#define WARN_OVERFLOW 15
#endif
#ifndef WARN_PACK
#define WARN_PACK 16
#endif
#ifndef WARN_PARENTHESIS
#define WARN_PARENTHESIS 32
#endif
#ifndef WARN_PIPE
#define WARN_PIPE 10
#endif
#ifndef WARN_PORTABLE
#define WARN_PORTABLE 17
#endif
#ifndef WARN_PRECEDENCE
#define WARN_PRECEDENCE 33
#endif
#ifndef WARN_PRINTF
#define WARN_PRINTF 34
#endif
#ifndef WARN_PROTOTYPE
#define WARN_PROTOTYPE 35
#endif
#ifndef WARN_QW
#define WARN_QW 36
#endif
#ifndef WARN_RECURSION
#define WARN_RECURSION 18
#endif
#ifndef WARN_REDEFINE
#define WARN_REDEFINE 19
#endif
#ifndef WARN_REDUNDANT
#define WARN_REDUNDANT 59
#endif
#ifndef WARN_REGEXP
#define WARN_REGEXP 20
#endif
#ifndef WARN_RESERVED
#define WARN_RESERVED 37
#endif
#ifndef WARN_SCALAR
#define WARN_SCALAR 71
#endif
#ifndef WARN_SEMICOLON
#define WARN_SEMICOLON 38
#endif
#ifndef WARN_SEVERE
#define WARN_SEVERE 21
#endif
#ifndef WARN_SHADOW
#define WARN_SHADOW 62
#endif
#ifndef WARN_SIGNAL
#define WARN_SIGNAL 26
#endif
#ifndef WARN_SUBSTR
#define WARN_SUBSTR 27
#endif
#ifndef WARN_SURROGATE
#define WARN_SURROGATE 51
#endif
#ifndef WARN_SYNTAX
#define WARN_SYNTAX 28
#endif
#ifndef WARN_SYSCALLS
#define WARN_SYSCALLS 54
#endif
#ifndef WARN_TAINT
#define WARN_TAINT 39
#endif
#ifndef WARN_THREADS
#define WARN_THREADS 40
#endif
#ifndef WARN_UNINITIALIZED
#define WARN_UNINITIALIZED 41
#endif
#ifndef WARN_UNOPENED
#define WARN_UNOPENED 11
#endif
#ifndef WARN_UNPACK
#define WARN_UNPACK 42
#endif
#ifndef WARN_UNTIE
#define WARN_UNTIE 43
#endif
#ifndef WARN_VOID
#define WARN_VOID 45
#endif
#endif /* GOPERL_XS_SDK_CONSTS_H */
