package json

import (
	"errors"
	"fmt"
	"unicode"
)

const (
	ObjStart = '{'
	ObjEnd   = '}'
	ArrStart = '['
	ArrEnd   = ']'
	SepColon = ':'
	SepComma = ','

	BoolT = 't'
	BoolF = 'f'

	NullStart = 'n'

	ControlCharacter = 0x20
)

const (
	QuotationMark         = '"'
	ReverseSolidus        = '\\'
	Solidus               = '/'
	Backspace             = 'b'
	FromFeed              = 'f'
	NEWLINE               = 'n'
	CarriageReturn        = 'r'
	HorizontalTab         = 't'
	FourHexadecimalDigits = 'u'
)

const (
	NumberDot   = '.'
	Numbere     = 'e'
	NumberE     = 'E'
	NumberPlus  = '+'
	NumberMinus = '-'
	NumberZero  = '0'
)

var (
	ErrInvalidJSON   = errors.New("invalid json format")
	ErrUnexpectedEOF = errors.New("unexpected end of JSON")
	ErrStringEscape  = errors.New("get an invalid escape character")
)

type JSON struct {
	jsonBytes   []byte
	position    uint
	maxPosition uint
}

func (j *JSON) len() int {
	return len(j.jsonBytes)
}

func (j *JSON) validateLen(x uint) {
	if j.maxPosition <= j.position {
		fmt.Println("asd")
		panic(ErrJSON{
			err:  ErrUnexpectedEOF,
			part: getPartOfJSON(j),
		})
	}
}

func (j *JSON) moveX(x uint) *JSON {
	if x == 0 {
		return j
	}

	j.validateLen(x)

	j.jsonBytes = j.jsonBytes[x:]
	j.position += x
	return j
}

func (j *JSON) moveOne() *JSON {
	return j.moveX(1)
}

func (j *JSON) byteX(x uint) byte {
	j.validateLen(x)

	return j.jsonBytes[x]
}

func (j *JSON) firstByte() byte {
	return j.byteX(0)
}

type ErrJSON struct {
	err        error
	additional string
	part       string
}

func (e ErrJSON) Error() string {
	return e.String()
}

func (e ErrJSON) String() string {
	return fmt.Sprintf("error:\n\t%s\nadditional:\n\t%s\n"+
		"occur at:\n\t %s\n", e.err, e.additional, e.part)
}

func Expect(b byte, data *JSON) {
	if data.firstByte() != b {
		panic(ErrJSON{
			err:        ErrInvalidJSON,
			additional: fmt.Sprintf("expect character: %c", b),
			part:       getPartOfJSON(data),
		})
	}
	TrimLeftSpace(data.moveOne())
	return
}

func Validate(jsonStr string) (err error) {
	defer func() {
		if e := recover(); e != nil {
			if e, ok := e.(error); ok {
				err = e.(error)
			} else {
				panic(e)
			}
		}
	}()

	data := &JSON{[]byte(jsonStr), 0, uint(len(jsonStr))}

	TrimLeftSpace(data)
	if data.firstByte() == ObjStart {
		ValidateObj(data)

		if TrimLeftSpace(data).len() == 0 {
			return nil
		}
	} else if data.firstByte() == ArrStart {
		ValidateArr(data)

		if TrimLeftSpace(data).len() == 0 {
			return nil
		}
	}

	return ErrJSON{
		err:        ErrInvalidJSON,
		additional: "extra characters after parsing",
		part:       getPartOfJSON(data),
	}
}

func ValidateObj(data *JSON) {
	Expect(ObjStart, data)

	if TrimLeftSpace(data).firstByte() == ObjEnd {
		data.moveOne()
		return
	}

	for {
		ValidateStr(TrimLeftSpace(data))

		Expect(SepColon, TrimLeftSpace(data))

		ValidateValue(TrimLeftSpace(data))

		TrimLeftSpace(data)

		if data.firstByte() == SepComma {
			data.moveOne()
		} else if data.firstByte() == ObjEnd {
			data.moveOne()
			return
		} else {
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: `expect any one of the following characters: ','  '}'`,
				part:       getPartOfJSON(data),
			})
		}
	}
}

func ValidateArr(data *JSON) {
	Expect(ArrStart, data)

	if TrimLeftSpace(data).firstByte() == ArrEnd {
		data.moveOne()
		return
	}

	for {
		ValidateValue(TrimLeftSpace(data))

		TrimLeftSpace(data)
		if data.firstByte() == SepComma {
			data.moveOne()
		} else if data.firstByte() == ArrEnd {
			data.moveOne()
			return
		} else {
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: `expect any one of the following characters: ','  ']'`,
				part:       getPartOfJSON(data),
			})
		}
	}
}

func ValidateStr(data *JSON) {
	Expect(QuotationMark, data)

	var needEsc bool

RE_VALID:
	for idx, r := range data.jsonBytes {
		if needEsc {
			ValidateEsc(data.moveX(uint(idx)))
			needEsc = false
			goto RE_VALID
		}

		switch {
		case r == QuotationMark:
			data.moveX(uint(idx + 1))
			return
		case r == ReverseSolidus:
			needEsc = true
		case r < ControlCharacter:
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: "control characters are not allowed in string type(< 0x20)",
				part:       getPartOfJSON(data),
			})
		}
	}

	panic(ErrJSON{
		err:  ErrUnexpectedEOF,
		part: getPartOfJSON(data),
	})
}

func ValidateEsc(data *JSON) {
	switch data.firstByte() {
	case QuotationMark, ReverseSolidus, Solidus, Backspace, FromFeed,
		NEWLINE, CarriageReturn, HorizontalTab:
		TrimLeftSpace(data.moveOne())
		return
	case FourHexadecimalDigits:
		for i := 1; i <= 4; i++ {
			switch {
			case data.byteX(uint(i)) >= '0' && data.byteX(uint(i)) <= '9':
			case data.byteX(uint(i)) >= 'A' && data.byteX(uint(i)) <= 'F':
			case data.byteX(uint(i)) >= 'a' && data.byteX(uint(i)) <= 'f':
			default:
				panic(ErrJSON{
					err:        ErrStringEscape,
					additional: `expect to get unicode characters consisting of \u and 4 hexadecimal digits`,
					part:       getPartOfJSON(data),
				})
			}
		}
		TrimLeftSpace(data.moveX(5))
	default:
		panic(ErrJSON{
			err:        ErrStringEscape,
			additional: `expect to get unicode characters consisting of \u and 4 hexadecimal digits, or any one of the following characters: '"'  '\'  '/'  'b'  'f'  'n'  'r'  't'`,
			part:       getPartOfJSON(data),
		})
	}
	return
}

func ValidateValue(data *JSON) {
	b := data.firstByte()
	switch {
	case b == QuotationMark:
		ValidateStr(data)
	case b == ObjStart:
		ValidateObj(data)
	case b == ArrStart:
		ValidateArr(data)
	case b == BoolT:
		if data.byteX(1) != 'r' || data.byteX(2) != 'u' ||
			data.byteX(3) != 'e' {
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: "expect a bool value: true",
				part:       getPartOfJSON(data),
			})
		}
		data.moveX(4)
		return
	case b == BoolF:
		if data.byteX(1) != 'a' || data.byteX(2) != 'l' ||
			data.byteX(3) != 's' || data.byteX(4) != 'e' {
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: "expect a bool value: false",
				part:       getPartOfJSON(data),
			})
		}
		data.moveX(5)
		return
	case b == NullStart:
		if data.byteX(1) != 'u' || data.byteX(2) != 'l' ||
			data.byteX(3) != 'l' {
			panic(ErrJSON{
				err:        ErrInvalidJSON,
				additional: "expect a null value: null",
				part:       getPartOfJSON(data),
			})
		}
		data.moveX(4)
		return
	case b == NumberMinus || b == NumberZero || (b >= '1' && b <= '9'):
		ValidateNumber(data)
	default:
		panic(ErrJSON{
			err:        ErrInvalidJSON,
			additional: `expect any one of the following characters: '"'  '{'  '['  't'  'f'  'n'  '-'  '0'  '1'  '2'  '3'  '4'  '5'  '6'  '7'  '8'  '9'`,
			part:       getPartOfJSON(data),
		})
	}

	return
}

func ValidateNumber(data *JSON) {
	if data.firstByte() == NumberMinus {
		data.moveOne()
	}

	if data.firstByte() == NumberZero {
		data.moveOne()
		// do nothing, maybe need read continuous '0' character
	} else if data.firstByte() >= '1' || data.firstByte() <= '9' {
		data.moveOne()

		if data.firstByte() >= '0' && data.firstByte() <= '9' {
			ValidateDigit(data)
		}
	} else {
		panic(ErrJSON{
			err:        ErrInvalidJSON,
			additional: `expect any one of the following characters: '-'  '0'  '1'  '2'  '3'  '4'  '5'  '6'  '7'  '8'  '9'`,
			part:       getPartOfJSON(data),
		})
	}

	if data.firstByte() == NumberDot {
		ValidateDigit(data.moveOne())
	}

	if data.firstByte() != Numbere && data.firstByte() != NumberE {
		return
	}

	data.moveOne()

	if data.firstByte() == NumberPlus || data.firstByte() == NumberMinus {
		data.moveOne()
	}

	ValidateDigit(data)

	return
}

func ValidateDigit(data *JSON) {
	if data.firstByte() < '0' || data.firstByte() > '9' {
		panic(ErrJSON{
			err:        ErrInvalidJSON,
			additional: "expect any one of the following characters: '0'  '1'  '2'  '3'  '4'  '5'  '6'  '7'  '8'  '9'",
			part:       getPartOfJSON(data),
		})
	}

	data.moveOne()

	for idx, b := range data.jsonBytes {
		if b < '0' || b > '9' {
			data.moveX(uint(idx))
			return
		}
	}

	panic(ErrJSON{
		err:  ErrUnexpectedEOF,
		part: getPartOfJSON(data),
	})
}

func TrimLeftSpace(data *JSON) *JSON {
	for idx, r := range data.jsonBytes {
		if !unicode.IsSpace(rune(r)) {
			return data.moveX(uint(idx))
		}
	}
	return data.moveX(uint(data.len()))
}

func getPartOfJSON(data *JSON) string {
	return string([]rune(string(data.jsonBytes[:160]))[:40])
}
