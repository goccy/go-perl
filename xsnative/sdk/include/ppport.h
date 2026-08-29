/* go-perl native XS SDK: ppport.h placeholder.
 *
 * Dists include "ppport.h" unconditionally (often generating it at build
 * time via Devel::PPPort). The real file back-ports REAL perl APIs across
 * perl versions - meaningless against this SDK, whose perl.h already
 * pre-defines the include guard (_P_P_PORTABILITY_H_). This file exists so
 * the #include resolves; a dist-bundled ppport.h is likewise neutralized by
 * the guard. */
