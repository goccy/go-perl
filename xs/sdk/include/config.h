/* go-perl native XS SDK -- config.h: what perl's Configure records about
 * the C environment, for code that asks the way it asks a perl (#ifdef
 * I_STDINT, HAS_STRTOL, ...). An XS module built against this SDK is
 * compiled for the host, a linux or darwin system with a C99 library, so
 * the answers are that system's; the sizes of perl's own types are the
 * embedded interpreter's (64-bit integers, double floating point). */
#ifndef GOPERL_XS_SDK_CONFIG_H
#define GOPERL_XS_SDK_CONFIG_H

/* headers */
#define I_STDLIB
#define I_STRING
#define I_STDARG 1
#define I_STDDEF
#define I_STDBOOL
#define I_STDINT
#define I_INTTYPES
#define I_LIMITS
#define I_FLOAT
#define I_MATH
#define I_LOCALE
#define I_TIME
#define I_UNISTD
#define I_FCNTL
#define I_DIRENT
#define I_DLFCN
#define I_GRP
#define I_PWD
#define I_NETDB
#define I_NETINET_IN
#define I_NETINET_TCP
#define I_ARPA_INET
#define I_POLL
#define I_PTHREAD
#define I_SYS_TYPES
#define I_SYS_STAT
#define I_SYS_TIME
#define I_SYS_TIMES
#define I_SYS_PARAM
#define I_SYS_WAIT
#define I_SYS_IOCTL
#define I_SYS_SELECT
#define I_SYS_SOCKET
#define I_SYS_UN
#define I_SYS_UIO
#define I_SYS_RESOURCE
#define I_SYS_FILE
#define I_SYSLOG
#define I_TERMIOS
#define I_UTIME
#define I_WCHAR
#define I_FENV
#define I_LANGINFO

/* functions of the C library */
#define HAS_MEMCHR
#define HAS_MEMCMP
#define HAS_MEMCPY
#define HAS_MEMMOVE
#define HAS_MEMSET
#define HAS_STRCHR
#define HAS_STRCOLL
#define HAS_STRERROR
#define HAS_STRFTIME
#define HAS_STRTOD
#define HAS_STRTOL
#define HAS_STRTOLL
#define HAS_STRTOUL
#define HAS_STRTOULL
#define HAS_STRXFRM
#define HAS_STRNLEN
#define HAS_SNPRINTF
#define HAS_VSNPRINTF
#define HAS_VPRINTF
#define HAS_ATOLL
#define HAS_ISNAN
#define HAS_ISINF
#define HAS_ISFINITE
#define HAS_FMOD
#define HAS_FREXP
#define HAS_LDEXP
#define HAS_MODFL
#define HAS_GETTIMEOFDAY
#define HAS_TIMES
#define HAS_SELECT
#define HAS_POLL
#define HAS_MKDIR
#define HAS_RMDIR
#define HAS_RENAME
#define HAS_LINK
#define HAS_SYMLINK
#define HAS_READLINK
#define HAS_LSTAT
#define HAS_FCHMOD
#define HAS_FCHOWN
#define HAS_FTRUNCATE
#define HAS_TRUNCATE
#define HAS_FSYNC
#define HAS_FLOCK
#define HAS_FCNTL
#define HAS_IOCTL
#define HAS_DUP2
#define HAS_PIPE
#define HAS_FORK
#define HAS_WAITPID
#define HAS_KILL
#define HAS_GETPPID
#define HAS_SETSID
#define HAS_UMASK
#define HAS_ACCESS
#define HAS_GETCWD
#define HAS_UNAME
#define HAS_GETHOSTNAME
#define HAS_MKSTEMP
#define HAS_MKTIME
#define HAS_LOCALTIME_R
#define HAS_GMTIME_R
#define HAS_NANOSLEEP
#define HAS_USLEEP
#define HAS_SIGACTION
#define HAS_SOCKET
#define HAS_SOCKETPAIR
#define HAS_GETADDRINFO
#define HAS_GETNAMEINFO
#define HAS_GAI_STRERROR
#define HAS_INETNTOP
#define HAS_INETPTON
#define HAS_HTONL
#define HAS_HTONS
#define HAS_NTOHL
#define HAS_NTOHS
#define HAS_SETLOCALE
#define HAS_LOCALECONV
#define HAS_MMAP
#define HAS_MUNMAP
#define HAS_MPROTECT
#define HAS_DLOPEN
#define HAS_DLERROR
#define HAS_SYSCONF
#define HAS_GETPAGESIZE

/* the compiler */
#define HASCONST
#define HASVOLATILE
#define CAN_PROTOTYPE
#define HAS_QUAD
#define HAS_LONG_LONG
#define HAS_LONG_DOUBLE
#define HAS_C99_VARIADIC_MACROS
#define HAS_BUILTIN_EXPECT
#define HAS_BUILTIN_CHOOSE_EXPR
#define HASATTRIBUTE_UNUSED
#define HASATTRIBUTE_NORETURN
#define HASATTRIBUTE_PURE
#define HASATTRIBUTE_FORMAT
#define HASATTRIBUTE_NONNULL
#define HASATTRIBUTE_MALLOC
#define HASATTRIBUTE_WARN_UNUSED_RESULT
#define HASATTRIBUTE_DEPRECATED
#define HASATTRIBUTE_ALWAYS_INLINE
#define HASATTRIBUTE_VISIBILITY
#define PERL_STATIC_INLINE static inline
#define STDCHAR char
#define CHARBITS 8
#define VOIDFLAGS 15
#define Malloc_t void *
#define Free_t void

/* sizes of the C types of the host */
#define SHORTSIZE __SIZEOF_SHORT__
#define INTSIZE __SIZEOF_INT__
#define LONGSIZE __SIZEOF_LONG__
#define LONGLONGSIZE 8
#define PTRSIZE __SIZEOF_POINTER__
#define DOUBLESIZE 8
#define SIZE_T_SIZE __SIZEOF_SIZE_T__
#define Size_t_size __SIZEOF_SIZE_T__
#define Size_t size_t
#define SIZEOF_SHORT 2
#define SIZEOF_INT 4
#define SIZEOF_LONG_LONG 8
#if defined(__LP64__) || defined(_LP64)
#define SIZEOF_LONG 8
#define SIZEOF_VOIDP 8
#define SIZEOF_SIZE_T 8
#else
#define SIZEOF_LONG 4
#define SIZEOF_VOIDP 4
#define SIZEOF_SIZE_T 4
#endif
#if defined(__BYTE_ORDER__) && __BYTE_ORDER__ == __ORDER_BIG_ENDIAN__
#define BYTEORDER 0x87654321
#else
#define BYTEORDER 0x12345678
#endif

/* perl's own types: the embedded interpreter's */
#define IVTYPE int64_t
#define UVTYPE uint64_t
#define NVTYPE double
#define I8TYPE int8_t
#define U8TYPE uint8_t
#define I16TYPE int16_t
#define U16TYPE uint16_t
#define I32TYPE int32_t
#define U32TYPE uint32_t
#define I64TYPE int64_t
#define U64TYPE uint64_t
#define Quad_t int64_t
#define Uquad_t uint64_t
#define QUADKIND 4
#define I8SIZE 1
#define U8SIZE 1
#define I16SIZE 2
#define U16SIZE 2
#define I32SIZE 4
#define U32SIZE 4
#define I64SIZE 8
#define U64SIZE 8
#define NVSIZE 8
#define NV_PRESERVES_UV_BITS 53
#define USE_64_BIT_INT
/* non-blocking I/O as perl's Configure finds it on this system: the flag
 * that makes a descriptor non-blocking, the error a read without data
 * gets, what read returns then, and that end of file is told apart from
 * no data (Filter::Util::Exec) */
#define VAL_O_NONBLOCK O_NONBLOCK
#define VAL_EAGAIN EAGAIN
#define RD_NODATA -1
#define EOF_NONBLOCK

#endif /* GOPERL_XS_SDK_CONFIG_H */
