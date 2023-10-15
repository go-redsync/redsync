//go:build !go1.21

package cmds

func Put(cs *CommandSlice) {
	for i := range cs.s {
		cs.s[i] = ""
	}
	cs.s = cs.s[:0]
	cs.l = -1
	cs.r = 0
	pool.Put(cs)
}
