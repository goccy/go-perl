/* The layout Layout.c's module says it was compiled with: entries of 32
 * bytes, as the SDK's HE had while it carried a field of its own. This
 * definition is a strong one and takes the place of the SDK's (weak). */
#include <stdint.h>

const uint32_t *__goperl_xs_layout(uint32_t *n) {
    static const uint32_t pairs[] = {23, 0, 24, 8, 25, 16, 26, 32};
    *n = 4;
    return pairs;
}
