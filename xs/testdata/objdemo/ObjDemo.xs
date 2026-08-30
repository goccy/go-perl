#include <string>

/* A compact C++ fixture pinning the SDK v2 surface: a T_PTROBJ-style native
 * object (host pointer crossing as a registry id inside a blessed IV ref),
 * plus AV/HV construction, hash stores/fetches, and sv_bless. */

class Counter {
public:
    Counter(const char *label) : label_(label), n_(0) {}
    long long incr(long long by) { n_ += by; return n_; }
    const std::string &label() const { return label_; }
    long long n() const { return n_; }

private:
    std::string label_;
    long long n_;
};

#ifdef __cplusplus
extern "C" {
#endif
#define PERL_NO_GET_CONTEXT
#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#ifdef __cplusplus
};
#endif

typedef Counter * Obj_Demo;

MODULE = Obj::Demo  PACKAGE = Obj::Demo
PROTOTYPES: DISABLE

Obj_Demo
_new(classname, label)
	char *classname
	const char *label
CODE:
{
	RETVAL = new Counter(label);
}
OUTPUT:
	RETVAL

void
DESTROY(self)
	Obj_Demo self
CODE:
{
	delete self;
}

IV
incr(self, by)
	Obj_Demo self
	IV by
CODE:
{
	RETVAL = self->incr(by);
}
OUTPUT:
	RETVAL

SV *
stats(self)
	Obj_Demo self
CODE:
{
	HV *hash = newHV();
	(void)hv_stores(hash, "label", newSVpv(self->label().c_str(), self->label().size()));
	(void)hv_stores(hash, "count", newSVuv((UV)self->n()));
	HV *stash = gv_stashpv("Obj::Demo::Stats", 1);
	RETVAL = SvREFCNT_inc(sv_bless(sv_2mortal(newRV_inc((SV*)hash)), stash));
}
OUTPUT:
	RETVAL
