package go_param_table

type metadata struct {
	hookup_start bufidx
	val_idx      validx
	type_code    ptype
	flags        pflags
	
}

type pflags uint8

const (
	_FLAG_IS_USED pflags = 1 << iota
	_FLAG_ALWAYS_UPDATE
	_FLAG_HAS_CHILDREN
	_FLAG_HAS_PARENT
	_FLAG_HAS_SIBLINGS
	_FLAG_HAS_CALCULATION
)

func (m metadata) is_free() bool {
	return m.flags&_FLAG_IS_USED == 0
}
func (m metadata) is_used() bool {
	return m.flags&_FLAG_IS_USED == _FLAG_IS_USED
}
func (m *metadata) set_free() {
	m.flags &= ^_FLAG_IS_USED
}
func (m *metadata) set_used() {
	m.flags |= _FLAG_IS_USED
}

func (m metadata) should_always_update() bool {
	return m.flags&_FLAG_ALWAYS_UPDATE == _FLAG_ALWAYS_UPDATE
}
func (m metadata) only_update_on_change() bool {
	return m.flags&_FLAG_ALWAYS_UPDATE == 0
}
func (m *metadata) set_always_update(state bool) {
	if state {
		m.flags |= _FLAG_ALWAYS_UPDATE
	} else {
		m.flags &= ^_FLAG_ALWAYS_UPDATE
	}
}

func (m metadata) has_children() bool {
	return m.flags&_FLAG_HAS_CHILDREN == _FLAG_HAS_CHILDREN
}
func (m metadata) no_children() bool {
	return m.flags&_FLAG_HAS_CHILDREN == 0
}
func (m *metadata) set_has_children(state bool) {
	if state {
		m.flags |= _FLAG_HAS_CHILDREN
	} else {
		m.flags &= ^_FLAG_HAS_CHILDREN
	}
}

func (m metadata) has_siblings() bool {
	return m.flags&_FLAG_HAS_SIBLINGS == _FLAG_HAS_SIBLINGS
}
func (m metadata) no_siblings() bool {
	return m.flags&_FLAG_HAS_SIBLINGS == 0
}
func (m *metadata) set_has_siblings(state bool) {
	if state {
		m.flags |= _FLAG_HAS_SIBLINGS
	} else {
		m.flags &= ^_FLAG_HAS_SIBLINGS
	}
}

func (m metadata) has_parent() bool {
	return m.flags&_FLAG_HAS_PARENT == _FLAG_HAS_PARENT
}
func (m metadata) no_parent() bool {
	return m.flags&_FLAG_HAS_PARENT == 0
}
func (m *metadata) set_has_parent(state bool) {
	if state {
		m.flags |= _FLAG_HAS_PARENT
	} else {
		m.flags &= ^_FLAG_HAS_PARENT
	}
}

func (m metadata) has_calc() bool {
	return m.flags&_FLAG_HAS_CALCULATION == _FLAG_HAS_CALCULATION
}
func (m metadata) no_calc() bool {
	return m.flags&_FLAG_HAS_CALCULATION == 0
}
func (m *metadata) set_has_calc(state bool) {
	if state {
		m.flags |= _FLAG_HAS_CALCULATION
	} else {
		m.flags &= ^_FLAG_HAS_CALCULATION
	}
}

func (m metadata) is_type(t ptype) bool {
	return t == m.type_code
}

func (t *ParamTable)