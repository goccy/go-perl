/* go-perl native XS SDK: the compiler's structures — pads, the fields of a
 * sub, the interpreter variables of a compilation, the hints of a COP — as
 * native code reaches them on perl: through their fields, assigning them,
 * and saving them for the interpreter to put back when a scope ends.
 *
 * Included at the end of goperl_perlapi.h; not meant to be included on its
 * own. */
#ifndef GOPERL_XS_SDK_COMPILER_H
#define GOPERL_XS_SDK_COMPILER_H
#pragma GCC system_header

#include <sys/mman.h>

/* INTERNAL selectors (mirror perl-wasm's perl_xs_helper). */
enum {
    GOPERL_INT_LAYOUT = 0,
    GOPERL_INT_SAVE_PLVAR = 1,
    GOPERL_INT_SAVE_MEM = 2,
    GOPERL_INT_SAVE_AVELEM = 3,
    GOPERL_INT_LEAVE_SCOPE = 4,
    GOPERL_INT_CV_GET = 5,
    GOPERL_INT_CV_SET = 6,
    GOPERL_INT_PAD_NEW = 7,
    GOPERL_INT_PADNAME_NEW = 8,
    GOPERL_INT_PADNAME_FREE = 9,
    GOPERL_INT_SAVE_FREEOP = 10,
    GOPERL_INT_RUNOPS = 11,
    GOPERL_INT_COP_LABEL = 12,
    GOPERL_INT_HINTHASH = 13,
    GOPERL_INT_HINTHASH_FETCH = 14,
    GOPERL_INT_HINTHASH_2HV = 15,
    GOPERL_INT_OBJ_FIELDS = 16,
    GOPERL_INT_OBJ_MAXFIELD_SET = 17,
    GOPERL_INT_OBJ_FIELDS_NEW = 18,
    GOPERL_INT_PARSER_NEW = 19,
    GOPERL_INT_HV_ARRAY = 20,
    GOPERL_INT_PERLIO_DEBUG = 21,
    GOPERL_INT_SV_PV_FREE = 22,
    GOPERL_INT_STATCACHE = 23,
    GOPERL_INT_SAVE_FIELD = 31, /* a: an SV; b: field<<8 | kind: a save of the field */
    GOPERL_INT_HASH = 32,       /* s+b: a key's bytes -> its hash (PERL_HASH) */
    GOPERL_INT_PUSHSTACK = 33,  /* b: the stackinfo type: PUSHSTACKi */
    GOPERL_INT_POPSTACK = 34,   /* POPSTACK */
    GOPERL_INT_DOCATCH = 43,    /* a: the running op: docatch with its pp function first */
    GOPERL_INT_SAVE_AELEM = 44, /* a: an array; b: flags<<32 | the index */
    GOPERL_INT_PADNAMELIST_NEW = 60, /* b: max -> newPADNAMELIST's list */
    GOPERL_INT_PADNAMELIST_FREE = 61, /* a: the list: PadnamelistREFCNT_dec */
    GOPERL_INT_PADNAME_OUTER = 62,  /* a: a pad name -> newPADNAMEouter's */
    GOPERL_INT_UTF8 = 42        /* a: the flags | (GOPERL_UTF8_* | GOPERL_UTF8_MSGS)
                                 * << 32; s+b: one character's bytes, or b: a
                                 * code point -> words */
};
/* The operations of INTERNAL UTF8 (perl-wasm's perl.cc). */
enum {
    GOPERL_UTF8_DECODE = 0,
    GOPERL_UTF8_DECODE_BUF = 1,
    GOPERL_UTF8_CHAR = 2,
    GOPERL_UTF8_PARTIAL = 3,
    GOPERL_UTF8_ENCODE = 4
};
#define GOPERL_UTF8_MSGS 0x100
/* The fields of a sub (INTERNAL CV_GET / CV_SET). */
enum {
    GOPERL_CVF_FLAGS = 0,
    GOPERL_CVF_ROOT = 1,
    GOPERL_CVF_START = 2,
    GOPERL_CVF_OUTSIDE_SEQ = 3,
    GOPERL_CVF_PADLIST = 4,
    GOPERL_CVF_DEPTH = 5,
    GOPERL_CVF_FILE = 6,
    GOPERL_CVF_STASH = 7,
    GOPERL_CVF_GV_U = 8,
    GOPERL_CVF_OUTSIDE = 9
};
/* (GOPERL_PL_PARSER, 94, is with the parser's copy in perl.h) */
/* PL_multideref_pc (an address of the interpreter's memory there, one in
 * this process here) and PL_nomemok. */
#define GOPERL_PL_MULTIDEREF_PC 95
#define GOPERL_PL_NOMEMOK 96
/* INTERNAL: a = the token (or the address), c = the selector, and in c's
 * high half a 32-bit value that goes with b (GOPERL_INTERNAL_HI). */
#define GOPERL_INTERNAL(f, sel, tok, b)                                     \
    goperl_do_op3((f), GOPERL_OP_INTERNAL, (uint64_t)(tok), (uint64_t)(b), (uint64_t)(sel), 0, 0)
#define GOPERL_INTERNAL_HI(f, sel, tok, b, hi)                              \
    goperl_do_op3((f), GOPERL_OP_INTERNAL, (uint64_t)(tok), (uint64_t)(b),  \
                  (uint64_t)(uint32_t)(sel) | ((uint64_t)(uint32_t)(hi) << 32), 0, 0)
#define GOPERL_INTERNAL_S(f, sel, tok, b, s, slen)                          \
    goperl_do_op3((f), GOPERL_OP_INTERNAL, (uint64_t)(tok), (uint64_t)(b), (uint64_t)(sel), (const char *)(s), (uint64_t)(slen))
/* PERL_HASH: the interpreter's hash of a key's bytes (see perl.h). */
static U32 goperl_key_hash(goperl_frame_t *f, const char *key, uint64_t len) {
    return (U32)GOPERL_INTERNAL_S(f, GOPERL_INT_HASH, 0, len, key, len);
}


/* What this SDK keeps per interpreter for the compiler's structures (the
 * shared block holds the pointer). */
typedef struct goperl_padent goperl_padent_t;
typedef struct goperl_compiler_state {
    uint32_t layout[32]; /* the offsets of the fields of the pad structures */
    int32_t have_layout;
    int32_t nbuckets;
    goperl_padent_t **buckets; /* the shadows of pad structures, by address */
    goperl_padent_t *active;   /* the ones written back: those native code
                                * has looked at (goperl_padent_fresh) */
    uint32_t gen;              /* the operations of the interpreter since the
                                * shadows began: a shadow is up to date when
                                * it was refreshed at this one */
    /* PL_curpad given values of native code's own, in its own memory
     * (Object::Pad's constructor gives the field hooks it runs a pad of
     * two values that way): the values, which the interpreter cannot
     * address, and the pad in use they stand in for (see goperl_curpad_ref) */
    SV **curpad_own;
    uint64_t curpad_own_pad;
} goperl_compiler_state_t;

static __attribute__((unused)) goperl_compiler_state_t *goperl_compiler(goperl_frame_t *f) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_compiler_state_t *c = (goperl_compiler_state_t *)sh->compiler;
    if (c) return c;
    c = (goperl_compiler_state_t *)calloc(1, sizeof(*c));
    if (!c) goperl_croakf(f, "Out of memory!");
    c->nbuckets = 256;
    c->buckets = (goperl_padent_t **)calloc((size_t)c->nbuckets, sizeof(void *));
    if (!c->buckets) goperl_croakf(f, "Out of memory!");
    sh->compiler = c;
    return c;
}

/* ---- saves the interpreter makes ------------------------------------------
 * A save of a location of the interpreter's is the interpreter's to make,
 * so that it puts the location back when its scope ends, as perl does —
 * also when that scope ends after native code has returned (a keyword
 * plugin saves PL_compcv and points it at a sub of its own, which the
 * interpreter compiles into until the block ends). Such a location is the
 * slot of an interpreter variable here (SAVESPTR(PL_comppad)), PL_curcop,
 * PL_op, or a location in the interpreter's memory: a byte of the name of a
 * pad entry (SAVEI8), or a slot of the body of an array as AvARRAY or
 * PL_curpad designates it (SAVESPTR(PAD_SVl(ix))). Every other location is
 * native code's own, saved here. */

/* A save of PL_curpad saves values of native code's own that stand in for
 * the interpreter's as well: the interpreter puts back its own, and these
 * are put back here when the scope ends. */
static __attribute__((unused)) void goperl_save_curpad_own(goperl_frame_t *f) {
    goperl_compiler_state_t *const c = goperl_compiler(f);
    goperl_hostsave_ptr(f, (void **)&c->curpad_own);
    goperl_hostsave_ptr(f, (void **)&c->curpad_own_pad);
}

static __attribute__((unused)) int goperl_save_routed(goperl_frame_t *f, void *loc, int kind, int width) {
    if (!f || !f->inst) return 0;
    const char *ps = (const char *)goperl_plshadow_v;
    if ((const char *)loc >= ps &&
        (const char *)loc < ps + sizeof(goperl_plshadow_t) * GOPERL_PLSHADOW_MAX) {
        const size_t off = (size_t)((const char *)loc - ps);
        if (off % sizeof(goperl_plshadow_t) == offsetof(goperl_plshadow_t, val)) {
            const int id = (int)(off / sizeof(goperl_plshadow_t));
            GOPERL_INTERNAL(f, GOPERL_INT_SAVE_PLVAR, id, kind);
            if (id == GOPERL_PL_CURPAD) goperl_save_curpad_own(f);
            return 1;
        }
    }
    if (f->inst && ((goperl_shared_t *)f->inst)->compiler &&
        loc == (void *)&goperl_compiler(f)->curpad_own) {
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_PLVAR, GOPERL_PL_CURPAD, kind);
        goperl_save_curpad_own(f);
        return 1;
    }
    if (loc == (void *)&goperl_curcop_slot_v) {
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_PLVAR, GOPERL_PL_CURCOP, kind);
        return 1;
    }
    if (loc == (void *)&f->plop) {
        /* the interpreter puts its PL_op back, and the op native code has
         * as PL_op is put back with it: a module that points PL_op at an op
         * of its own around a call (Syntax::Operator::In) returns the op
         * after its own when the scope has ended */
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_PLVAR, GOPERL_PL_OP, kind);
        goperl_hostsave_ptr(f, (void **)&f->plop);
        return 1;
    }
    if (loc == (void *)&goperl_parser_ptr_v) {
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_PLVAR, GOPERL_PL_PARSER, kind);
        return 1;
    }
    const uint64_t g = goperl_guest_addr(f, loc);
    if (g) {
        /* bytes, or a pointer (a slot of an array's body) */
        if (kind == GOPERL_SAVEK_INT ? (width != 1 && width != 2 && width != 4)
                                     : width != (int)sizeof(void *))
            goperl_croakf(f, "a save of %d bytes of the interpreter's memory as %s", width,
                          kind == GOPERL_SAVEK_INT ? "an integer" : "a pointer");
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_MEM, g, ((uint64_t)width << 8) | (uint64_t)kind);
        return 1;
    }
    /* the slot of a field an accessor designates (GvSV(gv), SvIVX(sv)):
     * the field is saved where it is, the slot being only this process's
     * copy of it */
    goperl_lvstate_t *st = goperl_lvstate_v;
    goperl_lvslot_t *m = 0;
    for (int32_t i = 0; st && !m && i < st->nuses; i++)
        if ((void *)&st->uses[i].slot->val == loc) m = st->uses[i].slot;
    for (int32_t i = 0; st && !m && i < st->nrecent; i++)
        if ((void *)&st->recent[i]->val == loc) m = st->recent[i];
    if (m) {
        switch (m->field) {
        case GOPERL_LV_GVSV:
        case GOPERL_LV_GVAV:
        case GOPERL_LV_GVHV:
        case GOPERL_LV_RV:
            if (kind == GOPERL_SAVEK_INT)
                goperl_croakf(f, "an integer save of a field that holds a pointer");
            break;
        case GOPERL_LV_IV:
        case GOPERL_LV_UV:
        case GOPERL_LV_NV:
            if (kind != GOPERL_SAVEK_INT || width != 8)
                goperl_croakf(f, "a save of a field that holds a number of 8 bytes as %d bytes", width);
            break;
        default:
            return 0;
        }
        /* the operation writes back what is pending, the slot first */
        GOPERL_INTERNAL(f, GOPERL_INT_SAVE_FIELD, m->svtok, ((uint64_t)m->field << 8) | (uint64_t)kind);
        return 1;
    }
    return 0;
}

/* LEAVE_SCOPE: the interpreter's saves above the index are undone. */
#define LEAVE_SCOPE(old) \
    ((void)GOPERL_INTERNAL(_gof, GOPERL_INT_LEAVE_SCOPE, 0, (uint64_t)(int64_t)(I32)(old)))
#define SAVEFREEOP(o)                                                       \
    ({ OP *const goperl_SAVEFREEOP_o_ = (OP *)(o);                          \
       (void)GOPERL_INTERNAL(_gof, GOPERL_INT_SAVE_FREEOP,                  \
                             goperl_SAVEFREEOP_o_ ? goperl_SAVEFREEOP_o_->gop : 0, 0); })
#define save_freeop(o) SAVEFREEOP(o)
/* save_aelem (scope.c): the interpreter saves the element of its array and
 * puts a new value in its place (the address native code gives is the one
 * av_fetch gave: the interpreter fetches the element itself) */
#define save_aelem_flags(av, idx, svp, flags)                                         \
    ((void)(svp), (void)GOPERL_INTERNAL(_gof, GOPERL_INT_SAVE_AELEM, GOPERL_TOK(av),   \
                                        ((uint64_t)(U32)(flags) << 32) |             \
                                            (uint64_t)(uint32_t)(I32)(idx)))
#define save_aelem(av, idx, svp) save_aelem_flags((av), (idx), (svp), SAVEf_SETMAGIC)
/* the length of a refcounted string of the interpreter's (cop.h's RCPV: a
 * COP's warnings), from its header before the bytes */
#define RCPV_LEN(pv) (((const STRLEN *)(const void *)(pv))[-1] - 1)

/* ---- interpreter variables of a compilation ---------------------------------
 * Lvalues like the others (goperl_plint_ref): read from the interpreter,
 * written back before its next operation. */
#define GOPERL_PL_CURPM 73
#define GOPERL_PL_PADIX 83
#define GOPERL_PL_CONSTPADIX 84
#define GOPERL_PL_COMPPAD_NAME_FILL 85
#define GOPERL_PL_MIN_INTRO_PENDING 86
#define GOPERL_PL_MAX_INTRO_PENDING 87
#define GOPERL_PL_CV_HAS_EVAL 88
#define GOPERL_PL_PAD_RESET_PENDING 89
#define GOPERL_PL_SUBLINE 90
#define GOPERL_PL_SUBNAME 91
#define GOPERL_PL_CURSTNAME 92
#define GOPERL_PL_PADIX_FLOOR 93
#define PL_padix (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_PADIX))
#define PL_constpadix (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_CONSTPADIX))
#define PL_comppad_name_fill \
    (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_COMPPAD_NAME_FILL))
#define PL_min_intro_pending \
    (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_MIN_INTRO_PENDING))
#define PL_max_intro_pending \
    (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_MAX_INTRO_PENDING))
#define PL_padix_floor (*(PADOFFSET *)goperl_plint_ref(_gof, GOPERL_PL_PADIX_FLOOR))
#define PL_cv_has_eval (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_CV_HAS_EVAL))
#define PL_pad_reset_pending \
    (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_PAD_RESET_PENDING))
#define PL_subline (*(I32 *)goperl_plint_ref(_gof, GOPERL_PL_SUBLINE))
#define PL_subname (*goperl_plsv_ref(_gof, GOPERL_PL_SUBNAME))
#define PL_curstname (*goperl_plsv_ref(_gof, GOPERL_PL_CURSTNAME))
#undef PL_cop_seqmax
#define PL_cop_seqmax (*(U32 *)goperl_plint_ref(_gof, GOPERL_PL_COP_SEQMAX))
#undef PL_compcv
#define PL_compcv (*(CV **)goperl_plsv_ref(_gof, GOPERL_PL_COMPCV))
#undef PL_comppad
/* The pad in use: assigning it makes its values PL_curpad as well. */
#define PL_comppad (*(PAD **)goperl_plsv_ref(_gof, GOPERL_PL_COMPPAD))

/* ---- pads ---------------------------------------------------------------------
 * A padlist, a list of pad names and a pad name are the interpreter's
 * structures, with perl's fields; here each is a shadow of the one in the
 * interpreter's memory, one per structure (pointers to them compare as
 * perl's do). The shadows a native call has reached are read again after
 * every operation of the interpreter, and what native code stored into
 * them is written back before the next one: a field as it is, an entry of
 * an array (the names of a list, the pads of a padlist) as the address of
 * the structure stored there. A pad (PAD) is an array, reached as arrays
 * are. */
typedef struct padlist PADLIST;
typedef struct padnamelist PADNAMELIST;
typedef struct padname PADNAME;
typedef AV PAD;
struct padname_fieldinfo;
struct padname {
    char *xpadn_pv;
    HV *xpadn_ourstash;
    union {
        HV *xpadn_typestash;
        CV *xpadn_protocv;
    } xpadn_type_u;
    struct padname_fieldinfo *xpadn_fieldinfo;
    U32 xpadn_low;
    U32 xpadn_high;
    U32 xpadn_refcnt;
    int xpadn_gen;
    U8 xpadn_len;
    U8 xpadn_flags;
};
struct padnamelist {
    SSize_t xpadnl_fill;
    PADNAME **xpadnl_alloc;
    SSize_t xpadnl_max;
    PADOFFSET xpadnl_max_named;
    U32 xpadnl_refcnt;
};
struct padlist {
    SSize_t xpadl_max;
    union {
        PAD **xpadlarr_alloc;
        void *xpadlarr_dbg;
    } xpadl_arr;
    U32 xpadl_id;
    U32 xpadl_outid;
};
enum { GOPERL_PADK_NAME = 0, GOPERL_PADK_NAMELIST = 1, GOPERL_PADK_PADLIST = 2 };
union goperl_padstruct {
    struct padname pn;
    struct padnamelist pnl;
    struct padlist pl;
};
struct goperl_padent {
    goperl_padent_t *next;        /* in its bucket */
    goperl_padent_t *active_next; /* among the ones refreshed and written back */
    uint64_t tok;                 /* the structure's address in the interpreter */
    uint8_t kind;
    uint8_t active;
    uint16_t pad0;
    uint32_t gen;       /* the generation it was refreshed at */
    uint64_t arr_g;     /* the address of its array there (lists), as loaded */
    int32_t cap;        /* the entries the arrays here have room for */
    int32_t nseen;      /* the entries loaded */
    uint64_t *arr_seen; /* the addresses in the array, as loaded */
    union goperl_padstruct seen; /* the fields, as loaded */
    union goperl_padstruct u;    /* last: what native code points at */
};
#define GOPERL_PADENT_OF(p) \
    ((goperl_padent_t *)((char *)(p) - offsetof(goperl_padent_t, u)))

enum {
    GOPERL_LAY_PN_SIZE = 0, GOPERL_LAY_PN_PV = 1, GOPERL_LAY_PN_OURSTASH = 2,
    GOPERL_LAY_PN_TYPE = 3, GOPERL_LAY_PN_FIELDINFO = 4, GOPERL_LAY_PN_LOW = 5,
    GOPERL_LAY_PN_HIGH = 6, GOPERL_LAY_PN_REFCNT = 7, GOPERL_LAY_PN_GEN = 8,
    GOPERL_LAY_PN_LEN = 9, GOPERL_LAY_PN_FLAGS = 10, GOPERL_LAY_PNL_FILL = 11,
    GOPERL_LAY_PNL_ALLOC = 12, GOPERL_LAY_PNL_MAX = 13, GOPERL_LAY_PNL_MAX_NAMED = 14,
    GOPERL_LAY_PNL_REFCNT = 15, GOPERL_LAY_PL_MAX = 16, GOPERL_LAY_PL_ARR = 17,
    GOPERL_LAY_PL_ID = 18, GOPERL_LAY_PL_OUTID = 19
};
static __attribute__((unused)) const uint32_t *goperl_pad_layout(goperl_frame_t *f) {
    goperl_compiler_state_t *c = goperl_compiler(f);
    if (!c->have_layout) {
        /* not an operation of the funnel: the shadows it refreshes after
         * every operation are read with this table */
        const uint32_t *w = (const uint32_t *)goperl_guest_mem(
            f, goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, 0, (uint64_t)(GOPERL_INT_LAYOUT), 0, 0, 0));
        if (!w) goperl_croakf(f, "the interpreter did not describe its pads");
        /* every count, index and pointer the structures hold (SSize_t,
         * PADOFFSET, a pointer) is a word of the interpreter's width */
        if (w[20] != GOPERL_GUEST_PTRSIZE || w[21] != GOPERL_GUEST_PTRSIZE ||
            w[22] != GOPERL_GUEST_PTRSIZE)
            goperl_croakf(f, "the pads of the interpreter have a layout this SDK does not read");
        memcpy(c->layout, w, sizeof(c->layout));
        c->have_layout = 1;
    }
    return c->layout;
}

static __attribute__((unused)) goperl_padent_t *goperl_padent_find(goperl_compiler_state_t *c, uint64_t tok) {
    for (goperl_padent_t *e = c->buckets[(tok >> 2) % (uint64_t)c->nbuckets]; e; e = e->next)
        if (e->tok == tok) return e;
    return 0;
}

static __attribute__((unused)) void goperl_padent_refresh(goperl_frame_t *f, goperl_padent_t *e);
static __attribute__((unused)) goperl_padent_t *goperl_padent_get(goperl_frame_t *f, uint64_t tok, int kind) {
    goperl_compiler_state_t *c = goperl_compiler(f);
    goperl_padent_t *e = goperl_padent_find(c, tok);
    if (e && e->kind != kind) {
        /* the memory of a structure that is gone is another one's now */
        free(e->arr_seen);
        if (e->kind == GOPERL_PADK_NAMELIST) free(e->u.pnl.xpadnl_alloc);
        if (e->kind == GOPERL_PADK_PADLIST) free(e->u.pl.xpadl_arr.xpadlarr_alloc);
        e->arr_seen = 0;
        e->cap = e->nseen = 0;
        e->arr_g = 0;
        memset(&e->u, 0, sizeof(e->u));
        memset(&e->seen, 0, sizeof(e->seen));
        e->kind = (uint8_t)kind;
    }
    if (!e) {
        e = (goperl_padent_t *)calloc(1, sizeof(*e));
        if (!e) goperl_croakf(f, "Out of memory!");
        e->tok = tok;
        e->kind = (uint8_t)kind;
        const uint32_t h = (uint32_t)((tok >> 2) % (uint64_t)c->nbuckets);
        e->next = c->buckets[h];
        c->buckets[h] = e;
    }
    /* read when native code first looks at it */
    if (!e->active) e->gen = c->gen - 1;
    return e;
}

static __attribute__((unused)) PADNAME *goperl_padname_shadow(goperl_frame_t *f, uint64_t tok) {
    return tok ? &goperl_padent_get(f, tok, GOPERL_PADK_NAME)->u.pn : (PADNAME *)0;
}
static __attribute__((unused)) PADNAMELIST *goperl_padnamelist_shadow(goperl_frame_t *f, uint64_t tok) {
    return tok ? &goperl_padent_get(f, tok, GOPERL_PADK_NAMELIST)->u.pnl
               : (PADNAMELIST *)0;
}
static __attribute__((unused)) PADLIST *goperl_padlist_shadow(goperl_frame_t *f, uint64_t tok) {
    return tok ? &goperl_padent_get(f, tok, GOPERL_PADK_PADLIST)->u.pl : (PADLIST *)0;
}
/* The number of a pad structure as B gives it (the number of a B::PADLIST,
 * B::PADNAMELIST or B::PADNAME object): its address in the interpreter with
 * bit 61 set and its kind (GOPERL_PADK_*) in bits 59 and 60, which no
 * address of this process has (perl-wasm scripts/wasi-configure.sh). A
 * module takes the number for the structure, by INT2PTR or by a cast of
 * the scalar pointer B::SV's typemap made of it (B::Flags), and every
 * accessor takes it to the structure's shadow (goperl_padent_fresh). */
#define GOPERL_PAD_NUMBER_TAG ((uint64_t)1 << 61)
#define GOPERL_PAD_NUMBER_KIND_SHIFT 59
#define GOPERL_PAD_NUMBER_ADDR_MASK (((uint64_t)1 << GOPERL_PAD_NUMBER_KIND_SHIFT) - 1)
static __attribute__((unused)) int goperl_is_pad_number(const void *p) {
    return ((uint64_t)(uintptr_t)p & GOPERL_PAD_NUMBER_TAG) != 0;
}
static __attribute__((unused)) void *goperl_pad_of_number(goperl_frame_t *f, const void *p) {
    const uint64_t n = (uint64_t)(uintptr_t)p;
    const int kind = (int)((n >> GOPERL_PAD_NUMBER_KIND_SHIFT) & 3);
    if (kind != GOPERL_PADK_NAME && kind != GOPERL_PADK_NAMELIST && kind != GOPERL_PADK_PADLIST)
        goperl_croakf(f, "the number of a pad structure has no kind");
    return &goperl_padent_get(f, n & GOPERL_PAD_NUMBER_ADDR_MASK, kind)->u;
}
/* The address in the interpreter of the structure a shadow (or a number
 * of B's) stands for. */
static __attribute__((unused)) uint64_t goperl_padstruct_tok(const void *p) {
    if (!p) return 0;
    if (goperl_is_pad_number(p)) return (uint64_t)(uintptr_t)p & GOPERL_PAD_NUMBER_ADDR_MASK;
    return GOPERL_PADENT_OF(p)->tok;
}

/* Room for n entries in the arrays of a list. */
static __attribute__((unused)) void goperl_padent_room(goperl_frame_t *f, goperl_padent_t *e, int32_t n, void ***arr) {
    if (n <= e->cap) return;
    int32_t ncap = e->cap ? e->cap : 8;
    while (ncap < n) ncap *= 2;
    void **na = (void **)realloc(*arr, (size_t)ncap * sizeof(void *));
    uint64_t *ns = (uint64_t *)realloc(e->arr_seen, (size_t)ncap * sizeof(uint64_t));
    if (!na || !ns) goperl_croakf(f, "Out of memory!");
    memset(na + e->cap, 0, (size_t)(ncap - e->cap) * sizeof(void *));
    memset(ns + e->cap, 0, (size_t)(ncap - e->cap) * sizeof(uint64_t));
    *arr = na;
    e->arr_seen = ns;
    e->cap = ncap;
}

/* The shadow takes what the interpreter holds now. The structures it
 * points to are shadows that are read when they are looked at. */
static __attribute__((unused)) void goperl_padent_refresh(goperl_frame_t *f, goperl_padent_t *e) {
    const uint32_t *L = goperl_pad_layout(f);
    const uint64_t t = e->tok;
    switch (e->kind) {
    case GOPERL_PADK_NAME: {
        struct padname *pn = &e->u.pn;
        const uint64_t pv = goperl_g64(f, t + L[GOPERL_LAY_PN_PV]);
        pn->xpadn_pv = pv ? (char *)(f->membase + pv) : 0;
        pn->xpadn_ourstash = (HV *)GOPERL_SV(goperl_g64(f, t + L[GOPERL_LAY_PN_OURSTASH]));
        pn->xpadn_type_u.xpadn_typestash =
            (HV *)GOPERL_SV(goperl_g64(f, t + L[GOPERL_LAY_PN_TYPE]));
        pn->xpadn_fieldinfo = (struct padname_fieldinfo *)(uintptr_t)goperl_g64(
            f, t + L[GOPERL_LAY_PN_FIELDINFO]);
        pn->xpadn_low = goperl_g32(f, t + L[GOPERL_LAY_PN_LOW]);
        pn->xpadn_high = goperl_g32(f, t + L[GOPERL_LAY_PN_HIGH]);
        pn->xpadn_refcnt = goperl_g32(f, t + L[GOPERL_LAY_PN_REFCNT]);
        pn->xpadn_gen = (int)goperl_g32(f, t + L[GOPERL_LAY_PN_GEN]);
        pn->xpadn_len = *(const U8 *)(f->membase + t + L[GOPERL_LAY_PN_LEN]);
        pn->xpadn_flags = *(const U8 *)(f->membase + t + L[GOPERL_LAY_PN_FLAGS]);
        e->seen.pn = *pn;
        break;
    }
    case GOPERL_PADK_NAMELIST: {
        struct padnamelist *pnl = &e->u.pnl;
        const int32_t fill = (int32_t)(int64_t)goperl_g64(f, t + L[GOPERL_LAY_PNL_FILL]);
        const int32_t max = (int32_t)(int64_t)goperl_g64(f, t + L[GOPERL_LAY_PNL_MAX]);
        e->arr_g = goperl_g64(f, t + L[GOPERL_LAY_PNL_ALLOC]);
        const int32_t n = (max > fill ? max : fill) + 1;
        goperl_padent_room(f, e, n, (void ***)&pnl->xpadnl_alloc);
        pnl->xpadnl_fill = fill;
        pnl->xpadnl_max = max;
        pnl->xpadnl_max_named = (PADOFFSET)goperl_g64(f, t + L[GOPERL_LAY_PNL_MAX_NAMED]);
        pnl->xpadnl_refcnt = goperl_g32(f, t + L[GOPERL_LAY_PNL_REFCNT]);
        /* the entries in use; the ones above are NULL until stored into */
        for (int32_t i = 0; i < e->cap; i++) {
            const uint64_t pt = (i <= fill && e->arr_g)
                                    ? goperl_g64(f, e->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i)
                                    : 0;
            e->arr_seen[i] = pt;
            pnl->xpadnl_alloc[i] = goperl_padname_shadow(f, pt);
        }
        e->nseen = n;
        e->seen.pnl = *pnl;
        break;
    }
    case GOPERL_PADK_PADLIST: {
        struct padlist *pl = &e->u.pl;
        const int32_t max = (int32_t)(int64_t)goperl_g64(f, t + L[GOPERL_LAY_PL_MAX]);
        e->arr_g = goperl_g64(f, t + L[GOPERL_LAY_PL_ARR]);
        goperl_padent_room(f, e, max + 1, (void ***)&pl->xpadl_arr.xpadlarr_alloc);
        pl->xpadl_max = max;
        pl->xpadl_id = goperl_g32(f, t + L[GOPERL_LAY_PL_ID]);
        pl->xpadl_outid = goperl_g32(f, t + L[GOPERL_LAY_PL_OUTID]);
        for (int32_t i = 0; i < e->cap; i++) {
            const uint64_t pt = (i <= max && e->arr_g)
                                    ? goperl_g64(f, e->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i)
                                    : 0;
            e->arr_seen[i] = pt;
            /* entry 0 is the list of names (PadlistNAMES), the others pads */
            pl->xpadl_arr.xpadlarr_alloc[i] =
                i == 0 ? (PAD *)(void *)goperl_padnamelist_shadow(f, pt) : (PAD *)GOPERL_SV(pt);
        }
        e->nseen = max + 1;
        e->seen.pl = *pl;
        break;
    }
    }
}

/* What native code stored into a shadow goes to the interpreter. */
static __attribute__((unused)) void goperl_padent_store(goperl_frame_t *f, goperl_padent_t *e) {
    const uint32_t *L = goperl_pad_layout(f);
    const uint64_t t = e->tok;
    switch (e->kind) {
    case GOPERL_PADK_NAME: {
        struct padname *pn = &e->u.pn, *s = &e->seen.pn;
        if (pn->xpadn_pv != s->xpadn_pv) {
            const uint64_t g = goperl_guest_addr(f, pn->xpadn_pv);
            if (pn->xpadn_pv && !g)
                goperl_croakf(f, "a pad name was given a name that is not in the "
                                 "interpreter's memory");
            goperl_g64_set(f, t + L[GOPERL_LAY_PN_PV], g);
        }
        if (pn->xpadn_ourstash != s->xpadn_ourstash)
            goperl_g64_set(f, t + L[GOPERL_LAY_PN_OURSTASH], GOPERL_TOK(pn->xpadn_ourstash));
        if (pn->xpadn_type_u.xpadn_typestash != s->xpadn_type_u.xpadn_typestash)
            goperl_g64_set(f, t + L[GOPERL_LAY_PN_TYPE],
                             GOPERL_TOK(pn->xpadn_type_u.xpadn_typestash));
        if (pn->xpadn_fieldinfo != s->xpadn_fieldinfo)
            goperl_g64_set(f, t + L[GOPERL_LAY_PN_FIELDINFO],
                             (uint64_t)(uintptr_t)pn->xpadn_fieldinfo);
        if (pn->xpadn_low != s->xpadn_low)
            goperl_g32_set(f, t + L[GOPERL_LAY_PN_LOW], pn->xpadn_low);
        if (pn->xpadn_high != s->xpadn_high)
            goperl_g32_set(f, t + L[GOPERL_LAY_PN_HIGH], pn->xpadn_high);
        if (pn->xpadn_refcnt != s->xpadn_refcnt)
            goperl_g32_set(f, t + L[GOPERL_LAY_PN_REFCNT], pn->xpadn_refcnt);
        if (pn->xpadn_gen != s->xpadn_gen)
            goperl_g32_set(f, t + L[GOPERL_LAY_PN_GEN], (uint32_t)pn->xpadn_gen);
        if (pn->xpadn_len != s->xpadn_len)
            *(U8 *)(f->membase + t + L[GOPERL_LAY_PN_LEN]) = pn->xpadn_len;
        if (pn->xpadn_flags != s->xpadn_flags)
            *(U8 *)(f->membase + t + L[GOPERL_LAY_PN_FLAGS]) = pn->xpadn_flags;
        *s = *pn;
        break;
    }
    case GOPERL_PADK_NAMELIST: {
        struct padnamelist *pnl = &e->u.pnl, *s = &e->seen.pnl;
        if (pnl->xpadnl_alloc != s->xpadnl_alloc || pnl->xpadnl_max != s->xpadnl_max) {
            *s = *pnl; /* told once */
            goperl_croakf(f, "the array of a list of pad names was replaced; "
                             "padnamelist_store makes room in it");
        }
        if (pnl->xpadnl_fill != s->xpadnl_fill)
            goperl_g64_set(f, t + L[GOPERL_LAY_PNL_FILL], (uint64_t)(int64_t)pnl->xpadnl_fill);
        if (pnl->xpadnl_max_named != s->xpadnl_max_named)
            goperl_g64_set(f, t + L[GOPERL_LAY_PNL_MAX_NAMED], (uint64_t)pnl->xpadnl_max_named);
        if (pnl->xpadnl_refcnt != s->xpadnl_refcnt)
            goperl_g32_set(f, t + L[GOPERL_LAY_PNL_REFCNT], pnl->xpadnl_refcnt);
        for (int32_t i = 0; i < e->nseen && i <= pnl->xpadnl_max; i++) {
            const uint64_t now = goperl_padstruct_tok(pnl->xpadnl_alloc[i]);
            if (now == e->arr_seen[i]) continue;
            goperl_g64_set(f, e->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i, now);
            e->arr_seen[i] = now;
        }
        *s = *pnl;
        break;
    }
    case GOPERL_PADK_PADLIST: {
        struct padlist *pl = &e->u.pl, *s = &e->seen.pl;
        if (pl->xpadl_arr.xpadlarr_alloc != s->xpadl_arr.xpadlarr_alloc ||
            pl->xpadl_max != s->xpadl_max) {
            *s = *pl; /* told once */
            goperl_croakf(f, "the array of a padlist was replaced");
        }
        if (pl->xpadl_id != s->xpadl_id) goperl_g32_set(f, t + L[GOPERL_LAY_PL_ID], pl->xpadl_id);
        if (pl->xpadl_outid != s->xpadl_outid)
            goperl_g32_set(f, t + L[GOPERL_LAY_PL_OUTID], pl->xpadl_outid);
        for (int32_t i = 0; i < e->nseen && i <= pl->xpadl_max; i++) {
            void *const p = (void *)pl->xpadl_arr.xpadlarr_alloc[i];
            const uint64_t now = i == 0 ? goperl_padstruct_tok(p) : GOPERL_TOK(p);
            if (now == e->arr_seen[i]) continue;
            goperl_g64_set(f, e->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i, now);
            e->arr_seen[i] = now;
        }
        *s = *pl;
        break;
    }
    }
}

/* Before an operation of the interpreter (goperl_flush). */
static __attribute__((unused)) void goperl_pads_flush(goperl_frame_t *f) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_compiler_state_t *c = sh ? (goperl_compiler_state_t *)sh->compiler : 0;
    if (!c) return;
    for (goperl_padent_t *e = c->active; e; e = e->active_next) goperl_padent_store(f, e);
}
/* After an operation of the interpreter, and when native code is entered:
 * every shadow takes what the interpreter holds when it is looked at next
 * (goperl_padent_fresh). Only then: the interpreter may have freed the
 * structure of a shadow that native code no longer looks at (a sub whose
 * pads it read is gone), whose memory holds something else now. */
static __attribute__((unused)) void goperl_pads_refresh(goperl_frame_t *f) {
    goperl_shared_t *const sh = f ? (goperl_shared_t *)f->inst : 0;
    goperl_compiler_state_t *c = sh ? (goperl_compiler_state_t *)sh->compiler : 0;
    if (c) c->gen++;
}
/* A shadow native code looks at, up to date. Every accessor of a pad
 * structure comes here, so a shadow native code stores into is one it has
 * looked at: from then on it is written back before every operation (a
 * list's entries have shadows as soon as the list is read, and are not
 * looked at before every operation for that). */
static __attribute__((unused)) void *goperl_padent_fresh(goperl_frame_t *f, const void *p) {
    if (!p) return (void *)(uintptr_t)p;
    if (goperl_is_pad_number(p)) p = goperl_pad_of_number(f, p);
    goperl_padent_t *const e = GOPERL_PADENT_OF(p);
    goperl_compiler_state_t *const c = goperl_compiler(f);
    if (e->gen != c->gen) {
        e->gen = c->gen;
        goperl_padent_refresh(f, e);
    }
    if (!e->active) {
        e->active = 1;
        e->active_next = c->active;
        c->active = e;
    }
    return (void *)(uintptr_t)p;
}
#define GOPERL_PADFRESH(p) ((__typeof__(&*(p)))goperl_padent_fresh(_gof, (const void *)(p)))
/* The outermost native call has returned: no shadow is refreshed or
 * written back until it is reached again. */
static __attribute__((unused)) void goperl_pads_rest(goperl_frame_t *f) {
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_compiler_state_t *c = sh ? (goperl_compiler_state_t *)sh->compiler : 0;
    if (!c) return;
    for (goperl_padent_t *e = c->active, *next; e; e = next) {
        next = e->active_next;
        e->active = 0;
        e->active_next = 0;
    }
    c->active = 0;
}

#define PadlistARRAY(pl) (GOPERL_PADFRESH(pl)->xpadl_arr.xpadlarr_alloc)
#define PadlistMAX(pl) (GOPERL_PADFRESH(pl)->xpadl_max)
#define PadlistNAMES(pl) (*((PADNAMELIST **)PadlistARRAY(pl)))
#define PadlistNAMESARRAY(pl) PadnamelistARRAY(PadlistNAMES(pl))
#define PadlistNAMESMAX(pl) PadnamelistMAX(PadlistNAMES(pl))
#define PadlistREFCNT(pl) 1
#define PadnamelistARRAY(pnl) (GOPERL_PADFRESH(pnl)->xpadnl_alloc)
#define PadnamelistMAX(pnl) (GOPERL_PADFRESH(pnl)->xpadnl_fill)
#define PadnamelistMAXNAMED(pnl) (GOPERL_PADFRESH(pnl)->xpadnl_max_named)
#define PadnamelistREFCNT(pnl) (GOPERL_PADFRESH(pnl)->xpadnl_refcnt)
#define padnamelist_fetch(pnl, key)                                  \
    ({ PADNAMELIST *const goperl_pnl_ = (pnl);                       \
       const SSize_t goperl_key_ = (SSize_t)(key);                   \
       goperl_key_ > PadnamelistMAX(goperl_pnl_)                     \
           ? (PADNAME *)0                                            \
           : PadnamelistARRAY(goperl_pnl_)[goperl_key_]; })
#define PadnamePV(pn) (GOPERL_PADFRESH(pn)->xpadn_pv)
#define PadnameLEN(pn) (GOPERL_PADFRESH(pn)->xpadn_len)
#define PadnameUTF8(pn) 1
#define PadnameSV(pn) \
    newSVpvn_flags(PadnamePV(pn), PadnameLEN(pn), SVs_TEMP | SVf_UTF8)
#define PadnameFLAGS(pn) (GOPERL_PADFRESH(pn)->xpadn_flags)
#define PadnameIsOUR(pn) cBOOL(PadnameOURSTASH(pn))
#define PadnameOURSTASH(pn) (GOPERL_PADFRESH(pn)->xpadn_ourstash)
#define PadnameTYPE(pn) (GOPERL_PADFRESH(pn)->xpadn_type_u.xpadn_typestash)
#define PadnameHasTYPE(pn) cBOOL(PadnameTYPE(pn))
#define PadnamePROTOCV(pn) (GOPERL_PADFRESH(pn)->xpadn_type_u.xpadn_protocv)
#define PadnameREFCNT(pn) (GOPERL_PADFRESH(pn)->xpadn_refcnt)
#define PadnameREFCNT_inc(pn) \
    ({ PADNAME *const goperl_pn_inc_ = (pn); PadnameREFCNT(goperl_pn_inc_)++; goperl_pn_inc_; })
#define PadnameREFCNT_dec(pn)                                              \
    ((void)GOPERL_INTERNAL(_gof, GOPERL_INT_PADNAME_FREE,                  \
                           goperl_padstruct_tok(pn), 0))
#define PadnameIsNULL(pn) (!(pn))
#define PadnameOUTER(pn) (PadnameFLAGS(pn) & PADNAMEf_OUTER)
#define PadnameIsSTATE(pn) (PadnameFLAGS(pn) & PADNAMEf_STATE)
#define PadnameLVALUE(pn) (PadnameFLAGS(pn) & PADNAMEf_LVALUE)
#define PadnameIsFIELD(pn) (PadnameFLAGS(pn) & PADNAMEf_FIELD)
#define PadnameFIELDINFO(pn) (GOPERL_PADFRESH(pn)->xpadn_fieldinfo)
#define PadnameLVALUE_on(pn) (PadnameFLAGS(pn) |= PADNAMEf_LVALUE)
#define PadnameIsSTATE_on(pn) (PadnameFLAGS(pn) |= PADNAMEf_STATE)
#define PadnameOUTER_on(pn) (PadnameFLAGS(pn) |= PADNAMEf_OUTER)
#define PadnameOUTER_off(pn) (PadnameFLAGS(pn) &= ~PADNAMEf_OUTER)
#define PadnameOURSTASH_set(pn, s) (PadnameOURSTASH(pn) = (s))
#define PadnameTYPE_set(pn, s) (PadnameTYPE(pn) = (s))
/* pad.h's names from before pad names were structures of their own (a
 * module that does not find PadnameIsOUR_on falls back to SvPAD_OUR_on,
 * and to SvFLAGS when that is missing too: Lexical::Var) */
#define SvPAD_STATE PadnameIsSTATE
#define SvPAD_STATE_on PadnameIsSTATE_on
#define SvPAD_TYPED_on(pn) (PadnameFLAGS(pn) |= PADNAMEf_TYPED)
#define SvPAD_OUR_on(pn) (PadnameFLAGS(pn) |= PADNAMEf_OUR)
#define SvOURSTASH PadnameOURSTASH
#define SvOURSTASH_set PadnameOURSTASH_set
#define PADNAMEt_OUTER PADNAMEf_OUTER
#define PADNAMEt_STATE PADNAMEf_STATE
#define PADNAMEt_LVALUE PADNAMEf_LVALUE
#define PADNAMEt_TYPED PADNAMEf_TYPED
#define PADNAMEt_OUR PADNAMEf_OUR
#define COP_SEQ_RANGE_LOW(pn) (GOPERL_PADFRESH(pn)->xpadn_low)
#define COP_SEQ_RANGE_HIGH(pn) (GOPERL_PADFRESH(pn)->xpadn_high)
#define PARENT_PAD_INDEX(pn) (GOPERL_PADFRESH(pn)->xpadn_low)
#define PARENT_FAKELEX_FLAGS(pn) (GOPERL_PADFRESH(pn)->xpadn_high)
#define PARENT_PAD_INDEX_set(pn, val) (PARENT_PAD_INDEX(pn) = (val))
#define PARENT_FAKELEX_FLAGS_set(pn, val) (PARENT_FAKELEX_FLAGS(pn) = (val))
#define PERL_PADSEQ_INTRO U32_MAX
#define PadnameIN_SCOPE(pn, seq)                                           \
    ({ const PADNAME *const goperl_pnis_ = (pn); const U32 goperl_seq_ = (U32)(seq); \
       COP_SEQ_RANGE_LOW(goperl_pnis_) != PERL_PADSEQ_INTRO &&             \
           COP_SEQ_RANGE_LOW(goperl_pnis_) < goperl_seq_ &&                \
           goperl_seq_ <= COP_SEQ_RANGE_HIGH(goperl_pnis_); })
/* A new pad name, the interpreter's: its own copy of the name. */
static __attribute__((unused)) PADNAME *goperl_newpadname_pvn(goperl_frame_t *f, const char *s, STRLEN len) {
    /* read now: a module sets its fields straight away (xpadn_low) */
    return (PADNAME *)goperl_padent_fresh(
        f, goperl_padname_shadow(
               f, GOPERL_INTERNAL_S(f, GOPERL_INT_PADNAME_NEW, 0, (uint64_t)len, s, len)));
}
#define newPADNAMEpvn(s, len) goperl_newpadname_pvn(_gof, (const char *)(s), (STRLEN)(len))
/* The list takes the name over, and makes room for it. */
static __attribute__((unused)) PADNAME **goperl_padnamelist_store(goperl_frame_t *f,
                                                                  PADNAMELIST *pnl, SSize_t key,
                                                                  PADNAME *val) {
    const goperl_gword_t pn = goperl_padstruct_tok(val);
    goperl_do_op3(f, GOPERL_OP_PAD, 11 /* GOPERL_PAD_NAMELIST_STORE */, goperl_padstruct_tok(pnl),
                  (uint64_t)(int64_t)key, (const char *)&pn, sizeof pn);
    return &PadnamelistARRAY(pnl)[key];
}
#define padnamelist_store(pnl, key, val) \
    goperl_padnamelist_store(_gof, (pnl), (SSize_t)(key), (val))
/* A new list of pad names, the interpreter's, with room for max + 1. */
#define newPADNAMELIST(max)                                                 \
    goperl_padnamelist_shadow(                                              \
        _gof, GOPERL_INTERNAL(_gof, GOPERL_INT_PADNAMELIST_NEW, 0, (uint64_t)(size_t)(max)))
#define PadnamelistREFCNT_dec(pnl)                                          \
    ((void)GOPERL_INTERNAL(_gof, GOPERL_INT_PADNAMELIST_FREE,               \
                           goperl_padstruct_tok(pnl), 0))
#define padnamelist_free(pnl) PadnamelistREFCNT_dec(pnl)
/* A pad name for a closure's copy of the outer one (newPADNAMEouter). */
#define newPADNAMEouter(outer)                                              \
    ((PADNAME *)goperl_padent_fresh(                                        \
        _gof, goperl_padname_shadow(                                        \
                  _gof, GOPERL_INTERNAL(_gof, GOPERL_INT_PADNAME_OUTER,     \
                                        goperl_padstruct_tok(outer), 0))))
/* pad_new: the flags of pad.h. */
#define padnew_CLONE 1
#define padnew_SAVE 2
#define padnew_SAVESUB 4
#define pad_new(flags) \
    goperl_padlist_shadow(_gof, GOPERL_INTERNAL(_gof, GOPERL_INT_PAD_NEW, 0, (int)(flags)))
#define Perl_pad_new(ctx, flags) pad_new(flags)
#define pad_findmy_pv(name, flags)                                         \
    ({ const char *const goperl_pfp_name_ = (name);                        \
       pad_findmy_pvn(goperl_pfp_name_, __builtin_strlen(goperl_pfp_name_), (flags)); })
#define pad_findmy_pvs(name, flags) pad_findmy_pvn("" name "", sizeof(name) - 1, (flags))

/* The list of names of the pad in use; an lvalue. */
static __attribute__((unused)) PADNAMELIST **goperl_comppad_name_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_COMPPAD_NAME];
    if (!s->live || s->val == s->shadow) {
        const uint64_t tok = goperl_do_op(f, GOPERL_OP_PLVAR_GET,
                                          (uint64_t)GOPERL_PL_COMPPAD_NAME, 0, 0, 0);
        s->val = s->shadow = (uint64_t)(uintptr_t)goperl_padnamelist_shadow(f, tok);
        goperl_plshadow_set_live(s);
    }
    return (PADNAMELIST **)&s->val;
}
#undef PL_comppad_name
#define PL_comppad_name (*goperl_comppad_name_ref(_gof))

/* PL_curpad: the values of the pad in use, as AvARRAY has them (stores
 * through it go to the pad); an lvalue, which a module points at the values
 * of the pad it makes the one in use (PL_curpad = AvARRAY(PL_comppad)), or
 * at values in its own memory, which the interpreter cannot address: those
 * stand in for the values of the pad in use as long as that pad is the one
 * that was in use when they were stored, and they are not the
 * interpreter's. A sub the interpreter runs meanwhile has its own pad in
 * use, with its own values, as with perl, where entersub makes that pad's
 * values PL_curpad and puts the module's back when the sub returns. */
static __attribute__((unused)) SV ***goperl_curpad_ref(goperl_frame_t *f, const void *site, uintptr_t frame) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_CURPAD];
    if (s->live && s->val != s->shadow) return (SV ***)&s->val; /* a store to write back */
    PAD *const pad = PL_comppad;
    goperl_compiler_state_t *const c =
        ((goperl_shared_t *)f->inst)->compiler ? goperl_compiler(f) : 0;
    if (c && c->curpad_own && c->curpad_own_pad == GOPERL_TOK(pad)) return &c->curpad_own;
    SV **const body =
        pad ? (SV **)(uintptr_t)*goperl_lv_ref(f, (SV *)pad, GOPERL_LV_AVBODY, site, frame)
            : (SV **)0;
    s->val = s->shadow = (uint64_t)(uintptr_t)body;
    goperl_plshadow_set_live(s);
    return (SV ***)&s->val;
}
#undef PL_curpad
#define PL_curpad (*goperl_curpad_ref(_gof, GOPERL_LV_AT))
#undef PAD_SVl
#define PAD_SVl(ix) (PL_curpad[ix])
/* the entry perl's save_padsv_and_mortalize pushes: the value (a reference
 * more), the pad, the index and the type */
static __attribute__((unused)) void goperl_save_padsv_and_mortalize(PADOFFSET off) {
    SV *const sv = PL_curpad[off];
    PAD *const pad = PL_comppad;
    SvREFCNT_inc_simple_void_NN(sv);
    dSS_ADD;
    SS_ADD_PTR(sv);
    SS_ADD_PTR(pad);
    SS_ADD_UV((UV)off);
    SS_ADD_UV(SAVEt_PADSV_AND_MORTALIZE);
    SS_ADD_END(4);
}
#define save_padsv_and_mortalize(off) goperl_save_padsv_and_mortalize((PADOFFSET)(off))
#define PadARRAY(pad) AvARRAY(pad)
#define PadMAX(pad) AvFILLp(pad)
/* What a stored value of PL_comppad_name or PL_curpad is to the
 * interpreter (see goperl_plshadow_flush): the list a shadow stands for,
 * the address of values in its memory (a pad's body, as AvARRAY gives it).
 * Values of PL_curpad in native code's own memory are not the
 * interpreter's: they are kept here, for the pad in use
 * (goperl_curpad_ref), and nothing is written back (0). */
static __attribute__((unused)) int goperl_curpad_guest(goperl_frame_t *f, uint64_t val, uint64_t *out) {
    *out = goperl_guest_addr(f, (const void *)(uintptr_t)val);
    return !val || *out;
}
static uint64_t goperl_runops_out(goperl_frame_t *f, uint64_t val);
static uint64_t goperl_sighandlerp_out(goperl_frame_t *f, uint64_t val);
static __attribute__((unused)) int goperl_plshadow_out(goperl_frame_t *f, int id, goperl_plshadow_t *s,
                                                       uint64_t *out) {
    if (id == GOPERL_PL_COMPPAD_NAME) {
        *out = goperl_padstruct_tok((const void *)(uintptr_t)s->val);
        return 1;
    }
    if (id == 74 /* GOPERL_PL_RESTARTOP: a shadow, its op */) {
        const OP *const o = (const OP *)(uintptr_t)s->val;
        if (o && !o->gop) goperl_croakf(f, "PL_restartop was given an op that is not the interpreter's");
        *out = o ? o->gop : 0;
        return 1;
    }
    if (id == 99 /* GOPERL_PL_RUNOPS */) {
        *out = goperl_runops_out(f, s->val);
        return 1;
    }
    if (id == 115 /* GOPERL_PL_SIGHANDLERP */) {
        *out = goperl_sighandlerp_out(f, s->val);
        return 1;
    }
    if (id == 103 /* GOPERL_PL_OP_MASK */) {
        /* a mask is the interpreter's bytes */
        *out = goperl_guest_addr(f, (const void *)(uintptr_t)s->val);
        if (s->val && !*out)
            goperl_croakf(f, "PL_op_mask points outside the interpreter's memory");
        return 1;
    }
    if (id == GOPERL_PL_MULTIDEREF_PC) {
        /* the items of an op's auxiliary data are the interpreter's */
        *out = goperl_guest_addr(f, (const void *)(uintptr_t)s->val);
        if (s->val && !*out)
            goperl_croakf(f, "PL_multideref_pc points outside the interpreter's memory");
        return 1;
    }
    if (id == GOPERL_PL_CURPAD) {
        goperl_compiler_state_t *const c = goperl_compiler(f);
        if (goperl_curpad_guest(f, s->val, out)) {
            c->curpad_own = 0;
            return 1;
        }
        c->curpad_own = (SV **)(uintptr_t)s->val;
        c->curpad_own_pad = goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_COMPPAD, 0, 0, 0);
        s->val = s->shadow;
        return 0;
    }
    *out = s->val;
    return 1;
}
/* Values of the pad in use stored where values of native code's own stood
 * in for them (PL_curpad = AvARRAY(pad) through that lvalue) are the
 * interpreter's again. */
static __attribute__((unused)) void goperl_curpad_own_flush(goperl_frame_t *f) {
    goperl_compiler_state_t *const c = (goperl_compiler_state_t *)((goperl_shared_t *)f->inst)->compiler;
    uint64_t g = 0;
    if (!c || !c->curpad_own || !goperl_curpad_guest(f, (uint64_t)(uintptr_t)c->curpad_own, &g)) return;
    c->curpad_own = 0;
    goperl_xs_op(f, GOPERL_OP_PLVAR_SET, GOPERL_PL_CURPAD, g, 0, 0);
    goperl_plshadow_v[GOPERL_PL_CURPAD].live = 0;
}
/* The pad of a padlist at a depth becomes the one in use, saved for the
 * interpreter to put back (PAD_SET_CUR) or not (PAD_SET_CUR_NOSAVE). */
#define PAD_SET_CUR_NOSAVE(padlist, nth) \
    (PL_comppad = PadlistARRAY(padlist)[nth])
#define PAD_SET_CUR(padlist, nth)                       \
    STMT_START {                                        \
        SAVESPTR(PL_comppad);                           \
        SAVESPTR(PL_curpad);                            \
        PAD_SET_CUR_NOSAVE((padlist), (nth));           \
    }                                                   \
    STMT_END
#define SAVECOMPPAD()                                   \
    STMT_START {                                        \
        SAVESPTR(PL_comppad);                           \
        SAVESPTR(PL_curpad);                            \
    }                                                   \
    STMT_END

/* save_clearsv is given the address of a value of the pad in use. */
static __attribute__((unused)) void goperl_save_clearsv(goperl_frame_t *f, SV **svp) __attribute__((unused));
static __attribute__((unused)) void goperl_save_clearsv(goperl_frame_t *f, SV **svp) {
    SV **const base = AvARRAY(PL_comppad);
    if (!base || svp < base || svp > base + AvFILLp(PL_comppad))
        goperl_croakf(f, "save_clearsv of a value that is not of the pad in use");
    goperl_do_op(f, GOPERL_OP_PAD, 10 /* GOPERL_PAD_SAVE_CLEARSV */, (uint64_t)(svp - base), 0, 0);
}
#undef save_clearsv
#define save_clearsv(svp) goperl_save_clearsv(_gof, (SV **)(svp))

/* ---- the fields of a sub ------------------------------------------------------
 * Lvalues (GOPERL_LV_): read from the interpreter, written back as they
 * are. The ops of a sub are op shadows here, its padlist a padlist's
 * shadow, its file a name as the file of a COP is (interned); a name that
 * is not in the interpreter's memory is copied there when it is stored, and
 * the sub owns the copy (CvDYNFILE). */
static __attribute__((unused)) uint64_t goperl_cvfield_load(goperl_frame_t *f, uint64_t tok, int field) {
    switch (field) {
    case GOPERL_LV_CVFLAGS:
        return GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_FLAGS);
    case GOPERL_LV_CVROOT:
        return (uint64_t)(uintptr_t)goperl_op_shadow_full(
            f, GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_ROOT));
    case GOPERL_LV_CVSTART:
        return (uint64_t)(uintptr_t)goperl_op_shadow_full(
            f, GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_START));
    case GOPERL_LV_CVOUTSIDE_SEQ:
        return GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_OUTSIDE_SEQ);
    case GOPERL_LV_CVPADLIST:
        /* a shadow native code looks at, up to date and written back: the
         * padlist's own fields are stored into through it
         * (`CvPADLIST(cv)->xpadl_id = ...`, Future::AsyncAwait's copy of a
         * suspended sub), with no accessor to tell the store */
        return (uint64_t)(uintptr_t)goperl_padent_fresh(
            f, goperl_padlist_shadow(f, GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok,
                                                        GOPERL_CVF_PADLIST)));
    case GOPERL_LV_CVDEPTH:
        return (uint64_t)(int64_t)(I32)GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_DEPTH);
    case GOPERL_LV_OBJMAXFIELD:
        /* OBJ_FIELDS' second value: the number of fields */
        return (uint64_t)((int64_t)goperl_op_second(f, GOPERL_OP_INTERNAL, tok, 0,
                                                    GOPERL_INT_OBJ_FIELDS) - 1);
    case GOPERL_LV_CVFILE: {
        /* interned, as the file of a COP is: the two compare as they do on
         * perl, where a sub's file is the file of the COP it was made at */
        const uint64_t g = GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, tok, GOPERL_CVF_FILE);
        if (!g) return 0;
        const char *const p = (const char *)(f->membase + g);
        return (uint64_t)(uintptr_t)goperl_intern(p, __builtin_strlen(p));
    }
    }
    goperl_croakf(f, "no such field of a sub (%d)", field);
}
/* An op stored into a field of the interpreter's structures is one of its
 * ops (an op shadow); 0 for NULL. */
static __attribute__((unused)) uint64_t goperl_op_field_tok(goperl_frame_t *f, uint64_t val) {
    const OP *const o = (const OP *)(uintptr_t)val;
    if (!o) return 0;
    if (!o->gop)
        goperl_croakf(f, "an op of native code's own making was stored into a "
                         "field of the interpreter's");
    return o->gop;
}
static __attribute__((unused)) uint64_t goperl_cv_file_tok(goperl_frame_t *f, uint64_t cvtok, uint64_t val) {
    const char *const p = (const char *)(uintptr_t)val;
    if (!p) return 0;
    const uint64_t g = goperl_guest_addr(f, p);
    if (g) return g;
    /* a name of this process's (a module's own copy, or the interned name
     * the field read as) */
    const size_t len = __builtin_strlen(p);
    const uint64_t copy = goperl_do_op(f, GOPERL_OP_MEM_REALLOC, 0, (uint64_t)len + 1, 0, 0);
    if (!copy) goperl_croakf(f, "Out of memory!");
    memcpy(f->membase + copy, p, len + 1);
    const uint32_t flags = (uint32_t)GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, cvtok, GOPERL_CVF_FLAGS);
    GOPERL_INTERNAL_HI(f, GOPERL_INT_CV_SET, cvtok, flags | CVf_DYNFILE, GOPERL_CVF_FLAGS);
    return copy;
}
static __attribute__((unused)) void goperl_cvfield_store(goperl_frame_t *f, uint64_t tok, int field, uint64_t val) {
    int sel = -1;
    uint64_t v = val;
    switch (field) {
    case GOPERL_LV_CVFLAGS: sel = GOPERL_CVF_FLAGS; break;
    case GOPERL_LV_CVROOT: sel = GOPERL_CVF_ROOT; v = goperl_op_field_tok(f, val); break;
    case GOPERL_LV_CVSTART: sel = GOPERL_CVF_START; v = goperl_op_field_tok(f, val); break;
    case GOPERL_LV_CVOUTSIDE_SEQ: sel = GOPERL_CVF_OUTSIDE_SEQ; break;
    case GOPERL_LV_CVPADLIST:
        sel = GOPERL_CVF_PADLIST;
        v = goperl_padstruct_tok((const void *)(uintptr_t)val);
        break;
    case GOPERL_LV_CVDEPTH: sel = GOPERL_CVF_DEPTH; break;
    case GOPERL_LV_CVFILE: sel = GOPERL_CVF_FILE; v = goperl_cv_file_tok(f, tok, val); break;
    case GOPERL_LV_OBJMAXFIELD:
        GOPERL_INTERNAL(f, GOPERL_INT_OBJ_MAXFIELD_SET, tok, val);
        return;
    }
    if (sel < 0) goperl_croakf(f, "no such field of a sub (%d)", field);
    /* CV_SET: b = the value, the field in c's high half */
    GOPERL_INTERNAL_HI(f, GOPERL_INT_CV_SET, tok, v, sel);
}
#define GOPERL_CV_LV(cv, field, type) \
    (*(type *)goperl_lv_ref(_gof, (SV *)(cv), (field), GOPERL_LV_AT))
#undef CvFLAGS
#define CvFLAGS(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVFLAGS, U32)
#undef CvROOT
#define CvROOT(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVROOT, OP *)
#undef CvSTART
#define CvSTART(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVSTART, OP *)
#define CvOUTSIDE_SEQ(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVOUTSIDE_SEQ, U32)
#undef CvPADLIST
#define CvPADLIST(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVPADLIST, PADLIST *)
#define CvPADLIST_set(cv, padlist) (CvPADLIST(cv) = (padlist))
#undef CvDEPTH
#define CvDEPTH(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVDEPTH, I32)
#undef CvFILE
#define CvFILE(cv) GOPERL_CV_LV(cv, GOPERL_LV_CVFILE, char *)
#define CvNAMED_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_NAMED, 0)
#define CvNAMED_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_NAMED)
#define CvLEXICAL_on(cv) GOPERL_CVFLAGS_CHANGE(cv, CVf_LEXICAL, 0)
#define CvLEXICAL_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_LEXICAL)
/* the sub of a string eval is unique and not fake (cv.h); unique and fake
 * is a special sub (CvSPECIAL: BEGIN, the main program's) */
#define CvEVAL(cv) (CvUNIQUE(cv) && !SvFAKE(cv))
#define CvEVAL_on(cv) (GOPERL_CVFLAGS_CHANGE(cv, CVf_UNIQUE, 0), SvFAKE_off(cv))
#define CvEVAL_off(cv) GOPERL_CVFLAGS_CHANGE(cv, 0, CVf_UNIQUE)

/* The body of a sub (SvANY of a CV), with perl's fields: loaded at every
 * evaluation, what native code stores into it written back as the fields
 * above are; the name of a sub that has one of its own (CvNAMED) is given
 * to the interpreter, which shares it. */
typedef union {
    OP *xcv_start;
    ANY xcv_xsubany;
} goperl_xcv_start_u;
typedef struct xpvcv {
    HV *xcv_stash;
    goperl_xcv_start_u xcv_start_u;
    union {
        OP *xcv_root;
        void (*xcv_xsub)(pTHX_ CV *);
    } xcv_root_u;
    union {
        GV *xcv_gv;
        HEK *xcv_hek;
    } xcv_gv_u;
    char *xcv_file;
    union {
        PADLIST *xcv_padlist;
        void *xcv_hscxt;
    } xcv_padlist_u;
    CV *xcv_outside;
    U32 xcv_outside_seq;
    U32 xcv_flags;
    I32 xcv_depth;
} XPVCV;
typedef struct goperl_cvbody {
    XPVCV body;
    XPVCV seen;
} goperl_cvbody_t;
static __attribute__((unused)) void goperl_cvbody_load(goperl_frame_t *f, goperl_lvslot_t *m) {
    goperl_cvbody_t *b = (goperl_cvbody_t *)(uintptr_t)m->val;
    if (!b) {
        b = (goperl_cvbody_t *)calloc(1, sizeof(*b));
        if (!b) goperl_croakf(f, "Out of memory!");
        m->val = m->shadow = (uint64_t)(uintptr_t)b;
    }
    const uint64_t t = m->svtok;
    XPVCV *x = &b->body;
    x->xcv_stash = (HV *)GOPERL_SV(GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, t, GOPERL_CVF_STASH));
    x->xcv_start_u.xcv_start = (OP *)(uintptr_t)goperl_cvfield_load(f, t, GOPERL_LV_CVSTART);
    x->xcv_root_u.xcv_root = (OP *)(uintptr_t)goperl_cvfield_load(f, t, GOPERL_LV_CVROOT);
    x->xcv_flags = (U32)goperl_cvfield_load(f, t, GOPERL_LV_CVFLAGS);
    x->xcv_gv_u.xcv_gv = (GV *)GOPERL_SV(GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, t, GOPERL_CVF_GV_U));
    if (x->xcv_flags & CVf_NAMED) x->xcv_gv_u.xcv_hek = CvNAME_HEK((CV *)GOPERL_SV(t));
    x->xcv_file = (char *)(uintptr_t)goperl_cvfield_load(f, t, GOPERL_LV_CVFILE);
    x->xcv_padlist_u.xcv_padlist =
        (PADLIST *)(uintptr_t)goperl_cvfield_load(f, t, GOPERL_LV_CVPADLIST);
    x->xcv_outside = (CV *)GOPERL_SV(GOPERL_INTERNAL(f, GOPERL_INT_CV_GET, t, GOPERL_CVF_OUTSIDE));
    x->xcv_outside_seq = (U32)goperl_cvfield_load(f, t, GOPERL_LV_CVOUTSIDE_SEQ);
    x->xcv_depth = (I32)goperl_cvfield_load(f, t, GOPERL_LV_CVDEPTH);
    b->seen = *x;
}
static __attribute__((unused)) void goperl_cvbody_store(goperl_frame_t *f, goperl_lvslot_t *m) {
    goperl_cvbody_t *b = (goperl_cvbody_t *)(uintptr_t)m->val;
    if (!b) return;
    const uint64_t t = m->svtok;
    /* the body is written back as it is now, and is clean from here on:
     * the stores below are operations, which flush (and come back here) */
    XPVCV now = b->body, was = b->seen;
    b->seen = b->body;
    XPVCV *x = &now, *s = &was;
    if (x->xcv_flags != s->xcv_flags) goperl_cvfield_store(f, t, GOPERL_LV_CVFLAGS, x->xcv_flags);
    if (x->xcv_stash != s->xcv_stash)
        GOPERL_INTERNAL_HI(f, GOPERL_INT_CV_SET, t, GOPERL_TOK(x->xcv_stash), GOPERL_CVF_STASH);
    if (x->xcv_start_u.xcv_start != s->xcv_start_u.xcv_start)
        goperl_cvfield_store(f, t, GOPERL_LV_CVSTART, (uint64_t)(uintptr_t)x->xcv_start_u.xcv_start);
    if (x->xcv_root_u.xcv_root != s->xcv_root_u.xcv_root)
        goperl_cvfield_store(f, t, GOPERL_LV_CVROOT, (uint64_t)(uintptr_t)x->xcv_root_u.xcv_root);
    if (x->xcv_gv_u.xcv_gv != s->xcv_gv_u.xcv_gv) {
        if (x->xcv_flags & CVf_NAMED) {
            if (x->xcv_gv_u.xcv_hek) CvNAME_HEK_set((CV *)GOPERL_SV(t), x->xcv_gv_u.xcv_hek);
        } else {
            GOPERL_INTERNAL_HI(f, GOPERL_INT_CV_SET, t, GOPERL_TOK(x->xcv_gv_u.xcv_gv),
                               GOPERL_CVF_GV_U);
        }
    }
    if (x->xcv_file != s->xcv_file)
        goperl_cvfield_store(f, t, GOPERL_LV_CVFILE, (uint64_t)(uintptr_t)x->xcv_file);
    if (x->xcv_padlist_u.xcv_padlist != s->xcv_padlist_u.xcv_padlist)
        goperl_cvfield_store(f, t, GOPERL_LV_CVPADLIST,
                             (uint64_t)(uintptr_t)x->xcv_padlist_u.xcv_padlist);
    if (x->xcv_outside != s->xcv_outside)
        GOPERL_INTERNAL_HI(f, GOPERL_INT_CV_SET, t, GOPERL_TOK(x->xcv_outside), GOPERL_CVF_OUTSIDE);
    if (x->xcv_outside_seq != s->xcv_outside_seq)
        goperl_cvfield_store(f, t, GOPERL_LV_CVOUTSIDE_SEQ, x->xcv_outside_seq);
    if (x->xcv_depth != s->xcv_depth)
        goperl_cvfield_store(f, t, GOPERL_LV_CVDEPTH, (uint64_t)(uint32_t)x->xcv_depth);
}
/* ---- the fields of an object (SVt_PVOBJ) ---------------------------------------
 * ObjectFIELDS is the object's array of fields, as AvARRAY is an array's
 * body: its elements are read at every evaluation and written back as they
 * are. A block of native code's own stored there (Newx(ObjectFIELDS(obj),
 * n, SV *), the way an object is made) becomes the object's array once its
 * size is known (ObjectMAXFIELD): the interpreter makes an array of that
 * size, the block's contents go there, and the block is freed, as perl's
 * sv_clear frees the array. */
typedef struct goperl_objbody {
    uint64_t arr_g; /* the array in the interpreter's memory */
    int32_t len;    /* the fields it has */
    int32_t cap;
    uint64_t *shadow;
    uint64_t slots[];
} goperl_objbody_t;
#define GOPERL_OBJBODY_OF(m) \
    ((goperl_objbody_t *)((char *)(uintptr_t)(m)->shadow - offsetof(goperl_objbody_t, slots)))
static __attribute__((unused)) void goperl_objbody_read(goperl_frame_t *f, goperl_lvslot_t *m) {
    /* -> the array of the fields; second: their number */
    uint64_t nfields = 0;
    const uint64_t arr = goperl_do_opx(f, GOPERL_OP_INTERNAL, m->svtok, 0, GOPERL_INT_OBJ_FIELDS,
                                       0, 0, &nfields);
    const int32_t n = arr ? (int32_t)nfields : 0;
    goperl_objbody_t *b = m->shadow ? GOPERL_OBJBODY_OF(m) : 0;
    if (!b || n > b->cap) {
        int32_t ncap = b ? b->cap : 8;
        while (ncap < n) ncap *= 2;
        uint64_t *osh = b ? b->shadow : 0;
        goperl_objbody_t *nb =
            (goperl_objbody_t *)realloc(b, sizeof(goperl_objbody_t) + (size_t)ncap * 8);
        if (!nb) goperl_croakf(f, "Out of memory for the fields of an object");
        uint64_t *nsh = (uint64_t *)realloc(osh, (size_t)ncap * 8);
        if (!nsh) goperl_croakf(f, "Out of memory for the fields of an object");
        b = nb;
        b->shadow = nsh;
        b->cap = ncap;
        m->shadow = (uint64_t)(uintptr_t)b->slots;
    }
    b->arr_g = arr;
    b->len = b->arr_g ? n : 0;
    for (int32_t i = 0; i < b->len; i++)
        b->slots[i] = b->shadow[i] = goperl_g64(f, b->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i);
    m->val = m->shadow;
}
/* The pending number of fields less one of the object, or the
 * interpreter's. */
static __attribute__((unused)) int64_t goperl_obj_maxfield_now(goperl_frame_t *f, uint64_t tok) {
    goperl_lvstate_t *st = goperl_lvstate_v;
    for (int32_t pass = 0; st && pass < 2; pass++) {
        const int32_t n = pass == 0 ? st->nuses : st->nrecent;
        for (int32_t i = 0; i < n; i++) {
            goperl_lvslot_t *m = pass == 0 ? st->uses[i].slot : st->recent[i];
            if (m->svtok == tok && m->field == GOPERL_LV_OBJMAXFIELD && m->val != m->shadow) {
                m->shadow = m->val; /* it goes with the new array */
                return (int64_t)m->val;
            }
        }
    }
    return (int64_t)goperl_op_second(f, GOPERL_OP_INTERNAL, tok, 0, GOPERL_INT_OBJ_FIELDS) - 1;
}
static __attribute__((unused)) void goperl_objbody_store(goperl_frame_t *f, goperl_lvslot_t *m) {
    if (m->val && m->val != m->shadow) {
        /* a block of native code's own; the slot is clean while the
         * operations below run (they flush, and come back here) */
        const uint64_t *own = (const uint64_t *)(uintptr_t)m->val;
        m->val = m->shadow;
        const int64_t max = goperl_obj_maxfield_now(f, m->svtok);
        if (max < 0) {
            m->val = (uint64_t)(uintptr_t)own; /* its size is not known yet */
            return;
        }
        const uint64_t arr = GOPERL_INTERNAL(f, GOPERL_INT_OBJ_FIELDS_NEW, m->svtok,
                                             (uint64_t)(max + 1));
        for (int64_t i = 0; arr && i <= max; i++)
            goperl_g64_set(f, arr + GOPERL_GUEST_PTRSIZE * (uint64_t)i, own[i]);
        free((void *)own);
        goperl_objbody_read(f, m);
        return;
    }
    if (!m->shadow) return;
    goperl_objbody_t *b = GOPERL_OBJBODY_OF(m);
    for (int32_t i = 0; i < b->len; i++) {
        if (b->slots[i] == b->shadow[i]) continue;
        b->shadow[i] = b->slots[i];
        goperl_g64_set(f, b->arr_g + GOPERL_GUEST_PTRSIZE * (uint64_t)i, b->slots[i]);
    }
}
#define ObjectFIELDS(obj) \
    (*(SV ***)goperl_lv_ref(_gof, (SV *)(obj), GOPERL_LV_OBJFIELDS, GOPERL_LV_AT))
#define ObjectMAXFIELD(obj) \
    (*(SSize_t *)goperl_lv_ref(_gof, (SV *)(obj), GOPERL_LV_OBJMAXFIELD, GOPERL_LV_AT))

/* The slots whose value is the address of a body here. */
static __attribute__((unused)) void goperl_lvbody_load(goperl_frame_t *f, goperl_lvslot_t *m) {
    if (m->field == GOPERL_LV_CVBODY) {
        goperl_cvbody_load(f, m);
        return;
    }
    /* a block of native code's own whose size is not known yet stays */
    if (m->val && m->val != m->shadow) return;
    goperl_objbody_read(f, m);
}
static __attribute__((unused)) void goperl_lvbody_store(goperl_frame_t *f, goperl_lvslot_t *m) {
    if (m->field == GOPERL_LV_CVBODY) goperl_cvbody_store(f, m);
    else goperl_objbody_store(f, m);
}
static __attribute__((unused)) void goperl_lvbody_release(goperl_lvslot_t *m) {
    if (m->field == GOPERL_LV_CVBODY) {
        free((void *)(uintptr_t)m->val);
    } else {
        if (m->val && m->val != m->shadow) free((void *)(uintptr_t)m->val);
        if (m->shadow) {
            goperl_objbody_t *b = GOPERL_OBJBODY_OF(m);
            free(b->shadow);
            free(b);
        }
    }
    m->val = m->shadow = 0;
}

/* The body of a regexp is the copy ReANY makes. The body of any other SV
 * is not modelled: SvANY says whether it has one (NULL for an SV of type
 * SVt_NULL, as perl's), and what it points to may not be touched (an
 * access ends the process at once, on a page that allows none). */
static __attribute__((unused)) void *goperl_no_body(void) {
    static void *page;
    if (!page) {
        page = mmap(0, (size_t)sysconf(_SC_PAGESIZE), PROT_NONE, MAP_PRIVATE | MAP_ANON, -1, 0);
        if (page == MAP_FAILED) abort();
    }
    return page;
}
/* What SvANY reads of an SV that is not a sub (the slot GOPERL_LV_SVANY
 * loads). */
static void *goperl_sv_any_value(goperl_frame_t *f, SV *sv) {
    const svtype t = (svtype)SvTYPE(sv);
    if (t == SVt_REGEXP) return goperl_reany(f, (REGEXP *)sv);
    if (t == SVt_NULL) {
        /* a head without a body has none (NULL), but for the header of an
         * SV arena, whose any is the next arena: what the head says, as
         * the interpreter's token of it */
        const uint64_t *const head = (const uint64_t *)goperl_guest_mem(f, GOPERL_TOK(sv));
        return head ? (void *)(uintptr_t)head[0] : 0;
    }
    return goperl_no_body();
}
/* SvANY, an lvalue as perl's is: the body of a sub, and for any other SV
 * the slot of what goperl_sv_any_value reads, whose store gives the SV a
 * body of the interpreter's memory (see GOPERL_LV_SVANY). An SV that is
 * NULL has a NULL body. */
static __attribute__((unused)) void **goperl_sv_any(goperl_frame_t *f, SV *sv, const void *site, uintptr_t frame) {
    static __thread void *none;
    if (!sv) {
        none = 0;
        return &none;
    }
    const svtype t = (svtype)SvTYPE(sv);
    if (t == SVt_PVCV || t == SVt_PVFM)
        return (void **)goperl_lv_ref(f, sv, GOPERL_LV_CVBODY, site, frame);
    return (void **)goperl_lv_ref(f, sv, GOPERL_LV_SVANY, site, frame);
}
#undef SvANY
#define SvANY(sv) (*goperl_sv_any(_gof, (SV *)(sv), GOPERL_LV_AT))
/* A key shared by the interpreter, shared once more, as hv.h's: a key of
 * the interpreter's memory (HeKEY_hek, SvSHARED_HEK_FROM_PV) counts one
 * more holder in the shared_he it is the key of, which the interpreter
 * takes back when it unshares the key (a scalar of LEN 0 that is IsCOW
 * and has the key as its string, when it is freed). A key of native
 * code's own (share_hek, CvNAME_HEK: a copy for the native call) is the
 * key itself; the interpreter shares it when it is given one. */
struct shared_he {
    struct goperl_he shared_he_he;
    struct hek shared_he_hek;
};
static inline __attribute__((unused)) HEK *goperl_share_hek_hek(goperl_frame_t *f, HEK *hek) {
    if (hek && goperl_in_memory(f, hek))
        ++((struct shared_he *)((char *)hek - STRUCT_OFFSET(struct shared_he, shared_he_hek)))
              ->shared_he_he.he_valu.hent_refcount;
    return hek;
}
#define share_hek_hek(hek) goperl_share_hek_hek(_gof, (HEK *)(hek))

/* ---- ops ----------------------------------------------------------------------- */
#define cPVOPo cPVOPx(o)
#define cPVOP cPVOPx(PL_op)
#define kGVOP_gv cGVOPx_gv(kid)
#define kSVOP_sv cSVOPx_sv(kid)
/* The reference count of the root of a sub's ops is its target, as perl
 * keeps it (OpREFCNT_inc on CvROOT when a sub shares another's ops). */
#define OpREFCNT_set(o, n) ((o)->op_targ = (n))
#define OpREFCNT_inc(o) \
    ({ OP *const goperl_opri_ = (OP *)(o); goperl_opri_ ? (++goperl_opri_->op_targ, goperl_opri_) : (OP *)0; })
#define OpREFCNT_dec(o) (--((OP *)(o))->op_targ)
#ifndef OP_REFCNT_LOCK
#define OP_REFCNT_LOCK NOOP
#define OP_REFCNT_UNLOCK NOOP
#endif
/* The label of a COP: its bytes in the interpreter's memory. */
static __attribute__((unused)) const char *goperl_cop_label(goperl_frame_t *f, const COP *cop, STRLEN *len, U32 *flags) {
    /* -> the label, or 0; second: utf8<<31 | length */
    uint64_t second = 0;
    const uint64_t r = cop && cop->gop ? goperl_do_opx(f, GOPERL_OP_INTERNAL, cop->gop, 0,
                                                       GOPERL_INT_COP_LABEL, 0, 0, &second)
                                       : 0;
    if (len) *len = r ? (STRLEN)(second & 0x7FFFFFFFu) : 0;
    if (flags) *flags = (r && (second & 0x80000000u)) ? SVf_UTF8 : 0;
    return r ? (const char *)goperl_guest_mem(f, r) : (const char *)0;
}
#define CopLABEL_len_flags(c, len, flags) \
    goperl_cop_label(_gof, (const COP *)(c), (len), (flags))
#define CopLABEL_len(c, len) goperl_cop_label(_gof, (const COP *)(c), (len), (U32 *)0)
#define CopLABEL(c) goperl_cop_label(_gof, (const COP *)(c), (STRLEN *)0, (U32 *)0)

/* CALLRUNOPS runs the ops from PL_op on, as the run loop does, on the
 * interpreter's stacks as native code left them; what the ops leave there
 * is native code's stack afterwards. PL_op may be an op of native code's
 * own (an entersub op on the C stack, as perl's amagic_call runs an
 * overloaded operator's sub): the interpreter runs its copy of it
 * (goperl_op_for_interp). PL_op is what the run loop left in it, NULL. */
static uint64_t goperl_op_for_interp(goperl_frame_t *f, const OP *o, int depth);
static __attribute__((unused)) void goperl_callrunops(goperl_frame_t *f) {
    OP *const start = f->plop;
    const uint64_t tok = goperl_op_for_interp(f, start, 0);
    goperl_optree_flush(f);
    GOPERL_INTERNAL(f, GOPERL_INT_RUNOPS, tok, (uint64_t)1 << 63);
    f->plop = f->plop_base = 0;
}
#define CALLRUNOPS(ctx) goperl_callrunops(_gof)
#define PL_runops_std goperl_callrunops

/* docatch (pp_ctl.c): a pp function, and the ops after it, in a run loop of
 * their own under a JMPENV of their own, which a pp function asks for when
 * CATCH_GET says the loop it runs in cannot go on after a die an eval
 * catches (Future::AsyncAwait's pp_await, resuming a sub it suspended
 * inside an eval): under the new JMPENV CATCH_GET is false, and an eval
 * entered there that catches a die goes on from its op in that loop. The
 * interpreter runs it, with the running op's pp function first, which the
 * function native code gives must be. -> NULL, as docatch's: the loop has
 * run the ops to their end. */
static __attribute__((unused)) OP *goperl_docatch(goperl_frame_t *f, Perl_ppaddr_t firstpp) {
    OP *const o = PL_op;
    if (!f->hook_op_tok)
        goperl_croakf(f, "docatch is called outside a pp function");
    if (!o || o->op_ppaddr != firstpp)
        goperl_croakf(f, "docatch is given a pp function that is not the running op's: gperl "
                         "runs docatch from the op");
    GOPERL_INTERNAL(f, GOPERL_INT_DOCATCH, goperl_op_field_tok(f, (uint64_t)(uintptr_t)o), 0);
    return NULL;
}
#define docatch(firstpp) goperl_docatch(_gof, (Perl_ppaddr_t)(firstpp))

/* ---- the hints of a COP -----------------------------------------------------------
 * The hints hash of a COP is the interpreter's chain of entries (a token
 * here); a value fetched from it is a new mortal, &PL_sv_placeholder when
 * the key is not there, and an exists-fetch is true or NULL, as perl's. */
typedef struct refcounted_he COPHH;
#define COPHH_KEY_UTF8 0x00000001
#define COPHH_EXISTS 0x00000002
#define CopHINTHASH_get(c)                                                 \
    ({ const COP *const goperl_chh_c_ = (const COP *)(c);                  \
       (COPHH *)(uintptr_t)(goperl_chh_c_ && goperl_chh_c_->gop            \
                                ? GOPERL_INTERNAL(_gof, GOPERL_INT_HINTHASH, \
                                                  goperl_chh_c_->gop, 0)   \
                                : 0); })
static __attribute__((unused)) SV *goperl_cophh_fetch(goperl_frame_t *f, const COPHH *cophh, const char *key,
                              STRLEN len, U32 flags) {
    return GOPERL_SV(GOPERL_INTERNAL_S(f, GOPERL_INT_HINTHASH_FETCH, (uintptr_t)cophh,
                                       ((uint64_t)flags << 32) | (uint32_t)len, key, len));
}
#define cophh_fetch_pvn(cophh, key, keylen, hash, flags) \
    goperl_cophh_fetch(_gof, (cophh), (key), (STRLEN)(keylen), (U32)(flags) & ~COPHH_EXISTS)
#define cophh_fetch_pvs(cophh, key, flags) \
    cophh_fetch_pvn((cophh), "" key "", sizeof(key) - 1, 0, (flags))
#define cophh_fetch_pv(cophh, key, hash, flags)                             \
    ({ const char *const goperl_cfp_k_ = (key);                             \
       cophh_fetch_pvn((cophh), goperl_cfp_k_, __builtin_strlen(goperl_cfp_k_), 0, (flags)); })
#define cophh_fetch_sv(cophh, keysv, hash, flags)                           \
    ({ STRLEN goperl_cfs_l_;                                                \
       SV *const goperl_cfs_k_ = (keysv);                                   \
       const char *const goperl_cfs_p_ = SvPV(goperl_cfs_k_, goperl_cfs_l_); \
       cophh_fetch_pvn((cophh), goperl_cfs_p_, goperl_cfs_l_, 0,           \
                       (flags) | (SvUTF8(goperl_cfs_k_) ? COPHH_KEY_UTF8 : 0)); })
#define cophh_exists_pvn(cophh, key, keylen, hash, flags) \
    cBOOL(goperl_cophh_fetch(_gof, (cophh), (key), (STRLEN)(keylen), (U32)(flags) | COPHH_EXISTS))
#define cophh_exists_pvs(cophh, key, flags) \
    cophh_exists_pvn((cophh), "" key "", sizeof(key) - 1, 0, (flags))
#define cophh_exists_pv(cophh, key, hash, flags)                            \
    ({ const char *const goperl_cep_k_ = (key);                             \
       cophh_exists_pvn((cophh), goperl_cep_k_, __builtin_strlen(goperl_cep_k_), 0, (flags)); })
#define cophh_exists_sv(cophh, keysv, hash, flags)                          \
    ({ STRLEN goperl_ces_l_;                                                \
       SV *const goperl_ces_k_ = (keysv);                                   \
       const char *const goperl_ces_p_ = SvPV(goperl_ces_k_, goperl_ces_l_); \
       cophh_exists_pvn((cophh), goperl_ces_p_, goperl_ces_l_, 0,          \
                        (flags) | (SvUTF8(goperl_ces_k_) ? COPHH_KEY_UTF8 : 0)); })
#define cophh_2hv(cophh, flags)                                             \
    ((HV *)GOPERL_SV(GOPERL_INTERNAL(_gof, GOPERL_INT_HINTHASH_2HV,         \
                                     (uintptr_t)(cophh), (U32)(flags))))

/* ---- methods, character classes, names ---------------------------------------- */
#define gv_fetchmeth_sv(stash, namesv, level, flags)                        \
    ({ STRLEN goperl_gfms_l_;                                               \
       SV *const goperl_gfms_n_ = (namesv);                                 \
       const char *const goperl_gfms_p_ = SvPV(goperl_gfms_n_, goperl_gfms_l_); \
       gv_fetchmeth_pvn((stash), goperl_gfms_p_, goperl_gfms_l_, (level),   \
                        (flags) | (SvUTF8(goperl_gfms_n_) ? SVf_UTF8 : 0)); })
/* The name of a stash, as a shared key. */
#define HvNAME_HEK(hv)                                                      \
    ({ HV *const goperl_hnh_hv_ = (HV *)(hv);                               \
       const char *const goperl_hnh_n_ = HvNAME(goperl_hnh_hv_);            \
       goperl_hnh_n_ ? goperl_hek_of(_gof, goperl_hnh_n_, (I32)HvNAMELEN(goperl_hnh_hv_), \
                                     HvNAMEUTF8(goperl_hnh_hv_) != 0, 0)    \
                     : (HEK *)0; })
#define HvENAME_HEK(hv) HvNAME_HEK(hv)
/* The name of a glob, as a shared key. */
#define GvNAME_HEK(gv)                                                      \
    ({ GV *const goperl_gnh_gv_ = (GV *)(gv);                               \
       const char *const goperl_gnh_n_ = GvNAME(goperl_gnh_gv_);            \
       goperl_gnh_n_ ? goperl_hek_of(_gof, goperl_gnh_n_, (I32)GvNAMELEN(goperl_gnh_gv_), \
                                     GvNAMEUTF8(goperl_gnh_gv_) != 0, 0)    \
                     : (HEK *)0; })

/* HvARRAY: the buckets of a hash as the interpreter has them when it is
 * evaluated, each its first entry (the interpreter's own: HeNEXT gives the
 * rest of the chain); NULL when the hash has no buckets yet (pp_defined's
 * test of a hash). The bucket array of the interpreter holds addresses of
 * its memory, so this is an array of this process's pointers to the same
 * entries, filled at each evaluation and kept until the outermost native
 * call returns: the same array for the same buckets, as perl's pointer is
 * the same until the hash is split (Sereal::Encoder tells a split during a
 * callback so). It is not an lvalue (a store into it would not reach the
 * hash). */
typedef struct goperl_hvarray {
    struct goperl_hvarray *next;
    uint64_t arr; /* the address of the buckets in the interpreter */
    uint64_t n;   /* their number */
    HE *ents[];
} goperl_hvarray_t;
static __attribute__((unused)) HE **goperl_hv_array(goperl_frame_t *f, HV *hv) {
    /* -> the buckets, or 0; second: their number */
    uint64_t nbuckets = 0;
    const uint64_t arr = goperl_do_opx(f, GOPERL_OP_INTERNAL, GOPERL_TOK(hv), 0, GOPERL_INT_HV_ARRAY,
                                       0, 0, &nbuckets);
    if (!arr) return (HE **)0;
    const uint64_t n = nbuckets;
    goperl_shared_t *const sh = (goperl_shared_t *)f->inst;
    goperl_hvarray_t *a = (goperl_hvarray_t *)sh->hv_arrays;
    while (a && (a->arr != arr || a->n != n)) a = a->next;
    if (!a) {
        a = (goperl_hvarray_t *)goperl_arena_alloc(sizeof(goperl_hvarray_t) +
                                                   sizeof(HE *) * (size_t)(n ? n : 1));
        if (!a) goperl_croakf(f, "Out of memory!");
        a->arr = arr;
        a->n = n;
        a->next = (goperl_hvarray_t *)sh->hv_arrays;
        sh->hv_arrays = a;
    }
    for (uint64_t i = 0; i < n; i++)
        a->ents[i] = goperl_he_at(f, goperl_g64(f, arr + GOPERL_GUEST_PTRSIZE * i));
    return a->ents;
}
#define HvARRAY(hv) goperl_hv_array(_gof, (HV *)(hv))

/* PerlIO_debug: the message goes to the interpreter's, which writes it
 * where PERLIO_DEBUG names, as perl's does. */
static __attribute__((unused)) void goperl_perlio_debug(goperl_frame_t *f, const char *fmt, ...) {
    goperl_text_t t;
    va_list ap;
    goperl_text_init(&t);
    va_start(ap, fmt);
    goperl_vtext(f, &t, fmt, ap);
    va_end(ap);
    goperl_text_put(&t, "", 1, 0);
    GOPERL_INTERNAL_S(f, GOPERL_INT_PERLIO_DEBUG, 0, 0, t.p, t.len);
    goperl_text_free(&t);
}
#define PerlIO_debug(...) goperl_perlio_debug(_gof, __VA_ARGS__)

/* SvPV_free: the string buffer the SV owns goes back to the interpreter's
 * allocator (a buffer whose start was chopped off as a whole); the length
 * and the pointer are the caller's to set next, as on perl. */
#define SvPV_free(sv)                                                        \
    ((void)GOPERL_INTERNAL(_gof, GOPERL_INT_SV_PV_FREE, GOPERL_TOK(sv), 0))

/* A compile error is queued: the parse goes on and the compilation fails
 * at its end with every queued message (perl's defer_error, and qerror,
 * its older name). */
#define GOPERL_CALL_QERROR 20
static __attribute__((unused)) void goperl_defer_error(goperl_frame_t *f, SV *err) {
    goperl_do_op3(f, GOPERL_OP_CALL, GOPERL_TOK(err), 0, (uint64_t)(GOPERL_CALL_QERROR), 0, 0);
}
#define defer_error(err) goperl_defer_error(_gof, (SV *)(err))
#define qerror(err) goperl_defer_error(_gof, (SV *)(err))
#define Perl_defer_error(...) goperl_defer_error(_gof, __VA_ARGS__)
#define Perl_qerror(...) goperl_defer_error(_gof, __VA_ARGS__)

/* PL_multideref_pc is the item of a multideref's auxiliary data being run,
 * which native code running a multideref op of its own keeps up to date
 * (the messages about undefined values name the element by it): in this
 * process, a pointer into the items an op's op_aux designates. */
static __attribute__((unused)) UNOP_AUX_item **goperl_multideref_pc_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_MULTIDEREF_PC];
    if (!s->live || s->val == s->shadow) {
        s->val = (uint64_t)(uintptr_t)goperl_guest_mem(
            f, goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_MULTIDEREF_PC, 0, 0, 0));
        s->shadow = s->val;
        goperl_plshadow_set_live(s);
    }
    return (UNOP_AUX_item **)&s->val;
}
#define PL_multideref_pc (*goperl_multideref_pc_ref(_gof))

/* PL_statcache, what the last stat found and -X _ reads, as a struct stat
 * of this process: read from the interpreter when native code first looks
 * at it after an operation, and what native code stored into it goes back
 * before the next one (EV fills it with what its own stat found). A field
 * native code stored into keeps what it stored until it has gone back, also
 * when an operation runs in between: in `PL_statcache.st_size =
 * SvIV(sv);` the compiler may take the field's address before SvIV runs
 * (gcc does), and the store then comes after the operation. The fields
 * nothing stored into are read again after an operation. */
typedef struct goperl_statcache {
    struct stat st;   /* as native code sees it */
    struct stat seen; /* as it was read */
    uint64_t gen;     /* GOPERL_SHARED->gen + 1 when it was read */
} goperl_statcache_t;
#define goperl_statcache_v (*(goperl_statcache_t **)&GOPERL_SHARED->statcache)
#define GOPERL_STATCACHE_WORDS 13
/* The fields of a struct stat as the interpreter exchanges them. */
static __attribute__((unused)) void goperl_statcache_words(const struct stat *st, uint64_t *w) {
    w[0] = (uint64_t)st->st_dev;
    w[1] = (uint64_t)st->st_ino;
    w[2] = (uint64_t)st->st_mode;
    w[3] = (uint64_t)st->st_nlink;
    w[4] = (uint64_t)st->st_uid;
    w[5] = (uint64_t)st->st_gid;
    w[6] = (uint64_t)st->st_rdev;
    w[7] = (uint64_t)(int64_t)st->st_size;
    w[8] = (uint64_t)(int64_t)st->st_atime;
    w[9] = (uint64_t)(int64_t)st->st_mtime;
    w[10] = (uint64_t)(int64_t)st->st_ctime;
    w[11] = (uint64_t)(int64_t)st->st_blksize;
    w[12] = (uint64_t)(int64_t)st->st_blocks;
}
static __attribute__((unused)) void goperl_statcache_fill(struct stat *st, const uint64_t *w) {
    memset(st, 0, sizeof *st);
    st->st_dev = (dev_t)w[0];
    st->st_ino = (ino_t)w[1];
    st->st_mode = (mode_t)w[2];
    st->st_nlink = (nlink_t)w[3];
    st->st_uid = (uid_t)w[4];
    st->st_gid = (gid_t)w[5];
    st->st_rdev = (dev_t)w[6];
    st->st_size = (off_t)(int64_t)w[7];
    st->st_atime = (time_t)(int64_t)w[8];
    st->st_mtime = (time_t)(int64_t)w[9];
    st->st_ctime = (time_t)(int64_t)w[10];
    st->st_blksize = (blksize_t)(int64_t)w[11];
    st->st_blocks = (blkcnt_t)(int64_t)w[12];
}
static __attribute__((unused)) struct stat *goperl_statcache_ref(goperl_frame_t *f) {
    goperl_statcache_t *c = goperl_statcache_v;
    if (!c) {
        c = (goperl_statcache_t *)calloc(1, sizeof *c);
        if (!c) goperl_croakf(f, "Out of memory!");
        goperl_statcache_v = c;
    }
    if (c->gen != GOPERL_SHARED->gen + 1) {
        uint64_t now[GOPERL_STATCACHE_WORDS], mine[GOPERL_STATCACHE_WORDS],
            was[GOPERL_STATCACHE_WORDS];
        memcpy(now, goperl_guest_mem(f, goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, 0, (uint64_t)(GOPERL_INT_STATCACHE), 0, 0, 0)),
               sizeof now);
        goperl_statcache_words(&c->st, mine);
        goperl_statcache_words(&c->seen, was);
        goperl_statcache_fill(&c->seen, now);
        /* a field stored into since it was read, not gone back yet */
        for (int i = 0; i < GOPERL_STATCACHE_WORDS; i++)
            if (mine[i] != was[i]) now[i] = mine[i];
        goperl_statcache_fill(&c->st, now);
        c->gen = GOPERL_SHARED->gen + 1;
    }
    return &c->st;
}
static void goperl_statcache_writeback(goperl_frame_t *f) {
    goperl_statcache_t *c = goperl_statcache_v;
    if (!c) return;
    uint64_t mine[GOPERL_STATCACHE_WORDS], was[GOPERL_STATCACHE_WORDS];
    goperl_statcache_words(&c->st, mine);
    goperl_statcache_words(&c->seen, was);
    if (!memcmp(mine, was, sizeof mine)) return;
    /* the fields nothing stored into, as the interpreter has them now */
    goperl_statcache_ref(f);
    goperl_statcache_words(&c->st, mine);
    c->seen = c->st;
    goperl_xs_opx(f, GOPERL_OP_INTERNAL, 0, 1, (uint64_t)(GOPERL_INT_STATCACHE), (const char *)mine, sizeof mine, 0);
}
#define PL_statcache (*goperl_statcache_ref(_gof))

/* ---- functions of native code the interpreter runs as its own ------------
 *
 * A source filter (filter_add), a run loop (PL_runops) and block hooks
 * (blockhook_register) are functions the interpreter calls where it would
 * call one of its own. It is given a function of its own that calls the
 * module's through the host-call channel, by the id it has here: the
 * index of the function in this instance's table, plus one. */
#define GOPERL_FNTAB_MAX 64 /* the interpreter has as many filter functions */
#define GOPERL_BHK_MAX 16   /* and as many block hooks */
typedef struct block_hooks {
    U32 bhk_flags;
    void (*bhk_start)(pTHX_ int full);
    void (*bhk_pre_end)(pTHX_ OP **seq);
    void (*bhk_post_end)(pTHX_ OP **seq);
    void (*bhk_eval)(pTHX_ OP *const saveop);
} BHK;
typedef struct goperl_fntab {
    void (*fns[GOPERL_FNTAB_MAX])(void);
    int32_t n;
    int32_t nbhk;
    BHK *bhks[GOPERL_BHK_MAX];
    U32 bhk_told[GOPERL_BHK_MAX]; /* the flags the interpreter was given */
    /* the interpreter's own run loop as native code sees it in PL_runops:
     * one function for every module, which the interpreter is told of as
     * its own loop when it is put back */
    int (*interp_loop)(pTHX);
} goperl_fntab_t;
#define goperl_fntab_v (*(goperl_fntab_t **)&GOPERL_SHARED->fntab)
static int goperl_runops_of_interp(pTHX);
static goperl_fntab_t *goperl_fntab(goperl_frame_t *f) {
    goperl_fntab_t *t = goperl_fntab_v;
    if (!t) {
        t = (goperl_fntab_t *)calloc(1, sizeof *t);
        if (!t) goperl_croakf(f, "Out of memory!");
        t->interp_loop = goperl_runops_of_interp;
        goperl_fntab_v = t;
    }
    return t;
}
/* The id of a function, given one the first time. */
static uint32_t goperl_fn_id(goperl_frame_t *f, void (*fn)(void)) {
    goperl_fntab_t *t = goperl_fntab(f);
    for (int32_t i = 0; i < t->n; i++)
        if (t->fns[i] == fn) return (uint32_t)i + 1;
    if (t->n >= GOPERL_FNTAB_MAX)
        goperl_croakf(f, "too many functions of native code given to the interpreter");
    t->fns[t->n++] = fn;
    return (uint32_t)t->n;
}
static void (*goperl_fn_of(goperl_frame_t *f, uint32_t id))(void) {
    const goperl_fntab_t *t = goperl_fntab_v;
    if (!t || id == 0 || (int32_t)id > t->n)
        goperl_croakf(f, "host call: no function %u of native code", (unsigned)id);
    return t->fns[id - 1];
}

/* Source filters. The data of a filter is the module's (FILTER_DATA); the
 * function a filter has (IoANY of the data) is the module's own function
 * for one it added, and one that runs the interpreter's filter in that
 * place for the others. */
typedef I32 (*filter_t)(pTHX_ int, SV *, int);
#define GOPERL_INT_FILTER_ADD 24
#define GOPERL_INT_FILTER_DEL 25
#define GOPERL_INT_FILTER_READ 26
#define GOPERL_INT_FILTER_ID 27
#define GOPERL_INT_SV_GETS 28
#define GOPERL_INT_BHK_REGISTER 29
#define GOPERL_INT_BHK_FLAGS 30
static __attribute__((unused)) SV *goperl_filter_add(goperl_frame_t *f, filter_t fn, SV *datasv) {
    if (!fn) return (SV *)0;
    const uint32_t id = goperl_fn_id(f, (void (*)(void))fn);
    return GOPERL_SV(GOPERL_INTERNAL(f, GOPERL_INT_FILTER_ADD, GOPERL_TOK(datasv), id));
}
static __attribute__((unused)) void goperl_filter_del(goperl_frame_t *f, filter_t fn) {
    const goperl_fntab_t *t = goperl_fntab_v;
    for (int32_t i = 0; t && i < t->n; i++)
        if (t->fns[i] == (void (*)(void))fn) {
            GOPERL_INTERNAL(f, GOPERL_INT_FILTER_DEL, 0, (uint64_t)i + 1);
            return;
        }
}
#define filter_add(fn, datasv) goperl_filter_add(_gof, (filter_t)(fn), (SV *)(datasv))
#define filter_del(fn) goperl_filter_del(_gof, (filter_t)(fn))
#define Perl_filter_add(...) goperl_filter_add(_gof, __VA_ARGS__)
#define Perl_filter_del(...) goperl_filter_del(_gof, __VA_ARGS__)
/* The next filter fills the buffer: the interpreter's filter_read. (A
 * module with a filter_read of its own, as Devel::Declare's copy of toke.c
 * has, gets the interpreter's for filter_read all the same, as it does on
 * perl, whose own definition takes the name's place.) */
static __attribute__((unused)) I32 goperl_filter_read(goperl_frame_t *f, int idx, SV *buf_sv,
                                                      int maxlen) {
    return (I32)(int32_t)(uint32_t)GOPERL_INTERNAL(
        f, GOPERL_INT_FILTER_READ, GOPERL_TOK(buf_sv),
        ((uint64_t)(uint32_t)idx << 32) | (uint32_t)maxlen);
}
#define filter_read(idx, sv, len) goperl_filter_read(_gof, (idx), (SV *)(sv), (len))
#define FILTER_READ(idx, sv, len) filter_read((idx), (sv), (len))
#define FILTER_DATA(idx) \
    (PL_parser ? AvARRAY(PL_parser->rsfp_filters)[idx] : (SV *)0)
#define FILTER_IS_ACTIVE(datasv) \
    (PL_parser && PL_parser->rsfp_filters && (datasv) == FILTER_DATA(AvFILLp(PL_parser->rsfp_filters)))
#define FILTER_ISREADER(idx) \
    (PL_parser && PL_parser->rsfp_filters && (idx) >= AvFILLp(PL_parser->rsfp_filters))
/* A filter of the interpreter's in the chain, run as native code would run
 * the function in that place: by the interpreter, at the same index. */
static I32 goperl_filter_of_interp(pTHX_ int idx, SV *buf_sv, int maxlen) {
    return (I32)(int32_t)(uint32_t)GOPERL_INTERNAL(
        _gof, GOPERL_INT_FILTER_READ, GOPERL_TOK(buf_sv),
        ((uint64_t)(uint32_t)idx << 32) | (uint32_t)maxlen);
}
static __attribute__((unused)) void *goperl_io_any(goperl_frame_t *f, SV *datasv) {
    const uint32_t id = (uint32_t)GOPERL_INTERNAL(f, GOPERL_INT_FILTER_ID, GOPERL_TOK(datasv), 0);
    return id ? (void *)goperl_fn_of(f, id) : (void *)goperl_filter_of_interp;
}
#define IoANY(sv) goperl_io_any(_gof, (SV *)(sv))

/* sv_gets: a line (or the rest) of a handle into the SV, appended at
 * append; the buffer, or NULL at the end. */
static __attribute__((unused)) char *goperl_sv_gets(goperl_frame_t *f, SV *sv, PerlIO *fp, I32 append) {
    /* b = the handle, append in c's high half */
    if (!GOPERL_INTERNAL_HI(f, GOPERL_INT_SV_GETS, GOPERL_TOK(sv), goperl_pio_tok(f, fp),
                            (uint32_t)append))
        return (char *)0;
    return SvPVX(sv);
}
#define sv_gets(sv, fp, append) goperl_sv_gets(_gof, (SV *)(sv), (PerlIO *)(fp), (I32)(append))
#define Perl_sv_gets(...) goperl_sv_gets(_gof, __VA_ARGS__)

/* The run loop, PL_runops: the interpreter's own (which runs from PL_op
 * when native code calls it), or one of native code. */
typedef int (*runops_proc_t)(pTHX);
#define GOPERL_PL_RUNOPS 99
static int goperl_runops_of_interp(pTHX) {
    goperl_callrunops(_gof);
    return 0;
}
static __attribute__((unused)) runops_proc_t *goperl_runops_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_RUNOPS];
    if (!s->live || s->val == s->shadow) {
        const uint32_t id = (uint32_t)goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_RUNOPS, 0, 0, 0);
        s->val = (uint64_t)(uintptr_t)(id ? (runops_proc_t)goperl_fn_of(f, id)
                                          : goperl_fntab(f)->interp_loop);
        s->shadow = s->val;
        goperl_plshadow_set_live(s);
    }
    return (runops_proc_t *)&s->val;
}
/* What PL_runops given a function is to the interpreter (goperl_plshadow_out):
 * its own loop, or the id of the module's. */
static uint64_t goperl_runops_out(goperl_frame_t *f, uint64_t val) {
    goperl_fntab_t *t = goperl_fntab(f);
    if (!val || val == (uint64_t)(uintptr_t)t->interp_loop) return 0;
    return goperl_fn_id(f, (void (*)(void))(uintptr_t)val);
}
#define PL_runops (*goperl_runops_ref(_gof))

/* PL_sortcop: the interpreter's, a word native code may keep what it likes
 * in while it sorts with a comparator of its own (Sort::Key keeps its keys
 * there for the comparator to find, under SAVEVPTR), as the interpreter's
 * sort keeps the op of a sort block there. What native code stores is kept
 * as it is, an address of this process's or the interpreter's. */
#define GOPERL_PL_SORTCOP 114
#define PL_sortcop (*(OP **)goperl_plint_ref(_gof, GOPERL_PL_SORTCOP))

/* The signal handler the interpreter calls for a signal that has arrived
 * (PL_sighandlerp): its own, which runs the %SIG handler, or a function of
 * native code (Tk's Event, which marks a handler of its own instead). The
 * function is given the host's number of the signal. */
#define GOPERL_PL_SIGHANDLERP 115
enum { GOPERL_HOST_SIGHANDLER = 10 };
#define GOPERL_INT_SIGHANDLER 37 /* b: a signal of the interpreter's: its
                                  * own handler runs for it */
static Signal_t goperl_sighandler_of_interp(int sig) {
    goperl_frame_t *const f = goperl_cur_frame_v;
    int guest = 0;
    for (int s = 1; s < 65 && !guest; s++)
        if (goperl_signal_host(s) == sig) guest = s;
    if (f && guest) GOPERL_INTERNAL(f, GOPERL_INT_SIGHANDLER, 0, (uint64_t)guest);
}
static __attribute__((unused)) Sighandler_t *goperl_sighandlerp_ref(goperl_frame_t *f) {
    goperl_plshadow_t *s = &goperl_plshadow_v[GOPERL_PL_SIGHANDLERP];
    if (!s->live || s->val == s->shadow) {
        const uint32_t id = (uint32_t)goperl_xs_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_SIGHANDLERP, 0, 0, 0);
        s->val = (uint64_t)(uintptr_t)(id ? (Sighandler_t)goperl_fn_of(f, id)
                                          : goperl_sighandler_of_interp);
        s->shadow = s->val;
        goperl_plshadow_set_live(s);
    }
    return (Sighandler_t *)&s->val;
}
/* What PL_sighandlerp given a function is to the interpreter
 * (goperl_plshadow_out): its own handler, or the id of the module's. */
static uint64_t goperl_sighandlerp_out(goperl_frame_t *f, uint64_t val) {
    if (!val || val == (uint64_t)(uintptr_t)goperl_sighandler_of_interp) return 0;
    return goperl_fn_id(f, (void (*)(void))(uintptr_t)val);
}
#define PL_sighandlerp (*goperl_sighandlerp_ref(_gof))
#define PL_sighandler1p PL_sighandlerp

/* my_exit and my_failure_exit: the program ends as `exit` ends it (END
 * blocks, global destruction), from inside the native call; neither
 * returns. */
#define GOPERL_INT_MY_EXIT 36 /* a: 0 for my_exit, with b the status; 1 for
                               * my_failure_exit */
static void goperl_my_exit(goperl_frame_t *f, int failure, U32 status) __attribute__((noreturn, unused));
static void goperl_my_exit(goperl_frame_t *f, int failure, U32 status) {
    GOPERL_INTERNAL(f, GOPERL_INT_MY_EXIT, (uint64_t)failure, (uint64_t)status);
    goperl_croakf(f, "%s: the program did not end", failure ? "my_failure_exit" : "my_exit");
}
#define my_exit(status) goperl_my_exit(_gof, 0, (U32)(status))
#define my_failure_exit() goperl_my_exit(_gof, 1, 0)
#define Perl_my_failure_exit(...) my_failure_exit()

/* rsignal, as perl's: the process's action for the signal becomes the
 * handler (through the module's sigaction, see perl.h), and the handler it
 * had is returned. */
static __attribute__((unused)) Sighandler_t goperl_rsignal(goperl_frame_t *f, int signo, Sighandler_t handler) {
    struct sigaction act, oact;
    memset(&act, 0, sizeof act);
    act.sa_handler = handler;
    sigemptyset(&act.sa_mask);
    act.sa_flags = 0;
#ifdef SA_RESTART
    if (f && (PL_signals & PERL_SIGNALS_UNSAFE_FLAG)) act.sa_flags |= SA_RESTART;
#endif
#if defined(SA_NOCLDWAIT) && !defined(__APPLE__)
    if (signo == SIGCHLD && handler == (Sighandler_t)SIG_IGN) act.sa_flags |= SA_NOCLDWAIT;
#endif
    if (sigaction(signo, &act, &oact) == -1) return (Sighandler_t)SIG_ERR;
    return (Sighandler_t)oact.sa_handler;
}
static __attribute__((unused)) Sighandler_t goperl_rsignal_state(int signo) {
    struct sigaction oact;
    if (sigaction(signo, (struct sigaction *)0, &oact) == -1) return (Sighandler_t)SIG_ERR;
    return (Sighandler_t)oact.sa_handler;
}
#define rsignal(signo, handler) goperl_rsignal(_gof, (signo), (Sighandler_t)(handler))
#define Perl_rsignal(...) rsignal(__VA_ARGS__)
#define rsignal_state(signo) goperl_rsignal_state(signo)
#define Perl_rsignal_state(...) rsignal_state(__VA_ARGS__)

/* save_svref: the value in a slot is saved, and a new one put there, as
 * `local` does to it. The slot must be one the interpreter knows the place
 * of: a slot of its memory (the SV** an hv_fetch or av_fetch returns, see
 * goperl_live_slot), which the interpreter saves itself, or a variable of
 * its own, a field that an accessor designates, a slot of an array's body;
 * a slot that is only this process's copy of a value cannot be saved where
 * the value lies, and the call dies saying so. */
#define GOPERL_INT_SAVE_SVREF 39 /* a: the slot's address -> the new SV */
static __attribute__((unused)) SV *goperl_save_svref(goperl_frame_t *f, SV **sptr) {
    if (!sptr) goperl_croakf(f, "save_svref of no slot");
    const uint64_t addr = goperl_guest_addr(f, (const void *)sptr);
    if (addr) return GOPERL_SV(GOPERL_INTERNAL(f, GOPERL_INT_SAVE_SVREF, addr, 0));
    /* the save takes the slot's reference to the value with it, and puts
     * the value back when the scope ends, releasing the new one */
    if (!goperl_save_routed(f, (void *)sptr, GOPERL_SAVEK_GENERIC, (int)sizeof(void *)))
        goperl_croakf(f, "save_svref: the slot is this process's copy of a value, not "
                         "the place the interpreter keeps it in (save_helem or "
                         "save_aelem saves an element of a hash or an array)");
    SV *const sv = newSV(0);
    *sptr = sv;
    return sv;
}
#define save_svref(sptr) goperl_save_svref(_gof, (SV **)(sptr))
#define Perl_save_svref(...) save_svref(__VA_ARGS__)

/* scan_vstring: the v-string spelled from s up to e (v1.22.333, or 1.22.333)
 * becomes sv, with its vstring magic, as the parser makes it; the end of
 * what was read is returned. The interpreter reads a copy of the spelling,
 * which ends there. */
#define GOPERL_INT_SCAN_VSTRING 38 /* a: the SV; s+b: the spelling -> the
                                    * number of bytes read */
static __attribute__((unused)) char *goperl_scan_vstring(goperl_frame_t *f, const char *s,
                                                         const char *e, SV *sv) {
    const uint64_t n = (uint64_t)(e - s);
    const uint64_t used = GOPERL_INTERNAL_S(f, GOPERL_INT_SCAN_VSTRING, GOPERL_TOK(sv), n, s, n);
    return (char *)s + (used <= n ? used : n);
}
#define scan_vstring(s, e, sv) goperl_scan_vstring(_gof, (const char *)(s), (const char *)(e), (SV *)(sv))

/* PL_stderrgv's handle, as perl's Perl_error_log is: the output handle of
 * STDERR as the program has it now (`open local(*STDERR), '>', \$buf`
 * captures what Test::LeakTrace reports), else the process's stderr */
static __attribute__((unused)) PerlIO *goperl_error_log(goperl_frame_t *f) {
    GV *const gv = PL_stderrgv;
    (void)f;
    if (gv && isGV(gv) && GvIOp(gv)) {
        PerlIO *const ofp = IoOFP(GvIOp(gv));
        if (ofp) return ofp;
    }
    return PerlIO_stderr();
}
#define Perl_error_log goperl_error_log(_gof)
/* every handle flushed, as before a fork or a command */
#define PERL_FLUSHALL_FOR_CHILD PerlIO_flush((PerlIO *)NULL)
#define Direntry_t struct dirent
/* a string the SV points at that is static (with SVf_IsCOW) */
#define SVppv_STATIC 0x40000000
/* the glob of a sub set (gv.c's cvgv_set, which CvGV_set is) */
#define cvgv_set(cv, gv) CvGV_set((cv), (gv))
/* a mask of the lowest n bits, and the largest number of n bits (handy.h) */
#define nBIT_MASK(n) (((n) == 0) ? 0 : (((UINTMAX_C(1) << ((n) - 1)) - 1) << 1) | 1)
#define nBIT_UMAX(n) nBIT_MASK(n)
#define CvISXSUB_on(cv) (CvFLAGS(cv) |= CVf_ISXSUB)
#define CvISXSUB_off(cv) (CvFLAGS(cv) &= ~CVf_ISXSUB)
/* TARG set to a number, as pp.h's (the cheap path of an IV target is the
 * same set) */
#ifndef TARGi
#define TARGi(i, do_taint) sv_setiv_mg(TARG, (IV)(i))
#define TARGu(u, do_taint) sv_setuv_mg(TARG, (UV)(u))
#define TARGn(n, do_taint) sv_setnv_mg(TARG, (NV)(n))
#endif

/* Block hooks. */
#define BHKf_bhk_start 0x01
#define BHKf_bhk_pre_end 0x02
#define BHKf_bhk_post_end 0x04
#define BHKf_bhk_eval 0x08
#define BhkFLAGS(hk) ((hk)->bhk_flags)
#define BhkENTRY(hk, which) ((BhkFLAGS(hk) & BHKf_##which) ? ((hk)->which) : NULL)
#define BhkENABLE(hk, which) ((void)(BhkFLAGS(hk) |= BHKf_##which))
#define BhkDISABLE(hk, which) ((void)(BhkFLAGS(hk) &= ~(BHKf_##which)))
#define BhkENTRY_set(hk, which, ptr) \
    ((void)((hk)->which = (ptr), BhkENABLE((hk), which)))
#define CALL_BLOCK_HOOKS(which, arg) \
    goperl_croakf(_gof, "CALL_BLOCK_HOOKS is the interpreter's")
static __attribute__((unused)) void goperl_blockhook_register(goperl_frame_t *f, BHK *hk) {
    goperl_fntab_t *t = goperl_fntab(f);
    if (t->nbhk >= GOPERL_BHK_MAX) goperl_croakf(f, "too many block hooks registered");
    const uint32_t id = (uint32_t)GOPERL_INTERNAL(f, GOPERL_INT_BHK_REGISTER, 0, hk->bhk_flags);
    if (id != (uint32_t)t->nbhk + 1) goperl_croakf(f, "block hooks out of step with the interpreter");
    t->bhks[t->nbhk] = hk;
    t->bhk_told[t->nbhk] = hk->bhk_flags;
    t->nbhk++;
}
#define Perl_blockhook_register(...) goperl_blockhook_register(_gof, __VA_ARGS__)
#define blockhook_register(hk) goperl_blockhook_register(_gof, (hk))
/* A hook enabled or disabled after it was registered (BhkDISABLE) is
 * called by the interpreter as it is now. */
static void goperl_bhk_sync(goperl_frame_t *f) {
    goperl_fntab_t *t = goperl_fntab_v;
    for (int32_t i = 0; t && i < t->nbhk; i++)
        if (t->bhks[i]->bhk_flags != t->bhk_told[i]) {
            t->bhk_told[i] = t->bhks[i]->bhk_flags;
            goperl_xs_opx(f, GOPERL_OP_INTERNAL, i + 1, t->bhk_told[i], (uint64_t)(GOPERL_INT_BHK_FLAGS), 0, 0, 0);
        }
}

/* The host-call kinds of these functions. */
enum { GOPERL_HOST_FILTER = 6, GOPERL_HOST_RUNOPS = 7, GOPERL_HOST_BHK = 8 };
static int32_t goperl_host_invoke_fn(goperl_frame_t *f, const goperl_gword_t *req,
                                     uint32_t reqlen, goperl_gword_t *resp) {
    /* req: [kind] and the kind's words; reqlen in bytes */
    switch (req[0]) {
    case GOPERL_HOST_FILTER: { /* [id][idx][buf][maxlen] -> [result] */
        if (reqlen < 40) goperl_croakf(f, "host call: short filter request");
        const filter_t fn = (filter_t)goperl_fn_of(f, (uint32_t)req[1]);
        goperl_flush(f);
        resp[0] = (goperl_gword_t)(int64_t)fn((int)(int64_t)req[2], GOPERL_SV(req[3]),
                                              (int)(int64_t)req[4]);
        return 8;
    }
    case GOPERL_HOST_RUNOPS: { /* [id] */
        if (reqlen < 16) goperl_croakf(f, "host call: short run loop request");
        const runops_proc_t fn = (runops_proc_t)goperl_fn_of(f, (uint32_t)req[1]);
        goperl_flush(f);
        /* The ops the loop runs work on the interpreter's stack, which the
         * frame has (the pp function of a custom op, LRU::Cache's, is the
         * module's own, which the loop calls itself) */
        fn();
        return 0;
    }
    case GOPERL_HOST_SIGHANDLER: { /* [id][signal of the interpreter's] */
        if (reqlen < 24) goperl_croakf(f, "host call: short signal handler request");
        const Sighandler_t fn = (Sighandler_t)goperl_fn_of(f, (uint32_t)req[1]);
        goperl_flush(f);
        fn(goperl_signal_host((int)(int64_t)req[2]));
        return 0;
    }
    case GOPERL_HOST_BHK: { /* [id][which][arg] -> [op] */
        const goperl_fntab_t *t = goperl_fntab_v;
        if (reqlen < 32 || !t || req[1] == 0 || (int64_t)req[1] > t->nbhk)
            goperl_croakf(f, "host call: no block hook for the request");
        BHK *hk = t->bhks[req[1] - 1];
        OP *o = (OP *)0;
        goperl_flush(f);
        switch (req[2]) {
        case 0:
            if (BhkENTRY(hk, bhk_start)) hk->bhk_start((int)(int64_t)req[3]);
            break;
        case 1:
        case 2:
            o = req[3] ? goperl_op_describe(f, req[3], GOPERL_OPDUMP_SUBTREE) : (OP *)0;
            if (req[2] == 1 && BhkENTRY(hk, bhk_pre_end)) hk->bhk_pre_end(&o);
            if (req[2] == 2 && BhkENTRY(hk, bhk_post_end)) hk->bhk_post_end(&o);
            if (o && !o->gop)
                goperl_croakf(f, "a block hook returned an op that is not the interpreter's");
            break;
        case 3:
            o = req[3] ? goperl_op_describe(f, req[3], GOPERL_OPDUMP_SUBTREE) : (OP *)0;
            if (BhkENTRY(hk, bhk_eval)) hk->bhk_eval(o);
            break;
        default:
            goperl_croakf(f, "host call: unknown block hook %u", (unsigned)req[2]);
        }
        goperl_optree_flush(f);
        resp[0] = o ? o->gop : 0;
        return 8;
    }
    }
    return -1;
}

/* The lexer's buffer, as the interpreter handles it. */
#define GOPERL_LEX_UNSTUFF 6
#define GOPERL_LEX_GROW_LINESTR 7
#define GOPERL_LEX_NEXT_CHUNK 8
#define GOPERL_LEX_DISCARD_TO 9
static __attribute__((unused)) void goperl_lex_at(goperl_frame_t *f, int sel, const char *ptr) {
    const uint64_t g = goperl_guest_addr(f, ptr);
    if (!g) goperl_croakf(f, "a lexer position outside of the buffer");
    goperl_do_op(f, GOPERL_OP_LEX, (uint64_t)sel, g, 0, 0);
}
#define lex_unstuff(ptr) goperl_lex_at(_gof, GOPERL_LEX_UNSTUFF, (const char *)(ptr))
#define lex_discard_to(ptr) goperl_lex_at(_gof, GOPERL_LEX_DISCARD_TO, (const char *)(ptr))
#define lex_grow_linestr(len) \
    ((char *)goperl_guest_mem(_gof, goperl_op0(GOPERL_OP_LEX, GOPERL_LEX_GROW_LINESTR, (STRLEN)(len))))
#define lex_next_chunk(flags) \
    ((bool)goperl_op0(GOPERL_OP_LEX, GOPERL_LEX_NEXT_CHUNK, (U32)(flags)))
#define Perl_lex_unstuff(ctx, ptr) lex_unstuff(ptr)

/* -n and -p, which the end of the program's text closes the loop of. */
#define GOPERL_PL_MINUS_N 97
#define GOPERL_PL_MINUS_P 98
#define PL_minus_n (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_MINUS_N))
#define PL_minus_p (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_MINUS_P))

/* The rest of what a copy of toke.c in a module uses. */
#define GOPERL_CALL_SV_CAT_DECODE 21
static __attribute__((unused)) bool goperl_sv_cat_decode(goperl_frame_t *f, SV *dsv, SV *encoding,
                                                         SV *ssv, int *offset, char *tstr, int tlen) {
    /* words [encoding][ssv][offset][tlen], then the terminator's bytes ->
     * the result; second: the offset */
    const goperl_gword_t w[4] = {GOPERL_TOK(encoding), GOPERL_TOK(ssv),
                                 (goperl_gword_t)(int64_t)*offset, (goperl_gword_t)(int64_t)tlen};
    const size_t n = sizeof w + (size_t)(tlen > 0 ? tlen : 0);
    char *req = (char *)malloc(n);
    if (!req) goperl_croakf(f, "Out of memory!");
    memcpy(req, w, sizeof w);
    if (tlen > 0) memcpy(req + sizeof w, tstr, (size_t)tlen);
    uint64_t off = 0;
    const uint64_t r = goperl_do_opx(f, GOPERL_OP_CALL, GOPERL_TOK(dsv), 0,
                                     GOPERL_CALL_SV_CAT_DECODE, req, n, &off);
    free(req);
    *offset = (int)(int64_t)off;
    return (r & 1) != 0;
}
#define sv_cat_decode(dsv, enc, ssv, offset, tstr, tlen) \
    goperl_sv_cat_decode(_gof, (SV *)(dsv), (SV *)(enc), (SV *)(ssv), (offset), (char *)(tstr), (int)(tlen))
#define Perl_sv_cat_decode(...) goperl_sv_cat_decode(_gof, __VA_ARGS__)
/* The encoding of the program's text: gone from perl (the encoding pragma
 * no longer sets it), as perl.h says outside the core. */
#define PL_encoding ((SV *)NULL)
/* -Dp's tracing: this perl is not a DEBUGGING build. */
#define DEBUG_P(a)
#define DEBUG_P_TEST 0
#ifndef isWORDCHAR_A
#define isWORDCHAR_A(c) isALNUM(c)
#endif
#ifndef isIDFIRST_lazy_if
#define isIDFIRST_lazy_if(p, UTF) \
    ((!(UTF) || UTF8_IS_INVARIANT(*(p))) ? isIDFIRST(*(p)) : isIDFIRST_utf8_safe((const U8 *)(p), (const U8 *)(p) + UTF8SKIP(p)))
#endif
#ifndef isALNUM_lazy_if
#define isALNUM_lazy_if(p, UTF) \
    ((!(UTF) || UTF8_IS_INVARIANT(*(p))) ? isALNUM(*(p)) : isIDCONT_utf8_safe((const U8 *)(p), (const U8 *)(p) + UTF8SKIP(p)))
#endif
#ifndef toCTRL
#define toCTRL(c) (toUPPER((U8)(c)) ^ 64)
#endif
/* The file of a COP is let go of (CopFILE_set gives it the next one). */
static __attribute__((unused)) void goperl_cop_file_free(goperl_frame_t *f, COP *c) {
    if (c->gop)
        goperl_do_op3(f, GOPERL_OP_CALL, c->gop, 0,
                      GOPERL_CALL_COP_SET | ((uint64_t)6 /* FILE_FREE */ << 32), 0, 0);
    c->cop_file = 0;
}
#define CopFILE_free(c) goperl_cop_file_free(_gof, (COP *)(c))
/* Whether the interpreter's allocator returns NULL instead of ending the
 * process when it has no memory. */
#define PL_nomemok (*(bool *)goperl_plint_ref(_gof, GOPERL_PL_NOMEMOK))

/* ---- case changes of characters (handy.h) ----------------------------------
 * The interpreter's: the result is the UTF-8 of the character in the other
 * case (possibly more than one character: the upper case of U+00DF is
 * "SS"), its length goes to *lenp, and the first code point of it is
 * returned. Nothing is written after the result, as with perl's: Tk
 * changes the case of a string in place, a character at a time, and a NUL
 * after the first character ended its string there (a search with
 * -nocase found nothing). */
#define GOPERL_CALL_CASEMAP 25
enum { GOPERL_CASE_UPPER = 0, GOPERL_CASE_LOWER = 1, GOPERL_CASE_TITLE = 2, GOPERL_CASE_FOLD = 3 };
static __attribute__((unused)) UV goperl_case_uvchr(goperl_frame_t *f, int which, UV cp,
                                                    U8 *s, STRLEN *lenp) {
    /* -> the address of [first][length] (two words) and the UTF-8 bytes */
    const goperl_gword_t *r = (const goperl_gword_t *)goperl_guest_mem(
        f, goperl_do_op3(f, GOPERL_OP_CALL, which, (uint64_t)cp, (uint64_t)(GOPERL_CALL_CASEMAP), 0, 0));
    if (!r) goperl_croakf(f, "the interpreter did not change the case of a character");
    const UV first = (UV)r[0];
    const STRLEN len = (STRLEN)r[1];
    memcpy(s, (const U8 *)(r + 2), len);
    if (lenp) *lenp = len;
    return first;
}
/* The same for the character at p (before e) of UTF-8 text. */
static __attribute__((unused)) UV goperl_case_utf8(goperl_frame_t *f, int which, const U8 *p,
                                                   const U8 *e, U8 *s, STRLEN *lenp) {
    if (p >= e) goperl_croakf(f, "Attempting case change past the end of a string");
    STRLEN clen = 0;
    const UV cp = utf8_to_uvchr_buf(p, e, &clen);
    return goperl_case_uvchr(f, which, cp, s, lenp);
}
#define toUPPER_uvchr(c, s, l) goperl_case_uvchr(_gof, GOPERL_CASE_UPPER, (UV)(c), (U8 *)(s), (l))
#define toLOWER_uvchr(c, s, l) goperl_case_uvchr(_gof, GOPERL_CASE_LOWER, (UV)(c), (U8 *)(s), (l))
#define toTITLE_uvchr(c, s, l) goperl_case_uvchr(_gof, GOPERL_CASE_TITLE, (UV)(c), (U8 *)(s), (l))
#define toFOLD_uvchr(c, s, l) goperl_case_uvchr(_gof, GOPERL_CASE_FOLD, (UV)(c), (U8 *)(s), (l))
#define to_uni_upper(c, s, l) toUPPER_uvchr((c), (s), (l))
#define to_uni_lower(c, s, l) toLOWER_uvchr((c), (s), (l))
#define to_uni_title(c, s, l) toTITLE_uvchr((c), (s), (l))
#define to_uni_fold(c, s, l) toFOLD_uvchr((c), (s), (l))
#define toUPPER_utf8_safe(p, e, s, l) \
    goperl_case_utf8(_gof, GOPERL_CASE_UPPER, (const U8 *)(p), (const U8 *)(e), (U8 *)(s), (l))
#define toLOWER_utf8_safe(p, e, s, l) \
    goperl_case_utf8(_gof, GOPERL_CASE_LOWER, (const U8 *)(p), (const U8 *)(e), (U8 *)(s), (l))
#define toTITLE_utf8_safe(p, e, s, l) \
    goperl_case_utf8(_gof, GOPERL_CASE_TITLE, (const U8 *)(p), (const U8 *)(e), (U8 *)(s), (l))
#define toFOLD_utf8_safe(p, e, s, l) \
    goperl_case_utf8(_gof, GOPERL_CASE_FOLD, (const U8 *)(p), (const U8 *)(e), (U8 *)(s), (l))
/* One character's UTF-8 by the interpreter's own functions: the bytes go
 * over up to the length the start byte announces (perl reads no further),
 * or the code point to encode, with the flags. Back come the words: the
 * result, the length (or -1), the errors, the messages (an array or a hash
 * the caller owns, or none), then the bytes of an encoded character. */
static __attribute__((noinline, cold)) const goperl_gword_t *
goperl_utf8_interp(goperl_frame_t *f, uint32_t what, U32 flags, const U8 *s, STRLEN n, uint64_t b) {
    const uint64_t r = GOPERL_INTERNAL_S(f, GOPERL_INT_UTF8, (uint64_t)flags | ((uint64_t)what << 32),
                                         b, s, s ? n : 0);
    const goperl_gword_t *const w = (const goperl_gword_t *)goperl_guest_mem(f, r);
    if (!w) goperl_croakf(f, "the interpreter did not hand back a UTF-8 character");
    return w;
}
static __attribute__((noinline, cold)) UV goperl_utf8_decode_interp(goperl_frame_t *f, const U8 *s,
                                                                STRLEN curlen, STRLEN *retlen,
                                                                U32 flags, U32 *errors, AV **msgs,
                                                                uint32_t mode) {
    const STRLEN skip = curlen ? UTF8SKIP(s) : 0;
    const STRLEN n = curlen < skip ? curlen : skip;
    const goperl_gword_t *const w =
        goperl_utf8_interp(f, (mode ? GOPERL_UTF8_DECODE_BUF : GOPERL_UTF8_DECODE) |
                                  (msgs ? GOPERL_UTF8_MSGS : 0),
                           flags, s, n, n);
    if (retlen) *retlen = (STRLEN)w[1];
    if (errors) *errors = (U32)w[2];
    if (msgs) *msgs = (AV *)GOPERL_SV(w[3]);
    return (UV)w[0];
}
/* isUTF8_CHAR_flags: the length of the character at s under the flags, or 0
 * when it is no complete, allowed one (inline.h) */
static inline __attribute__((always_inline, unused)) STRLEN goperl_isutf8_char_flags(const U8 *s,
                                                                                   const U8 *e,
                                                                                   U32 flags) {
    UV cp;
    STRLEN len;
    if (e <= s) return 0;
    if (goperl_utf8_strict(s, (STRLEN)(e - s), &cp, &len)) return len;
    {
        const STRLEN skip = UTF8SKIP(s);
        const STRLEN n = (STRLEN)(e - s) < skip ? (STRLEN)(e - s) : skip;
        return (STRLEN)goperl_utf8_interp(_gof, GOPERL_UTF8_CHAR, flags, s, n, n)[0];
    }
}
#define isUTF8_CHAR_flags(s, e, flags) \
    goperl_isutf8_char_flags((const U8 *)(s), (const U8 *)(e), (U32)(flags))
/* is_utf8_valid_partial_char_flags: whether s..e is the start of a
 * character allowed under the flags that goes on beyond e */
static inline __attribute__((unused)) bool goperl_utf8_valid_partial(const U8 *s, const U8 *e,
                                                                     U32 flags) {
    if (s >= e || (STRLEN)(e - s) >= UTF8SKIP(s)) return FALSE;
    return goperl_utf8_interp(_gof, GOPERL_UTF8_PARTIAL, flags, s, (STRLEN)(e - s),
                              (uint64_t)(e - s))[0] != 0;
}
#define is_utf8_valid_partial_char_flags(s, e, flags) \
    goperl_utf8_valid_partial((const U8 *)(s), (const U8 *)(e), (U32)(flags))
#define is_utf8_valid_partial_char(s, e) is_utf8_valid_partial_char_flags((s), (e), 0)
/* is_utf8_string_loclen_flags: whether the string is all characters allowed
 * under the flags, where the first that is not begins, and how many there
 * are before it; a length of 0 is the string's strlen, as perl has it */
static inline __attribute__((unused)) bool goperl_utf8_string_flags(const U8 *s, STRLEN len,
                                                                    const U8 **ep, STRLEN *el,
                                                                    U32 flags) {
    const U8 *p = s, *send;
    STRLEN count = 0;
    if (len == 0) len = strlen((const char *)s);
    send = s + len;
    while (p < send) {
        const STRLEN l = goperl_isutf8_char_flags(p, send, flags);
        if (!l) break;
        p += l;
        count++;
    }
    if (ep) *ep = p;
    if (el) *el = count;
    return p == send;
}
#define is_utf8_string_loclen_flags(s, len, ep, el, flags) \
    goperl_utf8_string_flags((const U8 *)(s), (STRLEN)(len), (const U8 **)(ep), (el), (U32)(flags))
#define is_utf8_string_loc_flags(s, len, ep, flags) \
    is_utf8_string_loclen_flags((s), (len), (ep), (STRLEN *)0, (flags))
#define is_utf8_string_flags(s, len, flags) \
    is_utf8_string_loclen_flags((s), (len), (const U8 **)0, (STRLEN *)0, (flags))
/* Strict UTF-8: no surrogate, no noncharacter and nothing above Unicode;
 * C9 strict: noncharacters allowed (Unicode Corrigendum #9). */
#define is_strict_utf8_string_loclen(s, len, ep, el) \
    is_utf8_string_loclen_flags((s), (len), (ep), (el), UTF8_DISALLOW_ILLEGAL_INTERCHANGE)
#define is_strict_utf8_string_loc(s, len, ep) is_strict_utf8_string_loclen((s), (len), (ep), (STRLEN *)0)
#define is_strict_utf8_string(s, len) is_strict_utf8_string_loclen((s), (len), (const U8 **)0, (STRLEN *)0)
#define is_c9strict_utf8_string_loclen(s, len, ep, el) \
    is_utf8_string_loclen_flags((s), (len), (ep), (el), UTF8_DISALLOW_ILLEGAL_C9_INTERCHANGE)
#define is_c9strict_utf8_string_loc(s, len, ep) \
    is_c9strict_utf8_string_loclen((s), (len), (ep), (STRLEN *)0)
#define is_c9strict_utf8_string(s, len) \
    is_c9strict_utf8_string_loclen((s), (len), (const U8 **)0, (STRLEN *)0)
/* The string is valid UTF-8 up to a character its end cuts off. */
#define is_utf8_fixed_width_buf_loclen_flags(s, len, ep, el, flags)          \
    ({ const U8 *const goperl_fwb_s_ = (const U8 *)(s);                     \
       const STRLEN goperl_fwb_len_ = (STRLEN)(len);                        \
       const U8 *goperl_fwb_e_;                                             \
       const U8 **const goperl_fwb_ep_ = (ep) ? (const U8 **)(ep) : &goperl_fwb_e_; \
       is_utf8_string_loclen_flags(goperl_fwb_s_, goperl_fwb_len_, goperl_fwb_ep_, (el), (flags)) || \
           is_utf8_valid_partial_char_flags(*goperl_fwb_ep_, goperl_fwb_s_ + goperl_fwb_len_, (flags)); })
#define is_utf8_fixed_width_buf_loc_flags(s, len, loc, flags) \
    is_utf8_fixed_width_buf_loclen_flags((s), (len), (loc), (STRLEN *)0, (flags))
#define is_utf8_fixed_width_buf_flags(s, len, flags) \
    is_utf8_fixed_width_buf_loclen_flags((s), (len), (const U8 **)0, (STRLEN *)0, (flags))
/* uvchr_to_utf8_flags_msgs: perl's encoder (utf8.c). A code point that is
 * no surrogate, no noncharacter and not above Unicode, or any one up to
 * IV_MAX under no flags, is encoded here; any other by the interpreter,
 * which warns, refuses (NULL) or gives the message under the flags as perl
 * does, and dies above IV_MAX. */
static inline __attribute__((unused)) U8 *goperl_uvchr_to_utf8_flags_msgs(U8 *d, UV uv, UV flags,
                                                                          HV **msgs) {
    if (msgs) *msgs = NULL;
    if ((flags == 0 && uv <= (UV)IV_MAX) || uv < 0xD800 || (uv >= 0xE000 && uv < 0xFDD0) ||
        (uv > 0xFDEF && uv <= 0x10FFFF && (uv & 0xFFFE) != 0xFFFE))
        return goperl_uvchr_to_utf8(d, uv);
    {
        const goperl_gword_t *const w =
            goperl_utf8_interp(_gof, GOPERL_UTF8_ENCODE | (msgs ? GOPERL_UTF8_MSGS : 0), (U32)flags,
                               (const U8 *)0, 0, (uint64_t)uv);
        const STRLEN n = (STRLEN)w[1];
        if (msgs) *msgs = (HV *)GOPERL_SV(w[3]);
        if (n == (STRLEN)-1) return (U8 *)0;
        memcpy(d, (const U8 *)(w + 4), n);
        return d + n;
    }
}
#define uvchr_to_utf8_flags_msgs(d, uv, flags, msgs) \
    goperl_uvchr_to_utf8_flags_msgs((U8 *)(d), (UV)(uv), (UV)(flags), (msgs))
#define uvchr_to_utf8_flags(d, uv, flags) uvchr_to_utf8_flags_msgs((d), (uv), (flags), (HV **)0)
/* foldEQ_utf8 (utf8.c): whether the leading portions of two strings,
 * either or both of which may be UTF-8, are the same case-insensitively,
 * how far into them given by l1/pe1 and l2/pe2 as perl has it, and on a
 * match where each scan ended into *pe1 and *pe2. It is perl's algorithm;
 * the full fold of a character past ASCII is the interpreter's
 * (toFOLD_uvchr). Locale folding (FOLDEQ_LOCALE) is not the SDK's: the
 * strings are compared as under no locale. */
#ifndef FOLDEQ_UTF8_NOMIX_ASCII
#define FOLDEQ_UTF8_NOMIX_ASCII (1 << 0)
#define FOLDEQ_LOCALE (1 << 1)
#define FOLDEQ_S1_ALREADY_FOLDED (1 << 2)
#define FOLDEQ_S2_ALREADY_FOLDED (1 << 3)
#define FOLDEQ_S1_FOLDS_SANE (1 << 4)
#define FOLDEQ_S2_FOLDS_SANE (1 << 5)
#endif
/* the fold of the character at p of a string (UTF-8 or not), in UTF-8 */
static __attribute__((unused)) const U8 *goperl_fold_char(goperl_frame_t *f, const U8 *p,
                                                          const U8 *e, bool u, U8 *buf,
                                                          STRLEN *n) {
    if (isASCII(*p)) {
        *buf = (U8)toFOLD(*p);
        *n = 1;
    } else if (u) {
        goperl_case_utf8(f, GOPERL_CASE_FOLD, p, e, buf, n);
    } else {
        goperl_case_uvchr(f, GOPERL_CASE_FOLD, (UV)*p, buf, n);
    }
    return buf;
}
static __attribute__((unused)) I32 goperl_foldEQ_utf8_flags(goperl_frame_t *f, const char *s1,
                                                            char **pe1, UV l1, bool u1,
                                                            const char *s2, char **pe2,
                                                            UV l2, bool u2, U32 flags) {
    const U8 *p1 = (const U8 *)s1, *p2 = (const U8 *)s2;
    const U8 *g1 = 0, *g2 = 0, *e1 = 0, *e2 = 0;
    const U8 *f1 = 0, *f2 = 0;
    STRLEN n1 = 0, n2 = 0;
    U8 foldbuf1[UTF8_MAXBYTES_CASE + 1], foldbuf2[UTF8_MAXBYTES_CASE + 1];
    if (pe1) e1 = *(U8 **)pe1;
    if (l1) g1 = (const U8 *)s1 + l1;
    if (pe2) e2 = *(U8 **)pe2;
    if (l2) g2 = (const U8 *)s2 + l2;
    if (!g1 && !g2) goperl_croakf(f, "foldEQ_utf8: neither string has a goal (l1, l2)");
    if (g1) e1 = g1;
    else if (!e1) goperl_croakf(f, "foldEQ_utf8: the first string has no end");
    if (g2) e2 = g2;
    else if (!e2) goperl_croakf(f, "foldEQ_utf8: the second string has no end");
    while (p1 < e1 && p2 < e2) {
        if (n1 == 0) {
            if (flags & FOLDEQ_S1_ALREADY_FOLDED) {
                f1 = p1;
                n1 = UTF8SKIP(f1);
            } else {
                if (isASCII(*p1) && (flags & FOLDEQ_UTF8_NOMIX_ASCII) && !isASCII(*p2))
                    return 0;
                f1 = goperl_fold_char(f, p1, e1, u1, foldbuf1, &n1);
            }
        }
        if (n2 == 0) {
            if (flags & FOLDEQ_S2_ALREADY_FOLDED) {
                if (UTF8_IS_INVARIANT(*p2)) {
                    f2 = p2;
                    n2 = 1;
                } else if (u2) {
                    f2 = p2;
                    n2 = UTF8SKIP(f2);
                } else {
                    foldbuf2[0] = UTF8_EIGHT_BIT_HI(*p2);
                    foldbuf2[1] = UTF8_EIGHT_BIT_LO(*p2);
                    f2 = foldbuf2;
                    n2 = 2;
                }
            } else {
                if (isASCII(*p2) && (flags & FOLDEQ_UTF8_NOMIX_ASCII) && !isASCII(*p1))
                    return 0;
                f2 = goperl_fold_char(f, p2, e2, u2, foldbuf2, &n2);
            }
        }
        /* the folds are compared a character at a time */
        while (n1 && n2) {
            const STRLEN fold_length = UTF8SKIP(f1);
            if (fold_length != (STRLEN)UTF8SKIP(f2) || memNE((const char *)f1, (const char *)f2, fold_length))
                return 0;
            n1 -= fold_length;
            f1 += fold_length;
            n2 -= fold_length;
            f2 += fold_length;
        }
        if (n1 == 0) p1 += u1 ? UTF8SKIP(p1) : 1;
        if (n2 == 0) p2 += u2 ? UTF8SKIP(p2) : 1;
    }
    /* a match: each scan with a goal reached it, and no fold is left half
     * matched */
    if (!((g1 == 0 || p1 == g1) && (g2 == 0 || p2 == g2)) || n1 || n2) return 0;
    if (pe1) *pe1 = (char *)p1;
    if (pe2) *pe2 = (char *)p2;
    return 1;
}
#define foldEQ_utf8_flags(s1, pe1, l1, u1, s2, pe2, l2, u2, flags)                         \
    goperl_foldEQ_utf8_flags(_gof, (const char *)(s1), (char **)(pe1), (UV)(l1), cBOOL(u1), \
                             (const char *)(s2), (char **)(pe2), (UV)(l2), cBOOL(u2), (U32)(flags))
#define foldEQ_utf8(s1, pe1, l1, u1, s2, pe2, l2, u2) \
    foldEQ_utf8_flags((s1), (pe1), (l1), (u1), (s2), (pe2), (l2), (u2), 0)
#define toUPPER_utf8(p, e, s, l) toUPPER_utf8_safe((p), (e), (s), (l))
#define toLOWER_utf8(p, e, s, l) toLOWER_utf8_safe((p), (e), (s), (l))
#define toTITLE_utf8(p, e, s, l) toTITLE_utf8_safe((p), (e), (s), (l))
#define toFOLD_utf8(p, e, s, l) toFOLD_utf8_safe((p), (e), (s), (l))

/* ---- more interpreter variables -------------------------------------------- */
#define GOPERL_PL_IN_CLEAN_OBJS 109
#define GOPERL_PL_IN_CLEAN_ALL 110
#define GOPERL_PL_ORIGARGC 111
#define GOPERL_PL_ORIGARGV 112
#define GOPERL_PL_ORS 113
/* The output record separator ($\) print appends, NULL for none: code
 * that prints for itself sets it aside (Text::CSV_XS: `ors = PL_ors_sv;
 * PL_ors_sv = NULL; ...; PL_ors_sv = ors`). */
#define PL_ors_sv (*goperl_plsv_ref(_gof, GOPERL_PL_ORS))
/* Global destruction: the objects are being destroyed (sv_clean_objs), or
 * every SV is (sv_clean_all). */
#define PL_in_clean_objs (GOPERL_PLVAR(GOPERL_PL_IN_CLEAN_OBJS) != 0)
#define PL_in_clean_all (GOPERL_PLVAR(GOPERL_PL_IN_CLEAN_ALL) != 0)
/* The argument vector the interpreter was started with: the strings are
 * the interpreter's own (writing into them is what an assignment to $0
 * does), the vector of their addresses here is a view made for the native
 * call. */
#define PL_origargc ((int)GOPERL_PLVAR(GOPERL_PL_ORIGARGC))
static __attribute__((unused)) char **goperl_origargv(goperl_frame_t *f) {
    const int argc = (int)goperl_do_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_ORIGARGC, 0, 0, 0);
    const uint64_t g = goperl_do_op(f, GOPERL_OP_PLVAR_GET, GOPERL_PL_ORIGARGV, 0, 0, 0);
    if (!g || argc < 0) return 0;
    char **v = (char **)goperl_arena_alloc(sizeof(char *) * (size_t)(argc + 1));
    if (!v) goperl_croakf(f, "Out of memory!");
    for (int i = 0; i < argc; i++)
        v[i] = (char *)goperl_guest_mem(
            f, goperl_g64(f, g + GOPERL_GUEST_PTRSIZE * (uint64_t)i));
    v[argc] = 0;
    return v;
}
#define PL_origargv goperl_origargv(_gof)

/* ---- the scanning functions perl gives B::Hooks::Parser ---------------------
 * (embed.h has them for PERL_CORE and PERL_EXT): they scan the lexer's
 * buffer from s, which is a place in it, and give the place after what they
 * read. */
#if defined(PERL_CORE) || defined(PERL_EXT)
#define GOPERL_LEX_SCAN_WORD 10
#define GOPERL_LEX_SCAN_STR 11
#define GOPERL_LEX_SKIPSPACE 12
/* A place in the lexer's buffer from the interpreter's address of it. */
static __attribute__((unused)) char *goperl_lex_place(goperl_frame_t *f, uint64_t g) {
    return g ? (char *)goperl_guest_mem(f, g) : 0;
}
static __attribute__((unused)) uint64_t goperl_lex_from(goperl_frame_t *f, const char *s) {
    const uint64_t g = goperl_guest_addr(f, s);
    if (!g) goperl_croakf(f, "a lexer position outside of the buffer");
    return g;
}
static __attribute__((unused)) char *goperl_scan_word(goperl_frame_t *f, char *s, char *dest,
                                                      STRLEN destlen, int allow_package,
                                                      STRLEN *slp) {
    /* -> the address of [the place after the word][its length] (words) and
     * the word */
    const goperl_gword_t w[2] = {(goperl_gword_t)destlen, (goperl_gword_t)(int64_t)allow_package};
    const uint64_t from = goperl_lex_from(f, s);
    const goperl_gword_t *r = (const goperl_gword_t *)goperl_guest_mem(
        f, (uint64_t)goperl_lex(f, GOPERL_LEX_SCAN_WORD, from, (const char *)w, sizeof w));
    if (!r) goperl_croakf(f, "the interpreter did not scan the word");
    const STRLEN len = (STRLEN)r[1];
    if (len < destlen) {
        memcpy(dest, (const char *)(r + 2), len);
        dest[len] = 0;
    }
    *slp = len;
    return goperl_lex_place(f, r[0]);
}
#define scan_word(s, dest, destlen, allow_package, slp) \
    goperl_scan_word(_gof, (s), (dest), (STRLEN)(destlen), (allow_package), (slp))
static __attribute__((unused)) char *goperl_scan_str(goperl_frame_t *f, char *start,
                                                     int keep_quoted, int keep_delims,
                                                     int re_reparse, char **delimp) {
    const goperl_gword_t w[3] = {(goperl_gword_t)(int64_t)keep_quoted,
                                 (goperl_gword_t)(int64_t)keep_delims,
                                 (goperl_gword_t)(int64_t)re_reparse};
    const uint64_t from = goperl_lex_from(f, start);
    const goperl_gword_t *r = (const goperl_gword_t *)goperl_guest_mem(
        f, (uint64_t)goperl_lex(f, GOPERL_LEX_SCAN_STR, from, (const char *)w, sizeof w));
    if (!r) goperl_croakf(f, "the interpreter did not scan the string");
    if (delimp) *delimp = goperl_lex_place(f, r[1]);
    return goperl_lex_place(f, r[0]);
}
#define scan_str(start, keep_quoted, keep_delims, re_reparse, delimp) \
    goperl_scan_str(_gof, (start), (keep_quoted), (keep_delims), (re_reparse), (delimp))
static __attribute__((unused)) char *goperl_skipspace_flags(goperl_frame_t *f, char *s,
                                                            U32 flags) {
    const goperl_gword_t w[1] = {(goperl_gword_t)flags};
    const uint64_t from = goperl_lex_from(f, s);
    return goperl_lex_place(
        f, (uint64_t)goperl_lex(f, GOPERL_LEX_SKIPSPACE, from, (const char *)w, sizeof w));
}
#define skipspace_flags(s, flags) goperl_skipspace_flags(_gof, (s), (U32)(flags))
#endif

#endif /* GOPERL_XS_SDK_COMPILER_H */
