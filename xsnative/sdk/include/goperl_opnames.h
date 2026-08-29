/* go-perl native XS SDK -- opcode numbers and names, pinned to the embedded
 * interpreter (perl v5.42.2). GENERATED from perl5's opnames.h/opcode.h; do
 * not edit by hand. */
#ifndef GOPERL_OPNAMES_H
#define GOPERL_OPNAMES_H

#define OP_NULL 0
#define OP_STUB 1
#define OP_SCALAR 2
#define OP_PUSHMARK 3
#define OP_WANTARRAY 4
#define OP_CONST 5
#define OP_GVSV 6
#define OP_GV 7
#define OP_GELEM 8
#define OP_PADSV 9
#define OP_PADSV_STORE 10
#define OP_PADAV 11
#define OP_PADHV 12
#define OP_PADANY 13
#define OP_RV2GV 14
#define OP_RV2SV 15
#define OP_AV2ARYLEN 16
#define OP_RV2CV 17
#define OP_ANONCODE 18
#define OP_PROTOTYPE 19
#define OP_REFGEN 20
#define OP_SREFGEN 21
#define OP_REF 22
#define OP_BLESS 23
#define OP_BACKTICK 24
#define OP_GLOB 25
#define OP_READLINE 26
#define OP_RCATLINE 27
#define OP_REGCMAYBE 28
#define OP_REGCRESET 29
#define OP_REGCOMP 30
#define OP_MATCH 31
#define OP_QR 32
#define OP_SUBST 33
#define OP_SUBSTCONT 34
#define OP_TRANS 35
#define OP_TRANSR 36
#define OP_SASSIGN 37
#define OP_AASSIGN 38
#define OP_CHOP 39
#define OP_SCHOP 40
#define OP_CHOMP 41
#define OP_SCHOMP 42
#define OP_DEFINED 43
#define OP_UNDEF 44
#define OP_STUDY 45
#define OP_POS 46
#define OP_PREINC 47
#define OP_I_PREINC 48
#define OP_PREDEC 49
#define OP_I_PREDEC 50
#define OP_POSTINC 51
#define OP_I_POSTINC 52
#define OP_POSTDEC 53
#define OP_I_POSTDEC 54
#define OP_POW 55
#define OP_MULTIPLY 56
#define OP_I_MULTIPLY 57
#define OP_DIVIDE 58
#define OP_I_DIVIDE 59
#define OP_MODULO 60
#define OP_I_MODULO 61
#define OP_REPEAT 62
#define OP_ADD 63
#define OP_I_ADD 64
#define OP_SUBTRACT 65
#define OP_I_SUBTRACT 66
#define OP_CONCAT 67
#define OP_MULTICONCAT 68
#define OP_STRINGIFY 69
#define OP_LEFT_SHIFT 70
#define OP_RIGHT_SHIFT 71
#define OP_LT 72
#define OP_I_LT 73
#define OP_GT 74
#define OP_I_GT 75
#define OP_LE 76
#define OP_I_LE 77
#define OP_GE 78
#define OP_I_GE 79
#define OP_EQ 80
#define OP_I_EQ 81
#define OP_NE 82
#define OP_I_NE 83
#define OP_NCMP 84
#define OP_I_NCMP 85
#define OP_SLT 86
#define OP_SGT 87
#define OP_SLE 88
#define OP_SGE 89
#define OP_SEQ 90
#define OP_SNE 91
#define OP_SCMP 92
#define OP_BIT_AND 93
#define OP_BIT_XOR 94
#define OP_BIT_OR 95
#define OP_NBIT_AND 96
#define OP_NBIT_XOR 97
#define OP_NBIT_OR 98
#define OP_SBIT_AND 99
#define OP_SBIT_XOR 100
#define OP_SBIT_OR 101
#define OP_NEGATE 102
#define OP_I_NEGATE 103
#define OP_NOT 104
#define OP_COMPLEMENT 105
#define OP_NCOMPLEMENT 106
#define OP_SCOMPLEMENT 107
#define OP_SMARTMATCH 108
#define OP_ATAN2 109
#define OP_SIN 110
#define OP_COS 111
#define OP_RAND 112
#define OP_SRAND 113
#define OP_EXP 114
#define OP_LOG 115
#define OP_SQRT 116
#define OP_INT 117
#define OP_HEX 118
#define OP_OCT 119
#define OP_ABS 120
#define OP_LENGTH 121
#define OP_SUBSTR 122
#define OP_SUBSTR_LEFT 123
#define OP_VEC 124
#define OP_INDEX 125
#define OP_RINDEX 126
#define OP_SPRINTF 127
#define OP_FORMLINE 128
#define OP_ORD 129
#define OP_CHR 130
#define OP_CRYPT 131
#define OP_UCFIRST 132
#define OP_LCFIRST 133
#define OP_UC 134
#define OP_LC 135
#define OP_QUOTEMETA 136
#define OP_RV2AV 137
#define OP_AELEMFAST 138
#define OP_AELEMFAST_LEX 139
#define OP_AELEMFASTLEX_STORE 140
#define OP_AELEM 141
#define OP_ASLICE 142
#define OP_KVASLICE 143
#define OP_AEACH 144
#define OP_AVALUES 145
#define OP_AKEYS 146
#define OP_EACH 147
#define OP_VALUES 148
#define OP_KEYS 149
#define OP_DELETE 150
#define OP_EXISTS 151
#define OP_RV2HV 152
#define OP_HELEM 153
#define OP_HSLICE 154
#define OP_KVHSLICE 155
#define OP_MULTIDEREF 156
#define OP_UNPACK 157
#define OP_PACK 158
#define OP_SPLIT 159
#define OP_JOIN 160
#define OP_LIST 161
#define OP_LSLICE 162
#define OP_ANONLIST 163
#define OP_ANONHASH 164
#define OP_EMPTYAVHV 165
#define OP_SPLICE 166
#define OP_PUSH 167
#define OP_POP 168
#define OP_SHIFT 169
#define OP_UNSHIFT 170
#define OP_SORT 171
#define OP_REVERSE 172
#define OP_GREPSTART 173
#define OP_GREPWHILE 174
#define OP_ANYSTART 175
#define OP_ALLSTART 176
#define OP_ANYWHILE 177
#define OP_MAPSTART 178
#define OP_MAPWHILE 179
#define OP_RANGE 180
#define OP_FLIP 181
#define OP_FLOP 182
#define OP_AND 183
#define OP_OR 184
#define OP_XOR 185
#define OP_DOR 186
#define OP_COND_EXPR 187
#define OP_ANDASSIGN 188
#define OP_ORASSIGN 189
#define OP_DORASSIGN 190
#define OP_ENTERSUB 191
#define OP_LEAVESUB 192
#define OP_LEAVESUBLV 193
#define OP_ARGCHECK 194
#define OP_ARGELEM 195
#define OP_ARGDEFELEM 196
#define OP_CALLER 197
#define OP_WARN 198
#define OP_DIE 199
#define OP_RESET 200
#define OP_LINESEQ 201
#define OP_NEXTSTATE 202
#define OP_DBSTATE 203
#define OP_UNSTACK 204
#define OP_ENTER 205
#define OP_LEAVE 206
#define OP_SCOPE 207
#define OP_ENTERITER 208
#define OP_ITER 209
#define OP_ENTERLOOP 210
#define OP_LEAVELOOP 211
#define OP_RETURN 212
#define OP_LAST 213
#define OP_NEXT 214
#define OP_REDO 215
#define OP_DUMP 216
#define OP_GOTO 217
#define OP_EXIT 218
#define OP_METHOD 219
#define OP_METHOD_NAMED 220
#define OP_METHOD_SUPER 221
#define OP_METHOD_REDIR 222
#define OP_METHOD_REDIR_SUPER 223
#define OP_ENTERGIVEN 224
#define OP_LEAVEGIVEN 225
#define OP_ENTERWHEN 226
#define OP_LEAVEWHEN 227
#define OP_BREAK 228
#define OP_CONTINUE 229
#define OP_OPEN 230
#define OP_CLOSE 231
#define OP_PIPE_OP 232
#define OP_FILENO 233
#define OP_UMASK 234
#define OP_BINMODE 235
#define OP_TIE 236
#define OP_UNTIE 237
#define OP_TIED 238
#define OP_DBMOPEN 239
#define OP_DBMCLOSE 240
#define OP_SSELECT 241
#define OP_SELECT 242
#define OP_GETC 243
#define OP_READ 244
#define OP_ENTERWRITE 245
#define OP_LEAVEWRITE 246
#define OP_PRTF 247
#define OP_PRINT 248
#define OP_SAY 249
#define OP_SYSOPEN 250
#define OP_SYSSEEK 251
#define OP_SYSREAD 252
#define OP_SYSWRITE 253
#define OP_EOF 254
#define OP_TELL 255
#define OP_SEEK 256
#define OP_TRUNCATE 257
#define OP_FCNTL 258
#define OP_IOCTL 259
#define OP_FLOCK 260
#define OP_SEND 261
#define OP_RECV 262
#define OP_SOCKET 263
#define OP_SOCKPAIR 264
#define OP_BIND 265
#define OP_CONNECT 266
#define OP_LISTEN 267
#define OP_ACCEPT 268
#define OP_SHUTDOWN 269
#define OP_GSOCKOPT 270
#define OP_SSOCKOPT 271
#define OP_GETSOCKNAME 272
#define OP_GETPEERNAME 273
#define OP_LSTAT 274
#define OP_STAT 275
#define OP_FTRREAD 276
#define OP_FTRWRITE 277
#define OP_FTREXEC 278
#define OP_FTEREAD 279
#define OP_FTEWRITE 280
#define OP_FTEEXEC 281
#define OP_FTIS 282
#define OP_FTSIZE 283
#define OP_FTMTIME 284
#define OP_FTATIME 285
#define OP_FTCTIME 286
#define OP_FTROWNED 287
#define OP_FTEOWNED 288
#define OP_FTZERO 289
#define OP_FTSOCK 290
#define OP_FTCHR 291
#define OP_FTBLK 292
#define OP_FTFILE 293
#define OP_FTDIR 294
#define OP_FTPIPE 295
#define OP_FTSUID 296
#define OP_FTSGID 297
#define OP_FTSVTX 298
#define OP_FTLINK 299
#define OP_FTTTY 300
#define OP_FTTEXT 301
#define OP_FTBINARY 302
#define OP_CHDIR 303
#define OP_CHOWN 304
#define OP_CHROOT 305
#define OP_UNLINK 306
#define OP_CHMOD 307
#define OP_UTIME 308
#define OP_RENAME 309
#define OP_LINK 310
#define OP_SYMLINK 311
#define OP_READLINK 312
#define OP_MKDIR 313
#define OP_RMDIR 314
#define OP_OPEN_DIR 315
#define OP_READDIR 316
#define OP_TELLDIR 317
#define OP_SEEKDIR 318
#define OP_REWINDDIR 319
#define OP_CLOSEDIR 320
#define OP_FORK 321
#define OP_WAIT 322
#define OP_WAITPID 323
#define OP_SYSTEM 324
#define OP_EXEC 325
#define OP_KILL 326
#define OP_GETPPID 327
#define OP_GETPGRP 328
#define OP_SETPGRP 329
#define OP_GETPRIORITY 330
#define OP_SETPRIORITY 331
#define OP_TIME 332
#define OP_TMS 333
#define OP_LOCALTIME 334
#define OP_GMTIME 335
#define OP_ALARM 336
#define OP_SLEEP 337
#define OP_SHMGET 338
#define OP_SHMCTL 339
#define OP_SHMREAD 340
#define OP_SHMWRITE 341
#define OP_MSGGET 342
#define OP_MSGCTL 343
#define OP_MSGSND 344
#define OP_MSGRCV 345
#define OP_SEMOP 346
#define OP_SEMGET 347
#define OP_SEMCTL 348
#define OP_REQUIRE 349
#define OP_DOFILE 350
#define OP_HINTSEVAL 351
#define OP_ENTEREVAL 352
#define OP_LEAVEEVAL 353
#define OP_ENTERTRY 354
#define OP_LEAVETRY 355
#define OP_GHBYNAME 356
#define OP_GHBYADDR 357
#define OP_GHOSTENT 358
#define OP_GNBYNAME 359
#define OP_GNBYADDR 360
#define OP_GNETENT 361
#define OP_GPBYNAME 362
#define OP_GPBYNUMBER 363
#define OP_GPROTOENT 364
#define OP_GSBYNAME 365
#define OP_GSBYPORT 366
#define OP_GSERVENT 367
#define OP_SHOSTENT 368
#define OP_SNETENT 369
#define OP_SPROTOENT 370
#define OP_SSERVENT 371
#define OP_EHOSTENT 372
#define OP_ENETENT 373
#define OP_EPROTOENT 374
#define OP_ESERVENT 375
#define OP_GPWNAM 376
#define OP_GPWUID 377
#define OP_GPWENT 378
#define OP_SPWENT 379
#define OP_EPWENT 380
#define OP_GGRNAM 381
#define OP_GGRGID 382
#define OP_GGRENT 383
#define OP_SGRENT 384
#define OP_EGRENT 385
#define OP_GETLOGIN 386
#define OP_SYSCALL 387
#define OP_LOCK 388
#define OP_ONCE 389
#define OP_CUSTOM 390
#define OP_COREARGS 391
#define OP_AVHVSWITCH 392
#define OP_RUNCV 393
#define OP_FC 394
#define OP_PADCV 395
#define OP_INTROCV 396
#define OP_CLONECV 397
#define OP_PADRANGE 398
#define OP_REFASSIGN 399
#define OP_LVREF 400
#define OP_LVREFSLICE 401
#define OP_LVAVREF 402
#define OP_ANONCONST 403
#define OP_ISA 404
#define OP_CMPCHAIN_AND 405
#define OP_CMPCHAIN_DUP 406
#define OP_ENTERTRYCATCH 407
#define OP_LEAVETRYCATCH 408
#define OP_POPTRY 409
#define OP_CATCH 410
#define OP_PUSHDEFER 411
#define OP_IS_BOOL 412
#define OP_IS_WEAK 413
#define OP_WEAKEN 414
#define OP_UNWEAKEN 415
#define OP_BLESSED 416
#define OP_REFADDR 417
#define OP_REFTYPE 418
#define OP_CEIL 419
#define OP_FLOOR 420
#define OP_IS_TAINTED 421
#define OP_HELEMEXISTSOR 422
#define OP_METHSTART 423
#define OP_INITFIELD 424
#define OP_CLASSNAME 425

#define OP_max 426
#ifndef MAXO
#define MAXO 426
#endif

static const char *const goperl_op_name_v[] = {
    "null",
    "stub",
    "scalar",
    "pushmark",
    "wantarray",
    "const",
    "gvsv",
    "gv",
    "gelem",
    "padsv",
    "padsv_store",
    "padav",
    "padhv",
    "padany",
    "rv2gv",
    "rv2sv",
    "av2arylen",
    "rv2cv",
    "anoncode",
    "prototype",
    "refgen",
    "srefgen",
    "ref",
    "bless",
    "backtick",
    "glob",
    "readline",
    "rcatline",
    "regcmaybe",
    "regcreset",
    "regcomp",
    "match",
    "qr",
    "subst",
    "substcont",
    "trans",
    "transr",
    "sassign",
    "aassign",
    "chop",
    "schop",
    "chomp",
    "schomp",
    "defined",
    "undef",
    "study",
    "pos",
    "preinc",
    "i_preinc",
    "predec",
    "i_predec",
    "postinc",
    "i_postinc",
    "postdec",
    "i_postdec",
    "pow",
    "multiply",
    "i_multiply",
    "divide",
    "i_divide",
    "modulo",
    "i_modulo",
    "repeat",
    "add",
    "i_add",
    "subtract",
    "i_subtract",
    "concat",
    "multiconcat",
    "stringify",
    "left_shift",
    "right_shift",
    "lt",
    "i_lt",
    "gt",
    "i_gt",
    "le",
    "i_le",
    "ge",
    "i_ge",
    "eq",
    "i_eq",
    "ne",
    "i_ne",
    "ncmp",
    "i_ncmp",
    "slt",
    "sgt",
    "sle",
    "sge",
    "seq",
    "sne",
    "scmp",
    "bit_and",
    "bit_xor",
    "bit_or",
    "nbit_and",
    "nbit_xor",
    "nbit_or",
    "sbit_and",
    "sbit_xor",
    "sbit_or",
    "negate",
    "i_negate",
    "not",
    "complement",
    "ncomplement",
    "scomplement",
    "smartmatch",
    "atan2",
    "sin",
    "cos",
    "rand",
    "srand",
    "exp",
    "log",
    "sqrt",
    "int",
    "hex",
    "oct",
    "abs",
    "length",
    "substr",
    "substr_left",
    "vec",
    "index",
    "rindex",
    "sprintf",
    "formline",
    "ord",
    "chr",
    "crypt",
    "ucfirst",
    "lcfirst",
    "uc",
    "lc",
    "quotemeta",
    "rv2av",
    "aelemfast",
    "aelemfast_lex",
    "aelemfastlex_store",
    "aelem",
    "aslice",
    "kvaslice",
    "aeach",
    "avalues",
    "akeys",
    "each",
    "values",
    "keys",
    "delete",
    "exists",
    "rv2hv",
    "helem",
    "hslice",
    "kvhslice",
    "multideref",
    "unpack",
    "pack",
    "split",
    "join",
    "list",
    "lslice",
    "anonlist",
    "anonhash",
    "emptyavhv",
    "splice",
    "push",
    "pop",
    "shift",
    "unshift",
    "sort",
    "reverse",
    "grepstart",
    "grepwhile",
    "anystart",
    "allstart",
    "anywhile",
    "mapstart",
    "mapwhile",
    "range",
    "flip",
    "flop",
    "and",
    "or",
    "xor",
    "dor",
    "cond_expr",
    "andassign",
    "orassign",
    "dorassign",
    "entersub",
    "leavesub",
    "leavesublv",
    "argcheck",
    "argelem",
    "argdefelem",
    "caller",
    "warn",
    "die",
    "reset",
    "lineseq",
    "nextstate",
    "dbstate",
    "unstack",
    "enter",
    "leave",
    "scope",
    "enteriter",
    "iter",
    "enterloop",
    "leaveloop",
    "return",
    "last",
    "next",
    "redo",
    "dump",
    "goto",
    "exit",
    "method",
    "method_named",
    "method_super",
    "method_redir",
    "method_redir_super",
    "entergiven",
    "leavegiven",
    "enterwhen",
    "leavewhen",
    "break",
    "continue",
    "open",
    "close",
    "pipe_op",
    "fileno",
    "umask",
    "binmode",
    "tie",
    "untie",
    "tied",
    "dbmopen",
    "dbmclose",
    "sselect",
    "select",
    "getc",
    "read",
    "enterwrite",
    "leavewrite",
    "prtf",
    "print",
    "say",
    "sysopen",
    "sysseek",
    "sysread",
    "syswrite",
    "eof",
    "tell",
    "seek",
    "truncate",
    "fcntl",
    "ioctl",
    "flock",
    "send",
    "recv",
    "socket",
    "sockpair",
    "bind",
    "connect",
    "listen",
    "accept",
    "shutdown",
    "gsockopt",
    "ssockopt",
    "getsockname",
    "getpeername",
    "lstat",
    "stat",
    "ftrread",
    "ftrwrite",
    "ftrexec",
    "fteread",
    "ftewrite",
    "fteexec",
    "ftis",
    "ftsize",
    "ftmtime",
    "ftatime",
    "ftctime",
    "ftrowned",
    "fteowned",
    "ftzero",
    "ftsock",
    "ftchr",
    "ftblk",
    "ftfile",
    "ftdir",
    "ftpipe",
    "ftsuid",
    "ftsgid",
    "ftsvtx",
    "ftlink",
    "fttty",
    "fttext",
    "ftbinary",
    "chdir",
    "chown",
    "chroot",
    "unlink",
    "chmod",
    "utime",
    "rename",
    "link",
    "symlink",
    "readlink",
    "mkdir",
    "rmdir",
    "open_dir",
    "readdir",
    "telldir",
    "seekdir",
    "rewinddir",
    "closedir",
    "fork",
    "wait",
    "waitpid",
    "system",
    "exec",
    "kill",
    "getppid",
    "getpgrp",
    "setpgrp",
    "getpriority",
    "setpriority",
    "time",
    "tms",
    "localtime",
    "gmtime",
    "alarm",
    "sleep",
    "shmget",
    "shmctl",
    "shmread",
    "shmwrite",
    "msgget",
    "msgctl",
    "msgsnd",
    "msgrcv",
    "semop",
    "semget",
    "semctl",
    "require",
    "dofile",
    "hintseval",
    "entereval",
    "leaveeval",
    "entertry",
    "leavetry",
    "ghbyname",
    "ghbyaddr",
    "ghostent",
    "gnbyname",
    "gnbyaddr",
    "gnetent",
    "gpbyname",
    "gpbynumber",
    "gprotoent",
    "gsbyname",
    "gsbyport",
    "gservent",
    "shostent",
    "snetent",
    "sprotoent",
    "sservent",
    "ehostent",
    "enetent",
    "eprotoent",
    "eservent",
    "gpwnam",
    "gpwuid",
    "gpwent",
    "spwent",
    "epwent",
    "ggrnam",
    "ggrgid",
    "ggrent",
    "sgrent",
    "egrent",
    "getlogin",
    "syscall",
    "lock",
    "once",
    "custom",
    "coreargs",
    "avhvswitch",
    "runcv",
    "fc",
    "padcv",
    "introcv",
    "clonecv",
    "padrange",
    "refassign",
    "lvref",
    "lvrefslice",
    "lvavref",
    "anonconst",
    "isa",
    "cmpchain_and",
    "cmpchain_dup",
    "entertrycatch",
    "leavetrycatch",
    "poptry",
    "catch",
    "pushdefer",
    "is_bool",
    "is_weak",
    "weaken",
    "unweaken",
    "blessed",
    "refaddr",
    "reftype",
    "ceil",
    "floor",
    "is_tainted",
    "helemexistsor",
    "methstart",
    "initfield",
    "classname",
    "freed",
};
#define PL_op_name goperl_op_name_v

#endif /* GOPERL_OPNAMES_H */
