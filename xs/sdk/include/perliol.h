/* perliol.h — the layer-author view of PerlIO for the goperl XS SDK.
 *
 * A module that includes this header defines layers (a PerlIO_funcs table
 * given to PerlIO_define_layer), walks the layers of handles, pushes and
 * pops layers, and calls perl's own layer functions, as on perl.
 *
 * A layer of native code's: the interpreter registers a class under the
 * module's name, with the module's size (0 for a pseudo-layer, which has no
 * struct of its own and does its work in Pushed), kind and slots. A slot
 * left NULL is NULL there; a slot that names one of perl's own layer
 * functions (PerlIOBuf_read, PerlIOBase_eof, ... below) is that function,
 * which the interpreter runs on the layer itself; every other slot is the
 * module's, and the interpreter calls it here, with the slot's arguments.
 *
 * The layer's struct (funcs->size bytes) is the module's, on this side: its
 * PerlIOl part and, for a layer whose struct is a PerlIOBuf (its table
 * names functions of perl's buffer layer, or it gives its layer to one of
 * them), its PerlIOBuf part are kept in step with the interpreter's layer
 * at every crossing: when a slot of the module's is called, around each
 * handle operation native code makes while the struct is here, and when
 * the outermost native call returns. PERLIO_K_BUFFERED alone says nothing
 * of the struct (PerlIO::Util's :dir keeps its buffer in its own struct).
 * The buffer is in the interpreter's memory, where perl reads and writes
 * it; a buffer the layer allocated itself (Newx in its Get_base) moves
 * there at the next crossing, and the layer's struct points at the moved
 * one from then on. The interpreter owns it then, as perl's
 * PerlIOBuf_close, which frees it, expects. A buffer inside the layer's
 * own struct is the interpreter's layer at the same place: a pointer into
 * the struct that a slot answers (Get_ptr, Get_base) points into the
 * interpreter's layer there, where the struct's bytes from there on go for
 * perl to read, and a pointer perl gives back into it (Set_ptrcnt) points
 * into the struct (perl writes no struct of a layer but through its PerlIOl
 * and PerlIOBuf parts). A layer popped while one of its own functions runs
 * (an "autopop" in Pushed) is freed when that function returns.
 *
 * Handles are views (goperl_pio_view): f and PerlIONext(f) stand for the
 * interpreter's handles, and a layer of a native code's class in a view is
 * that layer's struct. perl's own layer functions named here
 * (PerlIOBuf_fill, PerlIOBase_open, ...) and the slots of the tables of
 * the interpreter's layers (PerlIO_perlio, a table PerlIO_find_layer
 * gives) run in the interpreter, on its layers. */
#ifndef GOPERL_XS_SDK_PERLIOL_H
#define GOPERL_XS_SDK_PERLIOL_H
#pragma GCC system_header

#include "perl.h"
#include <stddef.h>

/* The layer-author view replaces the handle type, and the handle functions
 * that take a layered handle (defined further down). The other handle
 * functions of perl.h stay: they take any handle of the interpreter, and a
 * view stands for one. */
#undef PerlIO
#undef PerlIO_flush
#undef PerlIO_read
#undef PerlIO_eof
#undef PerlIO_error
#undef PerlIO_fill
#undef PerlIO_get_ptr
#undef PerlIO_get_cnt
#undef PerlIO_set_ptrcnt
#undef PerlIO_get_base
#undef PerlIO_fast_gets
#undef IoIFP
#undef IoOFP
#undef PerlIO_stdin
#undef PerlIO_stdout
#undef PerlIO_stderr

typedef struct _PerlIO PerlIOl;
typedef struct _PerlIO_funcs PerlIO_funcs;
typedef PerlIOl *PerlIO;
#define STDCHAR char
/* A list of layers of the interpreter's (the layers an Open gets,
 * PL_known_layers, PL_def_layerlist), as perl's: its entries' tables (views
 * of the interpreter's, or the module's own) and arguments, made for the
 * native call from the interpreter's list as it is then (goperl_pio_list);
 * the functions given one work on the interpreter's (gaddr). */
typedef struct goperl_perlio_list PerlIO_list_t;

struct _PerlIO {
    PerlIOl *next;
    PerlIO_funcs *tab;
    U32 flags;
    int err;
    PerlIOl *head;
};

struct _PerlIO_funcs {
    Size_t fsize;
    const char *name;
    Size_t size;
    U32 kind;
    IV (*Pushed)(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab);
    IV (*Popped)(pTHX_ PerlIO *f);
    PerlIO *(*Open)(pTHX_ PerlIO_funcs *tab, PerlIO_list_t *layers, IV n,
                    const char *mode, int fd, int imode, int perm, PerlIO *old,
                    int narg, SV **args);
    IV (*Binmode)(pTHX_ PerlIO *f);
    SV *(*Getarg)(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags);
    IV (*Fileno)(pTHX_ PerlIO *f);
    PerlIO *(*Dup)(pTHX_ PerlIO *f, PerlIO *o, CLONE_PARAMS *param, int flags);
    SSize_t (*Read)(pTHX_ PerlIO *f, void *vbuf, Size_t count);
    SSize_t (*Unread)(pTHX_ PerlIO *f, const void *vbuf, Size_t count);
    SSize_t (*Write)(pTHX_ PerlIO *f, const void *vbuf, Size_t count);
    IV (*Seek)(pTHX_ PerlIO *f, Off_t offset, int whence);
    Off_t (*Tell)(pTHX_ PerlIO *f);
    IV (*Close)(pTHX_ PerlIO *f);
    IV (*Flush)(pTHX_ PerlIO *f);
    IV (*Fill)(pTHX_ PerlIO *f);
    IV (*Eof)(pTHX_ PerlIO *f);
    IV (*Error)(pTHX_ PerlIO *f);
    void (*Clearerr)(pTHX_ PerlIO *f);
    void (*Setlinebuf)(pTHX_ PerlIO *f);
    STDCHAR *(*Get_base)(pTHX_ PerlIO *f);
    Size_t (*Get_bufsiz)(pTHX_ PerlIO *f);
    STDCHAR *(*Get_ptr)(pTHX_ PerlIO *f);
    SSize_t (*Get_cnt)(pTHX_ PerlIO *f);
    void (*Set_ptrcnt)(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt);
};

typedef struct {
    PerlIO_funcs *funcs;
    SV *arg;
} PerlIO_pair_t;
struct goperl_perlio_list {
    IV refcnt;
    IV cur;
    IV len;
    PerlIO_pair_t *array;
    uint64_t gaddr; /* the interpreter's list */
};
#define GOPERL_PIO_LIST_ADDR(l) ((l) ? ((const PerlIO_list_t *)(l))->gaddr : 0)

typedef struct {
    struct _PerlIO base;
    STDCHAR *buf;
    STDCHAR *end;
    STDCHAR *ptr;
    Off_t posn;
    Size_t bufsiz;
    IV oneword;
} PerlIOBuf;

#define PERLIO_K_RAW 0x00000001
#define PERLIO_K_BUFFERED 0x00000002
#define PERLIO_K_CANCRLF 0x00000004
#define PERLIO_K_FASTGETS 0x00000008
#define PERLIO_K_DUMMY 0x00000010
#define PERLIO_K_UTF8 0x00008000
#define PERLIO_K_DESTRUCT 0x00010000
#define PERLIO_K_MULTIARG 0x00020000

#define PERLIO_F_EOF 0x00000100
#define PERLIO_F_CANWRITE 0x00000200
#define PERLIO_F_CANREAD 0x00000400
#define PERLIO_F_ERROR 0x00000800
#define PERLIO_F_TRUNCATE 0x00001000
#define PERLIO_F_APPEND 0x00002000
#define PERLIO_F_CRLF 0x00004000
#define PERLIO_F_UTF8 0x00008000
#define PERLIO_F_UNBUF 0x00010000
#define PERLIO_F_WRBUF 0x00020000
#define PERLIO_F_RDBUF 0x00040000
#define PERLIO_F_LINEBUF 0x00080000
#define PERLIO_F_TEMP 0x00100000
#define PERLIO_F_OPEN 0x00200000
#define PERLIO_F_FASTGETS 0x00400000
#define PERLIO_F_TTY 0x00800000
#define PERLIO_F_NOTREG 0x01000000
#define PERLIO_F_CLEARED 0x02000000

#define PERLIOBUF_DEFAULT_BUFSIZ 8192
#define PERLIO_FUNCS_DECL(f) const PerlIO_funcs f
#define PERLIO_FUNCS_CAST(f) ((PerlIO_funcs *)(void *)(f))

#define PerlIOValid(f) \
    ({ GOPERL_ONCE(goperl_PerlIOValid_f_, f); \
       goperl_PerlIOValid_f_ && *goperl_PerlIOValid_f_; })
#define PerlIOBase(f) (*(f))
#define PerlIOSelf(f, type) ((type *)PerlIOBase(f))
#define PerlIONext(f) (&PerlIOBase(f)->next)

/* ---- the slots, and perl's own layer functions ----------------------------
 * The slots of PerlIO_funcs in their order, and the functions of perl's
 * layers by number, with their shapes and whether they take the layer's
 * struct for a PerlIOBuf: the interpreter (perl.cc, goperlio_stock) has the
 * same list in the same order. KEEP IN SYNC. */
enum {
    GOPERL_PIO_SLOT_PUSHED = 0,
    GOPERL_PIO_SLOT_POPPED,
    GOPERL_PIO_SLOT_OPEN,
    GOPERL_PIO_SLOT_BINMODE,
    GOPERL_PIO_SLOT_GETARG,
    GOPERL_PIO_SLOT_FILENO,
    GOPERL_PIO_SLOT_DUP,
    GOPERL_PIO_SLOT_READ,
    GOPERL_PIO_SLOT_UNREAD,
    GOPERL_PIO_SLOT_WRITE,
    GOPERL_PIO_SLOT_SEEK,
    GOPERL_PIO_SLOT_TELL,
    GOPERL_PIO_SLOT_CLOSE,
    GOPERL_PIO_SLOT_FLUSH,
    GOPERL_PIO_SLOT_FILL,
    GOPERL_PIO_SLOT_EOF,
    GOPERL_PIO_SLOT_ERROR,
    GOPERL_PIO_SLOT_CLEARERR,
    GOPERL_PIO_SLOT_SETLINEBUF,
    GOPERL_PIO_SLOT_GET_BASE,
    GOPERL_PIO_SLOT_GET_BUFSIZ,
    GOPERL_PIO_SLOT_GET_PTR,
    GOPERL_PIO_SLOT_GET_CNT,
    GOPERL_PIO_SLOT_SET_PTRCNT,
    GOPERL_PIO_SLOTS,
    /* the hook after a Popped of perl's (or none): the layer's struct goes */
    GOPERL_PIO_HOOK_DROP = GOPERL_PIO_SLOTS
};
#define GOPERL_PIO_SLOT_FN(tab, s) (((void (**)(void)) & (tab)->Pushed)[s])
#define GOPERL_PIO_STOCK(X)                                                                  \
    X(1, PerlIOBase_pushed, PUSHED, 0) X(2, PerlIOBuf_pushed, PUSHED, 1)                      \
    X(3, PerlIOCrlf_pushed, PUSHED, 1) X(4, PerlIOPending_pushed, PUSHED, 1)                  \
    X(5, PerlIOPop_pushed, PUSHED, 0) X(6, PerlIORaw_pushed, PUSHED, 0)                       \
    X(7, PerlIOStdio_pushed, PUSHED, 0) X(8, PerlIOUnix_pushed, PUSHED, 0)                    \
    X(9, PerlIOUtf8_pushed, PUSHED, 0) X(10, PerlIOBase_popped, F_IV, 0)                      \
    X(11, PerlIOBuf_popped, F_IV, 1) X(12, PerlIOBase_noop_ok, F_IV, 0)                       \
    X(13, PerlIOBase_noop_fail, F_IV, 0) X(14, PerlIOBase_binmode, F_IV, 0)                   \
    X(15, PerlIOCrlf_binmode, F_IV, 1) X(16, PerlIOBase_fileno, F_IV, 0)                      \
    X(17, PerlIOStdio_fileno, F_IV, 0) X(18, PerlIOUnix_fileno, F_IV, 0)                      \
    X(19, PerlIOBase_close, F_IV, 0) X(20, PerlIOBuf_close, F_IV, 1)                          \
    X(21, PerlIOPending_close, F_IV, 1) X(22, PerlIOStdio_close, F_IV, 0)                     \
    X(23, PerlIOUnix_close, F_IV, 0) X(24, PerlIOBuf_flush, F_IV, 1)                          \
    X(25, PerlIOCrlf_flush, F_IV, 1) X(26, PerlIOPending_flush, F_IV, 1)                      \
    X(27, PerlIOStdio_flush, F_IV, 0) X(28, PerlIOBuf_fill, F_IV, 1)                          \
    X(29, PerlIOPending_fill, F_IV, 1) X(30, PerlIOStdio_fill, F_IV, 0)                       \
    X(31, PerlIOBase_eof, F_IV, 0) X(32, PerlIOStdio_eof, F_IV, 0)                            \
    X(33, PerlIOBase_error, F_IV, 0) X(34, PerlIOStdio_error, F_IV, 0)                        \
    X(35, PerlIOBase_open, OPEN, 0) X(36, PerlIOBuf_open, OPEN, 1)                            \
    X(37, PerlIOStdio_open, OPEN, 0) X(38, PerlIOUnix_open, OPEN, 0)                          \
    X(39, PerlIOBase_dup, DUP, 0) X(40, PerlIOBuf_dup, DUP, 1)                                \
    X(41, PerlIOStdio_dup, DUP, 0) X(42, PerlIOUnix_dup, DUP, 0)                              \
    X(43, PerlIOBase_read, READ, 0) X(44, PerlIOBuf_read, READ, 1)                            \
    X(45, PerlIOPending_read, READ, 1) X(46, PerlIOStdio_read, READ, 0)                       \
    X(47, PerlIOUnix_read, READ, 0) X(48, PerlIOBase_unread, WRITE, 0)                        \
    X(49, PerlIOBuf_unread, WRITE, 1) X(50, PerlIOCrlf_unread, WRITE, 1)                      \
    X(51, PerlIOStdio_unread, WRITE, 0) X(52, PerlIOBuf_write, WRITE, 1)                      \
    X(53, PerlIOCrlf_write, WRITE, 1) X(54, PerlIOStdio_write, WRITE, 0)                      \
    X(55, PerlIOUnix_write, WRITE, 0) X(56, PerlIOBuf_seek, SEEK, 1)                          \
    X(57, PerlIOPending_seek, SEEK, 1) X(58, PerlIOStdio_seek, SEEK, 0)                       \
    X(59, PerlIOUnix_seek, SEEK, 0) X(60, PerlIOBuf_tell, F_OFF, 1)                           \
    X(61, PerlIOStdio_tell, F_OFF, 0) X(62, PerlIOUnix_tell, F_OFF, 0)                        \
    X(63, PerlIOBase_clearerr, F_VOID, 0) X(64, PerlIOStdio_clearerr, F_VOID, 0)              \
    X(65, PerlIOBase_setlinebuf, F_VOID, 0) X(66, PerlIOStdio_setlinebuf, F_VOID, 0)          \
    X(67, PerlIOBuf_get_base, F_PTR, 1) X(68, PerlIOBuf_get_ptr, F_PTR, 1)                    \
    X(69, PerlIOBuf_bufsiz, F_SIZE, 1) X(70, PerlIOBuf_get_cnt, F_SSIZE, 1)                   \
    X(71, PerlIOCrlf_get_cnt, F_SSIZE, 1) X(72, PerlIOBuf_set_ptrcnt, SETPTRCNT, 1)           \
    X(73, PerlIOCrlf_set_ptrcnt, SETPTRCNT, 1) X(74, PerlIOPending_set_ptrcnt, SETPTRCNT, 1)
/* a function of the list that takes the layer's struct for a PerlIOBuf */
#define GOPERL_PIO_STOCK_BUFFER 0x100

/* ---- the registry of the instance -----------------------------------------
 * The structs of layers of native code's, the views of the interpreter's
 * tables, and which structs are here now (activations, newest last: a
 * struct is here from its first activation to its last release). It hangs
 * off the instance's shared block (pio_layers), behind the head perl.h
 * knows, by which the handle operations of perl.h keep the structs in step
 * and the end of the outermost native call hands them back. */
typedef struct goperl_iol {
    struct goperl_iol *next;
    uint64_t ltok;     /* the interpreter's layer */
    PerlIO_funcs *tab; /* the module's table */
    int32_t act;       /* activations */
    int32_t dead;      /* popped: freed at its last release */
    int32_t buffered;  /* the struct starts with a PerlIOBuf */
    int32_t pad_;
    /* the table of the blocks of the module's allocators (its own copy of
     * goperl_block_forget), for a buffer the layer allocated itself */
    int (*block_forget)(const void *p, size_t *size);
    /* the fields as the interpreter's layer had them when the struct came
     * over (GOPERL_IOL_F_*): only those native code changed since go back,
     * so that what the interpreter did to the layer meanwhile (Perl code an
     * XSUB called read the handle) stays */
    uint64_t seen[7];
} __attribute__((aligned(16))) goperl_iol_t;
enum {
    GOPERL_IOL_F_FLAGS = 0,
    GOPERL_IOL_F_ERR,
    GOPERL_IOL_F_BUF,
    GOPERL_IOL_F_END,
    GOPERL_IOL_F_PTR,
    GOPERL_IOL_F_POSN,
    GOPERL_IOL_F_BUFSIZ
};
#define GOPERL_IOL_STRUCT(s) ((PerlIOl *)((goperl_iol_t *)(s) + 1))

/* A view of a table of the interpreter's: the table (its name, size and
 * kind, and for each slot that is not NULL a function that runs the slot
 * there) and the table's address. */
typedef struct goperl_pio_fview {
    PerlIO_funcs tab;
    uint64_t gtab;
    struct goperl_pio_fview *next;
} goperl_pio_fview_t;

/* A class of the module's: its table and the table of the blocks of the
 * allocators of the module that defined it (the module's own copy of
 * goperl_block_forget: a slot may run under another module's copy of this
 * code, as the loader has one entry for the layers of every module). */
typedef struct goperl_pio_class {
    struct goperl_pio_class *next;
    const PerlIO_funcs *tab;
    int (*block_forget)(const void *p, size_t *size);
} goperl_pio_class_t;

struct goperl_pio_kept;
typedef struct goperl_pio_reg {
    goperl_pio_layers_head_t head;
    goperl_pio_class_t *classes;
    goperl_iol_t *layers;
    goperl_iol_t *dying; /* popped while here: freed at their last release */
    goperl_iol_t **acts;
    int32_t nacts, capacts;
    goperl_pio_fview_t *fviews;
    struct goperl_pio_kept *kept; /* the handles native code may keep */
    /* the push native code makes of a class of native code's whose Pushed
     * is native code's, while the interpreter makes it: the class, and the
     * mode given, which that Pushed gets (goperl_pio_push) */
    const PerlIO_funcs *push_tab;
    const char *push_mode;
} goperl_pio_reg_t;

static goperl_pio_reg_t *goperl_pio_reg(goperl_frame_t *f);

/* The struct of a layer in step with the interpreter's layer (in: from
 * it, out: to it), and of a buffered layer the PerlIOBuf part too: the
 * buffer pointers are the interpreter's, the emergency buffer (oneword)
 * each side's own. */
#define GOPERL_IOL_ONEWORD offsetof(PerlIOBuf, oneword)
static void goperl_iol_in(goperl_frame_t *f, goperl_iol_t *s) {
    if (s->dead) return;
    const char *g = (const char *)goperl_guest_mem(f, s->ltok);
    PerlIOl *const l = GOPERL_IOL_STRUCT(s);
    memcpy(&l->flags, g + offsetof(PerlIOl, flags), sizeof l->flags);
    memcpy(&l->err, g + offsetof(PerlIOl, err), sizeof l->err);
    s->seen[GOPERL_IOL_F_FLAGS] = l->flags;
    s->seen[GOPERL_IOL_F_ERR] = (uint64_t)(int64_t)l->err;
    if (!s->buffered) return;
    PerlIOBuf *const b = (PerlIOBuf *)l;
    uint64_t p[3];
    memcpy(p, g + offsetof(PerlIOBuf, buf), sizeof p);
    STDCHAR **const dst[3] = {&b->buf, &b->end, &b->ptr};
    for (int i = 0; i < 3; i++) {
        s->seen[GOPERL_IOL_F_BUF + i] = p[i];
        if (!p[i])
            *dst[i] = 0;
        else if (p[i] >= s->ltok + GOPERL_IOL_ONEWORD && p[i] <= s->ltok + sizeof(PerlIOBuf))
            *dst[i] = (STDCHAR *)&b->oneword + (p[i] - (s->ltok + GOPERL_IOL_ONEWORD));
        else
            *dst[i] = (STDCHAR *)goperl_guest_mem(f, p[i]);
    }
    memcpy(&b->posn, g + offsetof(PerlIOBuf, posn), sizeof b->posn);
    memcpy(&b->bufsiz, g + offsetof(PerlIOBuf, bufsiz), sizeof b->bufsiz);
    s->seen[GOPERL_IOL_F_POSN] = (uint64_t)(int64_t)b->posn;
    s->seen[GOPERL_IOL_F_BUFSIZ] = (uint64_t)b->bufsiz;
}
/* A buffer the layer allocated itself (a block of the module's
 * allocators) moves into the interpreter's memory, which owns it from then
 * on, and the layer's pointers into it move with it. Any other memory of
 * native code's cannot be the interpreter's buffer. */
static void goperl_iol_place_buf(goperl_frame_t *f, goperl_iol_t *s) {
    PerlIOBuf *const b = (PerlIOBuf *)GOPERL_IOL_STRUCT(s);
    STDCHAR *const from = b->buf;
    if (!from || goperl_guest_addr(f, from) ||
        (from >= (STDCHAR *)&b->oneword && from <= (STDCHAR *)(b + 1)))
        return;
    size_t n = 0;
    if (!s->block_forget || !s->block_forget(from, &n))
        goperl_croakf(f,
                      "the buffer of the PerlIO layer '%s' is memory of native code's that no "
                      "allocator of the module gave (Newx): the interpreter cannot read it",
                      s->tab->name);
    const uint64_t g = goperl_xs_opx(f, GOPERL_OP_MEM_REALLOC, 0, n ? n : 1, 0, 0, 0, 0);
    if (!g) goperl_croakf(f, "Out of memory!");
    STDCHAR *const to = (STDCHAR *)goperl_guest_mem(f, g);
    if (n) memcpy(to, from, n);
    if (b->end >= from && b->end <= from + n) b->end = to + (b->end - from);
    if (b->ptr >= from && b->ptr <= from + n) b->ptr = to + (b->ptr - from);
    b->buf = to;
    free(from);
}
static uint64_t goperl_iol_ptr_out(goperl_frame_t *f, goperl_iol_t *s, const STDCHAR *p) {
    const PerlIOBuf *const b = (const PerlIOBuf *)GOPERL_IOL_STRUCT(s);
    if (!p) return 0;
    if (p >= (const STDCHAR *)&b->oneword && p <= (const STDCHAR *)(b + 1))
        return s->ltok + GOPERL_IOL_ONEWORD + (uint64_t)(p - (const STDCHAR *)&b->oneword);
    const uint64_t g = goperl_guest_addr(f, p);
    if (!g)
        goperl_croakf(f, "a buffer pointer of the PerlIO layer '%s' is outside the interpreter's "
                         "memory",
                      s->tab->name);
    return g;
}
/* A pointer into the layer's own struct past its PerlIOl part (a buffer
 * PerlIO::Util's :dir keeps there, which its Get_ptr answers) is the
 * interpreter's layer at the same place; the struct's bytes from there on
 * go there for perl to read (0: p is not one). A layer whose struct is a
 * PerlIOBuf keeps its buffer elsewhere (goperl_iol_place_buf). */
static uint64_t goperl_iol_inner_out(goperl_frame_t *f, goperl_iol_t *s, const STDCHAR *p) {
    const char *const st = (const char *)GOPERL_IOL_STRUCT(s);
    const char *const q = (const char *)p;
    if (s->buffered || q < st + sizeof(PerlIOl) || q > st + s->tab->size) return 0;
    const size_t off = (size_t)(q - st);
    memcpy((char *)goperl_guest_mem(f, s->ltok) + off, q, s->tab->size - off);
    return s->ltok + off;
}
static STDCHAR *goperl_pio_ptr_in(goperl_frame_t *f, uint64_t g);
/* a pointer perl gives a slot: into the interpreter's layer past its
 * PerlIOl part (one goperl_iol_inner_out answered) is into the struct */
static STDCHAR *goperl_iol_inner_in(goperl_frame_t *f, goperl_iol_t *s, uint64_t g) {
    if (s && !s->buffered && g >= s->ltok + sizeof(PerlIOl) && g <= s->ltok + s->tab->size)
        return (STDCHAR *)GOPERL_IOL_STRUCT(s) + (g - s->ltok);
    return goperl_pio_ptr_in(f, g);
}
/* a field native code changed goes to the interpreter's layer */
static void goperl_iol_put(goperl_iol_t *s, char *g, int field, size_t off, uint64_t v,
                           size_t size) {
    if (s->seen[field] == v) return;
    s->seen[field] = v;
    if (size == 4) {
        const uint32_t w = (uint32_t)v;
        memcpy(g + off, &w, 4);
    } else {
        memcpy(g + off, &v, 8);
    }
}
static void goperl_iol_out(goperl_frame_t *f, goperl_iol_t *s) {
    if (s->dead) return;
    char *const g = (char *)goperl_guest_mem(f, s->ltok);
    PerlIOl *const l = GOPERL_IOL_STRUCT(s);
    goperl_iol_put(s, g, GOPERL_IOL_F_FLAGS, offsetof(PerlIOl, flags), l->flags, 4);
    goperl_iol_put(s, g, GOPERL_IOL_F_ERR, offsetof(PerlIOl, err), (uint64_t)(int64_t)l->err, 4);
    if (!s->buffered) return;
    PerlIOBuf *const b = (PerlIOBuf *)l;
    goperl_iol_place_buf(f, s);
    goperl_iol_put(s, g, GOPERL_IOL_F_BUF, offsetof(PerlIOBuf, buf), goperl_iol_ptr_out(f, s, b->buf), 8);
    goperl_iol_put(s, g, GOPERL_IOL_F_END, offsetof(PerlIOBuf, end), goperl_iol_ptr_out(f, s, b->end), 8);
    goperl_iol_put(s, g, GOPERL_IOL_F_PTR, offsetof(PerlIOBuf, ptr), goperl_iol_ptr_out(f, s, b->ptr), 8);
    goperl_iol_put(s, g, GOPERL_IOL_F_POSN, offsetof(PerlIOBuf, posn), (uint64_t)(int64_t)b->posn, 8);
    goperl_iol_put(s, g, GOPERL_IOL_F_BUFSIZ, offsetof(PerlIOBuf, bufsiz), (uint64_t)b->bufsiz, 8);
}
/* every struct that is here, to the interpreter (out) or back (in): the
 * handle operations of perl.h call it around each operation */
static void goperl_iol_sync_all(goperl_frame_t *f, int out) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    for (int32_t i = 0; i < r->nacts; i++) {
        if (out)
            goperl_iol_out(f, r->acts[i]);
        else
            goperl_iol_in(f, r->acts[i]);
    }
}
static int goperl_pio_stock_id(void (*fn)(void));
/* The struct of the interpreter's layer ltok of the class of the module's
 * table tab, made (zeroed, as perl makes a layer) the first time. */
static goperl_iol_t *goperl_iol_get(goperl_frame_t *f, uint64_t ltok, PerlIO_funcs *tab) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    for (goperl_iol_t *s = r->layers; s; s = s->next)
        if (s->ltok == ltok) return s;
    if (tab->size < sizeof(PerlIOl))
        goperl_croakf(f, "PerlIO layer '%s' has a struct smaller than PerlIOl", tab->name);
    goperl_iol_t *const s = (goperl_iol_t *)calloc(1, sizeof(goperl_iol_t) + tab->size);
    if (!s) goperl_croakf(f, "Out of memory!");
    s->ltok = ltok;
    s->tab = tab;
    for (const goperl_pio_class_t *c = r->classes; c; c = c->next)
        if (c->tab == tab) s->block_forget = c->block_forget;
    /* buffered: its table names a function of perl's that takes its struct
     * for a PerlIOBuf (or native code gives its layer to one, see
     * goperl_pio_stock_buffered) */
    if (tab->size >= sizeof(PerlIOBuf)) {
        for (int i = 0; i < GOPERL_PIO_SLOTS && !s->buffered; i++)
            s->buffered =
                (goperl_pio_stock_id(GOPERL_PIO_SLOT_FN(tab, i)) & GOPERL_PIO_STOCK_BUFFER) != 0;
    }
    GOPERL_IOL_STRUCT(s)->tab = tab;
    s->next = r->layers;
    r->layers = s;
    return s;
}
/* The struct is here (from the interpreter, the first time) until the
 * matching release. */
static void goperl_iol_activate(goperl_frame_t *f, goperl_iol_t *s) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    if (r->nacts == r->capacts) {
        const int32_t cap = r->capacts ? 2 * r->capacts : 16;
        goperl_iol_t **const a = (goperl_iol_t **)realloc(r->acts, (size_t)cap * sizeof *a);
        if (!a) goperl_croakf(f, "Out of memory!");
        r->acts = a;
        r->capacts = cap;
    }
    r->acts[r->nacts++] = s;
    r->head.active = r->nacts;
    if (s->act++ == 0) goperl_iol_in(f, s);
}
/* The activations after the mark go: a struct no activation holds any more
 * goes back to the interpreter, or is freed when its layer was popped. */
static void goperl_iol_release_to(goperl_frame_t *f, int32_t mark) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    while (r->nacts > mark) {
        goperl_iol_t *const s = r->acts[--r->nacts];
        r->head.active = r->nacts;
        if (--s->act > 0) continue;
        if (!s->dead) {
            goperl_iol_out(f, s);
            continue;
        }
        for (goperl_iol_t **pp = &r->dying; *pp; pp = &(*pp)->next)
            if (*pp == s) {
                *pp = s->next;
                break;
            }
        free(s);
    }
}
/* The layer was popped (its Popped ran): the struct goes, at once or at
 * its last release, and no view finds it any more. */
static void goperl_iol_kill(goperl_frame_t *f, uint64_t ltok) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    for (goperl_iol_t **pp = &r->layers; *pp; pp = &(*pp)->next) {
        goperl_iol_t *const s = *pp;
        if (s->ltok != ltok) continue;
        *pp = s->next;
        if (s->act) {
            s->dead = 1;
            s->next = r->dying;
            r->dying = s;
        } else {
            free(s);
        }
        return;
    }
}
static void goperl_pio_keep(goperl_frame_t *f, goperl_pio_reg_t *r);
static void goperl_iol_rest(goperl_frame_t *f) {
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    goperl_pio_keep(f, r);
    goperl_iol_release_to(f, 0);
    r->push_tab = 0; /* a push a death cut short */
    r->push_mode = 0;
}
static void goperl_pio_kept_release(goperl_pio_reg_t *r);
/* The instance is gone: its classes, structs and views go with it. */
static void goperl_pio_reg_release(void *head) {
    goperl_pio_reg_t *const r = (goperl_pio_reg_t *)head;
    while (r->classes) {
        goperl_pio_class_t *const c = r->classes;
        r->classes = c->next;
        free(c);
    }
    goperl_iol_t *lists[2] = {r->layers, r->dying};
    for (int i = 0; i < 2; i++)
        while (lists[i]) {
            goperl_iol_t *const s = lists[i];
            lists[i] = s->next;
            free(s);
        }
    free(r->acts);
    while (r->fviews) {
        goperl_pio_fview_t *const v = r->fviews;
        r->fviews = v->next;
        free(v);
    }
    goperl_pio_kept_release(r);
    free(r);
}
static void *goperl_pio_reg_clone(const void *head);
static goperl_pio_reg_t *goperl_pio_reg_new(void) {
    goperl_pio_reg_t *const r = (goperl_pio_reg_t *)calloc(1, sizeof *r);
    if (!r) return 0;
    r->head.sync = goperl_iol_sync_all;
    r->head.rest = goperl_iol_rest;
    r->head.release = goperl_pio_reg_release;
    r->head.clone = goperl_pio_reg_clone;
    return r;
}
/* A copy of the instance (perl.Clone) has the classes; it has no structs
 * of layers: the layers it copied get theirs anew, zeroed, since a layer's
 * own state is memory of native code's, which the copies cannot share
 * (nor has it the handles native code kept, which are in that memory). */
static void *goperl_pio_reg_clone(const void *head) {
    const goperl_pio_reg_t *const src = (const goperl_pio_reg_t *)head;
    goperl_pio_reg_t *const r = goperl_pio_reg_new();
    if (!r) return 0;
    for (const goperl_pio_class_t *c = src->classes; c; c = c->next) {
        goperl_pio_class_t *const k = (goperl_pio_class_t *)malloc(sizeof *k);
        if (!k) {
            goperl_pio_reg_release(r);
            return 0;
        }
        *k = *c;
        k->next = r->classes;
        r->classes = k;
    }
    return r;
}
static goperl_pio_reg_t *goperl_pio_reg(goperl_frame_t *f) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_pio_reg_t *r = (goperl_pio_reg_t *)sh->pio_layers;
    if (r) return r;
    r = goperl_pio_reg_new();
    if (!r) goperl_croakf(f, "Out of memory!");
    sh->pio_layers = r;
    return r;
}

/* ---- views of the interpreter's handles -----------------------------------
 * A handle native code is given or finds (IoIFP, IoOFP, PerlIO_stdin, the
 * f of a slot, a handle a push or an open answers) is a view made from the
 * interpreter's layers as they are then: the handle, and the next field of
 * each layer, are registered with the interpreter's handle they stand for
 * (see goperl_pio_tok), so that every handle function given one of them, or
 * a PerlIONext of one, works on the interpreter's. A layer of perl's shows
 * its table (a view of the interpreter's, goperl_pio_tab_of), its flags,
 * and of a buffered one its buffer (buf, end and ptr point into the
 * interpreter's buffer, bufsiz and posn are its); a layer of a native
 * code's class is that layer's struct (here from then on, see
 * goperl_iol_activate). */
typedef struct {
    PerlIOl *top;
    goperl_pio_slot_t reg;
} goperl_pio_handle_view_t;
typedef struct {
    PerlIOBuf buf;         /* buf.base is the layer (PerlIOBase) */
    goperl_pio_slot_t reg; /* of the next field: the handle of the layer below */
} goperl_pio_layer_view_t;
/* The words of a layer in the answer of LAYERS with b = 1: [next][name]
 * [kind][flags][buf][end][ptr][bufsiz][posn][the module's table of a layer
 * of native code's, 0 for another][the layer's table]. */
#define GOPERL_PIO_LAYER_WORDS 11
static PerlIO_funcs *goperl_pio_tab_of(goperl_frame_t *f, uint64_t gtab);
/* What a buffered layer's view holds of its buffer, from the words of the
 * layer. */
static void goperl_pio_view_buf(goperl_frame_t *f, PerlIOBuf *b, const goperl_gword_t *l) {
    b->buf = l[4] ? (STDCHAR *)goperl_guest_mem(f, l[4]) : 0;
    b->end = l[5] ? (STDCHAR *)goperl_guest_mem(f, l[5]) : 0;
    b->ptr = l[6] ? (STDCHAR *)goperl_guest_mem(f, l[6]) : 0;
    b->bufsiz = (STRLEN)l[7];
    b->posn = (Off_t)(int64_t)l[8];
    b->oneword = 0;
}
static const goperl_gword_t *goperl_pio_layer_words(goperl_frame_t *f, uint64_t tok) {
    return (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_PERLIO, tok, 1, (uint64_t)(GOPERL_PIO_LAYERS), 0, 0));
}
static void goperl_pio_register(goperl_pio_slot_t *reg, const void *slot, uint64_t tok) {
    reg->slot = slot;
    reg->tok = tok;
    reg->next = goperl_pio_views_v;
    goperl_pio_views_v = reg;
}
/* The layers of the interpreter's handle tok as a chain of views: the top
 * one, or NULL for none. */
static PerlIOl *goperl_pio_chain(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    const goperl_gword_t *w = goperl_pio_layer_words(f, tok);
    const goperl_gword_t n = w ? w[0] : 0;
    if (!n) return 0;
    /* a copy: the tables and structs below make operations, which reuse
     * the answer's memory */
    const size_t nwords = (size_t)n * GOPERL_PIO_LAYER_WORDS;
    char *const m = (char *)goperl_arena_alloc(nwords * sizeof(goperl_gword_t) +
                                               (size_t)n * sizeof(goperl_pio_layer_view_t) +
                                               (size_t)n * sizeof(PerlIOl *));
    if (!m) goperl_croakf(f, "Out of memory!");
    goperl_gword_t *const c = (goperl_gword_t *)m;
    memcpy(c, w + 1, nwords * sizeof(goperl_gword_t));
    goperl_pio_layer_view_t *const lv = (goperl_pio_layer_view_t *)(c + nwords);
    PerlIOl **const layer = (PerlIOl **)(lv + n);
    for (goperl_gword_t i = 0; i < n; i++) {
        const goperl_gword_t *l = c + GOPERL_PIO_LAYER_WORDS * i;
        if (l[9]) {
            goperl_iol_t *const s = goperl_iol_get(f, l[0], (PerlIO_funcs *)(uintptr_t)l[9]);
            goperl_iol_activate(f, s);
            layer[i] = GOPERL_IOL_STRUCT(s);
        } else {
            lv[i].buf.base.tab = goperl_pio_tab_of(f, l[10]);
            lv[i].buf.base.flags = (U32)l[3];
            goperl_pio_view_buf(f, &lv[i].buf, l);
            layer[i] = &lv[i].buf.base;
        }
        goperl_pio_register(&lv[i].reg, &layer[i]->next, l[0]);
    }
    for (goperl_gword_t i = 0; i < n; i++) layer[i]->next = i + 1 < n ? layer[i + 1] : 0;
    return layer[0];
}
static PerlIO *goperl_pio_view(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    goperl_pio_handle_view_t *const hv =
        (goperl_pio_handle_view_t *)goperl_arena_alloc(sizeof(goperl_pio_handle_view_t));
    if (!hv) goperl_croakf(f, "Out of memory!");
    hv->top = goperl_pio_chain(f, tok);
    goperl_pio_register(&hv->reg, &hv->top, tok);
    return &hv->top;
}
/* ---- handles native code may keep -----------------------------------------
 * A handle native code asks for (IoIFP, IoOFP, PerlIO_stdout, an open, a
 * dup, PerlIO_allocate) it may keep after the native call returns, as a
 * PerlIO * of perl's is the handle until it is closed: PerlIO::Util's :tee
 * keeps the handle it writes to in its layer's struct, writes to it from
 * its Write slot and walks its layers from an XSUB. Such a handle is the
 * instance's, one for each handle of the interpreter's: the call makes its
 * layers as for any view, and when the outermost native call returns, a
 * copy of them (the views of the call are in the arena; a layer of a
 * native code's class is copied as the module left its struct) is what it
 * shows until it is made again. A handle function given it, or the next
 * field of a layer of the copy, works on the interpreter's handle then
 * too. */
static goperl_pio_fview_t *goperl_pio_fview_of(goperl_frame_t *f, const PerlIO_funcs *tab);
typedef struct goperl_pio_kept {
    PerlIOl *top;
    goperl_pio_slot_t reg; /* top stands for the interpreter's handle */
    struct goperl_pio_kept *next;
    char *copy;            /* the copy of its layers: n entries */
    size_t ncopy;
} goperl_pio_kept_t;
/* an entry of a copy: the slot of the layer's next field, then the layer
 * (size bytes, rounded up to 16) */
typedef struct {
    goperl_pio_slot_t reg;
    size_t size;
} goperl_pio_kept_layer_t;
#define GOPERL_PIO_KEPT_ROUND(n) (((n) + 15) & ~(size_t)15)
#define GOPERL_PIO_KEPT_STEP(e) (sizeof(goperl_pio_kept_layer_t) + GOPERL_PIO_KEPT_ROUND((e)->size))
#define GOPERL_PIO_KEPT_LAYER(e) ((PerlIOl *)(void *)((goperl_pio_kept_layer_t *)(e) + 1))
static PerlIO *goperl_pio_handle(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    goperl_pio_kept_t *k = r->kept;
    while (k && k->reg.tok != tok) k = k->next;
    if (!k) {
        k = (goperl_pio_kept_t *)calloc(1, sizeof *k);
        if (!k) goperl_croakf(f, "Out of memory!");
        k->next = r->kept;
        r->kept = k;
        goperl_pio_register(&k->reg, &k->top, tok);
    }
    k->top = goperl_pio_chain(f, tok);
    return &k->top;
}
/* Whether the layers a kept handle shows are its copy, as it was made. */
static int goperl_pio_kept_whole(const goperl_pio_kept_t *k) {
    const char *e = k->copy;
    const PerlIOl *l = k->top;
    for (size_t i = 0; i < k->ncopy; i++) {
        if (l != GOPERL_PIO_KEPT_LAYER(e)) return 0;
        l = l->next;
        e += GOPERL_PIO_KEPT_STEP((const goperl_pio_kept_layer_t *)e);
    }
    return l == 0;
}
/* The size of a layer of a view: a layer of perl's is a view (PerlIOBuf),
 * one of a native code's class that layer's struct. */
static size_t goperl_pio_kept_size(goperl_frame_t *f, const PerlIOl *l) {
    if (!l->tab || goperl_pio_fview_of(f, l->tab) || l->tab->size < sizeof(PerlIOl))
        return sizeof(PerlIOBuf);
    return (size_t)l->tab->size;
}
/* When the outermost native call returns (before the arena goes, and the
 * structs of layers of native code's go back): each kept handle shows a
 * copy of its layers, and the views registered are the kept handles and
 * the next fields of their copies. */
static void goperl_pio_keep(goperl_frame_t *f, goperl_pio_reg_t *r) {
    /* the copies replaced go once no view registered is in them */
    size_t nk = 0, nold = 0;
    for (const goperl_pio_kept_t *k = r->kept; k; k = k->next) nk++;
    char **old = 0;
    for (goperl_pio_kept_t *k = r->kept; k; k = k->next) {
        if (goperl_pio_kept_whole(k)) continue;
        if (!old && !(old = (char **)malloc(nk * sizeof(char *)))) goperl_croakf(f, "Out of memory!");
        size_t n = 0, bytes = 0;
        for (const PerlIOl *l = k->top; l; l = l->next) {
            n++;
            bytes += sizeof(goperl_pio_kept_layer_t) + GOPERL_PIO_KEPT_ROUND(goperl_pio_kept_size(f, l));
        }
        char *const copy = n ? (char *)malloc(bytes) : 0;
        if (n && !copy) goperl_croakf(f, "Out of memory!");
        char *e = copy;
        PerlIOl *prev = 0;
        for (const PerlIOl *l = k->top; l; l = l->next) {
            goperl_pio_kept_layer_t *const h = (goperl_pio_kept_layer_t *)e;
            h->size = goperl_pio_kept_size(f, l);
            const goperl_pio_slot_t *const m = goperl_pio_view_of(&l->next);
            h->reg.slot = &GOPERL_PIO_KEPT_LAYER(h)->next;
            h->reg.tok = m ? m->tok : 0;
            memcpy(GOPERL_PIO_KEPT_LAYER(h), l, h->size);
            if (prev) prev->next = GOPERL_PIO_KEPT_LAYER(h);
            prev = GOPERL_PIO_KEPT_LAYER(h);
            e += GOPERL_PIO_KEPT_STEP(h);
        }
        if (prev) prev->next = 0;
        old[nold++] = k->copy;
        k->copy = copy;
        k->ncopy = n;
        k->top = n ? GOPERL_PIO_KEPT_LAYER(copy) : 0;
    }
    goperl_pio_views_v = 0;
    for (goperl_pio_kept_t *k = r->kept; k; k = k->next) {
        goperl_pio_register(&k->reg, &k->top, k->reg.tok);
        char *e = k->copy;
        for (size_t i = 0; i < k->ncopy; i++) {
            goperl_pio_kept_layer_t *const h = (goperl_pio_kept_layer_t *)e;
            if (h->reg.tok) goperl_pio_register(&h->reg, h->reg.slot, h->reg.tok);
            e += GOPERL_PIO_KEPT_STEP(h);
        }
    }
    for (size_t i = 0; i < nold; i++) free(old[i]);
    free(old);
}
static void goperl_pio_kept_release(goperl_pio_reg_t *r) {
    while (r->kept) {
        goperl_pio_kept_t *const k = r->kept;
        r->kept = k->next;
        free(k->copy);
        free(k);
    }
}
/* the handles perl.h's opens answer, as kept handles */
#undef PerlIO_open
#undef PerlIO_fdopen
#undef PerlIO_importFILE
#define PerlIO_open(path, mode) \
    goperl_pio_handle(_gof, goperl_pio_tok(_gof, goperl_perlio_open(_gof, (path), (mode))))
#define PerlIO_fdopen(fd, mode) \
    goperl_pio_handle(_gof, goperl_pio_tok(_gof, goperl_pio_fdopen(_gof, (fd), (mode))))
#define PerlIO_importFILE(fp, mode) \
    goperl_pio_handle(_gof, goperl_pio_tok(_gof, goperl_pio_import(_gof, (fp), (mode))))
#define PerlIO_openn(...) goperl_pio_handle(_gof, goperl_pio_tok(_gof, (PerlIO_openn)(__VA_ARGS__)))

/* The layers under a handle of a view (a handle, or the next field of a
 * layer) changed (a push, a pop, layers applied): its slot shows them. */
static void goperl_pio_view_again(goperl_frame_t *f, PerlIO *h) {
    if (!goperl_pio_view_of(h)) return;
    *h = goperl_pio_chain(f, goperl_pio_tok(f, h));
}
#define IoIFP(io) goperl_pio_handle(_gof, goperl_pio(_gof, GOPERL_PIO_IFP, (io), 0, 0, 0))
#define IoOFP(io) goperl_pio_handle(_gof, goperl_pio(_gof, GOPERL_PIO_OFP, (io), 0, 0, 0))
#define PerlIO_stdin() goperl_pio_handle(_gof, goperl_pio(_gof, GOPERL_PIO_STDIN, 0, 0, 0, 0))
#define PerlIO_stdout() goperl_pio_handle(_gof, goperl_pio(_gof, GOPERL_PIO_STDOUT, 0, 0, 0, 0))
#define PerlIO_stderr() goperl_pio_handle(_gof, goperl_pio(_gof, GOPERL_PIO_STDERR, 0, 0, 0, 0))

/* utf8_strict walks PL_perlio to flush line-buffered TTY handles; there is
 * no host-visible handle table, and the walk exits immediately on NULL. */
GOPERL_C __attribute__((weak)) PerlIOl *goperl_pl_perlio_null_v = 0;
#define PL_perlio (goperl_pl_perlio_null_v)

/* ---- calls of layer functions in the interpreter --------------------------
 * PERLIO_CALL: [n words][which: 1 a function of perl's by number, 2 a slot
 * of the table of the handle's top layer, 3 a slot of a table][the number
 * or the slot][the table], the arguments as words, then bytes (see
 * goperlio_call in perl.cc). Around it the structs that are here go to the
 * interpreter and come back. */
enum { GOPERL_PIO_CALL_STOCK = 1, GOPERL_PIO_CALL_TOP = 2, GOPERL_PIO_CALL_TAB = 3 };
static uint64_t goperl_pio_call(goperl_frame_t *f, int which, uint64_t id, uint64_t gtab,
                                const goperl_gword_t *args, size_t nargs, const void *bytes,
                                size_t nbytes) {
    const size_t nw = 4 + nargs;
    const size_t len = nw * 8 + nbytes;
    char *const req = (char *)malloc(len);
    if (!req) goperl_croakf(f, "Out of memory!");
    const goperl_gword_t head[4] = {(goperl_gword_t)nw, (goperl_gword_t)which, id, gtab};
    memcpy(req, head, sizeof head);
    if (nargs) memcpy(req + sizeof head, args, nargs * 8);
    if (nbytes) memcpy(req + nw * 8, bytes, nbytes);
    goperl_pio_sync(f, 1);
    const uint64_t r = goperl_do_op3(f, GOPERL_OP_PERLIO_CALL, 0, len, 0, req, len);
    free(req);
    goperl_pio_sync(f, 0);
    return r;
}
/* Whether a table is a view of one of the interpreter's. */
static goperl_pio_fview_t *goperl_pio_fview_of(goperl_frame_t *f, const PerlIO_funcs *tab) {
    if (!tab || !((goperl_shared_t *)f->inst)->pio_layers) return 0;
    for (goperl_pio_fview_t *v = goperl_pio_reg(f)->fviews; v; v = v->next)
        if (&v->tab == tab) return v;
    return 0;
}
/* A table native code gives the interpreter: [kind][value], 0 none, 1 a
 * table of the interpreter's (a view of one), 2 a class of the module's by
 * its table. */
static void goperl_pio_tab_arg(goperl_frame_t *f, const PerlIO_funcs *tab, goperl_gword_t *w) {
    const goperl_pio_fview_t *const v = goperl_pio_fview_of(f, tab);
    w[0] = !tab ? 0 : v ? 1 : 2;
    w[1] = !tab ? 0 : v ? v->gtab : (goperl_gword_t)(uintptr_t)tab;
}
/* A handle the interpreter answers, as a kept handle; a handle native
 * code gives it, as the interpreter's. */
#define GOPERL_PIO_H_IN(f, tok) goperl_pio_handle((f), (tok))
#define GOPERL_PIO_H_OUT(f, h) ((goperl_gword_t)goperl_pio_tok((f), (h)))
static STDCHAR *goperl_pio_ptr_in(goperl_frame_t *f, uint64_t g) {
    return g ? (STDCHAR *)goperl_guest_mem(f, g) : 0;
}
static uint64_t goperl_pio_call_f(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f,
                                  const goperl_gword_t *more, size_t nmore, const void *bytes,
                                  size_t nbytes) {
    goperl_gword_t w[4];
    w[0] = GOPERL_PIO_H_OUT(gf, f);
    for (size_t i = 0; i < nmore; i++) w[1 + i] = more[i];
    return goperl_pio_call(gf, which, id, 0, w, 1 + nmore, bytes, nbytes);
}
/* The calls by shape (which, the number or slot, the table for 3). */
static IV goperl_pio_c_pushed(goperl_frame_t *gf, int which, uint64_t id, uint64_t gtab, PerlIO *f,
                              const char *mode, SV *arg, PerlIO_funcs *tab) {
    goperl_gword_t w[5] = {GOPERL_PIO_H_OUT(gf, f), GOPERL_TOK(arg), 0, 0, mode ? 1u : 0u};
    goperl_pio_tab_arg(gf, tab, w + 2);
    const IV r = (IV)(int64_t)goperl_pio_call(gf, which, id, gtab, w, 5, mode,
                                              mode ? __builtin_strlen(mode) + 1 : 0);
    if (r) goperl_errno_from_interp(gf);
    goperl_pio_view_again(gf, f);
    return r;
}
static PerlIO *goperl_pio_c_open(goperl_frame_t *gf, int which, uint64_t id, uint64_t gtab,
                                 PerlIO_funcs *tab, PerlIO_list_t *layers, IV n, const char *mode,
                                 int fd, int imode, int perm, PerlIO *old, int narg, SV **args) {
    const size_t inl = narg > 0 && args && !goperl_guest_addr(gf, args) ? (size_t)narg : 0;
    goperl_gword_t *const w = (goperl_gword_t *)malloc((12 + inl) * 8);
    if (!w) goperl_croakf(gf, "Out of memory!");
    goperl_pio_tab_arg(gf, tab, w);
    w[2] = GOPERL_PIO_LIST_ADDR(layers);
    w[3] = (goperl_gword_t)n;
    w[4] = (goperl_gword_t)(int64_t)fd;
    w[5] = (goperl_gword_t)(int64_t)goperl_oflags_to_guest(gf, imode);
    w[6] = (goperl_gword_t)(int64_t)perm;
    w[7] = old ? GOPERL_PIO_H_OUT(gf, old) : 0;
    w[8] = (goperl_gword_t)(int64_t)narg;
    w[9] = inl ? 2 : 1;
    w[10] = inl ? 0 : goperl_guest_addr(gf, args);
    w[11] = mode ? 1 : 0;
    for (size_t i = 0; i < inl; i++) w[12 + i] = GOPERL_TOK(args[i]);
    const uint64_t r = goperl_pio_call(gf, which, id, gtab, w, 12 + inl, mode,
                                       mode ? __builtin_strlen(mode) + 1 : 0);
    free(w);
    if (!r) goperl_errno_from_interp(gf);
    return GOPERL_PIO_H_IN(gf, r);
}
static IV goperl_pio_c_iv(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f) {
    const IV r = (IV)(int64_t)goperl_pio_call_f(gf, which, id, f, 0, 0, 0, 0);
    if (r == -1) goperl_errno_from_interp(gf);
    return r;
}
static SV *goperl_pio_c_getarg(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f, int flags) {
    const goperl_gword_t w[1] = {(goperl_gword_t)(int64_t)flags};
    return GOPERL_SV(goperl_pio_call_f(gf, which, id, f, w, 1, 0, 0));
}
static PerlIO *goperl_pio_c_dup(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f, PerlIO *o,
                                int flags) {
    const goperl_gword_t w[2] = {o ? GOPERL_PIO_H_OUT(gf, o) : 0, (goperl_gword_t)(int64_t)flags};
    return GOPERL_PIO_H_IN(gf, goperl_pio_call_f(gf, which, id, f, w, 2, 0, 0));
}
/* a buffer of native code's own for the interpreter to read into: a block
 * of the interpreter's for the call */
static SSize_t goperl_pio_c_read(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f, void *vbuf,
                                 Size_t count) {
    uint64_t g = goperl_guest_addr(gf, vbuf);
    const int scratch = vbuf && !g;
    if (scratch) {
        g = goperl_xs_opx(gf, GOPERL_OP_MEM_REALLOC, 0, count ? count : 1, 0, 0, 0, 0);
        if (!g) goperl_croakf(gf, "Out of memory!");
    }
    const goperl_gword_t w[2] = {g, (goperl_gword_t)count};
    const SSize_t r = (SSize_t)(int64_t)goperl_pio_call_f(gf, which, id, f, w, 2, 0, 0);
    if (scratch) {
        if (r > 0) memcpy(vbuf, goperl_guest_mem(gf, g), (size_t)r);
        goperl_xs_opx(gf, GOPERL_OP_MEM_FREE, g, 0, 0, 0, 0, 0);
    }
    if (r < 0) goperl_errno_from_interp(gf);
    return r;
}
/* bytes of native code's own go with the call */
static SSize_t goperl_pio_c_write(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f,
                                  const void *vbuf, Size_t count) {
    const uint64_t g = goperl_guest_addr(gf, vbuf);
    const goperl_gword_t w[2] = {g, (goperl_gword_t)count};
    const SSize_t r = (SSize_t)(int64_t)goperl_pio_call_f(gf, which, id, f, w, 2, g ? 0 : vbuf,
                                                          g ? 0 : count);
    if (r < 0) goperl_errno_from_interp(gf);
    return r;
}
static IV goperl_pio_c_seek(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f, Off_t off,
                            int whence) {
    const goperl_gword_t w[2] = {(goperl_gword_t)(int64_t)off, (goperl_gword_t)(int64_t)whence};
    const IV r = (IV)(int64_t)goperl_pio_call_f(gf, which, id, f, w, 2, 0, 0);
    if (r) goperl_errno_from_interp(gf);
    return r;
}
static void goperl_pio_c_setptrcnt(goperl_frame_t *gf, int which, uint64_t id, PerlIO *f,
                                   STDCHAR *ptr, SSize_t cnt) {
    const uint64_t g = goperl_guest_addr(gf, ptr);
    if (ptr && !g) goperl_croakf(gf, "Set_ptrcnt: a pointer outside the interpreter's buffer");
    const goperl_gword_t w[2] = {g, (goperl_gword_t)(int64_t)cnt};
    (void)goperl_pio_call_f(gf, which, id, f, w, 2, 0, 0);
}
#define GOPERL_PIO_C_F(gf, which, id, f) goperl_pio_call_f((gf), (which), (id), (f), 0, 0, 0, 0)

/* The trampolines of a view of an interpreter's table: each runs its slot
 * there, on the table of the handle's top layer, or on the table itself
 * for the slots that take one (Pushed, Open). */
#define GOPERL_PIO_GTAB(tab) (((const goperl_pio_fview_t *)(const void *)(tab))->gtab)
#define GOPERL_PIO_TOP GOPERL_PIO_CALL_TOP
static IV goperl_pio_tr_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {
    return goperl_pio_c_pushed(goperl_cur_frame_v, GOPERL_PIO_CALL_TAB, GOPERL_PIO_SLOT_PUSHED,
                               GOPERL_PIO_GTAB(tab), f, mode, arg, tab);
}
static PerlIO *goperl_pio_tr_open(pTHX_ PerlIO_funcs *tab, PerlIO_list_t *layers, IV n,
                                  const char *mode, int fd, int imode, int perm, PerlIO *old,
                                  int narg, SV **args) {
    return goperl_pio_c_open(goperl_cur_frame_v, GOPERL_PIO_CALL_TAB, GOPERL_PIO_SLOT_OPEN,
                             GOPERL_PIO_GTAB(tab), tab, layers, n, mode, fd, imode, perm, old,
                             narg, args);
}
#define GOPERL_PIO_TR_IV(name, slot)                                            \
    static IV name(pTHX_ PerlIO *f) {                                           \
        return goperl_pio_c_iv(goperl_cur_frame_v, GOPERL_PIO_TOP, (slot), f); \
    }
GOPERL_PIO_TR_IV(goperl_pio_tr_popped, GOPERL_PIO_SLOT_POPPED)
GOPERL_PIO_TR_IV(goperl_pio_tr_binmode, GOPERL_PIO_SLOT_BINMODE)
GOPERL_PIO_TR_IV(goperl_pio_tr_fileno, GOPERL_PIO_SLOT_FILENO)
GOPERL_PIO_TR_IV(goperl_pio_tr_close, GOPERL_PIO_SLOT_CLOSE)
GOPERL_PIO_TR_IV(goperl_pio_tr_flush, GOPERL_PIO_SLOT_FLUSH)
GOPERL_PIO_TR_IV(goperl_pio_tr_fill, GOPERL_PIO_SLOT_FILL)
GOPERL_PIO_TR_IV(goperl_pio_tr_eof, GOPERL_PIO_SLOT_EOF)
GOPERL_PIO_TR_IV(goperl_pio_tr_error, GOPERL_PIO_SLOT_ERROR)
static Off_t goperl_pio_tr_tell(pTHX_ PerlIO *f) {
    return (Off_t)(int64_t)GOPERL_PIO_C_F(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_TELL, f);
}
static void goperl_pio_tr_clearerr(pTHX_ PerlIO *f) {
    (void)GOPERL_PIO_C_F(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_CLEARERR, f);
}
static void goperl_pio_tr_setlinebuf(pTHX_ PerlIO *f) {
    (void)GOPERL_PIO_C_F(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_SETLINEBUF, f);
}
static STDCHAR *goperl_pio_tr_get_base(pTHX_ PerlIO *f) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    return goperl_pio_ptr_in(gf, GOPERL_PIO_C_F(gf, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_GET_BASE, f));
}
static Size_t goperl_pio_tr_get_bufsiz(pTHX_ PerlIO *f) {
    return (Size_t)GOPERL_PIO_C_F(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_GET_BUFSIZ, f);
}
static STDCHAR *goperl_pio_tr_get_ptr(pTHX_ PerlIO *f) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    return goperl_pio_ptr_in(gf, GOPERL_PIO_C_F(gf, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_GET_PTR, f));
}
static SSize_t goperl_pio_tr_get_cnt(pTHX_ PerlIO *f) {
    return (SSize_t)(int64_t)GOPERL_PIO_C_F(goperl_cur_frame_v, GOPERL_PIO_TOP,
                                            GOPERL_PIO_SLOT_GET_CNT, f);
}
static SV *goperl_pio_tr_getarg(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags) {
    (void)param;
    return goperl_pio_c_getarg(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_GETARG, f, flags);
}
static PerlIO *goperl_pio_tr_dup(pTHX_ PerlIO *f, PerlIO *o, CLONE_PARAMS *param, int flags) {
    (void)param;
    return goperl_pio_c_dup(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_DUP, f, o, flags);
}
static SSize_t goperl_pio_tr_read(pTHX_ PerlIO *f, void *vbuf, Size_t count) {
    return goperl_pio_c_read(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_READ, f, vbuf, count);
}
static SSize_t goperl_pio_tr_unread(pTHX_ PerlIO *f, const void *vbuf, Size_t count) {
    return goperl_pio_c_write(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_UNREAD, f, vbuf,
                              count);
}
static SSize_t goperl_pio_tr_write(pTHX_ PerlIO *f, const void *vbuf, Size_t count) {
    return goperl_pio_c_write(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_WRITE, f, vbuf,
                              count);
}
static IV goperl_pio_tr_seek(pTHX_ PerlIO *f, Off_t off, int whence) {
    return goperl_pio_c_seek(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_SEEK, f, off, whence);
}
static void goperl_pio_tr_set_ptrcnt(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    goperl_pio_c_setptrcnt(goperl_cur_frame_v, GOPERL_PIO_TOP, GOPERL_PIO_SLOT_SET_PTRCNT, f, ptr,
                           cnt);
}
static void (*const goperl_pio_trampolines[GOPERL_PIO_SLOTS])(void) = {
    (void (*)(void))goperl_pio_tr_pushed,     (void (*)(void))goperl_pio_tr_popped,
    (void (*)(void))goperl_pio_tr_open,       (void (*)(void))goperl_pio_tr_binmode,
    (void (*)(void))goperl_pio_tr_getarg,     (void (*)(void))goperl_pio_tr_fileno,
    (void (*)(void))goperl_pio_tr_dup,        (void (*)(void))goperl_pio_tr_read,
    (void (*)(void))goperl_pio_tr_unread,     (void (*)(void))goperl_pio_tr_write,
    (void (*)(void))goperl_pio_tr_seek,       (void (*)(void))goperl_pio_tr_tell,
    (void (*)(void))goperl_pio_tr_close,      (void (*)(void))goperl_pio_tr_flush,
    (void (*)(void))goperl_pio_tr_fill,       (void (*)(void))goperl_pio_tr_eof,
    (void (*)(void))goperl_pio_tr_error,      (void (*)(void))goperl_pio_tr_clearerr,
    (void (*)(void))goperl_pio_tr_setlinebuf, (void (*)(void))goperl_pio_tr_get_base,
    (void (*)(void))goperl_pio_tr_get_bufsiz, (void (*)(void))goperl_pio_tr_get_ptr,
    (void (*)(void))goperl_pio_tr_get_cnt,    (void (*)(void))goperl_pio_tr_set_ptrcnt,
};
/* A table of the interpreter's at gtab as native code sees it: the
 * module's own table for a class of native code's, or a view (made once). */
static PerlIO_funcs *goperl_pio_tab_of(goperl_frame_t *f, uint64_t gtab) {
    if (!gtab) return 0;
    goperl_pio_reg_t *const r = goperl_pio_reg(f);
    for (goperl_pio_fview_t *v = r->fviews; v; v = v->next)
        if (v->gtab == gtab) return &v->tab;
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_PERLIO, gtab, 0, (uint64_t)(GOPERL_PIO_FUNCS_INFO), 0, 0));
    if (!w) return 0;
    if (w[3]) return (PerlIO_funcs *)(uintptr_t)w[3];
    goperl_pio_fview_t *const v = (goperl_pio_fview_t *)calloc(1, sizeof *v);
    if (!v) goperl_croakf(f, "Out of memory!");
    v->gtab = gtab;
    v->tab.fsize = sizeof(PerlIO_funcs);
    v->tab.name = w[0] ? (const char *)goperl_guest_mem(f, w[0]) : "";
    v->tab.size = (Size_t)w[1];
    v->tab.kind = (U32)w[2];
    for (int i = 0; i < GOPERL_PIO_SLOTS; i++)
        if (w[4] & ((goperl_gword_t)1 << i)) GOPERL_PIO_SLOT_FN(&v->tab, i) = goperl_pio_trampolines[i];
    v->next = r->fviews;
    r->fviews = v;
    return &v->tab;
}

/* The interpreter's list of layers at g as a view (NULL for none). */
static PerlIO_list_t *goperl_pio_list(goperl_frame_t *f, uint64_t g) {
    if (!g) return 0;
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_PERLIO, g, 0, (uint64_t)GOPERL_PIO_LIST_INFO, 0, 0));
    if (!w) return 0;
    const IV cur = (IV)w[1];
    /* a copy: the tables below make operations, which reuse the answer */
    goperl_gword_t *const c = (goperl_gword_t *)goperl_arena_alloc((size_t)(3 + 2 * cur) * 8);
    PerlIO_list_t *const l = (PerlIO_list_t *)goperl_arena_alloc(
        sizeof(PerlIO_list_t) + (size_t)(cur > 0 ? cur : 1) * sizeof(PerlIO_pair_t));
    if (!c || !l) goperl_croakf(f, "Out of memory!");
    memcpy(c, w, (size_t)(3 + 2 * cur) * 8);
    l->refcnt = (IV)c[0];
    l->cur = cur;
    l->len = (IV)c[2];
    l->array = (PerlIO_pair_t *)(l + 1);
    l->gaddr = g;
    for (IV i = 0; i < cur; i++) {
        l->array[i].funcs = goperl_pio_tab_of(f, c[3 + 2 * i]);
        l->array[i].arg = GOPERL_SV(c[4 + 2 * i]);
    }
    return l;
}
#define PL_known_layers \
    goperl_pio_list(_gof, goperl_do_op3(_gof, GOPERL_OP_PERLIO, 0, 0, (uint64_t)GOPERL_PIO_KNOWN_LAYERS, 0, 0))
#define PL_def_layerlist \
    goperl_pio_list(_gof, goperl_do_op3(_gof, GOPERL_OP_PERLIO, 0, 0, (uint64_t)GOPERL_PIO_DEF_LAYERLIST, 0, 0))

/* ---- perl's own layer functions -------------------------------------------
 * Each runs perl's function of the name in the interpreter, on the
 * interpreter's layer (whose struct goes there and comes back, see
 * goperl_pio_call), and answers what perl's answers. One of perl's buffer
 * layer called on the struct of a layer of native code's makes it a
 * buffered one: the module takes it for a PerlIOBuf. These are also what
 * a table names for "perl's function". */
static void goperl_pio_stock_buffered(goperl_frame_t *gf, PerlIO *f) {
    if (!f || !*f || goperl_pio_fview_of(gf, (*f)->tab) || !((goperl_shared_t *)gf->inst)->pio_layers)
        return;
    goperl_pio_reg_t *const r = goperl_pio_reg(gf);
    for (int32_t i = 0; i < r->nacts; i++)
        if (GOPERL_IOL_STRUCT(r->acts[i]) == *f && r->acts[i]->tab->size >= sizeof(PerlIOBuf))
            r->acts[i]->buffered = 1;
}
#define GOPERL_PIO_FN GOPERL_C __attribute__((weak, visibility("hidden")))
#define GOPERL_PIO_PRE(buffered)                   \
    goperl_frame_t *const gf = goperl_cur_frame_v; \
    if (buffered) goperl_pio_stock_buffered(gf, f)
#define GOPERL_PIO_DEF_PUSHED(id, name, buffered)                                              \
    GOPERL_PIO_FN IV name(pTHX_ PerlIO *f, const char *mode, SV *arg, PerlIO_funcs *tab) {   \
        GOPERL_PIO_PRE(buffered);                                                              \
        return goperl_pio_c_pushed(gf, GOPERL_PIO_CALL_STOCK, id, 0, f, mode, arg, tab);       \
    }
#define GOPERL_PIO_DEF_F_IV(id, name, buffered)                     \
    GOPERL_PIO_FN IV name(pTHX_ PerlIO *f) {                         \
        GOPERL_PIO_PRE(buffered);                                    \
        return goperl_pio_c_iv(gf, GOPERL_PIO_CALL_STOCK, id, f);   \
    }
#define GOPERL_PIO_DEF_F_OFF(id, name, buffered)                                         \
    GOPERL_PIO_FN Off_t name(pTHX_ PerlIO *f) {                                           \
        GOPERL_PIO_PRE(buffered);                                                         \
        return (Off_t)(int64_t)GOPERL_PIO_C_F(gf, GOPERL_PIO_CALL_STOCK, id, f);         \
    }
#define GOPERL_PIO_DEF_F_VOID(id, name, buffered)                \
    GOPERL_PIO_FN void name(pTHX_ PerlIO *f) {                    \
        GOPERL_PIO_PRE(buffered);                                 \
        (void)GOPERL_PIO_C_F(gf, GOPERL_PIO_CALL_STOCK, id, f);   \
    }
#define GOPERL_PIO_DEF_F_PTR(id, name, buffered)                                         \
    GOPERL_PIO_FN STDCHAR *name(pTHX_ PerlIO *f) {                                        \
        GOPERL_PIO_PRE(buffered);                                                         \
        return goperl_pio_ptr_in(gf, GOPERL_PIO_C_F(gf, GOPERL_PIO_CALL_STOCK, id, f));  \
    }
#define GOPERL_PIO_DEF_F_SIZE(id, name, buffered)                             \
    GOPERL_PIO_FN Size_t name(pTHX_ PerlIO *f) {                               \
        GOPERL_PIO_PRE(buffered);                                              \
        return (Size_t)GOPERL_PIO_C_F(gf, GOPERL_PIO_CALL_STOCK, id, f);      \
    }
#define GOPERL_PIO_DEF_F_SSIZE(id, name, buffered)                                 \
    GOPERL_PIO_FN SSize_t name(pTHX_ PerlIO *f) {                                   \
        GOPERL_PIO_PRE(buffered);                                                   \
        return (SSize_t)(int64_t)GOPERL_PIO_C_F(gf, GOPERL_PIO_CALL_STOCK, id, f); \
    }
#define GOPERL_PIO_DEF_OPEN(id, name, buffered)                                                \
    GOPERL_PIO_FN PerlIO *name(pTHX_ PerlIO_funcs *tab, PerlIO_list_t *layers, IV n,          \
                               const char *mode, int fd, int imode, int perm, PerlIO *old,    \
                               int narg, SV **args) {                                         \
        return goperl_pio_c_open(goperl_cur_frame_v, GOPERL_PIO_CALL_STOCK, id, 0, tab, layers, \
                                 n, mode, fd, imode, perm, old, narg, args);                  \
    }
#define GOPERL_PIO_DEF_DUP(id, name, buffered)                                                 \
    GOPERL_PIO_FN PerlIO *name(pTHX_ PerlIO *f, PerlIO *o, CLONE_PARAMS *param, int flags) {  \
        GOPERL_PIO_PRE(buffered);                                                              \
        (void)param;                                                                           \
        return goperl_pio_c_dup(gf, GOPERL_PIO_CALL_STOCK, id, f, o, flags);                   \
    }
#define GOPERL_PIO_DEF_READ(id, name, buffered)                                  \
    GOPERL_PIO_FN SSize_t name(pTHX_ PerlIO *f, void *vbuf, Size_t count) {      \
        GOPERL_PIO_PRE(buffered);                                                \
        return goperl_pio_c_read(gf, GOPERL_PIO_CALL_STOCK, id, f, vbuf, count); \
    }
#define GOPERL_PIO_DEF_WRITE(id, name, buffered)                                  \
    GOPERL_PIO_FN SSize_t name(pTHX_ PerlIO *f, const void *vbuf, Size_t count) { \
        GOPERL_PIO_PRE(buffered);                                                 \
        return goperl_pio_c_write(gf, GOPERL_PIO_CALL_STOCK, id, f, vbuf, count); \
    }
#define GOPERL_PIO_DEF_SEEK(id, name, buffered)                                     \
    GOPERL_PIO_FN IV name(pTHX_ PerlIO *f, Off_t off, int whence) {                 \
        GOPERL_PIO_PRE(buffered);                                                   \
        return goperl_pio_c_seek(gf, GOPERL_PIO_CALL_STOCK, id, f, off, whence);   \
    }
#define GOPERL_PIO_DEF_SETPTRCNT(id, name, buffered)                             \
    GOPERL_PIO_FN void name(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt) {         \
        GOPERL_PIO_PRE(buffered);                                                 \
        goperl_pio_c_setptrcnt(gf, GOPERL_PIO_CALL_STOCK, id, f, ptr, cnt);       \
    }
#define GOPERL_PIO_DEF(id, name, shape, buffered) GOPERL_PIO_DEF_##shape(id, name, buffered)
GOPERL_PIO_STOCK(GOPERL_PIO_DEF)
#undef GOPERL_PIO_DEF

/* The number of perl's function a slot names, | GOPERL_PIO_STOCK_BUFFER
 * for one that takes the struct for a PerlIOBuf; 0 for a function of the
 * module's own. */
static int goperl_pio_stock_id(void (*fn)(void)) {
    if (!fn) return 0;
#define GOPERL_PIO_ID_OF(id, name, shape, buffered) \
    if (fn == (void (*)(void))name) return (id) | ((buffered) ? GOPERL_PIO_STOCK_BUFFER : 0);
    GOPERL_PIO_STOCK(GOPERL_PIO_ID_OF)
#undef GOPERL_PIO_ID_OF
    return 0;
}

/* ---- handle functions on layered handles ----------------------------------
 * A handle here is a view, or a handle of the interpreter's that perl.h
 * gave: either is the interpreter's to work on. After a function that may
 * change the layers under a view's slot, the slot shows the new ones. */
static SSize_t PerlIO_read(PerlIO *f, void *vbuf, Size_t count) {
    return goperl_pio_read(goperl_cur_frame_v, f, vbuf, count);
}
static IV PerlIO_fill(PerlIO *f) { return goperl_pio_status(goperl_cur_frame_v, GOPERL_PIO_FILL, f); }
static IV PerlIO_eof(PerlIO *f) { return (IV)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_EOF, f, 0, 0, 0); }
static IV PerlIO_error(PerlIO *f) {
    return (IV)(int64_t)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_ERROR, f, 0, 0, 0);
}
static IV PerlIO_flush(PerlIO *f) { return goperl_pio_flush(goperl_cur_frame_v, f); }
static STDCHAR *PerlIO_get_ptr(PerlIO *f) {
    return goperl_pio_bufptr(goperl_cur_frame_v, GOPERL_PIO_GET_PTR, f);
}
static SSize_t PerlIO_get_cnt(PerlIO *f) {
    return (SSize_t)(int64_t)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_GET_CNT, f, 0, 0, 0);
}
static void PerlIO_set_ptrcnt(PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    goperl_pio_set_ptrcnt(goperl_cur_frame_v, f, ptr, cnt);
}
/* the layer may get its buffer now: its view (or struct) shows it */
static STDCHAR *PerlIO_get_base(PerlIO *f) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    STDCHAR *const base = goperl_pio_bufptr(gf, GOPERL_PIO_GET_BASE, f);
    if (goperl_pio_view_of(f) && *f && goperl_pio_fview_of(gf, (*f)->tab) &&
        ((*f)->tab->kind & PERLIO_K_BUFFERED)) {
        const goperl_gword_t *w = goperl_pio_layer_words(gf, goperl_pio_tok(gf, f));
        if (w && w[0]) goperl_pio_view_buf(gf, PerlIOSelf(f, PerlIOBuf), w + 1);
    }
    return base;
}
static int PerlIO_fast_gets(PerlIO *f) {
    return (int)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_FAST_GETS, f, 0, 0, 0);
}
/* errno kept in the layer (its err field, which the struct of a layer of
 * native code's keeps in step with the interpreter's layer), and given
 * back, as perl's do */
static void goperl_pio_save_errno(PerlIO *f) {
    if (PerlIOValid(f)) PerlIOBase(f)->err = errno;
}
static void goperl_pio_restore_errno(PerlIO *f) {
    if (PerlIOValid(f)) errno = PerlIOBase(f)->err;
}
#define PerlIO_save_errno(f) goperl_pio_save_errno(f)
#define PerlIO_restore_errno(f) goperl_pio_restore_errno(f)
#define Perl_PerlIO_save_errno(f) goperl_pio_save_errno(f)
#define Perl_PerlIO_restore_errno(f) goperl_pio_restore_errno(f)
/* the mode of the handle as perl writes it into buf */
static char *PerlIO_modestr(PerlIO *f, char *buf) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    const char *const m =
        (const char *)goperl_guest_mem(gf, goperl_pio(gf, GOPERL_PIO_MODESTR, f, 0, 0, 0));
    size_t i = 0;
    if (m)
        for (; m[i] && i < 7; i++) buf[i] = m[i];
    buf[i] = 0;
    return buf;
}
static int goperl_pio_apply_layers_view(goperl_frame_t *gf, PerlIO *h, const char *mode,
                                        const char *names) {
    const int r = (PerlIO_apply_layers)(h, mode, names);
    goperl_pio_view_again(gf, h);
    return r;
}
/* variadic: an argument may expand to more than one, as with perl's */
#define PerlIO_apply_layers(...) goperl_pio_apply_layers_view(_gof, __VA_ARGS__)
static int goperl_pio_binmode_view(goperl_frame_t *gf, PerlIO *h, int iotype, int omode,
                                   const char *names) {
    const int r = (PerlIO_binmode)(h, iotype, omode, names);
    goperl_pio_view_again(gf, h);
    return r;
}
#define PerlIO_binmode(...) goperl_pio_binmode_view(_gof, __VA_ARGS__)

/* ---- the stack of layers ------------------------------------------------- */
/* Whether the table is a class native code defined (the classes are the
 * instance's: those of every module). */
static int goperl_pio_class_defined(goperl_frame_t *f, const PerlIO_funcs *tab) {
    if (!((goperl_shared_t *)f->inst)->pio_layers) return 0;
    for (const goperl_pio_class_t *c = goperl_pio_reg(f)->classes; c; c = c->next)
        if (c->tab == tab) return 1;
    return 0;
}
/* perl reads no mode of a push but to give it to the layer's Pushed (and
 * PerlIO_debug): of a class native code defined whose Pushed is native
 * code's, that Pushed gets the mode given here, as perl gives it the
 * pointer it was given, which need not be a string (PerlIO::Util's :tee
 * passes a struct of its own in its Dup), and the interpreter an empty
 * one. */
static PerlIO *goperl_pio_push(goperl_frame_t *gf, PerlIO *f, PerlIO_funcs *tab, const char *given,
                               SV *arg) {
    const int own = given && tab && goperl_pio_class_defined(gf, tab) && tab->Pushed &&
                    !goperl_pio_stock_id((void (*)(void))tab->Pushed);
    const char *const mode = own ? "" : given;
    goperl_pio_reg_t *const r = own ? goperl_pio_reg(gf) : 0;
    const PerlIO_funcs *const outer_tab = r ? r->push_tab : 0;
    const char *const outer_mode = r ? r->push_mode : 0;
    const size_t ml = mode ? __builtin_strlen(mode) + 1 : 0;
    char *const req = (char *)malloc(32 + ml + 1);
    if (!req) goperl_croakf(gf, "Out of memory!");
    goperl_gword_t w[4];
    goperl_pio_tab_arg(gf, tab, w);
    w[2] = GOPERL_TOK(arg);
    w[3] = mode ? 1 : 0;
    memcpy(req, w, sizeof w);
    if (ml) memcpy(req + 32, mode, ml);
    const uint64_t tok = goperl_pio_tok(gf, f);
    if (r) {
        r->push_tab = tab;
        r->push_mode = given;
    }
    const uint64_t h = goperl_pio(gf, GOPERL_PIO_PUSH, f, 0, req, 32 + ml);
    if (r) {
        r->push_tab = outer_tab;
        r->push_mode = outer_mode;
    }
    free(req);
    goperl_pio_view_again(gf, f);
    if (!h) {
        goperl_errno_from_interp(gf);
        return 0;
    }
    return h == tok ? f : goperl_pio_handle(gf, h);
}
static PerlIO *PerlIO_push(pTHX_ PerlIO *f, PerlIO_funcs *tab, const char *mode, SV *arg) {
    return goperl_pio_push(goperl_cur_frame_v, f, tab, mode, arg);
}
static void goperl_pio_pop(goperl_frame_t *gf, PerlIO *f) {
    (void)goperl_pio(gf, GOPERL_PIO_POP, f, 0, 0, 0);
    goperl_pio_view_again(gf, f);
}
static void PerlIO_pop(pTHX_ PerlIO *f) { goperl_pio_pop(goperl_cur_frame_v, f); }
static PerlIO_funcs *goperl_pio_find_layer(goperl_frame_t *gf, const char *name, STRLEN len,
                                           int load) {
    if ((SSize_t)len < 0) len = __builtin_strlen(name);
    return goperl_pio_tab_of(gf, goperl_pio(gf, GOPERL_PIO_FIND_LAYER, 0, (uint64_t)load, name, len));
}
static PerlIO_funcs *PerlIO_find_layer(pTHX_ const char *name, STRLEN len, int load) {
    return goperl_pio_find_layer(goperl_cur_frame_v, name, len, load);
}
static PerlIO_funcs *goperl_pio_layer_fetch(goperl_frame_t *gf, PerlIO_list_t *list, IV n,
                                            PerlIO_funcs *def) {
    goperl_gword_t w[2];
    goperl_pio_tab_arg(gf, def, w);
    return goperl_pio_tab_of(gf, goperl_do_op3(gf, GOPERL_OP_PERLIO, GOPERL_PIO_LIST_ADDR(list),
                                               (uint64_t)n, (uint64_t)(GOPERL_PIO_LAYER_FETCH),
                                               (const char *)w, sizeof w));
}
static PerlIO_funcs *PerlIO_layer_fetch(pTHX_ PerlIO_list_t *list, IV n, PerlIO_funcs *def) {
    return goperl_pio_layer_fetch(goperl_cur_frame_v, list, n, def);
}
static SV *PerlIO_arg_fetch(PerlIO_list_t *list, IV n) {
    return GOPERL_SV(goperl_do_op3(goperl_cur_frame_v, GOPERL_OP_PERLIO, GOPERL_PIO_LIST_ADDR(list),
                                   (uint64_t)n, (uint64_t)(GOPERL_PIO_ARG_FETCH), 0, 0));
}
/* PerlIO_sv_dup: the argument a Getarg slot answers, as perl's of an
 * interpreter without threads makes it (no interpreter is cloned here) */
static SV *PerlIO_sv_dup(pTHX_ SV *arg, CLONE_PARAMS *param) {
    (void)param;
    return arg ? newSVsv(arg) : 0;
}
/* PerlIO_allocate: a new handle of the interpreter's table */
static PerlIO *PerlIO_allocate(pTHX) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    return goperl_pio_handle(gf, goperl_pio(gf, GOPERL_PIO_ALLOCATE, 0, 0, 0, 0));
}
/* PerlIO_apply_layera: the layers n to max of the list on the handle */
static int PerlIO_apply_layera(pTHX_ PerlIO *f, const char *mode, PerlIO_list_t *layers, IV n,
                               IV max) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    const size_t ml = mode ? __builtin_strlen(mode) + 1 : 0;
    char *const req = (char *)malloc(32 + ml + 1);
    if (!req) goperl_croakf(gf, "Out of memory!");
    const goperl_gword_t w[4] = {GOPERL_PIO_LIST_ADDR(layers), (goperl_gword_t)n, (goperl_gword_t)max,
                                 mode ? 1u : 0u};
    memcpy(req, w, sizeof w);
    if (ml) memcpy(req + 32, mode, ml);
    const int r = (int)(int64_t)goperl_pio(gf, GOPERL_PIO_APPLY_LAYERA, f, 0, req, 32 + ml);
    free(req);
    goperl_pio_view_again(gf, f);
    if (r) goperl_errno_from_interp(gf);
    return r;
}
/* PerlIO_fdupopen: a handle on a duplicate of the handle's descriptor, with
 * its layers (no interpreter is cloned here: param is not used) */
static PerlIO *PerlIO_fdupopen(pTHX_ PerlIO *f, CLONE_PARAMS *param, int flags) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    (void)param;
    const uint64_t r = goperl_pio(gf, GOPERL_PIO_FDUPOPEN, f, (uint64_t)(int64_t)flags, 0, 0);
    if (!r) goperl_errno_from_interp(gf);
    return goperl_pio_handle(gf, r);
}
/* PerlIOUnix_oflags: the open flags of a mode, as this system's */
static int PerlIOUnix_oflags(const char *mode) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    const int64_t g = (int64_t)goperl_do_op3(gf, GOPERL_OP_PERLIO, 0, 0, (uint64_t)GOPERL_PIO_UNIX_OFLAGS,
                                             mode ? mode : "", mode ? __builtin_strlen(mode) : 0);
    if (g == -1) {
        errno = EINVAL;
        return -1;
    }
    return goperl_oflags_to_host(gf, g);
}
#define PerlIOArg PerlIO_arg_fetch(layers, n)
/* the tables of perl's own layers, as views of the interpreter's */
enum {
    GOPERL_PIO_STD_UNIX = 0,
    GOPERL_PIO_STD_PERLIO,
    GOPERL_PIO_STD_STDIO,
    GOPERL_PIO_STD_CRLF,
    GOPERL_PIO_STD_UTF8,
    GOPERL_PIO_STD_BYTE,
    GOPERL_PIO_STD_RAW,
    GOPERL_PIO_STD_PENDING
};
static PerlIO_funcs *goperl_pio_std(goperl_frame_t *f, int which) {
    PerlIO_funcs *const t = goperl_pio_tab_of(
        f, goperl_do_op3(f, GOPERL_OP_PERLIO, 0, (uint64_t)which, (uint64_t)(GOPERL_PIO_STD_FUNCS), 0, 0));
    if (!t) goperl_croakf(f, "the interpreter has no such PerlIO layer");
    return t;
}
#define PerlIO_unix (*goperl_pio_std(_gof, GOPERL_PIO_STD_UNIX))
#define PerlIO_perlio (*goperl_pio_std(_gof, GOPERL_PIO_STD_PERLIO))
#define PerlIO_stdio (*goperl_pio_std(_gof, GOPERL_PIO_STD_STDIO))
#define PerlIO_crlf (*goperl_pio_std(_gof, GOPERL_PIO_STD_CRLF))
#define PerlIO_utf8 (*goperl_pio_std(_gof, GOPERL_PIO_STD_UTF8))
#define PerlIO_byte (*goperl_pio_std(_gof, GOPERL_PIO_STD_BYTE))
#define PerlIO_raw (*goperl_pio_std(_gof, GOPERL_PIO_STD_RAW))
#define PerlIO_pending (*goperl_pio_std(_gof, GOPERL_PIO_STD_PENDING))

/* ---- registration ---------------------------------------------------------
 * PERLIO_DEFINE: [funcs id][the module's table][size][kind], a pair [0 NULL
 * | 1 perl's function | 2 the module's][the number of perl's] for each slot,
 * then the name. */
static void PerlIO_define_layer(pTHX_ PerlIO_funcs *tab) {
    goperl_frame_t *const gf = goperl_cur_frame_v;
    const size_t nl = __builtin_strlen(tab->name) + 1;
    goperl_gword_t w[4 + 2 * GOPERL_PIO_SLOTS];
    if (tab->fsize != sizeof(PerlIO_funcs))
        goperl_croakf(gf, "%s (%d) does not match %s (%d)", "PerlIO_funcs size", (int)tab->fsize,
                      "size expected by this perl", (int)sizeof(PerlIO_funcs));
    if (tab->size && tab->size < sizeof(PerlIOl))
        goperl_croakf(gf, "%s (%d) smaller than %s (%d)", "PerlIO layer size", (int)tab->size,
                      "size expected by this perl", (int)sizeof(PerlIOl));
    {
        goperl_pio_reg_t *const r = goperl_pio_reg(gf);
        goperl_pio_class_t *const c = (goperl_pio_class_t *)malloc(sizeof *c);
        if (!c) goperl_croakf(gf, "Out of memory!");
        c->tab = tab;
        c->block_forget = goperl_block_forget;
        c->next = r->classes;
        r->classes = c;
    }
    w[0] = goperl_api_perlio_def(gf, tab->name, (void *)tab);
    w[1] = (goperl_gword_t)(uintptr_t)tab;
    w[2] = (goperl_gword_t)tab->size;
    w[3] = (goperl_gword_t)tab->kind;
    for (int i = 0; i < GOPERL_PIO_SLOTS; i++) {
        void (*const fn)(void) = GOPERL_PIO_SLOT_FN(tab, i);
        const int id = goperl_pio_stock_id(fn) & ~GOPERL_PIO_STOCK_BUFFER;
        w[4 + 2 * i] = !fn ? 0 : id ? 1 : 2;
        w[5 + 2 * i] = (goperl_gword_t)id;
    }
    char *const req = (char *)malloc(sizeof w + nl);
    if (!req) goperl_croakf(gf, "Out of memory!");
    memcpy(req, w, sizeof w);
    memcpy(req + sizeof w, tab->name, nl);
    (void)goperl_do_op3(gf, GOPERL_OP_PERLIO_DEFINE, 0, sizeof w + nl, 0, req, sizeof w + nl);
    free(req);
}

/* ---- the slots of the module's layers (the loader calls this on method
 * -6) ---------------------------------------------------------------------
 * req: [slot][funcs id][f][layer][errno], the slot's words and its string (see
 * goperlio_hook in perl.cc); resp: [1][the slot's result]. A slot that
 * dies ends the frame as an XSUB that dies does, and the loader answers
 * for it. The call is an activation of native code as any other: the
 * structs here go back to the interpreter when it returns. */
static uint64_t goperl_pio_word(const unsigned char *req, uint32_t reqlen, uint32_t i) {
    goperl_gword_t w = 0;
    if ((size_t)reqlen >= 8 * ((size_t)i + 1)) memcpy(&w, req + 8 * (size_t)i, 8);
    return w;
}
GOPERL_C __attribute__((weak, visibility("default"), used)) int32_t
__goperl_perlio_invoke(goperl_frame_t *f, void *funcs_v, const unsigned char *req, uint32_t reqlen,
                       unsigned char *resp, uint32_t respcap) {
    PerlIO_funcs *const tab = (PerlIO_funcs *)funcs_v;
    jmp_buf jb;
    int32_t mark = -1;
    if (respcap < 1 + 8) return 0;
    f->jb = (void *)&jb;
    f->prev_frame = (void *)goperl_frame_enter(f);
    goperl_xs_depth_v++;
    goperl_opreg_enter();
    f->native_call = 1;
    f->hostsave_base = goperl_hostsave_n;
    if (GOPERL_SETJMP(jb)) {
        /* the frame says how the slot ended (a message, a thrown value, or
         * a jump of the interpreter's that goes on), which the loader
         * answers */
        goperl_hostsave_unwind_to(f, f->hostsave_base);
        if (mark >= 0) {
            goperl_iol_release_to(f, mark);
            goperl_pio_sync(f, 1);
        }
        goperl_pp_stack_give_back(f);
        goperl_xs_leave(f);
        return 0;
    }
    /* the slot works on the interpreter's stacks, as perl's would (a
     * layer's Fill that calls a sub, PerlIO::code's, pushes its arguments
     * there) */
    goperl_pp_stack_take(f);
    if (reqlen < 40) goperl_croakf(f, "perlio hook: short request");
    const uint64_t slot = goperl_pio_word(req, reqlen, 0);
    const uint64_t ftok = goperl_pio_word(req, reqlen, 2);
    const uint64_t ltok = goperl_pio_word(req, reqlen, 3);
#define A(i) goperl_pio_word(req, reqlen, 5 + (i))
    /* the slot's string follows its words */
#define STR(nw) ((reqlen > 8 * (5 + (nw))) ? (const char *)req + 8 * (5 + (nw)) : "")
    /* the interpreter's errno, as an XSUB starts with it (the loader put
     * it in the frame) */
    errno = goperl_errno_entry(f);
    mark = goperl_pio_reg(f)->nacts;
    goperl_pio_sync(f, 0); /* what the interpreter did to the structs here */
    goperl_iol_t *s = 0;
    if (ltok) {
        s = goperl_iol_get(f, ltok, tab);
        goperl_iol_activate(f, s);
    }
    PerlIO *const hf = ftok ? goperl_pio_view(f, ftok) : 0;
    goperl_gword_t ret = 0;
    switch (slot) {
    case GOPERL_PIO_SLOT_PUSHED: {
        /* the mode of a push native code makes (goperl_pio_push) */
        goperl_pio_reg_t *const r = goperl_pio_reg(f);
        const char *mode = A(1) ? STR(2) : 0;
        if (mode && r->push_tab == tab) {
            mode = r->push_mode;
            r->push_tab = 0;
        }
        ret = (goperl_gword_t)(int64_t)tab->Pushed(aTHX_ hf, mode, GOPERL_SV(A(0)), tab);
        break;
    }
    case GOPERL_PIO_SLOT_POPPED:
        ret = (goperl_gword_t)(int64_t)(tab->Popped ? tab->Popped(aTHX_ hf) : 0);
        goperl_iol_kill(f, ltok);
        break;
    case GOPERL_PIO_HOOK_DROP: goperl_iol_kill(f, ltok); break;
    case GOPERL_PIO_SLOT_OPEN: {
        PerlIO *const old = A(5) ? goperl_pio_view(f, A(5)) : 0;
        PerlIO *const h = tab->Open(aTHX_ tab, goperl_pio_list(f, A(0)), (IV)A(1),
                                    A(8) ? STR(9) : 0, (int)(int64_t)A(2),
                                    goperl_oflags_to_host(f, (int64_t)A(3)), (int)(int64_t)A(4), old,
                                    (int)(int64_t)A(6),
                                    A(7) ? (SV **)goperl_guest_mem(f, A(7)) : 0);
        ret = h ? goperl_pio_tok(f, h) : 0;
        break;
    }
    case GOPERL_PIO_SLOT_BINMODE: ret = (goperl_gword_t)(int64_t)tab->Binmode(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_GETARG:
        ret = GOPERL_TOK(tab->Getarg(aTHX_ hf, 0, (int)(int64_t)A(0)));
        break;
    case GOPERL_PIO_SLOT_FILENO: ret = (goperl_gword_t)(int64_t)tab->Fileno(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_DUP: {
        PerlIO *const h =
            tab->Dup(aTHX_ hf, A(0) ? goperl_pio_view(f, A(0)) : 0, 0, (int)(int64_t)A(1));
        ret = h ? goperl_pio_tok(f, h) : 0;
        break;
    }
    case GOPERL_PIO_SLOT_READ:
        ret = (goperl_gword_t)(int64_t)tab->Read(aTHX_ hf, goperl_pio_ptr_in(f, A(0)), (Size_t)A(1));
        break;
    case GOPERL_PIO_SLOT_UNREAD:
        ret = (goperl_gword_t)(int64_t)tab->Unread(aTHX_ hf, goperl_pio_ptr_in(f, A(0)), (Size_t)A(1));
        break;
    case GOPERL_PIO_SLOT_WRITE:
        ret = (goperl_gword_t)(int64_t)tab->Write(aTHX_ hf, goperl_pio_ptr_in(f, A(0)), (Size_t)A(1));
        break;
    case GOPERL_PIO_SLOT_SEEK:
        ret = (goperl_gword_t)(int64_t)tab->Seek(aTHX_ hf, (Off_t)(int64_t)A(0), (int)(int64_t)A(1));
        break;
    case GOPERL_PIO_SLOT_TELL: ret = (goperl_gword_t)(int64_t)tab->Tell(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_CLOSE: ret = (goperl_gword_t)(int64_t)tab->Close(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_FLUSH: ret = (goperl_gword_t)(int64_t)tab->Flush(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_FILL: ret = (goperl_gword_t)(int64_t)tab->Fill(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_EOF: ret = (goperl_gword_t)(int64_t)tab->Eof(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_ERROR: ret = (goperl_gword_t)(int64_t)tab->Error(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_CLEARERR: tab->Clearerr(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_SETLINEBUF: tab->Setlinebuf(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_GET_BASE:
    case GOPERL_PIO_SLOT_GET_PTR: {
        STDCHAR *const p =
            slot == GOPERL_PIO_SLOT_GET_BASE ? tab->Get_base(aTHX_ hf) : tab->Get_ptr(aTHX_ hf);
        if (s && s->buffered) {
            /* a buffer the layer allocated just now moves into the
             * interpreter's memory first: the answer points into the
             * moved one */
            PerlIOBuf *const b = (PerlIOBuf *)GOPERL_IOL_STRUCT(s);
            STDCHAR *const from = b->buf;
            goperl_iol_place_buf(f, s);
            ret = goperl_iol_ptr_out(f, s, p && from && b->buf != from && p >= from ? b->buf + (p - from) : p);
        } else {
            ret = s && p ? goperl_iol_inner_out(f, s, p) : 0;
            if (!ret) ret = goperl_guest_addr(f, p);
            if (p && !ret)
                goperl_croakf(f, "PerlIO layer '%s' answered a buffer outside the interpreter's "
                                 "memory",
                              tab->name);
        }
        break;
    }
    case GOPERL_PIO_SLOT_GET_BUFSIZ: ret = (goperl_gword_t)tab->Get_bufsiz(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_GET_CNT: ret = (goperl_gword_t)(int64_t)tab->Get_cnt(aTHX_ hf); break;
    case GOPERL_PIO_SLOT_SET_PTRCNT:
        tab->Set_ptrcnt(aTHX_ hf, goperl_iol_inner_in(f, s, A(0)), (SSize_t)(int64_t)A(1));
        break;
    default: goperl_croakf(f, "perlio hook: unknown slot %d", (int)slot);
    }
#undef A
#undef STR
    goperl_iol_release_to(f, mark);
    goperl_pio_sync(f, 1); /* the interpreter goes on */
    mark = -1;
    resp[0] = 1;
    memcpy(resp + 1, &ret, 8);
    /* what is pending is written back while the frame has the stacks */
    goperl_flush(f);
    goperl_pp_stack_give_back(f);
    goperl_xs_leave(f);
    return 9;
}

#endif /* GOPERL_XS_SDK_PERLIOL_H */
