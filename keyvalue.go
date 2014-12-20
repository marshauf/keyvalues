package keyvalues

import (
	"encoding/binary"
	"fmt"
	"io"
)

const (
	TokenType_String TokenType = iota
	TokenType_ChildStart
	TokenType_ChildEnd
)

type TokenType byte

type Token struct {
	Type  TokenType
	Value string
}

type KeyValue struct {
	Key      string
	Value    string
	children []*KeyValue
}

func (kv *KeyValue) SetChild(value *KeyValue) {
	for i, child := range kv.children {
		if child.Key == value.Key {
			kv.children[i] = value
			return
		}
	}
	kv.children = append(kv.children, value)
}

func (kv *KeyValue) GetChild(key string) *KeyValue {
	for _, child := range kv.children {
		if child.Key == key {
			return child
		}
	}
	return nil
}

func (kv *KeyValue) String() string {
	str := fmt.Sprintf("\"%s\" ", kv.Key)
	if len(kv.Value) == 0 {
		if kv.children != nil {
			str += "\n{\n"
			for _, t := range kv.children {
				str += t.String()
			}
			str += "}\n"
		}
	} else {
		str += fmt.Sprintf("\"%s\"\n", kv.Value)
	}
	return str
}

func Unmarshal(data []byte) (*KeyValue, error) {
	line := 1
	tokens := []*Token{}
	for i := 0; i < len(data); i++ {
		switch data[i] {
		case '"':
			j := i + 1
			for {
				if j >= len(data) {
					return nil, fmt.Errorf("EOF")
				}
				if data[j] == '"' {
					if data[j-1] == '\\' && data[j-2] != '\\' {
						j++
						continue
					}
					break
				} else {
					j++
				}
			}
			str := string(data[i+1 : j])
			tokens = append(tokens, &Token{Type: TokenType_String, Value: str})
			i = j + 1
		case ' ', '\t':
			continue
		case '\n', '\r':
			line++
			continue
		case '{':
			tokens = append(tokens, &Token{Type: TokenType_ChildStart, Value: "{"})
		case '}':
			tokens = append(tokens, &Token{Type: TokenType_ChildEnd, Value: "}"})
		case byte(0):
			break
		default:
			fmt.Printf("Last token: %v", tokens[len(tokens)-1])
			return nil, fmt.Errorf("Unhandled char \"%s\" at char %d in line %d\n", string(data[i]), i, line)
		}
	}
	/*
		for i := range tokens {
			fmt.Printf("tokens[%d]: %v\n", i, tokens[i])
		}
		return nil, nil
	*/

	root := &KeyValue{}
	readObject(tokens, root)

	return root.children[0], nil // Return root or first root children? Is it possible that multiple top level key, value pairs can exist?
}

const (
	TypeNone byte = iota
	TypeString
	TypeInt32
	TypeFloat32
	TypePointer
	TypeWideString
	TypeColor
	TypeUint64
	TypeEnd
)

func UnmarshalBinary(rd io.Reader) (*KeyValue, error) {
	root := &KeyValue{}
	err := readBinaryObject(rd, root)
	return root, err
}

func readBinaryObject(rd io.Reader, kv *KeyValue) error {
	t, err := ReadByte(rd)
	if err != nil {
		return err
	}
	for {
		if t == TypeEnd {
			break
		}

		current := &KeyValue{}

		// current.Key, err = rd.ReadString('\x00')
		current.Key, err = ReadString(rd)
		if err != nil {
			return err
		}

		switch t {
		case TypeNone:
			kv.SetChild(current)
			err = readBinaryObject(rd, current)
			if err != nil {
				return err
			}
		case TypeString:
			//current.Value, err = rd.ReadString('\x00')
			current.Value, err = ReadString(rd)
			if err != nil {
				return err
			}
		case TypeWideString: // https://github.com/ValveSoftware/source-sdk-2013/blob/master/mp/src/tier1/kvpacker.cpp#L206
			return fmt.Errorf("WideString not supported")
		case TypeInt32, TypeColor, TypePointer:
			n, err := ReadInt32(rd)
			if err != nil {
				return err
			}
			current.Value = fmt.Sprintf("%d", n)
		case TypeUint64:
			n, err := ReadUint32(rd)
			if err != nil {
				return err
			}
			current.Value = fmt.Sprintf("%d", n)
		case TypeFloat32:
			n, err := ReadFloat32(rd)
			if err != nil {
				return err
			}
			current.Value = fmt.Sprintf("%g", n)
		}

		t, err = ReadByte(rd)
		if err != nil {
			return err
		}

		if t == TypeEnd {
			break
		}

		//fmt.Printf("Type:\"%d\" Key:\"%s\" Value:\"%s\"\n", t, current.Key, current.Value)
		kv.SetChild(current)
	}
	return nil
}

func ReadFloat32(r io.Reader) (float32, error) {
	var c float32
	err := binary.Read(r, binary.LittleEndian, &c)
	return c, err
}

func ReadByte(r io.Reader) (byte, error) {
	var c byte
	err := binary.Read(r, binary.LittleEndian, &c)
	return c, err
}

func ReadInt32(r io.Reader) (int32, error) {
	var c int32
	err := binary.Read(r, binary.LittleEndian, &c)
	return c, err
}

func ReadUint32(r io.Reader) (uint32, error) {
	var c uint32
	err := binary.Read(r, binary.LittleEndian, &c)
	return c, err
}

func ReadString(r io.Reader) (string, error) {
	c := make([]byte, 0)
	var err error
	for {
		var b byte
		err = binary.Read(r, binary.LittleEndian, &b)
		if b == byte(0x0) || err != nil {
			break
		}
		c = append(c, b)
	}
	return string(c), err
}

func readObject(tokens []*Token, root *KeyValue) int {
	for i := 0; i < len(tokens); i++ {
		switch tokens[i].Type {
		case TokenType_String:
			// Peek ahead
			switch tokens[i+1].Type {
			case TokenType_String: // key, value (string)
				root.SetChild(&KeyValue{Key: tokens[i].Value, Value: tokens[i+1].Value})
				i++
			case TokenType_ChildStart: // key, value (object)
				child := &KeyValue{Key: tokens[i].Value}
				read := readObject(tokens[i+2:], child)
				root.SetChild(child)
				i += 2 + read
			}
		case TokenType_ChildEnd:
			return i
		default:
			panic(fmt.Errorf("Unknown token type. %v", tokens[i]))
		}
	}
	return len(tokens)
}
