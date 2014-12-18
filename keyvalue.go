package keyvalues

import (
	"fmt"
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

	return root.children[0], nil
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
