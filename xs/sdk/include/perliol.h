/* perliol.h — the layer-author view of PerlIO for the goperl XS SDK.
 *
 * A dist that includes this header defines a PerlIO_funcs table and
 * registers it with PerlIO_define_layer. The GUEST runs a PerlIOBuf-derived
 * proxy layer under the dist's layer name: every stock slot is the real
 * :perlio behavior executed natively in the guest, and only the slots the
 * dist actually customized (compared against the stock addresses below —
 * currently Pushed and Fill, the PerlIO::utf8_strict shape) round-trip to
 * the host. The buffer lives in guest memory, so fast_gets/readline stay
 * native; the host touches it through guest_mem during Fill.
 *
 * The hook runs against a HOST shadow of the layer struct (funcs->size
 * bytes, so PerlIOSelf() fields persist per instance) whose ->next is a
 * tail node forwarding PerlIO_read/eof/... to the guest layer BELOW the
 * proxy. A croak inside a hook propagates to the guest as a layer error.
 *
 * Layers customizing slots outside the supported set are rejected LOUDLY
 * at PerlIO_define_layer time rather than misbehaving quietly. */
#ifndef GOPERL_XS_SDK_PERLIOL_H
#define GOPERL_XS_SDK_PERLIOL_H

#include "perl.h"
#include <stddef.h>

/* The layer-author view replaces the handle type, and the handle functions
 * a layer calls on the layer below it (defined further down). The other
 * handle functions of perl.h stay: they take any handle of the interpreter. */
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

typedef struct _PerlIO PerlIOl;
typedef struct _PerlIO_funcs PerlIO_funcs;
typedef PerlIOl *PerlIO;
#define STDCHAR char
typedef struct goperl_perlio_list PerlIO_list_t; /* opaque */

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

#define PERLIOBUF_DEFAULT_BUFSIZ 8192
#define PERLIO_FUNCS_DECL(f) const PerlIO_funcs f
#define PERLIO_FUNCS_CAST(f) ((PerlIO_funcs *)(void *)(f))

#define PerlIOValid(f) \
    ({ GOPERL_ONCE(goperl_PerlIOValid_f_, f); \
       goperl_PerlIOValid_f_ && *goperl_PerlIOValid_f_; })
#define PerlIOBase(f) (*(f))
#define PerlIOSelf(f, type) ((type *)PerlIOBase(f))
#define PerlIONext(f) (&PerlIOBase(f)->next)

/* ---- views of the interpreter's handles -----------------------------------
 * Code that includes this header walks the layers of a handle it got from
 * IoIFP, IoOFP or PerlIO_stdin/stdout/stderr (File::Map takes only a handle
 * whose layers are all binary; PerlIO::Layers answers what they are):
 * PerlIOBase(f)->tab->name and kind, PerlIOBase(f)->flags, PerlIONext(f),
 * and of a buffered layer its buffer (PerlIOSelf(f, PerlIOBuf): buf, end
 * and ptr point into the interpreter's buffer, bufsiz and posn are its).
 * Those handles are views made for the native call from the interpreter's
 * layers as they are then: the handle, and the next field of each layer,
 * are registered with the interpreter's handle they stand for (see
 * goperl_pio_tok), so that every handle function given one of them, or a
 * PerlIONext of one, works on the interpreter's. The functions of a
 * layer's tab are the interpreter's to run and are not in the view. */
typedef struct {
    PerlIOl *top;
    goperl_pio_slot_t reg;
} goperl_pio_handle_view_t;
typedef struct {
    PerlIOBuf buf; /* buf.base is the layer (PerlIOBase) */
    PerlIO_funcs tab;
    goperl_pio_slot_t reg; /* of &buf.base.next: the handle of the layer below */
} goperl_pio_layer_view_t;
/* The words of a layer in the answer of LAYERS with b = 1. */
#define GOPERL_PIO_LAYER_WORDS 10
/* What a buffered layer's view holds of its buffer, from the words of the
 * layer. */
static void goperl_pio_view_buf(goperl_frame_t *f, PerlIOBuf *b, const goperl_gword_t *l) {
    b->buf = l[4] ? (STDCHAR *)goperl_guest_mem(f, l[4]) : 0;
    b->end = l[5] ? (STDCHAR *)goperl_guest_mem(f, l[5]) : 0;
    b->ptr = l[6] ? (STDCHAR *)goperl_guest_mem(f, l[6]) : 0;
    b->bufsiz = l[7];
    b->posn = (Off_t)(((uint64_t)l[9] << 32) | l[8]);
    b->oneword = 0;
}
static PerlIO *goperl_pio_view(goperl_frame_t *f, uint64_t tok) {
    if (!tok) return 0;
    const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op(f, GOPERL_OP_PERLIO, ((uint64_t)GOPERL_PIO_LAYERS << 32) | (uint32_t)tok,
                        1, 0, 0));
    const goperl_gword_t n = w ? w[0] : 0;
    goperl_pio_handle_view_t *hv = (goperl_pio_handle_view_t *)goperl_arena_alloc(
        sizeof(goperl_pio_handle_view_t) + (size_t)n * sizeof(goperl_pio_layer_view_t));
    if (!hv) goperl_croakf(f, "Out of memory!");
    goperl_pio_layer_view_t *lv = (goperl_pio_layer_view_t *)(hv + 1);
    hv->reg.slot = &hv->top;
    hv->reg.tok = (uint32_t)tok;
    hv->reg.next = goperl_pio_views_v;
    goperl_pio_views_v = &hv->reg;
    for (goperl_gword_t i = 0; i < n; i++) {
        const goperl_gword_t *l = w + 1 + GOPERL_PIO_LAYER_WORDS * i;
        lv[i].reg.slot = &lv[i].buf.base.next;
        lv[i].reg.tok = l[0];
        lv[i].reg.next = goperl_pio_views_v;
        goperl_pio_views_v = &lv[i].reg;
        lv[i].tab.fsize = sizeof(PerlIO_funcs);
        lv[i].tab.name = l[1] ? (const char *)goperl_guest_mem(f, l[1]) : "";
        lv[i].tab.size = (l[2] & PERLIO_K_BUFFERED) ? sizeof(PerlIOBuf) : sizeof(PerlIOl);
        lv[i].tab.kind = l[2];
        lv[i].buf.base.tab = &lv[i].tab;
        lv[i].buf.base.flags = l[3];
        lv[i].buf.base.next = i + 1 < n ? &lv[i + 1].buf.base : 0;
        goperl_pio_view_buf(f, &lv[i].buf, l);
    }
    hv->top = n ? &lv[0].buf.base : 0;
    return &hv->top;
}
#undef IoIFP
#undef IoOFP
#undef PerlIO_stdin
#undef PerlIO_stdout
#undef PerlIO_stderr
#define IoIFP(io) goperl_pio_view(_gof, goperl_pio(_gof, GOPERL_PIO_IFP, (io), 0, 0, 0))
#define IoOFP(io) goperl_pio_view(_gof, goperl_pio(_gof, GOPERL_PIO_OFP, (io), 0, 0, 0))
#define PerlIO_stdin() goperl_pio_view(_gof, goperl_pio(_gof, GOPERL_PIO_STDIN, 0, 0, 0, 0))
#define PerlIO_stdout() goperl_pio_view(_gof, goperl_pio(_gof, GOPERL_PIO_STDOUT, 0, 0, 0, 0))
#define PerlIO_stderr() goperl_pio_view(_gof, goperl_pio(_gof, GOPERL_PIO_STDERR, 0, 0, 0, 0))

/* utf8_strict walks PL_perlio to flush line-buffered TTY handles; there is
 * no host-visible handle table, and the walk exits immediately on NULL. */
__attribute__((weak)) PerlIOl *goperl_pl_perlio_null_v = 0;
#define PL_perlio (goperl_pl_perlio_null_v)

/* ---- host shadow instances ---------------------------------------------- */

/* One per pushed guest proxy layer: the dist's layer struct (funcs->size
 * bytes, PerlIOSelf state persists here), a handle slot for the dist node,
 * and a tail node whose funcs forward to the guest layer below. */
typedef struct goperl_iol_shadow {
    struct goperl_iol_shadow *next;
    uint32_t ltok; /* guest GoperlIOL* — the instance identity */
    uint32_t ff;   /* guest PerlIO* handle, valid during a hook call */
    PerlIOl *fslot;          /* handle: &fslot is the PerlIO* */
    struct {
        PerlIOl base;
        struct goperl_iol_shadow *owner;
    } tail;
    PerlIOl *tailslot;
    /* dist layer struct follows (funcs->size bytes) */
} goperl_iol_shadow_t;

__attribute__((weak)) goperl_iol_shadow_t *goperl_iol_shadows_v = 0;

/* ---- tail: the guest layer below the proxy ------------------------------ */

static goperl_iol_shadow_t *goperl_iol_owner(PerlIO *f) {
    /* Tail funcs receive &shadow->tailslot; the tail PerlIOl is embedded
     * in the shadow, so the owner is a fixed offset away. */
    PerlIOl *l = *f;
    return (goperl_iol_shadow_t *)((char *)l -
                                   offsetof(goperl_iol_shadow_t, tail));
}

static SSize_t goperl_tail_read(pTHX_ PerlIO *f, void *vbuf, Size_t count) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    /* vbuf is a host view into guest memory (the fill buffer); recover the
     * guest address from the current fill window. */
    extern uint32_t goperl_iol_gbuf_v;
    extern char *goperl_iol_hbuf_v;
    uint32_t gdst = goperl_iol_gbuf_v + (uint32_t)((char *)vbuf - goperl_iol_hbuf_v);
    int64_t got = (int64_t)goperl_do_op(
        goperl_cur_frame_v, GOPERL_OP_PERLIO_NEXT_READ, sh->ff,
        ((uint64_t)gdst << 32) | (uint32_t)count, 0, 0);
    return (SSize_t)got;
}
static IV goperl_tail_fill(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    return (IV)(int64_t)goperl_do_op(goperl_cur_frame_v,
                                     GOPERL_OP_PERLIO_NEXT_FILL, sh->ff, 0, 0,
                                     0);
}
static IV goperl_tail_eof(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    return (IV)(goperl_do_op(goperl_cur_frame_v, GOPERL_OP_PERLIO_STATE,
                             sh->ff, 0, 0, 0) &
                1);
}
static IV goperl_tail_error(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    return (IV)((goperl_do_op(goperl_cur_frame_v, GOPERL_OP_PERLIO_STATE,
                              sh->ff, 0, 0, 0) >>
                 1) &
                1);
}
static IV goperl_tail_flush(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    return (IV)(int64_t)goperl_do_op(goperl_cur_frame_v,
                                     GOPERL_OP_PERLIO_STATE, sh->ff, 1, 0, 0);
}
static STDCHAR *goperl_tail_get_ptr(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    extern uint32_t goperl_iol_gbuf_v;
    extern char *goperl_iol_hbuf_v;
    uint32_t gp = (uint32_t)goperl_do_op(
        goperl_cur_frame_v, GOPERL_OP_PERLIO_NEXT_GETPTR, sh->ff, 0, 0, 0);
    /* translate the guest pointer into host space */
    return (STDCHAR *)goperl_guest_mem(goperl_cur_frame_v, gp);
}
static SSize_t goperl_tail_get_cnt(pTHX_ PerlIO *f) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    return (SSize_t)(int64_t)goperl_do_op(goperl_cur_frame_v,
                                          GOPERL_OP_PERLIO_NEXT_GETCNT,
                                          sh->ff, 0, 0, 0);
}
static void goperl_tail_set_ptrcnt(pTHX_ PerlIO *f, STDCHAR *ptr,
                                   SSize_t cnt) {
    goperl_iol_shadow_t *sh = goperl_iol_owner(f);
    /* ptr is a host view of a guest pointer obtained from get_ptr; convert
     * back by asking the guest for its current ptr and offsetting. */
    uint32_t gp = (uint32_t)goperl_do_op(
        goperl_cur_frame_v, GOPERL_OP_PERLIO_NEXT_GETPTR, sh->ff, 0, 0, 0);
    char *hp = (char *)goperl_guest_mem(goperl_cur_frame_v, gp);
    uint32_t gnew = gp + (uint32_t)((char *)ptr - hp);
    goperl_do_op(goperl_cur_frame_v, GOPERL_OP_PERLIO_NEXT_SETPTRCNT, sh->ff,
                 ((uint64_t)gnew << 32) | (uint32_t)(int32_t)cnt, 0, 0);
}

/* PERLIO_K_FASTGETS on the tail advertises get_ptr/get_cnt support; the
 * actual answer comes from the guest per call. */
static IV goperl_tail_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg,
                             PerlIO_funcs *tab) {
    (void)f;
    (void)mode;
    (void)arg;
    (void)tab;
    return 0;
}

__attribute__((weak)) PerlIO_funcs goperl_tail_funcs_v = {
    sizeof(PerlIO_funcs), "goperl_tail", sizeof(PerlIOl),
    PERLIO_K_BUFFERED | PERLIO_K_FASTGETS,
    goperl_tail_pushed, 0, 0, 0, 0, 0, 0,
    goperl_tail_read, 0, 0, 0, 0, 0,
    goperl_tail_flush, goperl_tail_fill, goperl_tail_eof, goperl_tail_error,
    0, 0, 0, 0, goperl_tail_get_ptr, goperl_tail_get_cnt,
    goperl_tail_set_ptrcnt,
};

/* PerlIO_* dispatchers used by hook bodies (dispatch through the tab); a
 * view of a handle of the interpreter's (or of a layer of one) is the
 * interpreter's to work on. */
static int goperl_pio_is_view(const PerlIO *f) { return goperl_pio_view_of(f) != 0; }
static SSize_t PerlIO_read(PerlIO *f, void *vbuf, Size_t count) {
    if (goperl_pio_is_view(f)) return goperl_pio_read(goperl_cur_frame_v, f, vbuf, count);
    return PerlIOBase(f)->tab->Read(aTHX_ f, vbuf, count);
}
static IV PerlIO_fill(PerlIO *f) {
    if (goperl_pio_is_view(f)) return goperl_pio_status(goperl_cur_frame_v, GOPERL_PIO_FILL, f);
    return PerlIOBase(f)->tab->Fill(aTHX_ f);
}
static IV PerlIO_eof(PerlIO *f) {
    if (goperl_pio_is_view(f))
        return (IV)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_EOF, f, 0, 0, 0);
    return PerlIOBase(f)->tab->Eof(aTHX_ f);
}
static IV PerlIO_error(PerlIO *f) {
    if (goperl_pio_is_view(f))
        return (IV)(int64_t)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_ERROR, f, 0, 0, 0);
    return PerlIOBase(f)->tab->Error(aTHX_ f);
}
static IV PerlIO_flush(PerlIO *f) {
    if (goperl_pio_is_view(f)) return goperl_pio_flush(goperl_cur_frame_v, f);
    if (!PerlIOValid(f)) return 0;
    return PerlIOBase(f)->tab->Flush(aTHX_ f);
}
static STDCHAR *PerlIO_get_ptr(PerlIO *f) {
    if (goperl_pio_is_view(f))
        return goperl_pio_bufptr(goperl_cur_frame_v, GOPERL_PIO_GET_PTR, f);
    return PerlIOBase(f)->tab->Get_ptr(aTHX_ f);
}
static SSize_t PerlIO_get_cnt(PerlIO *f) {
    if (goperl_pio_is_view(f))
        return (SSize_t)(int64_t)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_GET_CNT, f, 0, 0, 0);
    return PerlIOBase(f)->tab->Get_cnt(aTHX_ f);
}
static void PerlIO_set_ptrcnt(PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    if (goperl_pio_is_view(f)) {
        goperl_pio_set_ptrcnt(goperl_cur_frame_v, f, ptr, cnt);
        return;
    }
    PerlIOBase(f)->tab->Set_ptrcnt(aTHX_ f, ptr, cnt);
}
static STDCHAR *PerlIO_get_base(PerlIO *f) {
    if (goperl_pio_is_view(f)) {
        /* the layer may get its buffer now: its view gets it too */
        goperl_frame_t *const gf = goperl_cur_frame_v;
        STDCHAR *const base = goperl_pio_bufptr(gf, GOPERL_PIO_GET_BASE, f);
        const goperl_gword_t *w = (const goperl_gword_t *)goperl_guest_mem(
            gf, goperl_do_op(gf, GOPERL_OP_PERLIO,
                             ((uint64_t)GOPERL_PIO_LAYERS << 32) | goperl_pio_tok(gf, f), 1, 0, 0));
        if (w && w[0] && *f && (PerlIOBase(f)->tab->kind & PERLIO_K_BUFFERED))
            goperl_pio_view_buf(gf, PerlIOSelf(f, PerlIOBuf), w + 1);
        return base;
    }
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    return b->buf;
}
static int PerlIO_fast_gets(PerlIO *f) {
    goperl_iol_shadow_t *sh;
    if (goperl_pio_is_view(f))
        return (int)goperl_pio(goperl_cur_frame_v, GOPERL_PIO_FAST_GETS, f, 0, 0, 0);
    if (!PerlIOValid(f)) return 0;
    if (PerlIOBase(f)->tab != &goperl_tail_funcs_v)
        return (PerlIOBase(f)->tab->kind & PERLIO_K_FASTGETS) != 0;
    sh = goperl_iol_owner(f);
    return (int)goperl_do_op(goperl_cur_frame_v,
                             GOPERL_OP_PERLIO_NEXT_FASTGETS, sh->ff, 0, 0, 0);
}

/* ---- stock slot implementations (identity anchors + host behavior) ------ */

static IV PerlIOBuf_pushed(pTHX_ PerlIO *f, const char *mode, SV *arg,
                           PerlIO_funcs *tab) {
    /* The GUEST proxy already ran the real PerlIOBuf_pushed before
     * forwarding; the host copy just initializes the shadow. */
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    (void)mode;
    (void)arg;
    b->base.tab = tab;
    b->buf = b->end = b->ptr = 0;
    b->bufsiz = 0;
    b->posn = 0;
    return 0;
}
static IV PerlIOBuf_popped(pTHX_ PerlIO *f) {
    (void)f;
    return 0;
}
static PerlIO *PerlIOBuf_open(pTHX_ PerlIO_funcs *tab, PerlIO_list_t *layers,
                              IV n, const char *mode, int fd, int imode,
                              int perm, PerlIO *old, int narg, SV **args) {
    (void)tab; (void)layers; (void)n; (void)mode; (void)fd; (void)imode;
    (void)perm; (void)old; (void)narg; (void)args;
    return 0;
}
static IV PerlIOBase_binmode(pTHX_ PerlIO *f) { (void)f; return 0; }
static IV PerlIOBase_fileno(pTHX_ PerlIO *f) { (void)f; return -1; }
static PerlIO *PerlIOBuf_dup(pTHX_ PerlIO *f, PerlIO *o, CLONE_PARAMS *param,
                             int flags) {
    (void)f; (void)o; (void)param; (void)flags;
    return 0;
}
static SSize_t PerlIOBuf_read(pTHX_ PerlIO *f, void *vbuf, Size_t count) {
    (void)f; (void)vbuf; (void)count;
    return -1;
}
static SSize_t PerlIOBase_unread(pTHX_ PerlIO *f, const void *vbuf,
                                 Size_t count) {
    (void)f; (void)vbuf; (void)count;
    return -1;
}
static SSize_t PerlIOBuf_write(pTHX_ PerlIO *f, const void *vbuf,
                               Size_t count) {
    (void)f; (void)vbuf; (void)count;
    return -1;
}
static IV PerlIOBuf_seek(pTHX_ PerlIO *f, Off_t offset, int whence) {
    (void)f; (void)offset; (void)whence;
    return -1;
}
static Off_t PerlIOBuf_tell(pTHX_ PerlIO *f) { (void)f; return 0; }
static IV PerlIOBuf_close(pTHX_ PerlIO *f) { (void)f; return 0; }
static IV PerlIOBuf_flush(pTHX_ PerlIO *f) { (void)f; return 0; }
static IV PerlIOBuf_fill(pTHX_ PerlIO *f) { (void)f; return -1; }
static IV PerlIOBase_eof(pTHX_ PerlIO *f) {
    return (PerlIOBase(f)->flags & PERLIO_F_EOF) != 0;
}
static IV PerlIOBase_error(pTHX_ PerlIO *f) {
    return (PerlIOBase(f)->flags & PERLIO_F_ERROR) != 0;
}
static void PerlIOBase_clearerr(pTHX_ PerlIO *f) {
    PerlIOBase(f)->flags &= ~(U32)(PERLIO_F_EOF | PERLIO_F_ERROR);
}
static void PerlIOBase_setlinebuf(pTHX_ PerlIO *f) { (void)f; }
static STDCHAR *PerlIOBuf_get_base(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOBuf)->buf;
}
static Size_t PerlIOBuf_bufsiz(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOBuf)->bufsiz;
}
static STDCHAR *PerlIOBuf_get_ptr(pTHX_ PerlIO *f) {
    return PerlIOSelf(f, PerlIOBuf)->ptr;
}
static SSize_t PerlIOBuf_get_cnt(pTHX_ PerlIO *f) {
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    return b->end - b->ptr;
}
static void PerlIOBuf_set_ptrcnt(pTHX_ PerlIO *f, STDCHAR *ptr, SSize_t cnt) {
    PerlIOBuf *b = PerlIOSelf(f, PerlIOBuf);
    b->ptr = ptr;
    b->end = ptr + cnt;
}

/* ---- registration -------------------------------------------------------- */

static void PerlIO_define_layer(pTHX_ PerlIO_funcs *tab) {
    uint32_t mask = 0;
    /* Every slot must be NULL, a stock address (guest handles it), or one
     * of the supported custom hooks. */
    if (tab->Fill && tab->Fill != PerlIOBuf_fill) mask |= 0x2;
    if (tab->Pushed && tab->Pushed != PerlIOBuf_pushed) mask |= 0x1;
#define GOPERL_IOL_REQ(slot, stock)                                        \
    if (tab->slot && (void *)tab->slot != (void *)stock)                   \
    goperl_croakf(goperl_cur_frame_v,                                      \
                  "PerlIO layer '%s': custom '" #slot                      \
                  "' slot is not supported by the goperl SDK yet",         \
                  tab->name)
    GOPERL_IOL_REQ(Popped, PerlIOBuf_popped);
    GOPERL_IOL_REQ(Open, PerlIOBuf_open);
    GOPERL_IOL_REQ(Binmode, PerlIOBase_binmode);
    GOPERL_IOL_REQ(Fileno, PerlIOBase_fileno);
    GOPERL_IOL_REQ(Dup, PerlIOBuf_dup);
    GOPERL_IOL_REQ(Read, PerlIOBuf_read);
    GOPERL_IOL_REQ(Unread, PerlIOBase_unread);
    GOPERL_IOL_REQ(Write, PerlIOBuf_write);
    GOPERL_IOL_REQ(Seek, PerlIOBuf_seek);
    GOPERL_IOL_REQ(Tell, PerlIOBuf_tell);
    GOPERL_IOL_REQ(Close, PerlIOBuf_close);
    GOPERL_IOL_REQ(Flush, PerlIOBuf_flush);
    GOPERL_IOL_REQ(Eof, PerlIOBase_eof);
    GOPERL_IOL_REQ(Error, PerlIOBase_error);
    GOPERL_IOL_REQ(Clearerr, PerlIOBase_clearerr);
    GOPERL_IOL_REQ(Setlinebuf, PerlIOBase_setlinebuf);
    GOPERL_IOL_REQ(Get_base, PerlIOBuf_get_base);
    GOPERL_IOL_REQ(Get_bufsiz, PerlIOBuf_bufsiz);
    GOPERL_IOL_REQ(Get_ptr, PerlIOBuf_get_ptr);
    GOPERL_IOL_REQ(Get_cnt, PerlIOBuf_get_cnt);
    GOPERL_IOL_REQ(Set_ptrcnt, PerlIOBuf_set_ptrcnt);
#undef GOPERL_IOL_REQ
    uint32_t id = goperl_api_perlio_def(goperl_cur_frame_v, tab->name,
                                           (void *)tab);
    goperl_do_op(goperl_cur_frame_v, GOPERL_OP_PERLIO_DEF_LAYER, id, mask,
                 tab->name, strlen(tab->name));
}

/* ---- the hook dispatcher (called by the loader on method -6) ------------ */

__attribute__((weak)) uint32_t goperl_iol_gbuf_v = 0;
__attribute__((weak)) char *goperl_iol_hbuf_v = 0;

static goperl_iol_shadow_t *goperl_iol_shadow(PerlIO_funcs *tab, uint32_t ltok,
                                              uint32_t ff) {
    goperl_iol_shadow_t *sh;
    for (sh = goperl_iol_shadows_v; sh; sh = sh->next)
        if (sh->ltok == ltok) {
            sh->ff = ff;
            return sh;
        }
    sh = (goperl_iol_shadow_t *)calloc(1, sizeof(goperl_iol_shadow_t) +
                                              tab->size);
    sh->ltok = ltok;
    sh->ff = ff;
    sh->fslot = (PerlIOl *)(sh + 1);
    sh->fslot->tab = tab;
    sh->fslot->next = &sh->tail.base;
    sh->tail.base.tab = &goperl_tail_funcs_v;
    sh->tail.owner = sh;
    sh->tailslot = &sh->tail.base;
    sh->next = goperl_iol_shadows_v;
    goperl_iol_shadows_v = sh;
    return sh;
}

__attribute__((weak, visibility("default"), used)) int32_t
__goperl_perlio_invoke(goperl_frame_t *f, void *funcs_v,
                       const unsigned char *req, uint32_t reqlen,
                       unsigned char *resp, uint32_t respcap) {
    PerlIO_funcs *tab = (PerlIO_funcs *)funcs_v;
    if (reqlen < 16 || respcap < 32) return 0;
    uint32_t hook, ltok, ff;
    memcpy(&hook, req, 4);
    memcpy(&ff, req + 8, 4);
    memcpy(&ltok, req + 12, 4);

    jmp_buf jb;
    void *prev_jb = f->jb;
    goperl_frame_t *prev_frame = goperl_cur_frame_v;
    f->jb = (void *)&jb;
    goperl_frame_enter(f);
    if (setjmp(jb)) {
        size_t ml = strlen(f->err);
        if (ml > respcap - 1) ml = respcap - 1;
        resp[0] = 0;
        memcpy(resp + 1, f->err, ml);
        f->failed = 0;
        f->jb = prev_jb;
        goperl_cur_frame_v = prev_frame;
        return (int32_t)(1 + ml);
    }

    int32_t rlen = 1;
    resp[0] = 1;
    switch (hook) {
    case 1: { /* PUSHED: extra = [u32 argtok][mode NUL] */
        goperl_iol_shadow_t *sh = goperl_iol_shadow(tab, ltok, ff);
        uint32_t argtok = 0;
        const char *mode = "r";
        if (reqlen >= 21) {
            memcpy(&argtok, req + 16, 4);
            mode = (const char *)req + 20;
        }
        PerlIO *hf = &sh->fslot;
        U32 before = PerlIOBase(hf)->flags;
        IV code = tab->Pushed(hf, mode, (SV *)(uintptr_t)argtok, tab);
        U32 after = PerlIOBase(hf)->flags;
        if (code != 0) goperl_croakf(f, "layer '%s' Pushed failed", tab->name);
        uint32_t setf = after & ~before, clrf = before & ~after;
        memcpy(resp + 1, &setf, 4);
        memcpy(resp + 5, &clrf, 4);
        rlen = 9;
        break;
    }
    case 2: { /* FILL: extra = [u32 gbuf][u32 bufsiz][u32 guest flags] */
        goperl_iol_shadow_t *sh = goperl_iol_shadow(tab, ltok, ff);
        uint32_t gbuf = 0, bufsiz = 0, gflags = 0;
        if (reqlen >= 28) {
            memcpy(&gbuf, req + 16, 4);
            memcpy(&bufsiz, req + 20, 4);
            memcpy(&gflags, req + 24, 4);
        }
        PerlIO *hf = &sh->fslot;
        PerlIOBuf *b = PerlIOSelf(hf, PerlIOBuf);
        goperl_iol_gbuf_v = gbuf;
        goperl_iol_hbuf_v = (char *)goperl_guest_mem(f, gbuf);
        b->buf = goperl_iol_hbuf_v;
        b->bufsiz = bufsiz;
        b->ptr = b->end = b->buf;
        /* mirror the guest's sticky flags, minus buffer state */
        PerlIOBase(hf)->flags =
            (gflags & ~(U32)(PERLIO_F_RDBUF | PERLIO_F_WRBUF));
        IV code = tab->Fill(hf);
        uint32_t status = code == 0 ? 0 : 1;
        uint32_t ptrOff = (uint32_t)(b->ptr - b->buf);
        uint32_t endOff = (uint32_t)(b->end - b->buf);
        uint32_t setf = PerlIOBase(hf)->flags &
                        (PERLIO_F_EOF | PERLIO_F_ERROR | PERLIO_F_RDBUF);
        memcpy(resp + 1, &status, 4);
        memcpy(resp + 5, &ptrOff, 4);
        memcpy(resp + 9, &endOff, 4);
        memcpy(resp + 13, &setf, 4);
        rlen = 17;
        break;
    }
    case 3: { /* POPPED: drop the shadow */
        goperl_iol_shadow_t **p = &goperl_iol_shadows_v;
        while (*p && (*p)->ltok != ltok) p = &(*p)->next;
        if (*p) {
            goperl_iol_shadow_t *dead = *p;
            *p = dead->next;
            free(dead);
        }
        break;
    }
    }
    f->jb = prev_jb;
    goperl_cur_frame_v = prev_frame;
    return rlen;
}

#endif /* GOPERL_XS_SDK_PERLIOL_H */
