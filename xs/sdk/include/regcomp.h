/* go-perl native XS SDK: perl's regcomp.h describes the program the regexp
 * compiler makes (the regnodes) and its private structures. On go-perl a
 * compiled regexp is the interpreter's, in its own memory, and native code
 * cannot walk its program, so a module that needs this header cannot be
 * built: the build stops here, saying so, instead of compiling code that
 * reads the program wrongly. */
#ifndef GOPERL_XS_SDK_REGCOMP_H
#define GOPERL_XS_SDK_REGCOMP_H
#error "regcomp.h: the program of a compiled regexp (its regnodes) is the interpreter's own on go-perl, and native code cannot walk it"
#endif
