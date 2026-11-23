package internal

// childSiginfo is darwin's siginfo_t (waitid(2)): si_signo, si_errno,
// si_code, si_pid, si_uid and si_status, within its 104 bytes.
type childSiginfo struct {
	signo, errno, code, pid, uid, status int32
	_                                    [104 - 24]byte
}
