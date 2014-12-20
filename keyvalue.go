package keyvalues

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
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
	if len(kv.Value) == 0 && kv.children != nil {
		str += "\n{\n"
		for _, t := range kv.children {
			str += t.String()
		}
		str += "}\n"
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
	err := readBinaryObject(bufio.NewReader(rd), root)
	return root, err
}

func readBinaryObject(rd *bufio.Reader, kv *KeyValue) error {
	t, err := rd.ReadByte()
	if err != nil {
		return err
	}
	for {
		if t == TypeEnd {
			break
		}

		current := &KeyValue{}
		current.Key, err = rd.ReadString('\000')
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
			current.Value, err = rd.ReadString('\000')
			if err != nil {
				return err
			}
		case TypeWideString: // https://github.com/ValveSoftware/source-sdk-2013/blob/master/mp/src/tier1/kvpacker.cpp#L206
			return fmt.Errorf("WideString not supported")
		case TypeInt32, TypeColor, TypePointer:
			b := make([]byte, 4)
			n, err := rd.Read(b)
			if err != nil {
				return err
			}
			if n != 4 {
				return fmt.Errorf("Read less than four bytes for 32bits.")
			}
			current.Value = strconv.Itoa(int(b[0]<<24 + b[1]<<16 + b[2]<<8 + b[3]))
		case TypeUint64:
			// TODO
			return fmt.Errorf("Uint64 Not supported")
		case TypeFloat32:
			// TODO
			return fmt.Errorf("Float32 not supported")
		}

		t, err = rd.ReadByte()
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
