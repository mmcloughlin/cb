// Package shield provides CPU isolation for benchmark execution.
package shield

import (
	"fmt"

	"github.com/mmcloughlin/cb/internal/errutil"
	"github.com/mmcloughlin/cb/pkg/cpuset"
	"github.com/mmcloughlin/cb/pkg/lg"
)

// Shield uses cpusets to setup exclusive access to some CPUs.
type Shield struct {
	root   string    // root cpuset
	shield string    // shield cpuset (relative to root)
	sys    string    // system cpuset name (relative to root)
	sysn   int       // number of cpus in system cpuset
	l      lg.Logger // logger

	deferred []func() error
}

// Option configures a shield.
type Option func(*Shield)

// NewShield builds a CPU shield.
func NewShield(opts ...Option) *Shield {
	s := &Shield{
		root:   "",
		shield: "shield",
		sys:    "sys",
		sysn:   1,
		l:      lg.Noop(),
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

// WithRoot configures the root cpuset.
func WithRoot(name string) Option {
	return func(s *Shield) { s.root = name }
}

// WithShieldName configures the name of the shield cpuset. Note this is interpreted relative to the root.
func WithShieldName(name string) Option {
	return func(s *Shield) { s.shield = name }
}

// WithSystemName configures the name of the system cpuset. Note this is interpreted
// relative to the root.
func WithSystemName(name string) Option {
	return func(s *Shield) { s.sys = name }
}

// WithSystemNumCPU configures the number of CPUs assigned to the system cpuset.
func WithSystemNumCPU(n int) Option {
	return func(s *Shield) { s.sysn = n }
}

// WithLogger configures the logger for CPU shield operations.
func WithLogger(l lg.Logger) Option {
	return func(s *Shield) { s.l = l }
}

// ShieldName returns the name of the shield cpuset.
func (s *Shield) ShieldName() string {
	return s.shield
}

// Available reports whether the shield mechanism can be applied. Note this is a
// rudimentary check that the environment supports cpusets at all, it is still
// possible that applying the shield would error.
func (s *Shield) Available() bool {
	root := cpuset.NewCPUSet(s.root)
	allcpu, err := root.CPUs()
	return err == nil && len(allcpu) > s.sysn
}

// Apply the configured shielding.
func (s *Shield) Apply() error {
	err := s.apply()
	// On error, attempt to cleanup.
	if err != nil {
		s.l.Printf("shield application failed: %s", err)
		s.l.Printf("attempting reset")
		_ = s.Reset()
	}
	return err
}

func (s *Shield) apply() error {
	// Determine available CPUs.
	root := cpuset.NewCPUSet(s.root)
	allcpu, err := root.CPUs()
	if err != nil {
		return err
	}
	lg.Param(s.l, "root_cpuset", root.Path())
	lg.Param(s.l, "allcpu", allcpu)

	if len(allcpu) <= s.sysn {
		return fmt.Errorf("not enough cpus: require %d for system but root has %d", s.sysn, len(allcpu))
	}

	// Pick CPUs for the system set.
	syscpu, err := pick(allcpu, s.sysn)
	if err != nil {
		return fmt.Errorf("could not pick system cpus: %w", err)
	}
	lg.Param(s.l, "syscpu", syscpu)

	// Create system cpuset.
	sys, err := cpuset.Create(s.sys)
	if err != nil {
		return err
	}
	s.cleanup(sys.Remove)

	// Assign CPUs.
	if err := sys.SetCPUs(syscpu); err != nil {
		return err
	}

	// Assign memory nodes.
	mems, err := root.Mems()
	if err != nil {
		return err
	}

	if err := sys.SetMems(mems); err != nil {
		return err
	}

	// Move all tasks from root to system.
	if err := s.movetasks(root, sys); err != nil {
		return err
	}
	s.cleanup(func() error {
		return s.movetasks(sys, root)
	})

	// Create shield cpuset.
	shield, err := cpuset.Create(s.shield)
	if err != nil {
		return err
	}
	s.cleanup(shield.Remove)

	// Assign CPUs for exclusive use.
	shieldcpu := allcpu.Difference(syscpu)
	if err := shield.SetCPUs(shieldcpu); err != nil {
		return err
	}
	lg.Param(s.l, "shieldcpu", shieldcpu)

	// Memory nodes.
	if err := shield.SetMems(mems); err != nil {
		return err
	}

	return nil
}

// cleanup adds an operation to be called on Reset(). Cleanup functions will be
// called in reverse order, similar to defer.
func (s *Shield) cleanup(f func() error) {
	s.deferred = append(s.deferred, f)
}

// Reset restores cpuset configuration to the state prior to shielding.
func (s *Shield) Reset() error {
	var errs errutil.Errors
	for i := len(s.deferred) - 1; i >= 0; i-- {
		if err := s.deferred[i](); err != nil {
			errs.Add(err)
		}
	}
	return errs.Err()
}

// movetasks moves all tasks form src to dst, with additional logging.
func (s *Shield) movetasks(src, dst *cpuset.CPUSet) error {
	s.l.Printf("move tasks from %s to %s", src.Path(), dst.Path())
	m, err := cpuset.MoveTasks(src, dst)
	if err != nil {
		return err
	}
	lg.Param(s.l, "num_moved", len(m.Moved))
	lg.Param(s.l, "num_nonexistent", len(m.Nonexistent))
	lg.Param(s.l, "num_invalid", len(m.Invalid))
	return nil
}

// pick an n-element subset of s.
func pick(s cpuset.Set, n int) (cpuset.Set, error) {
	if n > len(s) {
		return nil, fmt.Errorf("cannot pick an %d element subset of a set of size %d", n, len(s))
	}
	m := s.SortedMembers()
	return cpuset.NewSet(m[len(s)-n:]...), nil
}