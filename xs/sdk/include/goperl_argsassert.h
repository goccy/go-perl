/* Generated from perl's proto.h by the SDK's maintainers: the argument
 * checks of the interpreter's functions (PERL_ARGS_ASSERT_), for modules
 * that copy one of those functions. */
#ifndef GOPERL_XS_SDK_ARGSASSERT_H
#define GOPERL_XS_SDK_ARGSASSERT_H
#ifndef PERL_ARGS_ASSERT_ABORT_EXECUTION
#define PERL_ARGS_ASSERT_ABORT_EXECUTION assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_ALLOCMY
#define PERL_ARGS_ASSERT_ALLOCMY assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_ALLOC_LOGOP
#define PERL_ARGS_ASSERT_ALLOC_LOGOP
#endif
#ifndef PERL_ARGS_ASSERT_AMAGIC_APPLIES
#define PERL_ARGS_ASSERT_AMAGIC_APPLIES assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_AMAGIC_CALL
#define PERL_ARGS_ASSERT_AMAGIC_CALL assert(left); assert(right)
#endif
#ifndef PERL_ARGS_ASSERT_AMAGIC_DEREF_CALL
#define PERL_ARGS_ASSERT_AMAGIC_DEREF_CALL assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_AMAGIC_IS_ENABLED
#define PERL_ARGS_ASSERT_AMAGIC_IS_ENABLED
#endif
#ifndef PERL_ARGS_ASSERT_API_VERSION_ASSERT
#define PERL_ARGS_ASSERT_API_VERSION_ASSERT assert(api_version)
#endif
#ifndef PERL_ARGS_ASSERT_APPLY
#define PERL_ARGS_ASSERT_APPLY assert(mark); assert(sp)
#endif
#ifndef PERL_ARGS_ASSERT_APPLY_ATTRS_STRING
#define PERL_ARGS_ASSERT_APPLY_ATTRS_STRING assert(stashpv); assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(attrstr)
#endif
#ifndef PERL_ARGS_ASSERT_APPLY_BUILTIN_CV_ATTRIBUTES
#define PERL_ARGS_ASSERT_APPLY_BUILTIN_CV_ATTRIBUTES assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_ATFORK_CHILD
#define PERL_ARGS_ASSERT_ATFORK_CHILD
#endif
#ifndef PERL_ARGS_ASSERT_ATFORK_LOCK
#define PERL_ARGS_ASSERT_ATFORK_LOCK
#endif
#ifndef PERL_ARGS_ASSERT_ATFORK_UNLOCK
#define PERL_ARGS_ASSERT_ATFORK_UNLOCK
#endif
#ifndef PERL_ARGS_ASSERT_AV_ARYLEN_P
#define PERL_ARGS_ASSERT_AV_ARYLEN_P assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_CLEAR
#define PERL_ARGS_ASSERT_AV_CLEAR assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_CREATE_AND_PUSH
#define PERL_ARGS_ASSERT_AV_CREATE_AND_PUSH assert(avp); assert(val)
#endif
#ifndef PERL_ARGS_ASSERT_AV_CREATE_AND_UNSHIFT_ONE
#define PERL_ARGS_ASSERT_AV_CREATE_AND_UNSHIFT_ONE assert(avp); assert(val)
#endif
#ifndef PERL_ARGS_ASSERT_AV_DELETE
#define PERL_ARGS_ASSERT_AV_DELETE assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_DUMP
#define PERL_ARGS_ASSERT_AV_DUMP assert(!av || SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_EXISTS
#define PERL_ARGS_ASSERT_AV_EXISTS assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_EXTEND
#define PERL_ARGS_ASSERT_AV_EXTEND assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_EXTEND_GUTS
#define PERL_ARGS_ASSERT_AV_EXTEND_GUTS assert(!av || SvTYPE(av) == SVt_PVAV); assert(maxp); assert(allocp); assert(arrayp)
#endif
#ifndef PERL_ARGS_ASSERT_AV_FETCH
#define PERL_ARGS_ASSERT_AV_FETCH assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_FILL
#define PERL_ARGS_ASSERT_AV_FILL assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_ITER_P
#define PERL_ARGS_ASSERT_AV_ITER_P assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_LEN
#define PERL_ARGS_ASSERT_AV_LEN assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_MAKE
#define PERL_ARGS_ASSERT_AV_MAKE assert(strp)
#endif
#ifndef PERL_ARGS_ASSERT_AV_NONELEM
#define PERL_ARGS_ASSERT_AV_NONELEM assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_POP
#define PERL_ARGS_ASSERT_AV_POP assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_PUSH
#define PERL_ARGS_ASSERT_AV_PUSH assert(av); assert(SvTYPE(av) == SVt_PVAV); assert(val)
#endif
#ifndef PERL_ARGS_ASSERT_AV_REIFY
#define PERL_ARGS_ASSERT_AV_REIFY assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_SHIFT
#define PERL_ARGS_ASSERT_AV_SHIFT assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_STORE
#define PERL_ARGS_ASSERT_AV_STORE assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_UNDEF
#define PERL_ARGS_ASSERT_AV_UNDEF assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_AV_UNSHIFT
#define PERL_ARGS_ASSERT_AV_UNSHIFT assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_BIND_MATCH
#define PERL_ARGS_ASSERT_BIND_MATCH assert(left); assert(right)
#endif
#ifndef PERL_ARGS_ASSERT_BLOCKHOOK_REGISTER
#define PERL_ARGS_ASSERT_BLOCKHOOK_REGISTER assert(hk)
#endif
#ifndef PERL_ARGS_ASSERT_BLOCK_END
#define PERL_ARGS_ASSERT_BLOCK_END
#endif
#ifndef PERL_ARGS_ASSERT_BLOCK_GIMME
#define PERL_ARGS_ASSERT_BLOCK_GIMME
#endif
#ifndef PERL_ARGS_ASSERT_BLOCK_START
#define PERL_ARGS_ASSERT_BLOCK_START
#endif
#ifndef PERL_ARGS_ASSERT_BOOT_CORE_BUILTIN
#define PERL_ARGS_ASSERT_BOOT_CORE_BUILTIN
#endif
#ifndef PERL_ARGS_ASSERT_BOOT_CORE_MRO
#define PERL_ARGS_ASSERT_BOOT_CORE_MRO
#endif
#ifndef PERL_ARGS_ASSERT_BOOT_CORE_PERLIO
#define PERL_ARGS_ASSERT_BOOT_CORE_PERLIO
#endif
#ifndef PERL_ARGS_ASSERT_BOOT_CORE_UNIVERSAL
#define PERL_ARGS_ASSERT_BOOT_CORE_UNIVERSAL
#endif
#ifndef PERL_ARGS_ASSERT_BUILD_INFIX_PLUGIN
#define PERL_ARGS_ASSERT_BUILD_INFIX_PLUGIN assert(lhs); assert(rhs); assert(tokendata)
#endif
#ifndef PERL_ARGS_ASSERT_BYTES_CMP_UTF8
#define PERL_ARGS_ASSERT_BYTES_CMP_UTF8 assert(b); assert(u)
#endif
#ifndef PERL_ARGS_ASSERT_BYTES_FROM_UTF8
#define PERL_ARGS_ASSERT_BYTES_FROM_UTF8 assert(s); assert(lenp); assert(is_utf8p)
#endif
#ifndef PERL_ARGS_ASSERT_BYTES_TO_UTF8_FREE_ME
#define PERL_ARGS_ASSERT_BYTES_TO_UTF8_FREE_ME assert(s); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_BYTE_DUMP_STRING_
#define PERL_ARGS_ASSERT_BYTE_DUMP_STRING_
#endif
#ifndef PERL_ARGS_ASSERT_C9STRICT_UTF8_TO_UV
#define PERL_ARGS_ASSERT_C9STRICT_UTF8_TO_UV assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_CALLER_CX
#define PERL_ARGS_ASSERT_CALLER_CX
#endif
#ifndef PERL_ARGS_ASSERT_CALLOC
#define PERL_ARGS_ASSERT_CALLOC
#endif
#ifndef PERL_ARGS_ASSERT_CALL_ARGV
#define PERL_ARGS_ASSERT_CALL_ARGV assert(sub_name); assert(argv)
#endif
#ifndef PERL_ARGS_ASSERT_CALL_ATEXIT
#define PERL_ARGS_ASSERT_CALL_ATEXIT
#endif
#ifndef PERL_ARGS_ASSERT_CALL_LIST
#define PERL_ARGS_ASSERT_CALL_LIST assert(paramList); assert(SvTYPE(paramList) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_CALL_METHOD
#define PERL_ARGS_ASSERT_CALL_METHOD assert(methname)
#endif
#ifndef PERL_ARGS_ASSERT_CALL_PV
#define PERL_ARGS_ASSERT_CALL_PV assert(sub_name)
#endif
#ifndef PERL_ARGS_ASSERT_CALL_SV
#define PERL_ARGS_ASSERT_CALL_SV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_CANDO
#define PERL_ARGS_ASSERT_CANDO assert(statbufp)
#endif
#ifndef PERL_ARGS_ASSERT_CAST_I32
#define PERL_ARGS_ASSERT_CAST_I32
#endif
#ifndef PERL_ARGS_ASSERT_CAST_IV
#define PERL_ARGS_ASSERT_CAST_IV
#endif
#ifndef PERL_ARGS_ASSERT_CAST_ULONG
#define PERL_ARGS_ASSERT_CAST_ULONG
#endif
#ifndef PERL_ARGS_ASSERT_CAST_UV
#define PERL_ARGS_ASSERT_CAST_UV
#endif
#ifndef PERL_ARGS_ASSERT_CHECK_UTF8_PRINT
#define PERL_ARGS_ASSERT_CHECK_UTF8_PRINT assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_CKWARN
#define PERL_ARGS_ASSERT_CKWARN
#endif
#ifndef PERL_ARGS_ASSERT_CKWARN_D
#define PERL_ARGS_ASSERT_CKWARN_D
#endif
#ifndef PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_CORE
#define PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_CORE assert(entersubop); assert(namegv); assert(protosv)
#endif
#ifndef PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_LIST
#define PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_LIST assert(entersubop)
#endif
#ifndef PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_PROTO
#define PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_PROTO assert(entersubop); assert(namegv); assert(protosv)
#endif
#ifndef PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_PROTO_OR_LIST
#define PERL_ARGS_ASSERT_CK_ENTERSUB_ARGS_PROTO_OR_LIST assert(entersubop); assert(namegv); assert(protosv)
#endif
#ifndef PERL_ARGS_ASSERT_CK_WARNER
#define PERL_ARGS_ASSERT_CK_WARNER assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_CK_WARNER_D
#define PERL_ARGS_ASSERT_CK_WARNER_D assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_CLEAR_DEFARRAY
#define PERL_ARGS_ASSERT_CLEAR_DEFARRAY assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_CLOSEST_COP
#define PERL_ARGS_ASSERT_CLOSEST_COP assert(cop)
#endif
#ifndef PERL_ARGS_ASSERT_CMPCHAIN_EXTEND
#define PERL_ARGS_ASSERT_CMPCHAIN_EXTEND assert(ch)
#endif
#ifndef PERL_ARGS_ASSERT_CMPCHAIN_FINISH
#define PERL_ARGS_ASSERT_CMPCHAIN_FINISH assert(ch)
#endif
#ifndef PERL_ARGS_ASSERT_CMPCHAIN_START
#define PERL_ARGS_ASSERT_CMPCHAIN_START
#endif
#ifndef PERL_ARGS_ASSERT_CNTRL_TO_MNEMONIC
#define PERL_ARGS_ASSERT_CNTRL_TO_MNEMONIC
#endif
#ifndef PERL_ARGS_ASSERT_COP_DISABLE_WARNING
#define PERL_ARGS_ASSERT_COP_DISABLE_WARNING assert(cop)
#endif
#ifndef PERL_ARGS_ASSERT_COP_ENABLE_WARNING
#define PERL_ARGS_ASSERT_COP_ENABLE_WARNING assert(cop)
#endif
#ifndef PERL_ARGS_ASSERT_COP_FETCH_LABEL
#define PERL_ARGS_ASSERT_COP_FETCH_LABEL assert(cop)
#endif
#ifndef PERL_ARGS_ASSERT_COP_HAS_WARNING
#define PERL_ARGS_ASSERT_COP_HAS_WARNING assert(cop)
#endif
#ifndef PERL_ARGS_ASSERT_COP_STORE_LABEL
#define PERL_ARGS_ASSERT_COP_STORE_LABEL assert(cop); assert(label)
#endif
#ifndef PERL_ARGS_ASSERT_CORESUB_OP
#define PERL_ARGS_ASSERT_CORESUB_OP assert(coreargssv)
#endif
#ifndef PERL_ARGS_ASSERT_CORE_PROTOTYPE
#define PERL_ARGS_ASSERT_CORE_PROTOTYPE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_CREATE_EVAL_SCOPE
#define PERL_ARGS_ASSERT_CREATE_EVAL_SCOPE assert(sp)
#endif
#ifndef PERL_ARGS_ASSERT_CROAK
#define PERL_ARGS_ASSERT_CROAK
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_CALLER
#define PERL_ARGS_ASSERT_CROAK_CALLER
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_MEMORY_WRAP
#define PERL_ARGS_ASSERT_CROAK_MEMORY_WRAP
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_NO_MEM
#define PERL_ARGS_ASSERT_CROAK_NO_MEM
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_NO_MEM_EXT
#define PERL_ARGS_ASSERT_CROAK_NO_MEM_EXT assert(context)
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_NO_MODIFY
#define PERL_ARGS_ASSERT_CROAK_NO_MODIFY
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_POPSTACK
#define PERL_ARGS_ASSERT_CROAK_POPSTACK
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_SV
#define PERL_ARGS_ASSERT_CROAK_SV assert(baseex)
#endif
#ifndef PERL_ARGS_ASSERT_CROAK_XS_USAGE
#define PERL_ARGS_ASSERT_CROAK_XS_USAGE assert(cv); assert(params)
#endif
#ifndef PERL_ARGS_ASSERT_CSIGHANDLER1
#define PERL_ARGS_ASSERT_CSIGHANDLER1
#endif
#ifndef PERL_ARGS_ASSERT_CSIGHANDLER3
#define PERL_ARGS_ASSERT_CSIGHANDLER3
#endif
#ifndef PERL_ARGS_ASSERT_CURRENT_RE_ENGINE
#define PERL_ARGS_ASSERT_CURRENT_RE_ENGINE
#endif
#ifndef PERL_ARGS_ASSERT_CUSTOM_OP_GET_FIELD
#define PERL_ARGS_ASSERT_CUSTOM_OP_GET_FIELD assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_CUSTOM_OP_REGISTER
#define PERL_ARGS_ASSERT_CUSTOM_OP_REGISTER assert(ppaddr); assert(xop)
#endif
#ifndef PERL_ARGS_ASSERT_CVGV_FROM_HEK
#define PERL_ARGS_ASSERT_CVGV_FROM_HEK assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CVGV_SET
#define PERL_ARGS_ASSERT_CVGV_SET assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CVSTASH_SET
#define PERL_ARGS_ASSERT_CVSTASH_SET assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(!stash || SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_CV_CKPROTO_LEN_FLAGS
#define PERL_ARGS_ASSERT_CV_CKPROTO_LEN_FLAGS assert(cv)
#endif
#ifndef PERL_ARGS_ASSERT_CV_CLONE
#define PERL_ARGS_ASSERT_CV_CLONE assert(proto); assert(SvTYPE(proto) == SVt_PVCV || SvTYPE(proto) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CV_CLONE_INTO
#define PERL_ARGS_ASSERT_CV_CLONE_INTO assert(proto); assert(SvTYPE(proto) == SVt_PVCV || SvTYPE(proto) == SVt_PVFM); assert(target); assert(SvTYPE(target) == SVt_PVCV || SvTYPE(target) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CV_CONST_SV
#define PERL_ARGS_ASSERT_CV_CONST_SV
#endif
#ifndef PERL_ARGS_ASSERT_CV_CONST_SV_OR_AV
#define PERL_ARGS_ASSERT_CV_CONST_SV_OR_AV
#endif
#ifndef PERL_ARGS_ASSERT_CV_FORGET_SLAB
#define PERL_ARGS_ASSERT_CV_FORGET_SLAB assert(!cv || SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CV_GET_CALL_CHECKER
#define PERL_ARGS_ASSERT_CV_GET_CALL_CHECKER assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(ckfun_p); assert(ckobj_p)
#endif
#ifndef PERL_ARGS_ASSERT_CV_GET_CALL_CHECKER_FLAGS
#define PERL_ARGS_ASSERT_CV_GET_CALL_CHECKER_FLAGS assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(ckfun_p); assert(ckobj_p); assert(ckflags_p)
#endif
#ifndef PERL_ARGS_ASSERT_CV_NAME
#define PERL_ARGS_ASSERT_CV_NAME assert(cv)
#endif
#ifndef PERL_ARGS_ASSERT_CV_SET_CALL_CHECKER
#define PERL_ARGS_ASSERT_CV_SET_CALL_CHECKER assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(ckfun); assert(ckobj)
#endif
#ifndef PERL_ARGS_ASSERT_CV_SET_CALL_CHECKER_FLAGS
#define PERL_ARGS_ASSERT_CV_SET_CALL_CHECKER_FLAGS assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(ckfun); assert(ckobj)
#endif
#ifndef PERL_ARGS_ASSERT_CV_UNDEF
#define PERL_ARGS_ASSERT_CV_UNDEF assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CV_UNDEF_FLAGS
#define PERL_ARGS_ASSERT_CV_UNDEF_FLAGS assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_CXINC
#define PERL_ARGS_ASSERT_CXINC
#endif
#ifndef PERL_ARGS_ASSERT_CX_DUMP
#define PERL_ARGS_ASSERT_CX_DUMP assert(cx)
#endif
#ifndef PERL_ARGS_ASSERT_DEB
#define PERL_ARGS_ASSERT_DEB assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_DEBOP
#define PERL_ARGS_ASSERT_DEBOP assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_DEBPROFDUMP
#define PERL_ARGS_ASSERT_DEBPROFDUMP
#endif
#ifndef PERL_ARGS_ASSERT_DEBSTACK
#define PERL_ARGS_ASSERT_DEBSTACK
#endif
#ifndef PERL_ARGS_ASSERT_DEBSTACKPTRS
#define PERL_ARGS_ASSERT_DEBSTACKPTRS
#endif
#ifndef PERL_ARGS_ASSERT_DEBUG_HASH_SEED
#define PERL_ARGS_ASSERT_DEBUG_HASH_SEED
#endif
#ifndef PERL_ARGS_ASSERT_DEB_STACK_ALL
#define PERL_ARGS_ASSERT_DEB_STACK_ALL
#endif
#ifndef PERL_ARGS_ASSERT_DEFELEM_TARGET
#define PERL_ARGS_ASSERT_DEFELEM_TARGET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_DEFER_ERROR
#define PERL_ARGS_ASSERT_DEFER_ERROR
#endif
#ifndef PERL_ARGS_ASSERT_DELETE_EVAL_SCOPE
#define PERL_ARGS_ASSERT_DELETE_EVAL_SCOPE
#endif
#ifndef PERL_ARGS_ASSERT_DELIMCPY
#define PERL_ARGS_ASSERT_DELIMCPY assert(to); assert(to_end); assert(from); assert(from_end); assert(retlen); assert(to <= to_end); assert(from <= from_end)
#endif
#ifndef PERL_ARGS_ASSERT_DELIMCPY_NO_ESCAPE
#define PERL_ARGS_ASSERT_DELIMCPY_NO_ESCAPE assert(to); assert(to_end); assert(from); assert(from_end); assert(retlen); assert(to <= to_end); assert(from <= from_end)
#endif
#ifndef PERL_ARGS_ASSERT_DESPATCH_SIGNALS
#define PERL_ARGS_ASSERT_DESPATCH_SIGNALS
#endif
#ifndef PERL_ARGS_ASSERT_DIE
#define PERL_ARGS_ASSERT_DIE
#endif
#ifndef PERL_ARGS_ASSERT_DIE_SV
#define PERL_ARGS_ASSERT_DIE_SV assert(baseex)
#endif
#ifndef PERL_ARGS_ASSERT_DIE_UNWIND
#define PERL_ARGS_ASSERT_DIE_UNWIND assert(msv)
#endif
#ifndef PERL_ARGS_ASSERT_DOFILE
#define PERL_ARGS_ASSERT_DOFILE assert(term)
#endif
#ifndef PERL_ARGS_ASSERT_DOING_TAINT
#define PERL_ARGS_ASSERT_DOING_TAINT
#endif
#ifndef PERL_ARGS_ASSERT_DOREF
#define PERL_ARGS_ASSERT_DOREF assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_DOUNWIND
#define PERL_ARGS_ASSERT_DOUNWIND
#endif
#ifndef PERL_ARGS_ASSERT_DOWANTARRAY
#define PERL_ARGS_ASSERT_DOWANTARRAY
#endif
#ifndef PERL_ARGS_ASSERT_DO_AEXEC5
#define PERL_ARGS_ASSERT_DO_AEXEC5 assert(mark); assert(sp)
#endif
#ifndef PERL_ARGS_ASSERT_DO_CLOSE
#define PERL_ARGS_ASSERT_DO_CLOSE
#endif
#ifndef PERL_ARGS_ASSERT_DO_DUMP_PAD
#define PERL_ARGS_ASSERT_DO_DUMP_PAD assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_DO_EOF
#define PERL_ARGS_ASSERT_DO_EOF assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_GVGV_DUMP
#define PERL_ARGS_ASSERT_DO_GVGV_DUMP assert(file); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_DO_GV_DUMP
#define PERL_ARGS_ASSERT_DO_GV_DUMP assert(file); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_DO_HV_DUMP
#define PERL_ARGS_ASSERT_DO_HV_DUMP assert(file); assert(name); assert(!sv || SvTYPE(sv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_DO_JOIN
#define PERL_ARGS_ASSERT_DO_JOIN assert(sv); assert(delim); assert(mark); assert(sp)
#endif
#ifndef PERL_ARGS_ASSERT_DO_MAGIC_DUMP
#define PERL_ARGS_ASSERT_DO_MAGIC_DUMP assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_DO_NCMP
#define PERL_ARGS_ASSERT_DO_NCMP assert(left); assert(right)
#endif
#ifndef PERL_ARGS_ASSERT_DO_OPEN
#define PERL_ARGS_ASSERT_DO_OPEN assert(gv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_DO_OPEN6
#define PERL_ARGS_ASSERT_DO_OPEN6 assert(gv); assert(oname)
#endif
#ifndef PERL_ARGS_ASSERT_DO_OPENN
#define PERL_ARGS_ASSERT_DO_OPENN assert(gv); assert(oname)
#endif
#ifndef PERL_ARGS_ASSERT_DO_OPEN_RAW
#define PERL_ARGS_ASSERT_DO_OPEN_RAW assert(gv); assert(oname)
#endif
#ifndef PERL_ARGS_ASSERT_DO_OP_DUMP
#define PERL_ARGS_ASSERT_DO_OP_DUMP assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_DO_PMOP_DUMP
#define PERL_ARGS_ASSERT_DO_PMOP_DUMP assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_DO_PRINT
#define PERL_ARGS_ASSERT_DO_PRINT assert(fp)
#endif
#ifndef PERL_ARGS_ASSERT_DO_READLINE
#define PERL_ARGS_ASSERT_DO_READLINE
#endif
#ifndef PERL_ARGS_ASSERT_DO_SEEK
#define PERL_ARGS_ASSERT_DO_SEEK
#endif
#ifndef PERL_ARGS_ASSERT_DO_SPRINTF
#define PERL_ARGS_ASSERT_DO_SPRINTF assert(sv); assert(sarg)
#endif
#ifndef PERL_ARGS_ASSERT_DO_SV_DUMP
#define PERL_ARGS_ASSERT_DO_SV_DUMP assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_DO_SYSSEEK
#define PERL_ARGS_ASSERT_DO_SYSSEEK assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_TELL
#define PERL_ARGS_ASSERT_DO_TELL assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_TRANS
#define PERL_ARGS_ASSERT_DO_TRANS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_UNIPROP_MATCH
#define PERL_ARGS_ASSERT_DO_UNIPROP_MATCH assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_DO_VECGET
#define PERL_ARGS_ASSERT_DO_VECGET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_VECSET
#define PERL_ARGS_ASSERT_DO_VECSET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_DO_VOP
#define PERL_ARGS_ASSERT_DO_VOP assert(sv); assert(left); assert(right)
#endif
#ifndef PERL_ARGS_ASSERT_DRAND48_INIT_R
#define PERL_ARGS_ASSERT_DRAND48_INIT_R assert(random_state)
#endif
#ifndef PERL_ARGS_ASSERT_DRAND48_R
#define PERL_ARGS_ASSERT_DRAND48_R assert(random_state)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_ALL
#define PERL_ARGS_ASSERT_DUMP_ALL
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_ALL_PERL
#define PERL_ARGS_ASSERT_DUMP_ALL_PERL
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_EVAL
#define PERL_ARGS_ASSERT_DUMP_EVAL
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_FORM
#define PERL_ARGS_ASSERT_DUMP_FORM assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_INDENT
#define PERL_ARGS_ASSERT_DUMP_INDENT assert(file); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_PACKSUBS
#define PERL_ARGS_ASSERT_DUMP_PACKSUBS assert(stash)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_PACKSUBS_PERL
#define PERL_ARGS_ASSERT_DUMP_PACKSUBS_PERL assert(stash)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_SUB
#define PERL_ARGS_ASSERT_DUMP_SUB assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_SUB_PERL
#define PERL_ARGS_ASSERT_DUMP_SUB_PERL assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_DUMP_VINDENT
#define PERL_ARGS_ASSERT_DUMP_VINDENT assert(file); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_DUP_WARNINGS
#define PERL_ARGS_ASSERT_DUP_WARNINGS
#endif
#ifndef PERL_ARGS_ASSERT_EMULATE_COP_IO
#define PERL_ARGS_ASSERT_EMULATE_COP_IO assert(c); assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_EVAL_PV
#define PERL_ARGS_ASSERT_EVAL_PV assert(p)
#endif
#ifndef PERL_ARGS_ASSERT_EVAL_SV
#define PERL_ARGS_ASSERT_EVAL_SV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_EXPECTED_SIZE
#define PERL_ARGS_ASSERT_EXPECTED_SIZE
#endif
#ifndef PERL_ARGS_ASSERT_EXTENDED_UTF8_TO_UV
#define PERL_ARGS_ASSERT_EXTENDED_UTF8_TO_UV assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_FATAL_WARNER
#define PERL_ARGS_ASSERT_FATAL_WARNER assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_FBM_COMPILE
#define PERL_ARGS_ASSERT_FBM_COMPILE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_FBM_INSTR
#define PERL_ARGS_ASSERT_FBM_INSTR assert(big); assert(bigend); assert(littlestr); assert(big <= bigend)
#endif
#ifndef PERL_ARGS_ASSERT_FILTER_ADD
#define PERL_ARGS_ASSERT_FILTER_ADD
#endif
#ifndef PERL_ARGS_ASSERT_FILTER_DEL
#define PERL_ARGS_ASSERT_FILTER_DEL assert(funcp)
#endif
#ifndef PERL_ARGS_ASSERT_FILTER_READ
#define PERL_ARGS_ASSERT_FILTER_READ assert(buf_sv)
#endif
#ifndef PERL_ARGS_ASSERT_FIND_LEXICAL_CV
#define PERL_ARGS_ASSERT_FIND_LEXICAL_CV
#endif
#ifndef PERL_ARGS_ASSERT_FIND_RUNCV
#define PERL_ARGS_ASSERT_FIND_RUNCV
#endif
#ifndef PERL_ARGS_ASSERT_FIND_RUNCV_WHERE
#define PERL_ARGS_ASSERT_FIND_RUNCV_WHERE
#endif
#ifndef PERL_ARGS_ASSERT_FIND_RUNDEFSV
#define PERL_ARGS_ASSERT_FIND_RUNDEFSV
#endif
#ifndef PERL_ARGS_ASSERT_FIND_SCRIPT
#define PERL_ARGS_ASSERT_FIND_SCRIPT assert(scriptname)
#endif
#ifndef PERL_ARGS_ASSERT_FOLDEQ_UTF8
#define PERL_ARGS_ASSERT_FOLDEQ_UTF8 assert(s1); assert(s2)
#endif
#ifndef PERL_ARGS_ASSERT_FOLDEQ_UTF8_FLAGS
#define PERL_ARGS_ASSERT_FOLDEQ_UTF8_FLAGS assert(s1); assert(s2)
#endif
#ifndef PERL_ARGS_ASSERT_FORBID_OUTOFBLOCK_OPS
#define PERL_ARGS_ASSERT_FORBID_OUTOFBLOCK_OPS assert(o); assert(blockname)
#endif
#ifndef PERL_ARGS_ASSERT_FORCE_LOCALE_UNLOCK
#define PERL_ARGS_ASSERT_FORCE_LOCALE_UNLOCK
#endif
#ifndef PERL_ARGS_ASSERT_FORCE_OUT_MALFORMED_UTF8_MESSAGE_
#define PERL_ARGS_ASSERT_FORCE_OUT_MALFORMED_UTF8_MESSAGE_ assert(p); assert(e); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_FORM
#define PERL_ARGS_ASSERT_FORM assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_FREE_TIED_HV_POOL
#define PERL_ARGS_ASSERT_FREE_TIED_HV_POOL
#endif
#ifndef PERL_ARGS_ASSERT_FREE_TMPS
#define PERL_ARGS_ASSERT_FREE_TMPS
#endif
#ifndef PERL_ARGS_ASSERT_GETCWD_SV
#define PERL_ARGS_ASSERT_GETCWD_SV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_GET_AND_CHECK_BACKSLASH_N_NAME
#define PERL_ARGS_ASSERT_GET_AND_CHECK_BACKSLASH_N_NAME assert(s); assert(e); assert(error_msg); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_GET_AV
#define PERL_ARGS_ASSERT_GET_AV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GET_CV
#define PERL_ARGS_ASSERT_GET_CV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GET_CVN_FLAGS
#define PERL_ARGS_ASSERT_GET_CVN_FLAGS assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GET_DB_SUB
#define PERL_ARGS_ASSERT_GET_DB_SUB assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_GET_DEPRECATED_PROPERTY_MSG
#define PERL_ARGS_ASSERT_GET_DEPRECATED_PROPERTY_MSG
#endif
#ifndef PERL_ARGS_ASSERT_GET_EXTENDED_OS_ERRNO
#define PERL_ARGS_ASSERT_GET_EXTENDED_OS_ERRNO
#endif
#ifndef PERL_ARGS_ASSERT_GET_HASH_SEED
#define PERL_ARGS_ASSERT_GET_HASH_SEED assert(seed_buffer)
#endif
#ifndef PERL_ARGS_ASSERT_GET_HV
#define PERL_ARGS_ASSERT_GET_HV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GET_NO_MODIFY
#define PERL_ARGS_ASSERT_GET_NO_MODIFY
#endif
#ifndef PERL_ARGS_ASSERT_GET_OPARGS
#define PERL_ARGS_ASSERT_GET_OPARGS
#endif
#ifndef PERL_ARGS_ASSERT_GET_OP_DESCS
#define PERL_ARGS_ASSERT_GET_OP_DESCS
#endif
#ifndef PERL_ARGS_ASSERT_GET_OP_NAMES
#define PERL_ARGS_ASSERT_GET_OP_NAMES
#endif
#ifndef PERL_ARGS_ASSERT_GET_PPADDR
#define PERL_ARGS_ASSERT_GET_PPADDR
#endif
#ifndef PERL_ARGS_ASSERT_GET_PROP_DEFINITION
#define PERL_ARGS_ASSERT_GET_PROP_DEFINITION
#endif
#ifndef PERL_ARGS_ASSERT_GET_PROP_VALUES
#define PERL_ARGS_ASSERT_GET_PROP_VALUES
#endif
#ifndef PERL_ARGS_ASSERT_GET_RE_ARG
#define PERL_ARGS_ASSERT_GET_RE_ARG
#endif
#ifndef PERL_ARGS_ASSERT_GET_SV
#define PERL_ARGS_ASSERT_GET_SV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GP_FREE
#define PERL_ARGS_ASSERT_GP_FREE
#endif
#ifndef PERL_ARGS_ASSERT_GP_REF
#define PERL_ARGS_ASSERT_GP_REF
#endif
#ifndef PERL_ARGS_ASSERT_GROK_ATOUV
#define PERL_ARGS_ASSERT_GROK_ATOUV assert(pv); assert(valptr)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_BIN_HEX
#define PERL_ARGS_ASSERT_GROK_BIN_HEX assert(start); assert(len_p); assert(flags)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_INFNAN
#define PERL_ARGS_ASSERT_GROK_INFNAN assert(sp); assert(*sp); assert(send); assert(*sp <= send)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_NUMBER
#define PERL_ARGS_ASSERT_GROK_NUMBER assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_NUMBER_FLAGS
#define PERL_ARGS_ASSERT_GROK_NUMBER_FLAGS assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_NUMERIC_RADIX
#define PERL_ARGS_ASSERT_GROK_NUMERIC_RADIX assert(sp); assert(*sp); assert(send); assert(*sp <= send)
#endif
#ifndef PERL_ARGS_ASSERT_GROK_UINT_BY_BASE
#define PERL_ARGS_ASSERT_GROK_UINT_BY_BASE assert(start); assert(len_p); assert(flags)
#endif
#ifndef PERL_ARGS_ASSERT_GV_ADD_BY_TYPE
#define PERL_ARGS_ASSERT_GV_ADD_BY_TYPE
#endif
#ifndef PERL_ARGS_ASSERT_GV_AMUPDATE
#define PERL_ARGS_ASSERT_GV_AMUPDATE assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_GV_AUTOLOAD_PV
#define PERL_ARGS_ASSERT_GV_AUTOLOAD_PV assert(namepv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_AUTOLOAD_PVN
#define PERL_ARGS_ASSERT_GV_AUTOLOAD_PVN assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_AUTOLOAD_SV
#define PERL_ARGS_ASSERT_GV_AUTOLOAD_SV assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_AVADD
#define PERL_ARGS_ASSERT_GV_AVADD
#endif
#ifndef PERL_ARGS_ASSERT_GV_CHECK
#define PERL_ARGS_ASSERT_GV_CHECK assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_GV_CONST_SV
#define PERL_ARGS_ASSERT_GV_CONST_SV assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_DUMP
#define PERL_ARGS_ASSERT_GV_DUMP
#endif
#ifndef PERL_ARGS_ASSERT_GV_EFULLNAME3
#define PERL_ARGS_ASSERT_GV_EFULLNAME3 assert(sv); assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_EFULLNAME4
#define PERL_ARGS_ASSERT_GV_EFULLNAME4 assert(sv); assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHFILE
#define PERL_ARGS_ASSERT_GV_FETCHFILE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHFILE_FLAGS
#define PERL_ARGS_ASSERT_GV_FETCHFILE_FLAGS assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH
#define PERL_ARGS_ASSERT_GV_FETCHMETH assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETHOD
#define PERL_ARGS_ASSERT_GV_FETCHMETHOD assert(stash); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETHOD_AUTOLOAD
#define PERL_ARGS_ASSERT_GV_FETCHMETHOD_AUTOLOAD assert(stash); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETHOD_PVN_FLAGS
#define PERL_ARGS_ASSERT_GV_FETCHMETHOD_PVN_FLAGS assert(stash); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETHOD_PV_FLAGS
#define PERL_ARGS_ASSERT_GV_FETCHMETHOD_PV_FLAGS assert(stash); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETHOD_SV_FLAGS
#define PERL_ARGS_ASSERT_GV_FETCHMETHOD_SV_FLAGS assert(stash); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_AUTOLOAD
#define PERL_ARGS_ASSERT_GV_FETCHMETH_AUTOLOAD assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_PV
#define PERL_ARGS_ASSERT_GV_FETCHMETH_PV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_PVN
#define PERL_ARGS_ASSERT_GV_FETCHMETH_PVN assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_PVN_AUTOLOAD
#define PERL_ARGS_ASSERT_GV_FETCHMETH_PVN_AUTOLOAD assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_PV_AUTOLOAD
#define PERL_ARGS_ASSERT_GV_FETCHMETH_PV_AUTOLOAD assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_SV
#define PERL_ARGS_ASSERT_GV_FETCHMETH_SV assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHMETH_SV_AUTOLOAD
#define PERL_ARGS_ASSERT_GV_FETCHMETH_SV_AUTOLOAD assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHPV
#define PERL_ARGS_ASSERT_GV_FETCHPV assert(nambeg)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHPVN_FLAGS
#define PERL_ARGS_ASSERT_GV_FETCHPVN_FLAGS assert(nambeg)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FETCHSV
#define PERL_ARGS_ASSERT_GV_FETCHSV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FULLNAME3
#define PERL_ARGS_ASSERT_GV_FULLNAME3 assert(sv); assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_FULLNAME4
#define PERL_ARGS_ASSERT_GV_FULLNAME4 assert(sv); assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_HANDLER
#define PERL_ARGS_ASSERT_GV_HANDLER assert(!stash || SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_GV_HVADD
#define PERL_ARGS_ASSERT_GV_HVADD
#endif
#ifndef PERL_ARGS_ASSERT_GV_INIT
#define PERL_ARGS_ASSERT_GV_INIT assert(gv); assert(!stash || SvTYPE(stash) == SVt_PVHV); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_INIT_PV
#define PERL_ARGS_ASSERT_GV_INIT_PV assert(gv); assert(!stash || SvTYPE(stash) == SVt_PVHV); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_INIT_PVN
#define PERL_ARGS_ASSERT_GV_INIT_PVN assert(gv); assert(!stash || SvTYPE(stash) == SVt_PVHV); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_INIT_SV
#define PERL_ARGS_ASSERT_GV_INIT_SV assert(gv); assert(!stash || SvTYPE(stash) == SVt_PVHV); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_IOADD
#define PERL_ARGS_ASSERT_GV_IOADD
#endif
#ifndef PERL_ARGS_ASSERT_GV_NAME_SET
#define PERL_ARGS_ASSERT_GV_NAME_SET assert(gv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_OVERRIDE
#define PERL_ARGS_ASSERT_GV_OVERRIDE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_SETREF
#define PERL_ARGS_ASSERT_GV_SETREF assert(dsv); assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_STASHPV
#define PERL_ARGS_ASSERT_GV_STASHPV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_STASHPVN
#define PERL_ARGS_ASSERT_GV_STASHPVN assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_GV_STASHSV
#define PERL_ARGS_ASSERT_GV_STASHSV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_GV_SVADD
#define PERL_ARGS_ASSERT_GV_SVADD
#endif
#ifndef PERL_ARGS_ASSERT_GV_TRY_DOWNGRADE
#define PERL_ARGS_ASSERT_GV_TRY_DOWNGRADE assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_HV_AUXALLOC
#define PERL_ARGS_ASSERT_HV_AUXALLOC assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_BACKREFERENCES_P
#define PERL_ARGS_ASSERT_HV_BACKREFERENCES_P assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_BUCKET_RATIO
#define PERL_ARGS_ASSERT_HV_BUCKET_RATIO assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_CLEAR
#define PERL_ARGS_ASSERT_HV_CLEAR assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_CLEAR_PLACEHOLDERS
#define PERL_ARGS_ASSERT_HV_CLEAR_PLACEHOLDERS assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_COMMON
#define PERL_ARGS_ASSERT_HV_COMMON
#endif
#ifndef PERL_ARGS_ASSERT_HV_COPY_HINTS_HV
#define PERL_ARGS_ASSERT_HV_COPY_HINTS_HV assert(!ohv || SvTYPE(ohv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_DELAYFREE_ENT
#define PERL_ARGS_ASSERT_HV_DELAYFREE_ENT assert(!notused || SvTYPE(notused) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_DELETE
#define PERL_ARGS_ASSERT_HV_DELETE assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_HV_DELETE_ENT
#define PERL_ARGS_ASSERT_HV_DELETE_ENT assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(keysv)
#endif
#ifndef PERL_ARGS_ASSERT_HV_DUMP
#define PERL_ARGS_ASSERT_HV_DUMP assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_EITER_P
#define PERL_ARGS_ASSERT_HV_EITER_P assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_EITER_SET
#define PERL_ARGS_ASSERT_HV_EITER_SET assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ENAME_ADD
#define PERL_ARGS_ASSERT_HV_ENAME_ADD assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ENAME_DELETE
#define PERL_ARGS_ASSERT_HV_ENAME_DELETE assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_HV_EXISTS
#define PERL_ARGS_ASSERT_HV_EXISTS assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_HV_EXISTS_ENT
#define PERL_ARGS_ASSERT_HV_EXISTS_ENT assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(keysv)
#endif
#ifndef PERL_ARGS_ASSERT_HV_FETCH
#define PERL_ARGS_ASSERT_HV_FETCH assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_HV_FETCH_ENT
#define PERL_ARGS_ASSERT_HV_FETCH_ENT assert(!hv || SvTYPE(hv) == SVt_PVHV); assert(keysv)
#endif
#ifndef PERL_ARGS_ASSERT_HV_FILL
#define PERL_ARGS_ASSERT_HV_FILL assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_FREE_ENT
#define PERL_ARGS_ASSERT_HV_FREE_ENT assert(!notused || SvTYPE(notused) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERINIT
#define PERL_ARGS_ASSERT_HV_ITERINIT assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERKEY
#define PERL_ARGS_ASSERT_HV_ITERKEY assert(entry); assert(retlen)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERKEYSV
#define PERL_ARGS_ASSERT_HV_ITERKEYSV assert(entry)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERNEXT
#define PERL_ARGS_ASSERT_HV_ITERNEXT assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERNEXTSV
#define PERL_ARGS_ASSERT_HV_ITERNEXTSV assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(key); assert(retlen)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERNEXT_FLAGS
#define PERL_ARGS_ASSERT_HV_ITERNEXT_FLAGS assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_ITERVAL
#define PERL_ARGS_ASSERT_HV_ITERVAL assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(entry)
#endif
#ifndef PERL_ARGS_ASSERT_HV_KSPLIT
#define PERL_ARGS_ASSERT_HV_KSPLIT assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_MAGIC
#define PERL_ARGS_ASSERT_HV_MAGIC assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_NAME_SET
#define PERL_ARGS_ASSERT_HV_NAME_SET assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_PLACEHOLDERS_GET
#define PERL_ARGS_ASSERT_HV_PLACEHOLDERS_GET assert(hv)
#endif
#ifndef PERL_ARGS_ASSERT_HV_PLACEHOLDERS_P
#define PERL_ARGS_ASSERT_HV_PLACEHOLDERS_P assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_PLACEHOLDERS_SET
#define PERL_ARGS_ASSERT_HV_PLACEHOLDERS_SET assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_PUSHKV
#define PERL_ARGS_ASSERT_HV_PUSHKV assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_RAND_SET
#define PERL_ARGS_ASSERT_HV_RAND_SET assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_RITER_P
#define PERL_ARGS_ASSERT_HV_RITER_P assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_RITER_SET
#define PERL_ARGS_ASSERT_HV_RITER_SET assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_SCALAR
#define PERL_ARGS_ASSERT_HV_SCALAR assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_STORE
#define PERL_ARGS_ASSERT_HV_STORE assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_STORE_ENT
#define PERL_ARGS_ASSERT_HV_STORE_ENT assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_STORE_FLAGS
#define PERL_ARGS_ASSERT_HV_STORE_FLAGS assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_UNDEF
#define PERL_ARGS_ASSERT_HV_UNDEF assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_HV_UNDEF_FLAGS
#define PERL_ARGS_ASSERT_HV_UNDEF_FLAGS assert(!hv || SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_IBCMP
#define PERL_ARGS_ASSERT_IBCMP assert(a); assert(b)
#endif
#ifndef PERL_ARGS_ASSERT_IBCMP_LOCALE
#define PERL_ARGS_ASSERT_IBCMP_LOCALE assert(a); assert(b)
#endif
#ifndef PERL_ARGS_ASSERT_IBCMP_UTF8
#define PERL_ARGS_ASSERT_IBCMP_UTF8 assert(s1); assert(s2)
#endif
#ifndef PERL_ARGS_ASSERT_INFIX_PLUGIN_STANDARD
#define PERL_ARGS_ASSERT_INFIX_PLUGIN_STANDARD assert(operator_ptr); assert(def)
#endif
#ifndef PERL_ARGS_ASSERT_INIT_ARGV_SYMBOLS
#define PERL_ARGS_ASSERT_INIT_ARGV_SYMBOLS assert(argv)
#endif
#ifndef PERL_ARGS_ASSERT_INIT_CONSTANTS
#define PERL_ARGS_ASSERT_INIT_CONSTANTS
#endif
#ifndef PERL_ARGS_ASSERT_INIT_DBARGS
#define PERL_ARGS_ASSERT_INIT_DBARGS
#endif
#ifndef PERL_ARGS_ASSERT_INIT_DEBUGGER
#define PERL_ARGS_ASSERT_INIT_DEBUGGER
#endif
#ifndef PERL_ARGS_ASSERT_INIT_I18NL10N
#define PERL_ARGS_ASSERT_INIT_I18NL10N
#endif
#ifndef PERL_ARGS_ASSERT_INIT_NAMED_CV
#define PERL_ARGS_ASSERT_INIT_NAMED_CV assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM); assert(nameop)
#endif
#ifndef PERL_ARGS_ASSERT_INIT_STACKS
#define PERL_ARGS_ASSERT_INIT_STACKS
#endif
#ifndef PERL_ARGS_ASSERT_INIT_TM
#define PERL_ARGS_ASSERT_INIT_TM assert(ptm)
#endif
#ifndef PERL_ARGS_ASSERT_INIT_UNIPROPS
#define PERL_ARGS_ASSERT_INIT_UNIPROPS
#endif
#ifndef PERL_ARGS_ASSERT_INSTR
#define PERL_ARGS_ASSERT_INSTR assert(big); assert(little)
#endif
#ifndef PERL_ARGS_ASSERT_INTRO_MY
#define PERL_ARGS_ASSERT_INTRO_MY
#endif
#ifndef PERL_ARGS_ASSERT_INVERSE_FOLDS_
#define PERL_ARGS_ASSERT_INVERSE_FOLDS_ assert(first_folds_to); assert(remaining_folds_to)
#endif
#ifndef PERL_ARGS_ASSERT_INVERT
#define PERL_ARGS_ASSERT_INVERT
#endif
#ifndef PERL_ARGS_ASSERT_INVMAP_DUMP
#define PERL_ARGS_ASSERT_INVMAP_DUMP assert(invlist); assert(map)
#endif
#ifndef PERL_ARGS_ASSERT_IO_CLOSE
#define PERL_ARGS_ASSERT_IO_CLOSE assert(io)
#endif
#ifndef PERL_ARGS_ASSERT_ISINFNAN
#define PERL_ARGS_ASSERT_ISINFNAN
#endif
#ifndef PERL_ARGS_ASSERT_ISINFNANSV
#define PERL_ARGS_ASSERT_ISINFNANSV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_IS_C9STRICT_UTF8_STRING
#define PERL_ARGS_ASSERT_IS_C9STRICT_UTF8_STRING assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_IS_C9STRICT_UTF8_STRING_LOC
#define PERL_ARGS_ASSERT_IS_C9STRICT_UTF8_STRING_LOC assert(s); assert(ep)
#endif
#ifndef PERL_ARGS_ASSERT_IS_IN_LOCALE_CATEGORY_
#define PERL_ARGS_ASSERT_IS_IN_LOCALE_CATEGORY_
#endif
#ifndef PERL_ARGS_ASSERT_IS_LVALUE_SUB
#define PERL_ARGS_ASSERT_IS_LVALUE_SUB
#endif
#ifndef PERL_ARGS_ASSERT_IS_STRICT_UTF8_STRING
#define PERL_ARGS_ASSERT_IS_STRICT_UTF8_STRING assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_IS_STRICT_UTF8_STRING_LOC
#define PERL_ARGS_ASSERT_IS_STRICT_UTF8_STRING_LOC assert(s); assert(ep)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UNI_FOO_
#define PERL_ARGS_ASSERT_IS_UNI_FOO_
#endif
#ifndef PERL_ARGS_ASSERT_IS_UNI_PERL_IDCONT_
#define PERL_ARGS_ASSERT_IS_UNI_PERL_IDCONT_
#endif
#ifndef PERL_ARGS_ASSERT_IS_UNI_PERL_IDSTART_
#define PERL_ARGS_ASSERT_IS_UNI_PERL_IDSTART_
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_CHAR_BUF
#define PERL_ARGS_ASSERT_IS_UTF8_CHAR_BUF assert(buf); assert(buf_end); assert(buf <= buf_end)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_CHAR_HELPER_
#define PERL_ARGS_ASSERT_IS_UTF8_CHAR_HELPER_ assert(s); assert(e); assert(s < e)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_FF_HELPER_
#define PERL_ARGS_ASSERT_IS_UTF8_FF_HELPER_ assert(s0); assert(e); assert(s0 < e)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_FIXED_WIDTH_BUF_FLAGS
#define PERL_ARGS_ASSERT_IS_UTF8_FIXED_WIDTH_BUF_FLAGS assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_FIXED_WIDTH_BUF_LOC_FLAGS
#define PERL_ARGS_ASSERT_IS_UTF8_FIXED_WIDTH_BUF_LOC_FLAGS assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_FOO_
#define PERL_ARGS_ASSERT_IS_UTF8_FOO_ assert(p); assert(e); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_PERL_IDCONT_
#define PERL_ARGS_ASSERT_IS_UTF8_PERL_IDCONT_ assert(p); assert(e); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_PERL_IDSTART_
#define PERL_ARGS_ASSERT_IS_UTF8_PERL_IDSTART_ assert(p); assert(e); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_STRING
#define PERL_ARGS_ASSERT_IS_UTF8_STRING assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_STRING_LOC
#define PERL_ARGS_ASSERT_IS_UTF8_STRING_LOC assert(s); assert(ep)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_STRING_LOC_FLAGS
#define PERL_ARGS_ASSERT_IS_UTF8_STRING_LOC_FLAGS assert(s); assert(ep)
#endif
#ifndef PERL_ARGS_ASSERT_IS_UTF8_VALID_PARTIAL_CHAR
#define PERL_ARGS_ASSERT_IS_UTF8_VALID_PARTIAL_CHAR assert(s0); assert(e); assert(s0 < e)
#endif
#ifndef PERL_ARGS_ASSERT_JMAYBE
#define PERL_ARGS_ASSERT_JMAYBE assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_KEYWORD
#define PERL_ARGS_ASSERT_KEYWORD assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_KEYWORD_PLUGIN_STANDARD
#define PERL_ARGS_ASSERT_KEYWORD_PLUGIN_STANDARD assert(keyword_ptr); assert(op_ptr)
#endif
#ifndef PERL_ARGS_ASSERT_LEAVE_ADJUST_STACKS
#define PERL_ARGS_ASSERT_LEAVE_ADJUST_STACKS assert(from_sp); assert(to_sp)
#endif
#ifndef PERL_ARGS_ASSERT_LEAVE_SCOPE
#define PERL_ARGS_ASSERT_LEAVE_SCOPE
#endif
#ifndef PERL_ARGS_ASSERT_LEX_BUFUTF8
#define PERL_ARGS_ASSERT_LEX_BUFUTF8
#endif
#ifndef PERL_ARGS_ASSERT_LEX_DISCARD_TO
#define PERL_ARGS_ASSERT_LEX_DISCARD_TO assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_LEX_GROW_LINESTR
#define PERL_ARGS_ASSERT_LEX_GROW_LINESTR
#endif
#ifndef PERL_ARGS_ASSERT_LEX_NEXT_CHUNK
#define PERL_ARGS_ASSERT_LEX_NEXT_CHUNK
#endif
#ifndef PERL_ARGS_ASSERT_LEX_PEEK_UNICHAR
#define PERL_ARGS_ASSERT_LEX_PEEK_UNICHAR
#endif
#ifndef PERL_ARGS_ASSERT_LEX_READ_SPACE
#define PERL_ARGS_ASSERT_LEX_READ_SPACE
#endif
#ifndef PERL_ARGS_ASSERT_LEX_READ_TO
#define PERL_ARGS_ASSERT_LEX_READ_TO assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_LEX_READ_UNICHAR
#define PERL_ARGS_ASSERT_LEX_READ_UNICHAR
#endif
#ifndef PERL_ARGS_ASSERT_LEX_START
#define PERL_ARGS_ASSERT_LEX_START
#endif
#ifndef PERL_ARGS_ASSERT_LEX_STUFF_PV
#define PERL_ARGS_ASSERT_LEX_STUFF_PV assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_LEX_STUFF_PVN
#define PERL_ARGS_ASSERT_LEX_STUFF_PVN assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_LEX_STUFF_SV
#define PERL_ARGS_ASSERT_LEX_STUFF_SV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_LEX_UNSTUFF
#define PERL_ARGS_ASSERT_LEX_UNSTUFF assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_LIST
#define PERL_ARGS_ASSERT_LIST
#endif
#ifndef PERL_ARGS_ASSERT_LOAD_CHARNAMES
#define PERL_ARGS_ASSERT_LOAD_CHARNAMES assert(char_name); assert(context); assert(error_msg)
#endif
#ifndef PERL_ARGS_ASSERT_LOAD_MODULE
#define PERL_ARGS_ASSERT_LOAD_MODULE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_LOCALE_PANIC
#define PERL_ARGS_ASSERT_LOCALE_PANIC assert(msg); assert(higher_caller_file)
#endif
#ifndef PERL_ARGS_ASSERT_LOCALIZE
#define PERL_ARGS_ASSERT_LOCALIZE assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_LONG_VALID_UTF8_TO_UV
#define PERL_ARGS_ASSERT_LONG_VALID_UTF8_TO_UV assert(s); assert(e); assert(s < e)
#endif
#ifndef PERL_ARGS_ASSERT_LOOKS_LIKE_NUMBER
#define PERL_ARGS_ASSERT_LOOKS_LIKE_NUMBER assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARARYLEN_P
#define PERL_ARGS_ASSERT_MAGIC_CLEARARYLEN_P assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARENV
#define PERL_ARGS_ASSERT_MAGIC_CLEARENV assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARHINT
#define PERL_ARGS_ASSERT_MAGIC_CLEARHINT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARHINTS
#define PERL_ARGS_ASSERT_MAGIC_CLEARHINTS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARHOOK
#define PERL_ARGS_ASSERT_MAGIC_CLEARHOOK assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARHOOKALL
#define PERL_ARGS_ASSERT_MAGIC_CLEARHOOKALL assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARISA
#define PERL_ARGS_ASSERT_MAGIC_CLEARISA assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARPACK
#define PERL_ARGS_ASSERT_MAGIC_CLEARPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEARSIG
#define PERL_ARGS_ASSERT_MAGIC_CLEARSIG assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_CLEAR_ALL_ENV
#define PERL_ARGS_ASSERT_MAGIC_CLEAR_ALL_ENV assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_COPYCALLCHECKER
#define PERL_ARGS_ASSERT_MAGIC_COPYCALLCHECKER assert(sv); assert(mg); assert(nsv)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_DUMP
#define PERL_ARGS_ASSERT_MAGIC_DUMP
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_EXISTSPACK
#define PERL_ARGS_ASSERT_MAGIC_EXISTSPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_FREEARYLEN_P
#define PERL_ARGS_ASSERT_MAGIC_FREEARYLEN_P assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_FREEDESTRUCT
#define PERL_ARGS_ASSERT_MAGIC_FREEDESTRUCT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_FREEMGLOB
#define PERL_ARGS_ASSERT_MAGIC_FREEMGLOB assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_FREEOVRLD
#define PERL_ARGS_ASSERT_MAGIC_FREEOVRLD assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_FREEUTF8
#define PERL_ARGS_ASSERT_MAGIC_FREEUTF8 assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GET
#define PERL_ARGS_ASSERT_MAGIC_GET assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETARYLEN
#define PERL_ARGS_ASSERT_MAGIC_GETARYLEN assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETDEBUGVAR
#define PERL_ARGS_ASSERT_MAGIC_GETDEBUGVAR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETDEFELEM
#define PERL_ARGS_ASSERT_MAGIC_GETDEFELEM assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETNKEYS
#define PERL_ARGS_ASSERT_MAGIC_GETNKEYS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETPACK
#define PERL_ARGS_ASSERT_MAGIC_GETPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETPOS
#define PERL_ARGS_ASSERT_MAGIC_GETPOS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETSIG
#define PERL_ARGS_ASSERT_MAGIC_GETSIG assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETSUBSTR
#define PERL_ARGS_ASSERT_MAGIC_GETSUBSTR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETTAINT
#define PERL_ARGS_ASSERT_MAGIC_GETTAINT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETUVAR
#define PERL_ARGS_ASSERT_MAGIC_GETUVAR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_GETVEC
#define PERL_ARGS_ASSERT_MAGIC_GETVEC assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_KILLBACKREFS
#define PERL_ARGS_ASSERT_MAGIC_KILLBACKREFS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_METHCALL
#define PERL_ARGS_ASSERT_MAGIC_METHCALL assert(sv); assert(mg); assert(meth)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_NEXTPACK
#define PERL_ARGS_ASSERT_MAGIC_NEXTPACK assert(sv); assert(mg); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_REGDATA_CNT
#define PERL_ARGS_ASSERT_MAGIC_REGDATA_CNT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_REGDATUM_GET
#define PERL_ARGS_ASSERT_MAGIC_REGDATUM_GET assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_REGDATUM_SET
#define PERL_ARGS_ASSERT_MAGIC_REGDATUM_SET assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SCALARPACK
#define PERL_ARGS_ASSERT_MAGIC_SCALARPACK assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SET
#define PERL_ARGS_ASSERT_MAGIC_SET assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETARYLEN
#define PERL_ARGS_ASSERT_MAGIC_SETARYLEN assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETDBLINE
#define PERL_ARGS_ASSERT_MAGIC_SETDBLINE assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETDEBUGVAR
#define PERL_ARGS_ASSERT_MAGIC_SETDEBUGVAR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETDEFELEM
#define PERL_ARGS_ASSERT_MAGIC_SETDEFELEM assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETENV
#define PERL_ARGS_ASSERT_MAGIC_SETENV assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETHINT
#define PERL_ARGS_ASSERT_MAGIC_SETHINT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETHOOK
#define PERL_ARGS_ASSERT_MAGIC_SETHOOK assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETHOOKALL
#define PERL_ARGS_ASSERT_MAGIC_SETHOOKALL assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETISA
#define PERL_ARGS_ASSERT_MAGIC_SETISA assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETLVREF
#define PERL_ARGS_ASSERT_MAGIC_SETLVREF assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETMGLOB
#define PERL_ARGS_ASSERT_MAGIC_SETMGLOB assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETNKEYS
#define PERL_ARGS_ASSERT_MAGIC_SETNKEYS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETNONELEM
#define PERL_ARGS_ASSERT_MAGIC_SETNONELEM assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETPACK
#define PERL_ARGS_ASSERT_MAGIC_SETPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETPOS
#define PERL_ARGS_ASSERT_MAGIC_SETPOS assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETREGEXP
#define PERL_ARGS_ASSERT_MAGIC_SETREGEXP assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETSIG
#define PERL_ARGS_ASSERT_MAGIC_SETSIG assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETSIGALL
#define PERL_ARGS_ASSERT_MAGIC_SETSIGALL assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETSUBSTR
#define PERL_ARGS_ASSERT_MAGIC_SETSUBSTR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETTAINT
#define PERL_ARGS_ASSERT_MAGIC_SETTAINT assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETUTF8
#define PERL_ARGS_ASSERT_MAGIC_SETUTF8 assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETUVAR
#define PERL_ARGS_ASSERT_MAGIC_SETUVAR assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SETVEC
#define PERL_ARGS_ASSERT_MAGIC_SETVEC assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SET_ALL_ENV
#define PERL_ARGS_ASSERT_MAGIC_SET_ALL_ENV assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_SIZEPACK
#define PERL_ARGS_ASSERT_MAGIC_SIZEPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MAGIC_WIPEPACK
#define PERL_ARGS_ASSERT_MAGIC_WIPEPACK assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_MALLOC
#define PERL_ARGS_ASSERT_MALLOC
#endif
#ifndef PERL_ARGS_ASSERT_MARKSTACK_GROW
#define PERL_ARGS_ASSERT_MARKSTACK_GROW
#endif
#ifndef PERL_ARGS_ASSERT_MBTOWC_
#define PERL_ARGS_ASSERT_MBTOWC_
#endif
#ifndef PERL_ARGS_ASSERT_MESS
#define PERL_ARGS_ASSERT_MESS assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_MESS_SV
#define PERL_ARGS_ASSERT_MESS_SV assert(basemsg)
#endif
#ifndef PERL_ARGS_ASSERT_MFREE
#define PERL_ARGS_ASSERT_MFREE
#endif
#ifndef PERL_ARGS_ASSERT_MG_CLEAR
#define PERL_ARGS_ASSERT_MG_CLEAR assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_COPY
#define PERL_ARGS_ASSERT_MG_COPY assert(sv); assert(nsv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_FIND
#define PERL_ARGS_ASSERT_MG_FIND
#endif
#ifndef PERL_ARGS_ASSERT_MG_FINDEXT
#define PERL_ARGS_ASSERT_MG_FINDEXT
#endif
#ifndef PERL_ARGS_ASSERT_MG_FIND_MGLOB
#define PERL_ARGS_ASSERT_MG_FIND_MGLOB assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_FREE
#define PERL_ARGS_ASSERT_MG_FREE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_FREEEXT
#define PERL_ARGS_ASSERT_MG_FREEEXT assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_FREE_TYPE
#define PERL_ARGS_ASSERT_MG_FREE_TYPE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_GET
#define PERL_ARGS_ASSERT_MG_GET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_LOCALIZE
#define PERL_ARGS_ASSERT_MG_LOCALIZE assert(sv); assert(nsv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_MAGICAL
#define PERL_ARGS_ASSERT_MG_MAGICAL assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_SET
#define PERL_ARGS_ASSERT_MG_SET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MG_SIZE
#define PERL_ARGS_ASSERT_MG_SIZE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_MINI_MKTIME
#define PERL_ARGS_ASSERT_MINI_MKTIME assert(ptm)
#endif
#ifndef PERL_ARGS_ASSERT_MODE_FROM_DISCIPLINE
#define PERL_ARGS_ASSERT_MODE_FROM_DISCIPLINE
#endif
#ifndef PERL_ARGS_ASSERT_MORESWITCHES
#define PERL_ARGS_ASSERT_MORESWITCHES assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_MORE_BODIES
#define PERL_ARGS_ASSERT_MORE_BODIES
#endif
#ifndef PERL_ARGS_ASSERT_MORE_SV
#define PERL_ARGS_ASSERT_MORE_SV
#endif
#ifndef PERL_ARGS_ASSERT_MORTAL_DESTRUCTOR_SV
#define PERL_ARGS_ASSERT_MORTAL_DESTRUCTOR_SV assert(coderef)
#endif
#ifndef PERL_ARGS_ASSERT_MORTAL_SVFUNC_X
#define PERL_ARGS_ASSERT_MORTAL_SVFUNC_X
#endif
#ifndef PERL_ARGS_ASSERT_MRO_GET_FROM_NAME
#define PERL_ARGS_ASSERT_MRO_GET_FROM_NAME assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_GET_LINEAR_ISA
#define PERL_ARGS_ASSERT_MRO_GET_LINEAR_ISA assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_GET_PRIVATE_DATA
#define PERL_ARGS_ASSERT_MRO_GET_PRIVATE_DATA assert(smeta); assert(which)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_ISA_CHANGED_IN
#define PERL_ARGS_ASSERT_MRO_ISA_CHANGED_IN assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_META_INIT
#define PERL_ARGS_ASSERT_MRO_META_INIT assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_METHOD_CHANGED_IN
#define PERL_ARGS_ASSERT_MRO_METHOD_CHANGED_IN assert(stash); assert(SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_PACKAGE_MOVED
#define PERL_ARGS_ASSERT_MRO_PACKAGE_MOVED assert(!stash || SvTYPE(stash) == SVt_PVHV); assert(!oldstash || SvTYPE(oldstash) == SVt_PVHV); assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_REGISTER
#define PERL_ARGS_ASSERT_MRO_REGISTER assert(mro)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_SET_MRO
#define PERL_ARGS_ASSERT_MRO_SET_MRO assert(meta); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_MRO_SET_PRIVATE_DATA
#define PERL_ARGS_ASSERT_MRO_SET_PRIVATE_DATA assert(smeta); assert(which); assert(data)
#endif
#ifndef PERL_ARGS_ASSERT_MULTICONCAT_STRINGIFY
#define PERL_ARGS_ASSERT_MULTICONCAT_STRINGIFY assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_MULTIDEREF_STRINGIFY
#define PERL_ARGS_ASSERT_MULTIDEREF_STRINGIFY assert(o); assert(!cv || SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_MY_ATOF
#define PERL_ARGS_ASSERT_MY_ATOF assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_MY_ATOF2
#define PERL_ARGS_ASSERT_MY_ATOF2 assert(orig); assert(value)
#endif
#ifndef PERL_ARGS_ASSERT_MY_ATOF3
#define PERL_ARGS_ASSERT_MY_ATOF3 assert(orig); assert(value)
#endif
#ifndef PERL_ARGS_ASSERT_MY_ATTRS
#define PERL_ARGS_ASSERT_MY_ATTRS assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_MY_CLEARENV
#define PERL_ARGS_ASSERT_MY_CLEARENV
#endif
#ifndef PERL_ARGS_ASSERT_MY_DIRFD
#define PERL_ARGS_ASSERT_MY_DIRFD
#endif
#ifndef PERL_ARGS_ASSERT_MY_EXIT
#define PERL_ARGS_ASSERT_MY_EXIT
#endif
#ifndef PERL_ARGS_ASSERT_MY_FAILURE_EXIT
#define PERL_ARGS_ASSERT_MY_FAILURE_EXIT
#endif
#ifndef PERL_ARGS_ASSERT_MY_FFLUSH_ALL
#define PERL_ARGS_ASSERT_MY_FFLUSH_ALL
#endif
#ifndef PERL_ARGS_ASSERT_MY_FORK
#define PERL_ARGS_ASSERT_MY_FORK
#endif
#ifndef PERL_ARGS_ASSERT_MY_LSTAT_FLAGS
#define PERL_ARGS_ASSERT_MY_LSTAT_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_MY_MKOSTEMP_CLOEXEC
#define PERL_ARGS_ASSERT_MY_MKOSTEMP_CLOEXEC assert(templte)
#endif
#ifndef PERL_ARGS_ASSERT_MY_MKSTEMP_CLOEXEC
#define PERL_ARGS_ASSERT_MY_MKSTEMP_CLOEXEC assert(templte)
#endif
#ifndef PERL_ARGS_ASSERT_MY_POPEN_LIST
#define PERL_ARGS_ASSERT_MY_POPEN_LIST assert(mode); assert(args)
#endif
#ifndef PERL_ARGS_ASSERT_MY_SETENV
#define PERL_ARGS_ASSERT_MY_SETENV
#endif
#ifndef PERL_ARGS_ASSERT_MY_SNPRINTF
#define PERL_ARGS_ASSERT_MY_SNPRINTF assert(buffer); assert(format)
#endif
#ifndef PERL_ARGS_ASSERT_MY_SOCKETPAIR
#define PERL_ARGS_ASSERT_MY_SOCKETPAIR
#endif
#ifndef PERL_ARGS_ASSERT_MY_STAT_FLAGS
#define PERL_ARGS_ASSERT_MY_STAT_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_MY_STRERROR
#define PERL_ARGS_ASSERT_MY_STRERROR assert(utf8ness)
#endif
#ifndef PERL_ARGS_ASSERT_MY_STRFTIME
#define PERL_ARGS_ASSERT_MY_STRFTIME assert(fmt)
#endif
#ifndef PERL_ARGS_ASSERT_MY_STRTOD
#define PERL_ARGS_ASSERT_MY_STRTOD assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_MY_UNEXEC
#define PERL_ARGS_ASSERT_MY_UNEXEC
#endif
#ifndef PERL_ARGS_ASSERT_MY_VSNPRINTF
#define PERL_ARGS_ASSERT_MY_VSNPRINTF assert(buffer); assert(format)
#endif
#ifndef PERL_ARGS_ASSERT_NEWANONATTRSUB
#define PERL_ARGS_ASSERT_NEWANONATTRSUB
#endif
#ifndef PERL_ARGS_ASSERT_NEWANONHASH
#define PERL_ARGS_ASSERT_NEWANONHASH
#endif
#ifndef PERL_ARGS_ASSERT_NEWANONLIST
#define PERL_ARGS_ASSERT_NEWANONLIST
#endif
#ifndef PERL_ARGS_ASSERT_NEWANONSUB
#define PERL_ARGS_ASSERT_NEWANONSUB
#endif
#ifndef PERL_ARGS_ASSERT_NEWARGDEFELEMOP
#define PERL_ARGS_ASSERT_NEWARGDEFELEMOP assert(expr)
#endif
#ifndef PERL_ARGS_ASSERT_NEWASSIGNOP
#define PERL_ARGS_ASSERT_NEWASSIGNOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWATTRSUB
#define PERL_ARGS_ASSERT_NEWATTRSUB
#endif
#ifndef PERL_ARGS_ASSERT_NEWATTRSUB_X
#define PERL_ARGS_ASSERT_NEWATTRSUB_X
#endif
#ifndef PERL_ARGS_ASSERT_NEWAV
#define PERL_ARGS_ASSERT_NEWAV
#endif
#ifndef PERL_ARGS_ASSERT_NEWAVAV
#define PERL_ARGS_ASSERT_NEWAVAV assert(!oav || SvTYPE(oav) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWAVHV
#define PERL_ARGS_ASSERT_NEWAVHV assert(!ohv || SvTYPE(ohv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWAVREF
#define PERL_ARGS_ASSERT_NEWAVREF assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_NEWAV_ALLOC_X
#define PERL_ARGS_ASSERT_NEWAV_ALLOC_X
#endif
#ifndef PERL_ARGS_ASSERT_NEWAV_ALLOC_XZ
#define PERL_ARGS_ASSERT_NEWAV_ALLOC_XZ
#endif
#ifndef PERL_ARGS_ASSERT_NEWAV_MORTAL
#define PERL_ARGS_ASSERT_NEWAV_MORTAL
#endif
#ifndef PERL_ARGS_ASSERT_NEWBINOP
#define PERL_ARGS_ASSERT_NEWBINOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWCONDOP
#define PERL_ARGS_ASSERT_NEWCONDOP assert(first)
#endif
#ifndef PERL_ARGS_ASSERT_NEWCONSTSUB
#define PERL_ARGS_ASSERT_NEWCONSTSUB assert(!stash || SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWCONSTSUB_FLAGS
#define PERL_ARGS_ASSERT_NEWCONSTSUB_FLAGS assert(!stash || SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWCVREF
#define PERL_ARGS_ASSERT_NEWCVREF
#endif
#ifndef PERL_ARGS_ASSERT_NEWDEFEROP
#define PERL_ARGS_ASSERT_NEWDEFEROP assert(block)
#endif
#ifndef PERL_ARGS_ASSERT_NEWDEFSVOP
#define PERL_ARGS_ASSERT_NEWDEFSVOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWFORM
#define PERL_ARGS_ASSERT_NEWFORM
#endif
#ifndef PERL_ARGS_ASSERT_NEWFOROP
#define PERL_ARGS_ASSERT_NEWFOROP assert(expr)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGIVENOP
#define PERL_ARGS_ASSERT_NEWGIVENOP assert(cond); assert(block)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGP
#define PERL_ARGS_ASSERT_NEWGP assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGVGEN
#define PERL_ARGS_ASSERT_NEWGVGEN assert(pack)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGVGEN_FLAGS
#define PERL_ARGS_ASSERT_NEWGVGEN_FLAGS assert(pack)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGVOP
#define PERL_ARGS_ASSERT_NEWGVOP assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWGVREF
#define PERL_ARGS_ASSERT_NEWGVREF
#endif
#ifndef PERL_ARGS_ASSERT_NEWHV
#define PERL_ARGS_ASSERT_NEWHV
#endif
#ifndef PERL_ARGS_ASSERT_NEWHVHV
#define PERL_ARGS_ASSERT_NEWHVHV assert(!ohv || SvTYPE(ohv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWHVREF
#define PERL_ARGS_ASSERT_NEWHVREF assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_NEWIO
#define PERL_ARGS_ASSERT_NEWIO
#endif
#ifndef PERL_ARGS_ASSERT_NEWLISTOP
#define PERL_ARGS_ASSERT_NEWLISTOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWLISTOPN
#define PERL_ARGS_ASSERT_NEWLISTOPN
#endif
#ifndef PERL_ARGS_ASSERT_NEWLOGOP
#define PERL_ARGS_ASSERT_NEWLOGOP assert(first); assert(other)
#endif
#ifndef PERL_ARGS_ASSERT_NEWLOOPEX
#define PERL_ARGS_ASSERT_NEWLOOPEX assert(label)
#endif
#ifndef PERL_ARGS_ASSERT_NEWLOOPOP
#define PERL_ARGS_ASSERT_NEWLOOPOP assert(expr)
#endif
#ifndef PERL_ARGS_ASSERT_NEWMETHOP
#define PERL_ARGS_ASSERT_NEWMETHOP assert(dynamic_meth)
#endif
#ifndef PERL_ARGS_ASSERT_NEWMETHOP_NAMED
#define PERL_ARGS_ASSERT_NEWMETHOP_NAMED assert(const_meth)
#endif
#ifndef PERL_ARGS_ASSERT_NEWMYSUB
#define PERL_ARGS_ASSERT_NEWMYSUB assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_NEWNULLLIST
#define PERL_ARGS_ASSERT_NEWNULLLIST
#endif
#ifndef PERL_ARGS_ASSERT_NEWOP
#define PERL_ARGS_ASSERT_NEWOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWPADNAMELIST
#define PERL_ARGS_ASSERT_NEWPADNAMELIST
#endif
#ifndef PERL_ARGS_ASSERT_NEWPADNAMEOUTER
#define PERL_ARGS_ASSERT_NEWPADNAMEOUTER assert(outer)
#endif
#ifndef PERL_ARGS_ASSERT_NEWPADNAMEPVN
#define PERL_ARGS_ASSERT_NEWPADNAMEPVN assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_NEWPMOP
#define PERL_ARGS_ASSERT_NEWPMOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWPROG
#define PERL_ARGS_ASSERT_NEWPROG assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_NEWPVOP
#define PERL_ARGS_ASSERT_NEWPVOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWRANGE
#define PERL_ARGS_ASSERT_NEWRANGE assert(left); assert(right)
#endif
#ifndef PERL_ARGS_ASSERT_NEWRV
#define PERL_ARGS_ASSERT_NEWRV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSLICEOP
#define PERL_ARGS_ASSERT_NEWSLICEOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWSTATEOP
#define PERL_ARGS_ASSERT_NEWSTATEOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWSTUB
#define PERL_ARGS_ASSERT_NEWSTUB assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSUB
#define PERL_ARGS_ASSERT_NEWSUB
#endif
#ifndef PERL_ARGS_ASSERT_NEWSV
#define PERL_ARGS_ASSERT_NEWSV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVAVDEFELEM
#define PERL_ARGS_ASSERT_NEWSVAVDEFELEM assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVBOOL
#define PERL_ARGS_ASSERT_NEWSVBOOL
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVHEK
#define PERL_ARGS_ASSERT_NEWSVHEK
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVHEK_MORTAL
#define PERL_ARGS_ASSERT_NEWSVHEK_MORTAL
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVIV
#define PERL_ARGS_ASSERT_NEWSVIV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVNV
#define PERL_ARGS_ASSERT_NEWSVNV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVOP
#define PERL_ARGS_ASSERT_NEWSVOP assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPV
#define PERL_ARGS_ASSERT_NEWSVPV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPVF
#define PERL_ARGS_ASSERT_NEWSVPVF assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPVN
#define PERL_ARGS_ASSERT_NEWSVPVN
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPVN_FLAGS
#define PERL_ARGS_ASSERT_NEWSVPVN_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPVN_SHARE
#define PERL_ARGS_ASSERT_NEWSVPVN_SHARE
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPVZ
#define PERL_ARGS_ASSERT_NEWSVPVZ
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVPV_SHARE
#define PERL_ARGS_ASSERT_NEWSVPV_SHARE
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVREF
#define PERL_ARGS_ASSERT_NEWSVREF assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVRV
#define PERL_ARGS_ASSERT_NEWSVRV assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVSV
#define PERL_ARGS_ASSERT_NEWSVSV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVSV_FLAGS_NN
#define PERL_ARGS_ASSERT_NEWSVSV_FLAGS_NN assert(old)
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVSV_NOMG
#define PERL_ARGS_ASSERT_NEWSVSV_NOMG
#endif
#ifndef PERL_ARGS_ASSERT_NEWSVUV
#define PERL_ARGS_ASSERT_NEWSVUV
#endif
#ifndef PERL_ARGS_ASSERT_NEWSV_FALSE
#define PERL_ARGS_ASSERT_NEWSV_FALSE
#endif
#ifndef PERL_ARGS_ASSERT_NEWSV_TRUE
#define PERL_ARGS_ASSERT_NEWSV_TRUE
#endif
#ifndef PERL_ARGS_ASSERT_NEWTRYCATCHOP
#define PERL_ARGS_ASSERT_NEWTRYCATCHOP assert(tryblock); assert(catchvar); assert(catchblock)
#endif
#ifndef PERL_ARGS_ASSERT_NEWUNOP
#define PERL_ARGS_ASSERT_NEWUNOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWUNOP_AUX
#define PERL_ARGS_ASSERT_NEWUNOP_AUX
#endif
#ifndef PERL_ARGS_ASSERT_NEWWHENOP
#define PERL_ARGS_ASSERT_NEWWHENOP assert(block)
#endif
#ifndef PERL_ARGS_ASSERT_NEWWHILEOP
#define PERL_ARGS_ASSERT_NEWWHILEOP
#endif
#ifndef PERL_ARGS_ASSERT_NEWXS
#define PERL_ARGS_ASSERT_NEWXS assert(subaddr); assert(filename)
#endif
#ifndef PERL_ARGS_ASSERT_NEWXS_DEFFILE
#define PERL_ARGS_ASSERT_NEWXS_DEFFILE assert(name); assert(subaddr)
#endif
#ifndef PERL_ARGS_ASSERT_NEWXS_FLAGS
#define PERL_ARGS_ASSERT_NEWXS_FLAGS assert(subaddr); assert(filename)
#endif
#ifndef PERL_ARGS_ASSERT_NEWXS_LEN_FLAGS
#define PERL_ARGS_ASSERT_NEWXS_LEN_FLAGS assert(subaddr)
#endif
#ifndef PERL_ARGS_ASSERT_NEW_BLOCK_STATEMENT
#define PERL_ARGS_ASSERT_NEW_BLOCK_STATEMENT assert(block)
#endif
#ifndef PERL_ARGS_ASSERT_NEW_STACKINFO
#define PERL_ARGS_ASSERT_NEW_STACKINFO
#endif
#ifndef PERL_ARGS_ASSERT_NEW_STACKINFO_FLAGS
#define PERL_ARGS_ASSERT_NEW_STACKINFO_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_NEW_VERSION
#define PERL_ARGS_ASSERT_NEW_VERSION assert(ver)
#endif
#ifndef PERL_ARGS_ASSERT_NEW_WARNINGS_BITFIELD
#define PERL_ARGS_ASSERT_NEW_WARNINGS_BITFIELD assert(bits)
#endif
#ifndef PERL_ARGS_ASSERT_NEXTARGV
#define PERL_ARGS_ASSERT_NEXTARGV assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_NINSTR
#define PERL_ARGS_ASSERT_NINSTR assert(big); assert(bigend); assert(little); assert(lend); assert(big <= bigend); assert(little <= lend)
#endif
#ifndef PERL_ARGS_ASSERT_NOPERL_DIE
#define PERL_ARGS_ASSERT_NOPERL_DIE assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_NOSHUTDOWNHOOK
#define PERL_ARGS_ASSERT_NOSHUTDOWNHOOK
#endif
#ifndef PERL_ARGS_ASSERT_NOTHREADHOOK
#define PERL_ARGS_ASSERT_NOTHREADHOOK
#endif
#ifndef PERL_ARGS_ASSERT_NOTIFY_PARSER_THAT_ENCODING_CHANGED
#define PERL_ARGS_ASSERT_NOTIFY_PARSER_THAT_ENCODING_CHANGED
#endif
#ifndef PERL_ARGS_ASSERT_NO_BAREWORD_FILEHANDLE
#define PERL_ARGS_ASSERT_NO_BAREWORD_FILEHANDLE assert(fhname)
#endif
#ifndef PERL_ARGS_ASSERT_OOPSAV
#define PERL_ARGS_ASSERT_OOPSAV assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OOPSHV
#define PERL_ARGS_ASSERT_OOPSHV assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OPDUMP_PRINTF
#define PERL_ARGS_ASSERT_OPDUMP_PRINTF assert(ctx); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_OP_APPEND_ELEM
#define PERL_ARGS_ASSERT_OP_APPEND_ELEM
#endif
#ifndef PERL_ARGS_ASSERT_OP_APPEND_LIST
#define PERL_ARGS_ASSERT_OP_APPEND_LIST
#endif
#ifndef PERL_ARGS_ASSERT_OP_CLASS
#define PERL_ARGS_ASSERT_OP_CLASS
#endif
#ifndef PERL_ARGS_ASSERT_OP_CLEAR
#define PERL_ARGS_ASSERT_OP_CLEAR assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_CONTEXTUALIZE
#define PERL_ARGS_ASSERT_OP_CONTEXTUALIZE assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_CONVERT_LIST
#define PERL_ARGS_ASSERT_OP_CONVERT_LIST
#endif
#ifndef PERL_ARGS_ASSERT_OP_DUMP
#define PERL_ARGS_ASSERT_OP_DUMP assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_FORCE_LIST
#define PERL_ARGS_ASSERT_OP_FORCE_LIST
#endif
#ifndef PERL_ARGS_ASSERT_OP_FREE
#define PERL_ARGS_ASSERT_OP_FREE
#endif
#ifndef PERL_ARGS_ASSERT_OP_LINKLIST
#define PERL_ARGS_ASSERT_OP_LINKLIST assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_LVALUE
#define PERL_ARGS_ASSERT_OP_LVALUE
#endif
#ifndef PERL_ARGS_ASSERT_OP_LVALUE_FLAGS
#define PERL_ARGS_ASSERT_OP_LVALUE_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_OP_NULL
#define PERL_ARGS_ASSERT_OP_NULL assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_PARENT
#define PERL_ARGS_ASSERT_OP_PARENT assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_OP_PREPEND_ELEM
#define PERL_ARGS_ASSERT_OP_PREPEND_ELEM
#endif
#ifndef PERL_ARGS_ASSERT_OP_REFCNT_LOCK
#define PERL_ARGS_ASSERT_OP_REFCNT_LOCK
#endif
#ifndef PERL_ARGS_ASSERT_OP_REFCNT_UNLOCK
#define PERL_ARGS_ASSERT_OP_REFCNT_UNLOCK
#endif
#ifndef PERL_ARGS_ASSERT_OP_SCOPE
#define PERL_ARGS_ASSERT_OP_SCOPE
#endif
#ifndef PERL_ARGS_ASSERT_OP_SIBLING_SPLICE
#define PERL_ARGS_ASSERT_OP_SIBLING_SPLICE
#endif
#ifndef PERL_ARGS_ASSERT_OP_UNSCOPE
#define PERL_ARGS_ASSERT_OP_UNSCOPE
#endif
#ifndef PERL_ARGS_ASSERT_OP_WRAP_FINALLY
#define PERL_ARGS_ASSERT_OP_WRAP_FINALLY assert(block); assert(finally)
#endif
#ifndef PERL_ARGS_ASSERT_OUTPUT_NON_PORTABLE
#define PERL_ARGS_ASSERT_OUTPUT_NON_PORTABLE
#endif
#ifndef PERL_ARGS_ASSERT_PACKAGE
#define PERL_ARGS_ASSERT_PACKAGE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_PACKLIST
#define PERL_ARGS_ASSERT_PACKLIST assert(cat); assert(pat); assert(patend); assert(beglist); assert(endlist); assert(pat <= patend); assert(*patend == '\0')
#endif
#ifndef PERL_ARGS_ASSERT_PADLIST_STORE
#define PERL_ARGS_ASSERT_PADLIST_STORE assert(padlist)
#endif
#ifndef PERL_ARGS_ASSERT_PADNAMELIST_FETCH
#define PERL_ARGS_ASSERT_PADNAMELIST_FETCH assert(pnl)
#endif
#ifndef PERL_ARGS_ASSERT_PADNAMELIST_FREE
#define PERL_ARGS_ASSERT_PADNAMELIST_FREE assert(pnl)
#endif
#ifndef PERL_ARGS_ASSERT_PADNAMELIST_STORE
#define PERL_ARGS_ASSERT_PADNAMELIST_STORE assert(pnl)
#endif
#ifndef PERL_ARGS_ASSERT_PADNAME_FREE
#define PERL_ARGS_ASSERT_PADNAME_FREE assert(pn)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ADD_ANON
#define PERL_ARGS_ASSERT_PAD_ADD_ANON assert(func); assert(SvTYPE(func) == SVt_PVCV || SvTYPE(func) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ADD_NAME_PV
#define PERL_ARGS_ASSERT_PAD_ADD_NAME_PV assert(name); assert(!typestash || SvTYPE(typestash) == SVt_PVHV); assert(!ourstash || SvTYPE(ourstash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ADD_NAME_PVN
#define PERL_ARGS_ASSERT_PAD_ADD_NAME_PVN assert(namepv); assert(!typestash || SvTYPE(typestash) == SVt_PVHV); assert(!ourstash || SvTYPE(ourstash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ADD_NAME_SV
#define PERL_ARGS_ASSERT_PAD_ADD_NAME_SV assert(name); assert(!typestash || SvTYPE(typestash) == SVt_PVHV); assert(!ourstash || SvTYPE(ourstash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ADD_WEAKREF
#define PERL_ARGS_ASSERT_PAD_ADD_WEAKREF assert(func); assert(SvTYPE(func) == SVt_PVCV || SvTYPE(func) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_ALLOC
#define PERL_ARGS_ASSERT_PAD_ALLOC
#endif
#ifndef PERL_ARGS_ASSERT_PAD_BLOCK_START
#define PERL_ARGS_ASSERT_PAD_BLOCK_START
#endif
#ifndef PERL_ARGS_ASSERT_PAD_FINDMY_PV
#define PERL_ARGS_ASSERT_PAD_FINDMY_PV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_FINDMY_PVN
#define PERL_ARGS_ASSERT_PAD_FINDMY_PVN assert(namepv)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_FINDMY_SV
#define PERL_ARGS_ASSERT_PAD_FINDMY_SV assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_FIXUP_INNER_ANONS
#define PERL_ARGS_ASSERT_PAD_FIXUP_INNER_ANONS assert(padlist); assert(old_cv); assert(SvTYPE(old_cv) == SVt_PVCV || SvTYPE(old_cv) == SVt_PVFM); assert(new_cv); assert(SvTYPE(new_cv) == SVt_PVCV || SvTYPE(new_cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_FREE
#define PERL_ARGS_ASSERT_PAD_FREE
#endif
#ifndef PERL_ARGS_ASSERT_PAD_LEAVEMY
#define PERL_ARGS_ASSERT_PAD_LEAVEMY
#endif
#ifndef PERL_ARGS_ASSERT_PAD_NEW
#define PERL_ARGS_ASSERT_PAD_NEW
#endif
#ifndef PERL_ARGS_ASSERT_PAD_PUSH
#define PERL_ARGS_ASSERT_PAD_PUSH assert(padlist)
#endif
#ifndef PERL_ARGS_ASSERT_PAD_SWIPE
#define PERL_ARGS_ASSERT_PAD_SWIPE
#endif
#ifndef PERL_ARGS_ASSERT_PAD_TIDY
#define PERL_ARGS_ASSERT_PAD_TIDY
#endif
#ifndef PERL_ARGS_ASSERT_PARSER_FREE
#define PERL_ARGS_ASSERT_PARSER_FREE assert(parser)
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_ARITHEXPR
#define PERL_ARGS_ASSERT_PARSE_ARITHEXPR
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_BARESTMT
#define PERL_ARGS_ASSERT_PARSE_BARESTMT
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_BLOCK
#define PERL_ARGS_ASSERT_PARSE_BLOCK
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_FULLEXPR
#define PERL_ARGS_ASSERT_PARSE_FULLEXPR
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_FULLSTMT
#define PERL_ARGS_ASSERT_PARSE_FULLSTMT
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_IDENT_MSG
#define PERL_ARGS_ASSERT_PARSE_IDENT_MSG assert(s); assert(end); assert(s < end)
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_LABEL
#define PERL_ARGS_ASSERT_PARSE_LABEL
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_LISTEXPR
#define PERL_ARGS_ASSERT_PARSE_LISTEXPR
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_STMTSEQ
#define PERL_ARGS_ASSERT_PARSE_STMTSEQ
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_SUBSIGNATURE
#define PERL_ARGS_ASSERT_PARSE_SUBSIGNATURE
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_TERMEXPR
#define PERL_ARGS_ASSERT_PARSE_TERMEXPR
#endif
#ifndef PERL_ARGS_ASSERT_PARSE_UNICODE_OPTS
#define PERL_ARGS_ASSERT_PARSE_UNICODE_OPTS assert(popt)
#endif
#ifndef PERL_ARGS_ASSERT_PEEP
#define PERL_ARGS_ASSERT_PEEP
#endif
#ifndef PERL_ARGS_ASSERT_PERLIO_CONTEXT_LAYERS
#define PERL_ARGS_ASSERT_PERLIO_CONTEXT_LAYERS
#endif
#ifndef PERL_ARGS_ASSERT_PERLLIO_DUP2_CLOEXEC
#define PERL_ARGS_ASSERT_PERLLIO_DUP2_CLOEXEC
#endif
#ifndef PERL_ARGS_ASSERT_PERLLIO_DUP_CLOEXEC
#define PERL_ARGS_ASSERT_PERLLIO_DUP_CLOEXEC
#endif
#ifndef PERL_ARGS_ASSERT_PERLLIO_OPEN3_CLOEXEC
#define PERL_ARGS_ASSERT_PERLLIO_OPEN3_CLOEXEC assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_PERLLIO_OPEN_CLOEXEC
#define PERL_ARGS_ASSERT_PERLLIO_OPEN_CLOEXEC assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_PERLY_SIGHANDLER
#define PERL_ARGS_ASSERT_PERLY_SIGHANDLER
#endif
#ifndef PERL_ARGS_ASSERT_PERL_ALLOC
#define PERL_ARGS_ASSERT_PERL_ALLOC
#endif
#ifndef PERL_ARGS_ASSERT_PERL_CONSTRUCT
#define PERL_ARGS_ASSERT_PERL_CONSTRUCT assert(my_perl)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_DESTRUCT
#define PERL_ARGS_ASSERT_PERL_DESTRUCT assert(my_perl)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_FREE
#define PERL_ARGS_ASSERT_PERL_FREE assert(my_perl)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_LANGINFO
#define PERL_ARGS_ASSERT_PERL_LANGINFO
#endif
#ifndef PERL_ARGS_ASSERT_PERL_LANGINFO8
#define PERL_ARGS_ASSERT_PERL_LANGINFO8 assert(utf8ness)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_LOCALECONV
#define PERL_ARGS_ASSERT_PERL_LOCALECONV
#endif
#ifndef PERL_ARGS_ASSERT_PERL_PARSE
#define PERL_ARGS_ASSERT_PERL_PARSE assert(my_perl)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_RUN
#define PERL_ARGS_ASSERT_PERL_RUN assert(my_perl)
#endif
#ifndef PERL_ARGS_ASSERT_PERL_SETLOCALE
#define PERL_ARGS_ASSERT_PERL_SETLOCALE
#endif
#ifndef PERL_ARGS_ASSERT_PHASE_NAME
#define PERL_ARGS_ASSERT_PHASE_NAME
#endif
#ifndef PERL_ARGS_ASSERT_PMOP_DUMP
#define PERL_ARGS_ASSERT_PMOP_DUMP
#endif
#ifndef PERL_ARGS_ASSERT_PMRUNTIME
#define PERL_ARGS_ASSERT_PMRUNTIME assert(o); assert(expr)
#endif
#ifndef PERL_ARGS_ASSERT_POPULATE_ISA
#define PERL_ARGS_ASSERT_POPULATE_ISA assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_POP_SCOPE
#define PERL_ARGS_ASSERT_POP_SCOPE
#endif
#ifndef PERL_ARGS_ASSERT_PREGCOMP
#define PERL_ARGS_ASSERT_PREGCOMP assert(pattern)
#endif
#ifndef PERL_ARGS_ASSERT_PREGEXEC
#define PERL_ARGS_ASSERT_PREGEXEC assert(prog); assert(stringarg); assert(strend); assert(strbeg); assert(screamer); assert(strbeg <= stringarg); assert(stringarg <= strend)
#endif
#ifndef PERL_ARGS_ASSERT_PREGFREE
#define PERL_ARGS_ASSERT_PREGFREE
#endif
#ifndef PERL_ARGS_ASSERT_PREGFREE2
#define PERL_ARGS_ASSERT_PREGFREE2 assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_PRESCAN_VERSION
#define PERL_ARGS_ASSERT_PRESCAN_VERSION assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_PTR_TABLE_FETCH
#define PERL_ARGS_ASSERT_PTR_TABLE_FETCH assert(tbl)
#endif
#ifndef PERL_ARGS_ASSERT_PTR_TABLE_FREE
#define PERL_ARGS_ASSERT_PTR_TABLE_FREE
#endif
#ifndef PERL_ARGS_ASSERT_PTR_TABLE_NEW
#define PERL_ARGS_ASSERT_PTR_TABLE_NEW
#endif
#ifndef PERL_ARGS_ASSERT_PTR_TABLE_SPLIT
#define PERL_ARGS_ASSERT_PTR_TABLE_SPLIT assert(tbl)
#endif
#ifndef PERL_ARGS_ASSERT_PTR_TABLE_STORE
#define PERL_ARGS_ASSERT_PTR_TABLE_STORE assert(tbl); assert(newsv)
#endif
#ifndef PERL_ARGS_ASSERT_PUSH_SCOPE
#define PERL_ARGS_ASSERT_PUSH_SCOPE
#endif
#ifndef PERL_ARGS_ASSERT_PV_DISPLAY
#define PERL_ARGS_ASSERT_PV_DISPLAY assert(dsv); assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_PV_ESCAPE
#define PERL_ARGS_ASSERT_PV_ESCAPE assert(str)
#endif
#ifndef PERL_ARGS_ASSERT_PV_PRETTY
#define PERL_ARGS_ASSERT_PV_PRETTY assert(dsv); assert(str)
#endif
#ifndef PERL_ARGS_ASSERT_PV_UNI_DISPLAY
#define PERL_ARGS_ASSERT_PV_UNI_DISPLAY assert(dsv); assert(spv)
#endif
#ifndef PERL_ARGS_ASSERT_QERROR
#define PERL_ARGS_ASSERT_QERROR
#endif
#ifndef PERL_ARGS_ASSERT_RCPV_COPY
#define PERL_ARGS_ASSERT_RCPV_COPY
#endif
#ifndef PERL_ARGS_ASSERT_RCPV_FREE
#define PERL_ARGS_ASSERT_RCPV_FREE
#endif
#ifndef PERL_ARGS_ASSERT_RCPV_NEW
#define PERL_ARGS_ASSERT_RCPV_NEW
#endif
#ifndef PERL_ARGS_ASSERT_REALLOC
#define PERL_ARGS_ASSERT_REALLOC
#endif
#ifndef PERL_ARGS_ASSERT_REENTRANT_FREE
#define PERL_ARGS_ASSERT_REENTRANT_FREE
#endif
#ifndef PERL_ARGS_ASSERT_REENTRANT_INIT
#define PERL_ARGS_ASSERT_REENTRANT_INIT
#endif
#ifndef PERL_ARGS_ASSERT_REENTRANT_RETRY
#define PERL_ARGS_ASSERT_REENTRANT_RETRY assert(f)
#endif
#ifndef PERL_ARGS_ASSERT_REENTRANT_SIZE
#define PERL_ARGS_ASSERT_REENTRANT_SIZE
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_CHAIN_2HV
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_CHAIN_2HV
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_PV
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_PV assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_PVN
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_PVN assert(keypv)
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_SV
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_FETCH_SV assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_FREE
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_FREE
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_INC
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_INC
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_PV
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_PV assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_PVN
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_PVN assert(keypv)
#endif
#ifndef PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_SV
#define PERL_ARGS_ASSERT_REFCOUNTED_HE_NEW_SV assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_REGDUMP
#define PERL_ARGS_ASSERT_REGDUMP assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_REGEXEC_FLAGS
#define PERL_ARGS_ASSERT_REGEXEC_FLAGS assert(rx); assert(stringarg); assert(strend); assert(strbeg); assert(sv); assert(strbeg <= stringarg); assert(stringarg <= strend)
#endif
#ifndef PERL_ARGS_ASSERT_REGFREE_INTERNAL
#define PERL_ARGS_ASSERT_REGFREE_INTERNAL assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_REGINITCOLORS
#define PERL_ARGS_ASSERT_REGINITCOLORS
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_ALL
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_ALL assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_EXISTS
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_EXISTS assert(r); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_FETCH
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_FETCH assert(r); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_FIRSTKEY
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_FIRSTKEY assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_ITER
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_ITER assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_NEXTKEY
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_NEXTKEY assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NAMED_BUFF_SCALAR
#define PERL_ARGS_ASSERT_REG_NAMED_BUFF_SCALAR assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_FETCH
#define PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_FETCH assert(re)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_FETCH_FLAGS
#define PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_FETCH_FLAGS assert(re)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_LENGTH
#define PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_LENGTH assert(r); assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_STORE
#define PERL_ARGS_ASSERT_REG_NUMBERED_BUFF_STORE assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_REG_QR_PACKAGE
#define PERL_ARGS_ASSERT_REG_QR_PACKAGE assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_REG_TEMP_COPY
#define PERL_ARGS_ASSERT_REG_TEMP_COPY assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_RELEASE_REXC_STATE
#define PERL_ARGS_ASSERT_RELEASE_REXC_STATE assert(vstate)
#endif
#ifndef PERL_ARGS_ASSERT_REPEATCPY
#define PERL_ARGS_ASSERT_REPEATCPY assert(to); assert(from)
#endif
#ifndef PERL_ARGS_ASSERT_REPORT_EVIL_FH
#define PERL_ARGS_ASSERT_REPORT_EVIL_FH
#endif
#ifndef PERL_ARGS_ASSERT_REPORT_UNINIT
#define PERL_ARGS_ASSERT_REPORT_UNINIT
#endif
#ifndef PERL_ARGS_ASSERT_REPORT_WRONGWAY_FH
#define PERL_ARGS_ASSERT_REPORT_WRONGWAY_FH
#endif
#ifndef PERL_ARGS_ASSERT_REQUIRE_PV
#define PERL_ARGS_ASSERT_REQUIRE_PV assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_RESUME_COMPCV
#define PERL_ARGS_ASSERT_RESUME_COMPCV assert(buffer)
#endif
#ifndef PERL_ARGS_ASSERT_RESUME_COMPCV_AND_SAVE
#define PERL_ARGS_ASSERT_RESUME_COMPCV_AND_SAVE assert(buffer)
#endif
#ifndef PERL_ARGS_ASSERT_RESUME_COMPCV_FINAL
#define PERL_ARGS_ASSERT_RESUME_COMPCV_FINAL assert(buffer)
#endif
#ifndef PERL_ARGS_ASSERT_RE_COMPILE
#define PERL_ARGS_ASSERT_RE_COMPILE assert(pattern)
#endif
#ifndef PERL_ARGS_ASSERT_RE_INTUIT_START
#define PERL_ARGS_ASSERT_RE_INTUIT_START assert(rx); assert(strbeg); assert(strpos); assert(strend); assert(strbeg <= strpos); assert(strpos <= strend)
#endif
#ifndef PERL_ARGS_ASSERT_RE_INTUIT_STRING
#define PERL_ARGS_ASSERT_RE_INTUIT_STRING assert(r)
#endif
#ifndef PERL_ARGS_ASSERT_RE_OP_COMPILE
#define PERL_ARGS_ASSERT_RE_OP_COMPILE assert(eng)
#endif
#ifndef PERL_ARGS_ASSERT_RNINSTR
#define PERL_ARGS_ASSERT_RNINSTR assert(big); assert(bigend); assert(little); assert(lend); assert(big <= bigend); assert(little <= lend)
#endif
#ifndef PERL_ARGS_ASSERT_RPEEP
#define PERL_ARGS_ASSERT_RPEEP
#endif
#ifndef PERL_ARGS_ASSERT_RPP_FREE_2_
#define PERL_ARGS_ASSERT_RPP_FREE_2_ assert(sv1); assert(sv2)
#endif
#ifndef PERL_ARGS_ASSERT_RPP_OBLITERATE_STACK_TO
#define PERL_ARGS_ASSERT_RPP_OBLITERATE_STACK_TO
#endif
#ifndef PERL_ARGS_ASSERT_RSIGNAL
#define PERL_ARGS_ASSERT_RSIGNAL
#endif
#ifndef PERL_ARGS_ASSERT_RSIGNAL_RESTORE
#define PERL_ARGS_ASSERT_RSIGNAL_RESTORE
#endif
#ifndef PERL_ARGS_ASSERT_RSIGNAL_SAVE
#define PERL_ARGS_ASSERT_RSIGNAL_SAVE assert(save)
#endif
#ifndef PERL_ARGS_ASSERT_RSIGNAL_STATE
#define PERL_ARGS_ASSERT_RSIGNAL_STATE
#endif
#ifndef PERL_ARGS_ASSERT_RUNOPS_DEBUG
#define PERL_ARGS_ASSERT_RUNOPS_DEBUG
#endif
#ifndef PERL_ARGS_ASSERT_RUNOPS_STANDARD
#define PERL_ARGS_ASSERT_RUNOPS_STANDARD
#endif
#ifndef PERL_ARGS_ASSERT_RV2CV_OP_CV
#define PERL_ARGS_ASSERT_RV2CV_OP_CV assert(cvop)
#endif
#ifndef PERL_ARGS_ASSERT_RXRES_SAVE
#define PERL_ARGS_ASSERT_RXRES_SAVE assert(rsp); assert(rx)
#endif
#ifndef PERL_ARGS_ASSERT_SAFESYSCALLOC
#define PERL_ARGS_ASSERT_SAFESYSCALLOC
#endif
#ifndef PERL_ARGS_ASSERT_SAFESYSFREE
#define PERL_ARGS_ASSERT_SAFESYSFREE
#endif
#ifndef PERL_ARGS_ASSERT_SAFESYSMALLOC
#define PERL_ARGS_ASSERT_SAFESYSMALLOC
#endif
#ifndef PERL_ARGS_ASSERT_SAFESYSREALLOC
#define PERL_ARGS_ASSERT_SAFESYSREALLOC
#endif
#ifndef PERL_ARGS_ASSERT_SAVESHAREDPV
#define PERL_ARGS_ASSERT_SAVESHAREDPV
#endif
#ifndef PERL_ARGS_ASSERT_SAVESHAREDPVN
#define PERL_ARGS_ASSERT_SAVESHAREDPVN assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVESTACK_GROW
#define PERL_ARGS_ASSERT_SAVESTACK_GROW
#endif
#ifndef PERL_ARGS_ASSERT_SAVESTACK_GROW_CNT
#define PERL_ARGS_ASSERT_SAVESTACK_GROW_CNT
#endif
#ifndef PERL_ARGS_ASSERT_SAVETMPS
#define PERL_ARGS_ASSERT_SAVETMPS
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_ADELETE
#define PERL_ARGS_ASSERT_SAVE_ADELETE assert(av); assert(SvTYPE(av) == SVt_PVAV)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_AELEM
#define PERL_ARGS_ASSERT_SAVE_AELEM assert(av); assert(SvTYPE(av) == SVt_PVAV); assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_AELEM_FLAGS
#define PERL_ARGS_ASSERT_SAVE_AELEM_FLAGS assert(av); assert(SvTYPE(av) == SVt_PVAV); assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_ALLOC
#define PERL_ARGS_ASSERT_SAVE_ALLOC
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_APTR
#define PERL_ARGS_ASSERT_SAVE_APTR assert(aptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_ARY
#define PERL_ARGS_ASSERT_SAVE_ARY assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_BOOL
#define PERL_ARGS_ASSERT_SAVE_BOOL assert(boolp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_CLEARSV
#define PERL_ARGS_ASSERT_SAVE_CLEARSV assert(svp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_DELETE
#define PERL_ARGS_ASSERT_SAVE_DELETE assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(key)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_DESTRUCTOR
#define PERL_ARGS_ASSERT_SAVE_DESTRUCTOR assert(p)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_DESTRUCTOR_X
#define PERL_ARGS_ASSERT_SAVE_DESTRUCTOR_X
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_FREEOP
#define PERL_ARGS_ASSERT_SAVE_FREEOP
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_FREEPV
#define PERL_ARGS_ASSERT_SAVE_FREEPV
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_FREERCPV
#define PERL_ARGS_ASSERT_SAVE_FREERCPV assert(rcpv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_FREESV
#define PERL_ARGS_ASSERT_SAVE_FREESV
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_GENERIC_PVREF
#define PERL_ARGS_ASSERT_SAVE_GENERIC_PVREF assert(str)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_GENERIC_SVREF
#define PERL_ARGS_ASSERT_SAVE_GENERIC_SVREF assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_GP
#define PERL_ARGS_ASSERT_SAVE_GP assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HASH
#define PERL_ARGS_ASSERT_SAVE_HASH assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HDELETE
#define PERL_ARGS_ASSERT_SAVE_HDELETE assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(keysv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HELEM
#define PERL_ARGS_ASSERT_SAVE_HELEM assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(key); assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HELEM_FLAGS
#define PERL_ARGS_ASSERT_SAVE_HELEM_FLAGS assert(hv); assert(SvTYPE(hv) == SVt_PVHV); assert(key); assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HINTS
#define PERL_ARGS_ASSERT_SAVE_HINTS
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_HPTR
#define PERL_ARGS_ASSERT_SAVE_HPTR assert(hptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_I16
#define PERL_ARGS_ASSERT_SAVE_I16 assert(intp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_I32
#define PERL_ARGS_ASSERT_SAVE_I32 assert(intp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_I8
#define PERL_ARGS_ASSERT_SAVE_I8 assert(bytep)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_INT
#define PERL_ARGS_ASSERT_SAVE_INT assert(intp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_ITEM
#define PERL_ARGS_ASSERT_SAVE_ITEM assert(item)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_IV
#define PERL_ARGS_ASSERT_SAVE_IV assert(ivp)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_MORTALIZESV
#define PERL_ARGS_ASSERT_SAVE_MORTALIZESV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_OP
#define PERL_ARGS_ASSERT_SAVE_OP
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PADSV
#define PERL_ARGS_ASSERT_SAVE_PADSV
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PADSV_AND_MORTALIZE
#define PERL_ARGS_ASSERT_SAVE_PADSV_AND_MORTALIZE
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PPTR
#define PERL_ARGS_ASSERT_SAVE_PPTR assert(pptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PUSHI32PTR
#define PERL_ARGS_ASSERT_SAVE_PUSHI32PTR
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PUSHPTR
#define PERL_ARGS_ASSERT_SAVE_PUSHPTR
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_PUSHPTRPTR
#define PERL_ARGS_ASSERT_SAVE_PUSHPTRPTR
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_RCPV
#define PERL_ARGS_ASSERT_SAVE_RCPV assert(prcpv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_RE_CONTEXT
#define PERL_ARGS_ASSERT_SAVE_RE_CONTEXT
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_SCALAR
#define PERL_ARGS_ASSERT_SAVE_SCALAR assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_SET_SVFLAGS
#define PERL_ARGS_ASSERT_SAVE_SET_SVFLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_SHARED_PVREF
#define PERL_ARGS_ASSERT_SAVE_SHARED_PVREF assert(str)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_SPTR
#define PERL_ARGS_ASSERT_SAVE_SPTR assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_STRLEN
#define PERL_ARGS_ASSERT_SAVE_STRLEN assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_SVREF
#define PERL_ARGS_ASSERT_SAVE_SVREF assert(sptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAVE_VPTR
#define PERL_ARGS_ASSERT_SAVE_VPTR assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_SAWPARENS
#define PERL_ARGS_ASSERT_SAWPARENS
#endif
#ifndef PERL_ARGS_ASSERT_SCALAR
#define PERL_ARGS_ASSERT_SCALAR
#endif
#ifndef PERL_ARGS_ASSERT_SCALARVOID
#define PERL_ARGS_ASSERT_SCALARVOID assert(arg)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_BIN
#define PERL_ARGS_ASSERT_SCAN_BIN assert(start); assert(retlen)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_HEX
#define PERL_ARGS_ASSERT_SCAN_HEX assert(start); assert(retlen)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_NUM
#define PERL_ARGS_ASSERT_SCAN_NUM assert(start); assert(lvalp)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_OCT
#define PERL_ARGS_ASSERT_SCAN_OCT assert(start); assert(retlen)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_STR
#define PERL_ARGS_ASSERT_SCAN_STR assert(start)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_VERSION
#define PERL_ARGS_ASSERT_SCAN_VERSION assert(s); assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_VSTRING
#define PERL_ARGS_ASSERT_SCAN_VSTRING assert(s); assert(e); assert(sv); assert(s < e)
#endif
#ifndef PERL_ARGS_ASSERT_SCAN_WORD
#define PERL_ARGS_ASSERT_SCAN_WORD assert(s); assert(dest); assert(slp)
#endif
#ifndef PERL_ARGS_ASSERT_SEED
#define PERL_ARGS_ASSERT_SEED
#endif
#ifndef PERL_ARGS_ASSERT_SETDEFOUT
#define PERL_ARGS_ASSERT_SETDEFOUT assert(gv)
#endif
#ifndef PERL_ARGS_ASSERT_SETFD_CLOEXEC
#define PERL_ARGS_ASSERT_SETFD_CLOEXEC
#endif
#ifndef PERL_ARGS_ASSERT_SETFD_CLOEXEC_FOR_NONSYSFD
#define PERL_ARGS_ASSERT_SETFD_CLOEXEC_FOR_NONSYSFD
#endif
#ifndef PERL_ARGS_ASSERT_SETFD_CLOEXEC_OR_INHEXEC_BY_SYSFDNESS
#define PERL_ARGS_ASSERT_SETFD_CLOEXEC_OR_INHEXEC_BY_SYSFDNESS
#endif
#ifndef PERL_ARGS_ASSERT_SETFD_INHEXEC
#define PERL_ARGS_ASSERT_SETFD_INHEXEC
#endif
#ifndef PERL_ARGS_ASSERT_SETFD_INHEXEC_FOR_SYSFD
#define PERL_ARGS_ASSERT_SETFD_INHEXEC_FOR_SYSFD
#endif
#ifndef PERL_ARGS_ASSERT_SET_CARET_X
#define PERL_ARGS_ASSERT_SET_CARET_X
#endif
#ifndef PERL_ARGS_ASSERT_SET_CONTEXT
#define PERL_ARGS_ASSERT_SET_CONTEXT assert(t)
#endif
#ifndef PERL_ARGS_ASSERT_SET_NUMERIC_STANDARD
#define PERL_ARGS_ASSERT_SET_NUMERIC_STANDARD assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_SET_NUMERIC_UNDERLYING
#define PERL_ARGS_ASSERT_SET_NUMERIC_UNDERLYING assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_SHARE_HEK
#define PERL_ARGS_ASSERT_SHARE_HEK assert(str)
#endif
#ifndef PERL_ARGS_ASSERT_SIGHANDLER1
#define PERL_ARGS_ASSERT_SIGHANDLER1
#endif
#ifndef PERL_ARGS_ASSERT_SIGHANDLER3
#define PERL_ARGS_ASSERT_SIGHANDLER3
#endif
#ifndef PERL_ARGS_ASSERT_SKIPSPACE_FLAGS
#define PERL_ARGS_ASSERT_SKIPSPACE_FLAGS assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_SLAB_ALLOC
#define PERL_ARGS_ASSERT_SLAB_ALLOC
#endif
#ifndef PERL_ARGS_ASSERT_SLAB_FREE
#define PERL_ARGS_ASSERT_SLAB_FREE assert(op)
#endif
#ifndef PERL_ARGS_ASSERT_SORTSV
#define PERL_ARGS_ASSERT_SORTSV assert(cmp)
#endif
#ifndef PERL_ARGS_ASSERT_SORTSV_FLAGS
#define PERL_ARGS_ASSERT_SORTSV_FLAGS assert(cmp)
#endif
#ifndef PERL_ARGS_ASSERT_STACK_GROW
#define PERL_ARGS_ASSERT_STACK_GROW assert(sp); assert(p)
#endif
#ifndef PERL_ARGS_ASSERT_START_GLOB
#define PERL_ARGS_ASSERT_START_GLOB assert(tmpglob); assert(io)
#endif
#ifndef PERL_ARGS_ASSERT_START_SUBPARSE
#define PERL_ARGS_ASSERT_START_SUBPARSE
#endif
#ifndef PERL_ARGS_ASSERT_STRICT_UTF8_TO_UV
#define PERL_ARGS_ASSERT_STRICT_UTF8_TO_UV assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_STR_TO_VERSION
#define PERL_ARGS_ASSERT_STR_TO_VERSION assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_FENCE_OP
#define PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_FENCE_OP assert(o)
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_NAMED
#define PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_NAMED assert(paramname)
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_POSITIONAL
#define PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_POSITIONAL
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_SLURPY
#define PERL_ARGS_ASSERT_SUBSIGNATURE_APPEND_SLURPY
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_FINISH
#define PERL_ARGS_ASSERT_SUBSIGNATURE_FINISH
#endif
#ifndef PERL_ARGS_ASSERT_SUBSIGNATURE_START
#define PERL_ARGS_ASSERT_SUBSIGNATURE_START
#endif
#ifndef PERL_ARGS_ASSERT_SUB_CRUSH_DEPTH
#define PERL_ARGS_ASSERT_SUB_CRUSH_DEPTH assert(cv); assert(SvTYPE(cv) == SVt_PVCV || SvTYPE(cv) == SVt_PVFM)
#endif
#ifndef PERL_ARGS_ASSERT_SUSPEND_COMPCV
#define PERL_ARGS_ASSERT_SUSPEND_COMPCV assert(buffer)
#endif
#ifndef PERL_ARGS_ASSERT_SVREFCNT_DEC_SET_NULL
#define PERL_ARGS_ASSERT_SVREFCNT_DEC_SET_NULL
#endif
#ifndef PERL_ARGS_ASSERT_SV_2BOOL
#define PERL_ARGS_ASSERT_SV_2BOOL assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2BOOL_FLAGS
#define PERL_ARGS_ASSERT_SV_2BOOL_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2CV
#define PERL_ARGS_ASSERT_SV_2CV assert(st); assert(gvp)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2IO
#define PERL_ARGS_ASSERT_SV_2IO assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2IV
#define PERL_ARGS_ASSERT_SV_2IV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2IV_FLAGS
#define PERL_ARGS_ASSERT_SV_2IV_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2MORTAL
#define PERL_ARGS_ASSERT_SV_2MORTAL
#endif
#ifndef PERL_ARGS_ASSERT_SV_2NUM_FLAGS
#define PERL_ARGS_ASSERT_SV_2NUM_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2NV_FLAGS
#define PERL_ARGS_ASSERT_SV_2NV_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PV
#define PERL_ARGS_ASSERT_SV_2PV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVBYTE
#define PERL_ARGS_ASSERT_SV_2PVBYTE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVBYTE_FLAGS
#define PERL_ARGS_ASSERT_SV_2PVBYTE_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVBYTE_NOLEN
#define PERL_ARGS_ASSERT_SV_2PVBYTE_NOLEN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVUTF8
#define PERL_ARGS_ASSERT_SV_2PVUTF8 assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVUTF8_FLAGS
#define PERL_ARGS_ASSERT_SV_2PVUTF8_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PVUTF8_NOLEN
#define PERL_ARGS_ASSERT_SV_2PVUTF8_NOLEN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PV_FLAGS
#define PERL_ARGS_ASSERT_SV_2PV_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2PV_NOLEN
#define PERL_ARGS_ASSERT_SV_2PV_NOLEN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2UV
#define PERL_ARGS_ASSERT_SV_2UV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_2UV_FLAGS
#define PERL_ARGS_ASSERT_SV_2UV_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_BACKOFF
#define PERL_ARGS_ASSERT_SV_BACKOFF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_BLESS
#define PERL_ARGS_ASSERT_SV_BLESS assert(sv); assert(stash)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CAN_SWIPE_PV_BUF
#define PERL_ARGS_ASSERT_SV_CAN_SWIPE_PV_BUF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPV
#define PERL_ARGS_ASSERT_SV_CATPV assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPVF
#define PERL_ARGS_ASSERT_SV_CATPVF assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPVF_MG
#define PERL_ARGS_ASSERT_SV_CATPVF_MG assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPVN
#define PERL_ARGS_ASSERT_SV_CATPVN assert(dsv); assert(sstr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPVN_FLAGS
#define PERL_ARGS_ASSERT_SV_CATPVN_FLAGS assert(dsv); assert(sstr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPVN_MG
#define PERL_ARGS_ASSERT_SV_CATPVN_MG assert(dsv); assert(sstr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPV_FLAGS
#define PERL_ARGS_ASSERT_SV_CATPV_FLAGS assert(dsv); assert(sstr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATPV_MG
#define PERL_ARGS_ASSERT_SV_CATPV_MG assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATSV
#define PERL_ARGS_ASSERT_SV_CATSV assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATSV_FLAGS
#define PERL_ARGS_ASSERT_SV_CATSV_FLAGS assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CATSV_MG
#define PERL_ARGS_ASSERT_SV_CATSV_MG assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CAT_DECODE
#define PERL_ARGS_ASSERT_SV_CAT_DECODE assert(dsv); assert(encoding); assert(ssv); assert(offset); assert(tstr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CHOP
#define PERL_ARGS_ASSERT_SV_CHOP assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CLEAN_ALL
#define PERL_ARGS_ASSERT_SV_CLEAN_ALL
#endif
#ifndef PERL_ARGS_ASSERT_SV_CLEAN_OBJS
#define PERL_ARGS_ASSERT_SV_CLEAN_OBJS
#endif
#ifndef PERL_ARGS_ASSERT_SV_CLEAR
#define PERL_ARGS_ASSERT_SV_CLEAR assert(orig_sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_CMP
#define PERL_ARGS_ASSERT_SV_CMP
#endif
#ifndef PERL_ARGS_ASSERT_SV_CMP_FLAGS
#define PERL_ARGS_ASSERT_SV_CMP_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_CMP_LOCALE
#define PERL_ARGS_ASSERT_SV_CMP_LOCALE
#endif
#ifndef PERL_ARGS_ASSERT_SV_CMP_LOCALE_FLAGS
#define PERL_ARGS_ASSERT_SV_CMP_LOCALE_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_COPYPV
#define PERL_ARGS_ASSERT_SV_COPYPV assert(dsv); assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_COPYPV_FLAGS
#define PERL_ARGS_ASSERT_SV_COPYPV_FLAGS assert(dsv); assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_COPYPV_NOMG
#define PERL_ARGS_ASSERT_SV_COPYPV_NOMG assert(dsv); assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DEC
#define PERL_ARGS_ASSERT_SV_DEC
#endif
#ifndef PERL_ARGS_ASSERT_SV_DEC_NOMG
#define PERL_ARGS_ASSERT_SV_DEC_NOMG
#endif
#ifndef PERL_ARGS_ASSERT_SV_DEL_BACKREF
#define PERL_ARGS_ASSERT_SV_DEL_BACKREF assert(tsv); assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DERIVED_FROM
#define PERL_ARGS_ASSERT_SV_DERIVED_FROM assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DERIVED_FROM_HV
#define PERL_ARGS_ASSERT_SV_DERIVED_FROM_HV assert(sv); assert(hv); assert(SvTYPE(hv) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DERIVED_FROM_PV
#define PERL_ARGS_ASSERT_SV_DERIVED_FROM_PV assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DERIVED_FROM_PVN
#define PERL_ARGS_ASSERT_SV_DERIVED_FROM_PVN assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DERIVED_FROM_SV
#define PERL_ARGS_ASSERT_SV_DERIVED_FROM_SV assert(sv); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DESTROYABLE
#define PERL_ARGS_ASSERT_SV_DESTROYABLE
#endif
#ifndef PERL_ARGS_ASSERT_SV_DOES
#define PERL_ARGS_ASSERT_SV_DOES assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DOES_PV
#define PERL_ARGS_ASSERT_SV_DOES_PV assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DOES_PVN
#define PERL_ARGS_ASSERT_SV_DOES_PVN assert(sv); assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DOES_SV
#define PERL_ARGS_ASSERT_SV_DOES_SV assert(sv); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_DUMP
#define PERL_ARGS_ASSERT_SV_DUMP
#endif
#ifndef PERL_ARGS_ASSERT_SV_DUMP_DEPTH
#define PERL_ARGS_ASSERT_SV_DUMP_DEPTH
#endif
#ifndef PERL_ARGS_ASSERT_SV_EQ
#define PERL_ARGS_ASSERT_SV_EQ
#endif
#ifndef PERL_ARGS_ASSERT_SV_EQ_FLAGS
#define PERL_ARGS_ASSERT_SV_EQ_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_FORCE_NORMAL
#define PERL_ARGS_ASSERT_SV_FORCE_NORMAL assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_FORCE_NORMAL_FLAGS
#define PERL_ARGS_ASSERT_SV_FORCE_NORMAL_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_FREE
#define PERL_ARGS_ASSERT_SV_FREE
#endif
#ifndef PERL_ARGS_ASSERT_SV_FREE2
#define PERL_ARGS_ASSERT_SV_FREE2 assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_FREE_ARENAS
#define PERL_ARGS_ASSERT_SV_FREE_ARENAS
#endif
#ifndef PERL_ARGS_ASSERT_SV_GETS
#define PERL_ARGS_ASSERT_SV_GETS assert(sv); assert(fp)
#endif
#ifndef PERL_ARGS_ASSERT_SV_GET_BACKREFS
#define PERL_ARGS_ASSERT_SV_GET_BACKREFS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_GROW
#define PERL_ARGS_ASSERT_SV_GROW assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_GROW_FRESH
#define PERL_ARGS_ASSERT_SV_GROW_FRESH assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_INC
#define PERL_ARGS_ASSERT_SV_INC
#endif
#ifndef PERL_ARGS_ASSERT_SV_INC_NOMG
#define PERL_ARGS_ASSERT_SV_INC_NOMG
#endif
#ifndef PERL_ARGS_ASSERT_SV_INSERT
#define PERL_ARGS_ASSERT_SV_INSERT assert(bigstr); assert(little)
#endif
#ifndef PERL_ARGS_ASSERT_SV_INSERT_FLAGS
#define PERL_ARGS_ASSERT_SV_INSERT_FLAGS assert(bigstr); assert(little)
#endif
#ifndef PERL_ARGS_ASSERT_SV_ISA
#define PERL_ARGS_ASSERT_SV_ISA assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_SV_ISA_SV
#define PERL_ARGS_ASSERT_SV_ISA_SV assert(sv); assert(namesv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_ISOBJECT
#define PERL_ARGS_ASSERT_SV_ISOBJECT
#endif
#ifndef PERL_ARGS_ASSERT_SV_LANGINFO
#define PERL_ARGS_ASSERT_SV_LANGINFO
#endif
#ifndef PERL_ARGS_ASSERT_SV_LEN
#define PERL_ARGS_ASSERT_SV_LEN
#endif
#ifndef PERL_ARGS_ASSERT_SV_LEN_UTF8
#define PERL_ARGS_ASSERT_SV_LEN_UTF8
#endif
#ifndef PERL_ARGS_ASSERT_SV_LEN_UTF8_NOMG
#define PERL_ARGS_ASSERT_SV_LEN_UTF8_NOMG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_MAGIC
#define PERL_ARGS_ASSERT_SV_MAGIC assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_MAGICEXT
#define PERL_ARGS_ASSERT_SV_MAGICEXT assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_MAGICEXT_MGLOB
#define PERL_ARGS_ASSERT_SV_MAGICEXT_MGLOB assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_MORTALCOPY
#define PERL_ARGS_ASSERT_SV_MORTALCOPY
#endif
#ifndef PERL_ARGS_ASSERT_SV_MORTALCOPY_FLAGS
#define PERL_ARGS_ASSERT_SV_MORTALCOPY_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NEWMORTAL
#define PERL_ARGS_ASSERT_SV_NEWMORTAL
#endif
#ifndef PERL_ARGS_ASSERT_SV_NEWREF
#define PERL_ARGS_ASSERT_SV_NEWREF
#endif
#ifndef PERL_ARGS_ASSERT_SV_NOSHARING
#define PERL_ARGS_ASSERT_SV_NOSHARING
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMCMP
#define PERL_ARGS_ASSERT_SV_NUMCMP
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMCMP_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMCMP_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMEQ
#define PERL_ARGS_ASSERT_SV_NUMEQ
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMEQ_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMEQ_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMGE
#define PERL_ARGS_ASSERT_SV_NUMGE
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMGE_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMGE_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMGT
#define PERL_ARGS_ASSERT_SV_NUMGT
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMGT_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMGT_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMLE
#define PERL_ARGS_ASSERT_SV_NUMLE
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMLE_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMLE_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMLT
#define PERL_ARGS_ASSERT_SV_NUMLT
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMLT_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMLT_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMNE
#define PERL_ARGS_ASSERT_SV_NUMNE
#endif
#ifndef PERL_ARGS_ASSERT_SV_NUMNE_FLAGS
#define PERL_ARGS_ASSERT_SV_NUMNE_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_PEEK
#define PERL_ARGS_ASSERT_SV_PEEK
#endif
#ifndef PERL_ARGS_ASSERT_SV_POS_B2U
#define PERL_ARGS_ASSERT_SV_POS_B2U assert(offsetp)
#endif
#ifndef PERL_ARGS_ASSERT_SV_POS_B2U_FLAGS
#define PERL_ARGS_ASSERT_SV_POS_B2U_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_POS_U2B
#define PERL_ARGS_ASSERT_SV_POS_U2B assert(offsetp)
#endif
#ifndef PERL_ARGS_ASSERT_SV_POS_U2B_FLAGS
#define PERL_ARGS_ASSERT_SV_POS_U2B_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PV
#define PERL_ARGS_ASSERT_SV_PV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVBYTE
#define PERL_ARGS_ASSERT_SV_PVBYTE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVBYTEN_FORCE
#define PERL_ARGS_ASSERT_SV_PVBYTEN_FORCE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVN_FORCE
#define PERL_ARGS_ASSERT_SV_PVN_FORCE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVN_FORCE_FLAGS
#define PERL_ARGS_ASSERT_SV_PVN_FORCE_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVUTF8
#define PERL_ARGS_ASSERT_SV_PVUTF8 assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_PVUTF8N_FORCE
#define PERL_ARGS_ASSERT_SV_PVUTF8N_FORCE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_RECODE_TO_UTF8
#define PERL_ARGS_ASSERT_SV_RECODE_TO_UTF8 assert(sv); assert(encoding)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REF
#define PERL_ARGS_ASSERT_SV_REF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REFTYPE
#define PERL_ARGS_ASSERT_SV_REFTYPE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_CLEAR
#define PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_CLEAR assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_GET
#define PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_GET assert(sv); assert(posp)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_SET
#define PERL_ARGS_ASSERT_SV_REGEX_GLOBAL_POS_SET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REPLACE
#define PERL_ARGS_ASSERT_SV_REPLACE assert(sv); assert(nsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_REPORT_USED
#define PERL_ARGS_ASSERT_SV_REPORT_USED
#endif
#ifndef PERL_ARGS_ASSERT_SV_RESET
#define PERL_ARGS_ASSERT_SV_RESET assert(s); assert(!stash || SvTYPE(stash) == SVt_PVHV)
#endif
#ifndef PERL_ARGS_ASSERT_SV_RESETPVN
#define PERL_ARGS_ASSERT_SV_RESETPVN
#endif
#ifndef PERL_ARGS_ASSERT_SV_RVUNWEAKEN
#define PERL_ARGS_ASSERT_SV_RVUNWEAKEN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_RVWEAKEN
#define PERL_ARGS_ASSERT_SV_RVWEAKEN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETHEK
#define PERL_ARGS_ASSERT_SV_SETHEK assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETIV
#define PERL_ARGS_ASSERT_SV_SETIV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETIV_MG
#define PERL_ARGS_ASSERT_SV_SETIV_MG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETNV
#define PERL_ARGS_ASSERT_SV_SETNV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETNV_MG
#define PERL_ARGS_ASSERT_SV_SETNV_MG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPV
#define PERL_ARGS_ASSERT_SV_SETPV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPVF
#define PERL_ARGS_ASSERT_SV_SETPVF assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPVF_MG
#define PERL_ARGS_ASSERT_SV_SETPVF_MG assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPVN
#define PERL_ARGS_ASSERT_SV_SETPVN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPVN_FRESH
#define PERL_ARGS_ASSERT_SV_SETPVN_FRESH assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPVN_MG
#define PERL_ARGS_ASSERT_SV_SETPVN_MG assert(sv); assert(ptr)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPV_BUFSIZE
#define PERL_ARGS_ASSERT_SV_SETPV_BUFSIZE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETPV_MG
#define PERL_ARGS_ASSERT_SV_SETPV_MG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETREF_IV
#define PERL_ARGS_ASSERT_SV_SETREF_IV assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETREF_NV
#define PERL_ARGS_ASSERT_SV_SETREF_NV assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETREF_PV
#define PERL_ARGS_ASSERT_SV_SETREF_PV assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETREF_PVN
#define PERL_ARGS_ASSERT_SV_SETREF_PVN assert(rv); assert(pv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETREF_UV
#define PERL_ARGS_ASSERT_SV_SETREF_UV assert(rv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETRV_INC
#define PERL_ARGS_ASSERT_SV_SETRV_INC assert(sv); assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETRV_INC_MG
#define PERL_ARGS_ASSERT_SV_SETRV_INC_MG assert(sv); assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETRV_NOINC
#define PERL_ARGS_ASSERT_SV_SETRV_NOINC assert(sv); assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETRV_NOINC_MG
#define PERL_ARGS_ASSERT_SV_SETRV_NOINC_MG assert(sv); assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETSV
#define PERL_ARGS_ASSERT_SV_SETSV assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETSV_FLAGS
#define PERL_ARGS_ASSERT_SV_SETSV_FLAGS assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETSV_MG
#define PERL_ARGS_ASSERT_SV_SETSV_MG assert(dsv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETUV
#define PERL_ARGS_ASSERT_SV_SETUV assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SETUV_MG
#define PERL_ARGS_ASSERT_SV_SETUV_MG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SET_BOOL
#define PERL_ARGS_ASSERT_SV_SET_BOOL assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SET_FALSE
#define PERL_ARGS_ASSERT_SV_SET_FALSE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SET_TRUE
#define PERL_ARGS_ASSERT_SV_SET_TRUE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_SET_UNDEF
#define PERL_ARGS_ASSERT_SV_SET_UNDEF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_STREQ
#define PERL_ARGS_ASSERT_SV_STREQ
#endif
#ifndef PERL_ARGS_ASSERT_SV_STREQ_FLAGS
#define PERL_ARGS_ASSERT_SV_STREQ_FLAGS
#endif
#ifndef PERL_ARGS_ASSERT_SV_STRFTIME_INTS
#define PERL_ARGS_ASSERT_SV_STRFTIME_INTS assert(fmt)
#endif
#ifndef PERL_ARGS_ASSERT_SV_STRFTIME_TM
#define PERL_ARGS_ASSERT_SV_STRFTIME_TM assert(fmt); assert(mytm)
#endif
#ifndef PERL_ARGS_ASSERT_SV_STRING_FROM_ERRNUM
#define PERL_ARGS_ASSERT_SV_STRING_FROM_ERRNUM
#endif
#ifndef PERL_ARGS_ASSERT_SV_TAINT
#define PERL_ARGS_ASSERT_SV_TAINT assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_TAINTED
#define PERL_ARGS_ASSERT_SV_TAINTED assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_TRUE
#define PERL_ARGS_ASSERT_SV_TRUE
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNI_DISPLAY
#define PERL_ARGS_ASSERT_SV_UNI_DISPLAY assert(dsv); assert(ssv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNMAGIC
#define PERL_ARGS_ASSERT_SV_UNMAGIC assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNMAGICEXT
#define PERL_ARGS_ASSERT_SV_UNMAGICEXT assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNREF
#define PERL_ARGS_ASSERT_SV_UNREF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNREF_FLAGS
#define PERL_ARGS_ASSERT_SV_UNREF_FLAGS assert(ref)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UNTAINT
#define PERL_ARGS_ASSERT_SV_UNTAINT assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UPGRADE
#define PERL_ARGS_ASSERT_SV_UPGRADE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_USEPVN
#define PERL_ARGS_ASSERT_SV_USEPVN assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_USEPVN_FLAGS
#define PERL_ARGS_ASSERT_SV_USEPVN_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_USEPVN_MG
#define PERL_ARGS_ASSERT_SV_USEPVN_MG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_DECODE
#define PERL_ARGS_ASSERT_SV_UTF8_DECODE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE
#define PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE_FLAGS
#define PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE_NOMG
#define PERL_ARGS_ASSERT_SV_UTF8_DOWNGRADE_NOMG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_ENCODE
#define PERL_ARGS_ASSERT_SV_UTF8_ENCODE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_UPGRADE
#define PERL_ARGS_ASSERT_SV_UTF8_UPGRADE assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_FLAGS
#define PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_FLAGS assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_FLAGS_GROW
#define PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_FLAGS_GROW assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_NOMG
#define PERL_ARGS_ASSERT_SV_UTF8_UPGRADE_NOMG assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VCATPVF
#define PERL_ARGS_ASSERT_SV_VCATPVF assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VCATPVFN
#define PERL_ARGS_ASSERT_SV_VCATPVFN assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VCATPVFN_FLAGS
#define PERL_ARGS_ASSERT_SV_VCATPVFN_FLAGS assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VCATPVF_MG
#define PERL_ARGS_ASSERT_SV_VCATPVF_MG assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VSETPVF
#define PERL_ARGS_ASSERT_SV_VSETPVF assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VSETPVFN
#define PERL_ARGS_ASSERT_SV_VSETPVFN assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VSETPVF_MG
#define PERL_ARGS_ASSERT_SV_VSETPVF_MG assert(sv); assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_SV_VSTRING_GET
#define PERL_ARGS_ASSERT_SV_VSTRING_GET assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_SWITCH_TO_GLOBAL_LOCALE
#define PERL_ARGS_ASSERT_SWITCH_TO_GLOBAL_LOCALE
#endif
#ifndef PERL_ARGS_ASSERT_SYNC_LOCALE
#define PERL_ARGS_ASSERT_SYNC_LOCALE
#endif
#ifndef PERL_ARGS_ASSERT_SYS_INIT
#define PERL_ARGS_ASSERT_SYS_INIT assert(argc); assert(argv)
#endif
#ifndef PERL_ARGS_ASSERT_SYS_INIT3
#define PERL_ARGS_ASSERT_SYS_INIT3 assert(argc); assert(argv); assert(env)
#endif
#ifndef PERL_ARGS_ASSERT_SYS_TERM
#define PERL_ARGS_ASSERT_SYS_TERM
#endif
#ifndef PERL_ARGS_ASSERT_TAINT_ENV
#define PERL_ARGS_ASSERT_TAINT_ENV
#endif
#ifndef PERL_ARGS_ASSERT_TAINT_PROPER
#define PERL_ARGS_ASSERT_TAINT_PROPER assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_TIED_METHOD
#define PERL_ARGS_ASSERT_TIED_METHOD assert(methname); assert(mark); assert(sv); assert(mg)
#endif
#ifndef PERL_ARGS_ASSERT_TMPS_GROW_P
#define PERL_ARGS_ASSERT_TMPS_GROW_P
#endif
#ifndef PERL_ARGS_ASSERT_TO_UNI_FOLD
#define PERL_ARGS_ASSERT_TO_UNI_FOLD assert(p); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UNI_FOLD_FLAGS_
#define PERL_ARGS_ASSERT_TO_UNI_FOLD_FLAGS_ assert(p); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UNI_LOWER
#define PERL_ARGS_ASSERT_TO_UNI_LOWER assert(p); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UNI_TITLE
#define PERL_ARGS_ASSERT_TO_UNI_TITLE assert(p); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UNI_UPPER
#define PERL_ARGS_ASSERT_TO_UNI_UPPER assert(p); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UTF8_FOLD_FLAGS_
#define PERL_ARGS_ASSERT_TO_UTF8_FOLD_FLAGS_ assert(p); assert(e); assert(ustrp); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UTF8_LOWER_FLAGS_
#define PERL_ARGS_ASSERT_TO_UTF8_LOWER_FLAGS_ assert(p); assert(e); assert(ustrp); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UTF8_TITLE_FLAGS_
#define PERL_ARGS_ASSERT_TO_UTF8_TITLE_FLAGS_ assert(p); assert(e); assert(ustrp); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_TO_UTF8_UPPER_FLAGS_
#define PERL_ARGS_ASSERT_TO_UTF8_UPPER_FLAGS_ assert(p); assert(e); assert(ustrp); assert(p < e)
#endif
#ifndef PERL_ARGS_ASSERT_TRY_AMAGIC_BIN
#define PERL_ARGS_ASSERT_TRY_AMAGIC_BIN
#endif
#ifndef PERL_ARGS_ASSERT_TRY_AMAGIC_UN
#define PERL_ARGS_ASSERT_TRY_AMAGIC_UN
#endif
#ifndef PERL_ARGS_ASSERT_UIV_2BUF
#define PERL_ARGS_ASSERT_UIV_2BUF assert(buf); assert(peob)
#endif
#ifndef PERL_ARGS_ASSERT_UNPACKSTRING
#define PERL_ARGS_ASSERT_UNPACKSTRING assert(pat); assert(patend); assert(s); assert(strend); assert(pat <= patend); assert(s <= strend)
#endif
#ifndef PERL_ARGS_ASSERT_UNSHAREPVN
#define PERL_ARGS_ASSERT_UNSHAREPVN
#endif
#ifndef PERL_ARGS_ASSERT_UNSHARE_HEK
#define PERL_ARGS_ASSERT_UNSHARE_HEK
#endif
#ifndef PERL_ARGS_ASSERT_UPG_VERSION
#define PERL_ARGS_ASSERT_UPG_VERSION assert(ver)
#endif
#ifndef PERL_ARGS_ASSERT_UTF16_TO_UTF8_BASE
#define PERL_ARGS_ASSERT_UTF16_TO_UTF8_BASE assert(p); assert(d); assert(newlen)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8N_TO_UVCHR
#define PERL_ARGS_ASSERT_UTF8N_TO_UVCHR assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8N_TO_UVCHR_ERROR
#define PERL_ARGS_ASSERT_UTF8N_TO_UVCHR_ERROR assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_HOP_BACK
#define PERL_ARGS_ASSERT_UTF8_HOP_BACK assert(s); assert(start); assert(start <= s)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_HOP_FORWARD
#define PERL_ARGS_ASSERT_UTF8_HOP_FORWARD assert(s); assert(end); assert(s <= end)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_HOP_SAFE
#define PERL_ARGS_ASSERT_UTF8_HOP_SAFE assert(s); assert(start); assert(end); assert(start <= s); assert(s <= end)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_LENGTH
#define PERL_ARGS_ASSERT_UTF8_LENGTH assert(s0); assert(e); assert(s0 <= e)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_BYTES
#define PERL_ARGS_ASSERT_UTF8_TO_BYTES assert(s); assert(lenp)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_BYTES_
#define PERL_ARGS_ASSERT_UTF8_TO_BYTES_ assert(s_ptr); assert(lenp); assert(free_me)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_UTF16_BASE
#define PERL_ARGS_ASSERT_UTF8_TO_UTF16_BASE assert(s); assert(d); assert(newlen)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_UV
#define PERL_ARGS_ASSERT_UTF8_TO_UV assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_UV_ERRORS
#define PERL_ARGS_ASSERT_UTF8_TO_UV_ERRORS assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_UV_FLAGS
#define PERL_ARGS_ASSERT_UTF8_TO_UV_FLAGS assert(s); assert(e); assert(cp_p); assert(s <= e)
#endif
#ifndef PERL_ARGS_ASSERT_UTF8_TO_UV_MSGS_HELPER_
#define PERL_ARGS_ASSERT_UTF8_TO_UV_MSGS_HELPER_ assert(s0); assert(e); assert(cp_p); assert(s0 <= e)
#endif
#ifndef PERL_ARGS_ASSERT_UTILIZE
#define PERL_ARGS_ASSERT_UTILIZE assert(idop)
#endif
#ifndef PERL_ARGS_ASSERT_UVCHR_TO_UTF8
#define PERL_ARGS_ASSERT_UVCHR_TO_UTF8 assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_UVCHR_TO_UTF8_FLAGS
#define PERL_ARGS_ASSERT_UVCHR_TO_UTF8_FLAGS assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_UVCHR_TO_UTF8_FLAGS_MSGS
#define PERL_ARGS_ASSERT_UVCHR_TO_UTF8_FLAGS_MSGS assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_UVOFFUNI_TO_UTF8_FLAGS
#define PERL_ARGS_ASSERT_UVOFFUNI_TO_UTF8_FLAGS assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_UVOFFUNI_TO_UTF8_FLAGS_MSGS
#define PERL_ARGS_ASSERT_UVOFFUNI_TO_UTF8_FLAGS_MSGS assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_UV_TO_UTF8_MSGS
#define PERL_ARGS_ASSERT_UV_TO_UTF8_MSGS assert(d)
#endif
#ifndef PERL_ARGS_ASSERT_VALIDATE_PROTO
#define PERL_ARGS_ASSERT_VALIDATE_PROTO assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_VALID_IDENTIFIER_PVE
#define PERL_ARGS_ASSERT_VALID_IDENTIFIER_PVE assert(s); assert(end); assert(s <= end)
#endif
#ifndef PERL_ARGS_ASSERT_VALID_IDENTIFIER_PVN
#define PERL_ARGS_ASSERT_VALID_IDENTIFIER_PVN assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_VALID_IDENTIFIER_SV
#define PERL_ARGS_ASSERT_VALID_IDENTIFIER_SV
#endif
#ifndef PERL_ARGS_ASSERT_VALID_UTF8_TO_UVCHR
#define PERL_ARGS_ASSERT_VALID_UTF8_TO_UVCHR assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_VCMP
#define PERL_ARGS_ASSERT_VCMP assert(lhv); assert(rhv)
#endif
#ifndef PERL_ARGS_ASSERT_VCROAK
#define PERL_ARGS_ASSERT_VCROAK
#endif
#ifndef PERL_ARGS_ASSERT_VDEB
#define PERL_ARGS_ASSERT_VDEB assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VFATAL_WARNER
#define PERL_ARGS_ASSERT_VFATAL_WARNER assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VFORM
#define PERL_ARGS_ASSERT_VFORM assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VIVIFY_DEFELEM
#define PERL_ARGS_ASSERT_VIVIFY_DEFELEM assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_VIVIFY_REF
#define PERL_ARGS_ASSERT_VIVIFY_REF assert(sv)
#endif
#ifndef PERL_ARGS_ASSERT_VLOAD_MODULE
#define PERL_ARGS_ASSERT_VLOAD_MODULE assert(name)
#endif
#ifndef PERL_ARGS_ASSERT_VMESS
#define PERL_ARGS_ASSERT_VMESS assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VNEWSVPVF
#define PERL_ARGS_ASSERT_VNEWSVPVF assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VNORMAL
#define PERL_ARGS_ASSERT_VNORMAL assert(vs)
#endif
#ifndef PERL_ARGS_ASSERT_VNUMIFY
#define PERL_ARGS_ASSERT_VNUMIFY assert(vs)
#endif
#ifndef PERL_ARGS_ASSERT_VSTRINGIFY
#define PERL_ARGS_ASSERT_VSTRINGIFY assert(vs)
#endif
#ifndef PERL_ARGS_ASSERT_VVERIFY
#define PERL_ARGS_ASSERT_VVERIFY assert(vs)
#endif
#ifndef PERL_ARGS_ASSERT_VWARN
#define PERL_ARGS_ASSERT_VWARN assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_VWARNER
#define PERL_ARGS_ASSERT_VWARNER assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_WAIT4PID
#define PERL_ARGS_ASSERT_WAIT4PID assert(statusp)
#endif
#ifndef PERL_ARGS_ASSERT_WARN
#define PERL_ARGS_ASSERT_WARN assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_WARNER
#define PERL_ARGS_ASSERT_WARNER assert(pat)
#endif
#ifndef PERL_ARGS_ASSERT_WARN_SV
#define PERL_ARGS_ASSERT_WARN_SV assert(baseex)
#endif
#ifndef PERL_ARGS_ASSERT_WAS_LVALUE_SUB
#define PERL_ARGS_ASSERT_WAS_LVALUE_SUB
#endif
#ifndef PERL_ARGS_ASSERT_WATCH
#define PERL_ARGS_ASSERT_WATCH assert(addr)
#endif
#ifndef PERL_ARGS_ASSERT_WHICHSIG
#define PERL_ARGS_ASSERT_WHICHSIG assert(sig)
#endif
#ifndef PERL_ARGS_ASSERT_WHICHSIG_PV
#define PERL_ARGS_ASSERT_WHICHSIG_PV assert(sig)
#endif
#ifndef PERL_ARGS_ASSERT_WHICHSIG_PVN
#define PERL_ARGS_ASSERT_WHICHSIG_PVN assert(sig)
#endif
#ifndef PERL_ARGS_ASSERT_WHICHSIG_SV
#define PERL_ARGS_ASSERT_WHICHSIG_SV assert(sigsv)
#endif
#ifndef PERL_ARGS_ASSERT_WRAP_INFIX_PLUGIN
#define PERL_ARGS_ASSERT_WRAP_INFIX_PLUGIN assert(new_plugin); assert(old_plugin_p)
#endif
#ifndef PERL_ARGS_ASSERT_WRAP_KEYWORD_PLUGIN
#define PERL_ARGS_ASSERT_WRAP_KEYWORD_PLUGIN assert(new_plugin); assert(old_plugin_p)
#endif
#ifndef PERL_ARGS_ASSERT_WRAP_OP_CHECKER
#define PERL_ARGS_ASSERT_WRAP_OP_CHECKER assert(new_checker); assert(old_checker_p)
#endif
#ifndef PERL_ARGS_ASSERT_WRITE_TO_STDERR
#define PERL_ARGS_ASSERT_WRITE_TO_STDERR assert(msv)
#endif
#ifndef PERL_ARGS_ASSERT_XS_BOOT_EPILOG
#define PERL_ARGS_ASSERT_XS_BOOT_EPILOG
#endif
#ifndef PERL_ARGS_ASSERT_XS_HANDSHAKE
#define PERL_ARGS_ASSERT_XS_HANDSHAKE assert(v_my_perl); assert(file)
#endif
#ifndef PERL_ARGS_ASSERT_YYERROR
#define PERL_ARGS_ASSERT_YYERROR assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_YYERROR_PV
#define PERL_ARGS_ASSERT_YYERROR_PV assert(s)
#endif
#ifndef PERL_ARGS_ASSERT_YYERROR_PVN
#define PERL_ARGS_ASSERT_YYERROR_PVN
#endif
#ifndef PERL_ARGS_ASSERT_YYLEX
#define PERL_ARGS_ASSERT_YYLEX
#endif
#ifndef PERL_ARGS_ASSERT_YYPARSE
#define PERL_ARGS_ASSERT_YYPARSE
#endif
#ifndef PERL_ARGS_ASSERT_YYQUIT
#define PERL_ARGS_ASSERT_YYQUIT
#endif
#ifndef PERL_ARGS_ASSERT_YYUNLEX
#define PERL_ARGS_ASSERT_YYUNLEX
#endif
#endif /* GOPERL_XS_SDK_ARGSASSERT_H */
