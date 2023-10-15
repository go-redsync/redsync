// Code generated DO NOT EDIT

package cmds

import "strconv"

type Hdel Incomplete

func (b Builder) Hdel() (c Hdel) {
	c = Hdel{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HDEL")
	return c
}

func (c Hdel) Key(key string) HdelKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HdelKey)(c)
}

type HdelField Incomplete

func (c HdelField) Field(field ...string) HdelField {
	c.cs.s = append(c.cs.s, field...)
	return c
}

func (c HdelField) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HdelKey Incomplete

func (c HdelKey) Field(field ...string) HdelField {
	c.cs.s = append(c.cs.s, field...)
	return (HdelField)(c)
}

type Hexists Incomplete

func (b Builder) Hexists() (c Hexists) {
	c = Hexists{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HEXISTS")
	return c
}

func (c Hexists) Key(key string) HexistsKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HexistsKey)(c)
}

type HexistsField Incomplete

func (c HexistsField) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HexistsField) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HexistsKey Incomplete

func (c HexistsKey) Field(field string) HexistsField {
	c.cs.s = append(c.cs.s, field)
	return (HexistsField)(c)
}

type Hget Incomplete

func (b Builder) Hget() (c Hget) {
	c = Hget{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HGET")
	return c
}

func (c Hget) Key(key string) HgetKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HgetKey)(c)
}

type HgetField Incomplete

func (c HgetField) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HgetField) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HgetKey Incomplete

func (c HgetKey) Field(field string) HgetField {
	c.cs.s = append(c.cs.s, field)
	return (HgetField)(c)
}

type Hgetall Incomplete

func (b Builder) Hgetall() (c Hgetall) {
	c = Hgetall{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HGETALL")
	return c
}

func (c Hgetall) Key(key string) HgetallKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HgetallKey)(c)
}

type HgetallKey Incomplete

func (c HgetallKey) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HgetallKey) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hincrby Incomplete

func (b Builder) Hincrby() (c Hincrby) {
	c = Hincrby{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HINCRBY")
	return c
}

func (c Hincrby) Key(key string) HincrbyKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HincrbyKey)(c)
}

type HincrbyField Incomplete

func (c HincrbyField) Increment(increment int64) HincrbyIncrement {
	c.cs.s = append(c.cs.s, strconv.FormatInt(increment, 10))
	return (HincrbyIncrement)(c)
}

type HincrbyIncrement Incomplete

func (c HincrbyIncrement) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HincrbyKey Incomplete

func (c HincrbyKey) Field(field string) HincrbyField {
	c.cs.s = append(c.cs.s, field)
	return (HincrbyField)(c)
}

type Hincrbyfloat Incomplete

func (b Builder) Hincrbyfloat() (c Hincrbyfloat) {
	c = Hincrbyfloat{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HINCRBYFLOAT")
	return c
}

func (c Hincrbyfloat) Key(key string) HincrbyfloatKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HincrbyfloatKey)(c)
}

type HincrbyfloatField Incomplete

func (c HincrbyfloatField) Increment(increment float64) HincrbyfloatIncrement {
	c.cs.s = append(c.cs.s, strconv.FormatFloat(increment, 'f', -1, 64))
	return (HincrbyfloatIncrement)(c)
}

type HincrbyfloatIncrement Incomplete

func (c HincrbyfloatIncrement) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HincrbyfloatKey Incomplete

func (c HincrbyfloatKey) Field(field string) HincrbyfloatField {
	c.cs.s = append(c.cs.s, field)
	return (HincrbyfloatField)(c)
}

type Hkeys Incomplete

func (b Builder) Hkeys() (c Hkeys) {
	c = Hkeys{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HKEYS")
	return c
}

func (c Hkeys) Key(key string) HkeysKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HkeysKey)(c)
}

type HkeysKey Incomplete

func (c HkeysKey) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HkeysKey) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hlen Incomplete

func (b Builder) Hlen() (c Hlen) {
	c = Hlen{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HLEN")
	return c
}

func (c Hlen) Key(key string) HlenKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HlenKey)(c)
}

type HlenKey Incomplete

func (c HlenKey) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HlenKey) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hmget Incomplete

func (b Builder) Hmget() (c Hmget) {
	c = Hmget{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HMGET")
	return c
}

func (c Hmget) Key(key string) HmgetKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HmgetKey)(c)
}

type HmgetField Incomplete

func (c HmgetField) Field(field ...string) HmgetField {
	c.cs.s = append(c.cs.s, field...)
	return c
}

func (c HmgetField) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HmgetField) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HmgetKey Incomplete

func (c HmgetKey) Field(field ...string) HmgetField {
	c.cs.s = append(c.cs.s, field...)
	return (HmgetField)(c)
}

type Hmset Incomplete

func (b Builder) Hmset() (c Hmset) {
	c = Hmset{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HMSET")
	return c
}

func (c Hmset) Key(key string) HmsetKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HmsetKey)(c)
}

type HmsetFieldValue Incomplete

func (c HmsetFieldValue) FieldValue(field string, value string) HmsetFieldValue {
	c.cs.s = append(c.cs.s, field, value)
	return c
}

func (c HmsetFieldValue) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HmsetKey Incomplete

func (c HmsetKey) FieldValue() HmsetFieldValue {
	return (HmsetFieldValue)(c)
}

type Hrandfield Incomplete

func (b Builder) Hrandfield() (c Hrandfield) {
	c = Hrandfield{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HRANDFIELD")
	return c
}

func (c Hrandfield) Key(key string) HrandfieldKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HrandfieldKey)(c)
}

type HrandfieldKey Incomplete

func (c HrandfieldKey) Count(count int64) HrandfieldOptionsCount {
	c.cs.s = append(c.cs.s, strconv.FormatInt(count, 10))
	return (HrandfieldOptionsCount)(c)
}

func (c HrandfieldKey) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HrandfieldOptionsCount Incomplete

func (c HrandfieldOptionsCount) Withvalues() HrandfieldOptionsWithvalues {
	c.cs.s = append(c.cs.s, "WITHVALUES")
	return (HrandfieldOptionsWithvalues)(c)
}

func (c HrandfieldOptionsCount) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HrandfieldOptionsWithvalues Incomplete

func (c HrandfieldOptionsWithvalues) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hscan Incomplete

func (b Builder) Hscan() (c Hscan) {
	c = Hscan{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HSCAN")
	return c
}

func (c Hscan) Key(key string) HscanKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HscanKey)(c)
}

type HscanCount Incomplete

func (c HscanCount) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HscanCursor Incomplete

func (c HscanCursor) Match(pattern string) HscanMatch {
	c.cs.s = append(c.cs.s, "MATCH", pattern)
	return (HscanMatch)(c)
}

func (c HscanCursor) Count(count int64) HscanCount {
	c.cs.s = append(c.cs.s, "COUNT", strconv.FormatInt(count, 10))
	return (HscanCount)(c)
}

func (c HscanCursor) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HscanKey Incomplete

func (c HscanKey) Cursor(cursor uint64) HscanCursor {
	c.cs.s = append(c.cs.s, strconv.FormatUint(cursor, 10))
	return (HscanCursor)(c)
}

type HscanMatch Incomplete

func (c HscanMatch) Count(count int64) HscanCount {
	c.cs.s = append(c.cs.s, "COUNT", strconv.FormatInt(count, 10))
	return (HscanCount)(c)
}

func (c HscanMatch) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hset Incomplete

func (b Builder) Hset() (c Hset) {
	c = Hset{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HSET")
	return c
}

func (c Hset) Key(key string) HsetKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HsetKey)(c)
}

type HsetFieldValue Incomplete

func (c HsetFieldValue) FieldValue(field string, value string) HsetFieldValue {
	c.cs.s = append(c.cs.s, field, value)
	return c
}

func (c HsetFieldValue) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HsetKey Incomplete

func (c HsetKey) FieldValue() HsetFieldValue {
	return (HsetFieldValue)(c)
}

type Hsetnx Incomplete

func (b Builder) Hsetnx() (c Hsetnx) {
	c = Hsetnx{cs: get(), ks: b.ks}
	c.cs.s = append(c.cs.s, "HSETNX")
	return c
}

func (c Hsetnx) Key(key string) HsetnxKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HsetnxKey)(c)
}

type HsetnxField Incomplete

func (c HsetnxField) Value(value string) HsetnxValue {
	c.cs.s = append(c.cs.s, value)
	return (HsetnxValue)(c)
}

type HsetnxKey Incomplete

func (c HsetnxKey) Field(field string) HsetnxField {
	c.cs.s = append(c.cs.s, field)
	return (HsetnxField)(c)
}

type HsetnxValue Incomplete

func (c HsetnxValue) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type Hstrlen Incomplete

func (b Builder) Hstrlen() (c Hstrlen) {
	c = Hstrlen{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HSTRLEN")
	return c
}

func (c Hstrlen) Key(key string) HstrlenKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HstrlenKey)(c)
}

type HstrlenField Incomplete

func (c HstrlenField) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HstrlenField) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

type HstrlenKey Incomplete

func (c HstrlenKey) Field(field string) HstrlenField {
	c.cs.s = append(c.cs.s, field)
	return (HstrlenField)(c)
}

type Hvals Incomplete

func (b Builder) Hvals() (c Hvals) {
	c = Hvals{cs: get(), ks: b.ks, cf: int16(readonly)}
	c.cs.s = append(c.cs.s, "HVALS")
	return c
}

func (c Hvals) Key(key string) HvalsKey {
	if c.ks&NoSlot == NoSlot {
		c.ks = NoSlot | slot(key)
	} else {
		c.ks = check(c.ks, slot(key))
	}
	c.cs.s = append(c.cs.s, key)
	return (HvalsKey)(c)
}

type HvalsKey Incomplete

func (c HvalsKey) Build() Completed {
	c.cs.Build()
	return Completed{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}

func (c HvalsKey) Cache() Cacheable {
	c.cs.Build()
	return Cacheable{cs: c.cs, cf: uint16(c.cf), ks: c.ks}
}
