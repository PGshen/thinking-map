package utils

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// PathMatcherCallback 路径匹配器的回调函数类型
type PathMatcherCallback func(value interface{}, path []interface{})

// PathPattern 路径匹配模式类型
type PathPattern struct {
	Tokens   []interface{} // 解析后的标记数组
	Original string        // 原始模式字符串
	Callback PathMatcherCallback
}

// SimplePathMatcher 简化版 JSON 路径匹配系统
type SimplePathMatcher struct {
	patterns []PathPattern // 存储所有注册的模式和回调
}

// NewSimplePathMatcher 创建新的路径匹配器
func NewSimplePathMatcher() *SimplePathMatcher {
	return &SimplePathMatcher{
		patterns: make([]PathPattern, 0),
	}
}

// On 注册一个路径模式和对应的回调函数
func (m *SimplePathMatcher) On(pattern string, callback PathMatcherCallback) *SimplePathMatcher {
	// 解析路径模式为标记数组
	parsedPattern := m.parsePath(pattern)
	m.patterns = append(m.patterns, PathPattern{
		Tokens:   parsedPattern,
		Original: pattern,
		Callback: callback,
	})
	return m
}

// parsePath 解析路径字符串为标记数组
func (m *SimplePathMatcher) parsePath(path string) []interface{} {
	if path == "" || path == "$" {
		return []interface{}{"$"}
	}

	// 移除开头的 $ 和 . 符号
	path = strings.TrimPrefix(path, "$")
	path = strings.TrimPrefix(path, ".")

	// 分割路径
	parts := make([]interface{}, 0)
	currentPart := ""
	inBrackets := false

	for _, char := range path {
		switch char {
		case '.':
			if !inBrackets {
				if currentPart != "" {
					parts = append(parts, currentPart)
					currentPart = ""
				}
			} else {
				currentPart += string(char)
			}
		case '[':
			if currentPart != "" {
				parts = append(parts, currentPart)
				currentPart = ""
			}
			inBrackets = true
		case ']':
			if currentPart == "*" {
				parts = append(parts, "*")
			} else if num, err := strconv.Atoi(currentPart); err == nil {
				parts = append(parts, num)
			}
			currentPart = ""
			inBrackets = false
		default:
			currentPart += string(char)
		}
	}

	if currentPart != "" {
		parts = append(parts, currentPart)
	}
	return parts
}

// CheckPatterns 检查当前路径是否匹配任何注册的模式
func (m *SimplePathMatcher) CheckPatterns(path []interface{}, value interface{}) {
	for _, pattern := range m.patterns {
		if m.matchPath(path, pattern.Tokens) {
			// 如果匹配，调用回调函数
			pattern.Callback(value, path)
		}
	}
}

// matchPath 检查路径是否匹配模式
func (m *SimplePathMatcher) matchPath(path []interface{}, pattern []interface{}) bool {
	// 路径长度必须与模式长度完全匹配（精确匹配）
	if len(pattern) != len(path) {
		return false
	}

	// 逐个比较路径元素
	for i := 0; i < len(pattern); i++ {
		patternPart := pattern[i]
		pathPart := path[i]

		// 处理通配符
		if patternPart == "*" {
			continue
		}

		// 处理数组索引
		if patternInt, ok := patternPart.(int); ok {
			if pathInt, ok := pathPart.(int); ok {
				if patternInt != pathInt {
					return false
				}
				continue
			} else {
				return false
			}
		}

		// 处理属性名
		if patternPart != pathPart {
			return false
		}
	}

	return true
}

// ParserState JSON 解析器的状态类型
type ParserState int

const (
	VALUE ParserState = iota
	KEY_OR_END
	KEY
	COLON
	COMMA
	VALUE_OR_END
	NUMBER
	TRUE1
	TRUE2
	TRUE3
	FALSE1
	FALSE2
	FALSE3
	FALSE4
	NULL1
	NULL2
	NULL3
)

// StreamingJsonParser 真实的流式 JSON 解析器
type StreamingJsonParser struct {
	matcher      *SimplePathMatcher
	realtime     bool
	incremental  bool // 新增：控制是返回增量内容还是累积内容
	stack        []interface{}
	path         []interface{}
	state        ParserState
	buffer       string
	isEscaped    bool
	isInString   bool
	currentKey   *string
	arrayIndexes []int
	lastSentPos  map[string]int // 新增：记录每个路径上次发送的位置
}

// NewStreamingJsonParser 创建新的流式JSON解析器
// realtime: 控制是否实时返回解析结果
// incremental: 控制是返回增量内容(true)还是累积内容(false)
func NewStreamingJsonParser(matcher *SimplePathMatcher, realtime bool, incremental bool) *StreamingJsonParser {
	parser := &StreamingJsonParser{
		matcher:     matcher,
		realtime:    realtime,
		incremental: incremental,
	}
	parser.Reset()
	return parser
}

// Reset 重置解析器状态
func (p *StreamingJsonParser) Reset() {
	p.stack = make([]interface{}, 0)
	p.path = make([]interface{}, 0)
	p.state = VALUE
	p.buffer = ""
	p.isEscaped = false
	p.isInString = false
	p.currentKey = nil
	p.arrayIndexes = make([]int, 0)
	p.lastSentPos = make(map[string]int)
}

// Write 逐字符处理输入流
func (p *StreamingJsonParser) Write(chunk string) error {
	for _, char := range chunk {
		if err := p.processChar(char); err != nil {
			return err
		}
	}
	return nil
}

// processChar 处理单个字符
func (p *StreamingJsonParser) processChar(char rune) error {
	// 处理字符串中的转义
	if p.isInString {
		if p.isEscaped {
			// 处理转义字符
			switch char {
			case 'n':
				p.buffer += "\n"
			case 't':
				p.buffer += "\t"
			case 'r':
				p.buffer += "\r"
			case '\\':
				p.buffer += "\\"
			case '"':
				p.buffer += "\""
			case '/':
				p.buffer += "/"
			case 'b':
				p.buffer += "\b"
			case 'f':
				p.buffer += "\f"
			default:
				// 对于不认识的转义字符，保持原样
				p.buffer += "\\" + string(char)
			}
			p.isEscaped = false
			return nil
		}

		if char == '\\' {
			p.isEscaped = true
			return nil
		}

		if char == '"' {
			p.isInString = false

			if p.state == KEY {
				// 复制buffer的值而不是引用
				keyValue := p.buffer
				p.currentKey = &keyValue
				p.buffer = ""
				p.state = COLON
			} else if p.state == VALUE {
				// 字符串值完成时，检查是否已经发送过增量内容
				hasIncremental := false
				if p.realtime && p.incremental {
					pathKey := p.getPathKey()
					_, hasIncremental = p.lastSentPos[pathKey]
					delete(p.lastSentPos, pathKey)
				}
				// 只有在非实时增量模式或者没有发送过增量内容时才调用addValue
				if !(p.realtime && p.incremental && hasIncremental) {
					p.addValue(p.buffer)
				}
				p.buffer = ""
				p.state = COMMA
			}

			return nil
		}

		p.buffer += string(char)
		// 实时触发回调
		if p.realtime && p.state == VALUE && p.buffer != "" {
			if p.incremental {
				// 增量模式：只发送新增的字符
				pathKey := p.getPathKey()
				lastPos := p.lastSentPos[pathKey]
				if len(p.buffer) > lastPos {
					incrementalContent := p.buffer[lastPos:]
					p.matcher.CheckPatterns(p.path, incrementalContent)
					p.lastSentPos[pathKey] = len(p.buffer)
				}
			} else {
				// 累积模式：发送完整内容
				p.matcher.CheckPatterns(p.path, p.buffer)
			}
		}
		return nil
	}

	// 处理非字符串状态
	switch p.state {
	case VALUE:
		return p.handleValueState(char)
	case KEY_OR_END:
		return p.handleKeyOrEndState(char)
	case KEY:
		return p.handleKeyState(char)
	case COLON:
		return p.handleColonState(char)
	case COMMA:
		return p.handleCommaState(char)
	case VALUE_OR_END:
		return p.handleValueOrEndState(char)
	case NUMBER:
		return p.handleNumberState(char)
	case TRUE1, TRUE2, TRUE3:
		return p.handleTrueState(char)
	case FALSE1, FALSE2, FALSE3, FALSE4:
		return p.handleFalseState(char)
	case NULL1, NULL2, NULL3:
		return p.handleNullState(char)
	}

	return nil
}

// handleValueState 处理VALUE状态
func (p *StreamingJsonParser) handleValueState(char rune) error {
	switch char {
	case '{':
		// 开始对象
		obj := make(map[string]interface{})
		p.addValue(obj)
		p.stack = append(p.stack, obj)
		p.state = KEY_OR_END
	case '[':
		// 开始数组
		arr := make([]interface{}, 0)
		p.addValue(&arr)
		p.stack = append(p.stack, &arr)
		p.arrayIndexes = append(p.arrayIndexes, 0)
		p.path = append(p.path, 0)
		p.state = VALUE_OR_END
	case '"':
		// 开始字符串
		p.isInString = true
		p.buffer = ""
	case 't':
		// 可能是 true
		p.buffer = "t"
		p.state = TRUE1
	case 'f':
		// 可能是 false
		p.buffer = "f"
		p.state = FALSE1
	case 'n':
		// 可能是 null
		p.buffer = "n"
		p.state = NULL1
	case '-':
		fallthrough
	default:
		if char >= '0' && char <= '9' || char == '-' {
			// 开始数字
			p.buffer = string(char)
			p.state = NUMBER
		} else if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
			return fmt.Errorf("unexpected character in VALUE state: %c", char)
		}
	}
	return nil
}

// handleKeyOrEndState 处理KEY_OR_END状态
func (p *StreamingJsonParser) handleKeyOrEndState(char rune) error {
	switch char {
	case '}':
		// 结束对象
		p.endObject()
		p.state = COMMA
	case '"':
		// 开始键名
		p.isInString = true
		p.buffer = ""
		p.state = KEY
	default:
		if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
			return fmt.Errorf("unexpected character in KEY_OR_END state: %c", char)
		}
	}
	return nil
}

// handleKeyState 处理KEY状态
func (p *StreamingJsonParser) handleKeyState(char rune) error {
	if char == '"' {
		// 开始字符串
		p.isInString = true
		p.buffer = ""
	} else if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
		return fmt.Errorf("unexpected character in KEY state: %c", char)
	}
	return nil
}

// handleColonState 处理COLON状态
func (p *StreamingJsonParser) handleColonState(char rune) error {
	if char == ':' {
		p.state = VALUE
		// 更新路径 - 添加当前键到路径
		if p.currentKey != nil {
			p.path = append(p.path, *p.currentKey)
		}
	} else if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
		return fmt.Errorf("unexpected character in COLON state: %c", char)
	}
	return nil
}

// handleCommaState 处理COMMA状态
func (p *StreamingJsonParser) handleCommaState(char rune) error {
	switch char {
	case ',':
		if len(p.stack) > 0 {
			if _, isArray := p.stack[len(p.stack)-1].(*[]interface{}); isArray {
				// 数组中的下一个元素
				if len(p.arrayIndexes) > 0 {
					p.arrayIndexes[len(p.arrayIndexes)-1]++
					p.path[len(p.path)-1] = p.arrayIndexes[len(p.arrayIndexes)-1]
				}
				p.state = VALUE
			} else {
				// 对象中的下一个键 - 移除当前键
				if len(p.path) > 0 {
					p.path = p.path[:len(p.path)-1]
				}
				p.state = KEY
			}
		}
	case '}':
		// 结束对象
		p.endObject()
	case ']':
		// 结束数组
		p.endArray()
	default:
		if char != ' ' && char != '\t' && char != '\n' && char != '\r' {
			return fmt.Errorf("unexpected character in COMMA state: %c", char)
		}
	}
	return nil
}

// handleValueOrEndState 处理VALUE_OR_END状态
func (p *StreamingJsonParser) handleValueOrEndState(char rune) error {
	if char == ']' {
		// 空数组
		p.endArray()
		p.state = COMMA
	} else {
		// 回到 VALUE 状态处理这个字符
		p.state = VALUE
		return p.processChar(char)
	}
	return nil
}

// handleNumberState 处理NUMBER状态
func (p *StreamingJsonParser) handleNumberState(char rune) error {
	if (char >= '0' && char <= '9') || char == '.' || char == 'e' || char == 'E' || char == '+' || char == '-' {
		p.buffer += string(char)
	} else {
		// 数字结束
		if num, err := strconv.ParseFloat(p.buffer, 64); err == nil {
			p.addValue(num)
		} else {
			return fmt.Errorf("invalid number: %s", p.buffer)
		}
		p.buffer = ""
		p.state = COMMA
		// 重新处理当前字符
		return p.processChar(char)
	}
	return nil
}

// handleTrueState 处理TRUE状态
func (p *StreamingJsonParser) handleTrueState(char rune) error {
	switch p.state {
	case TRUE1:
		if char == 'r' {
			p.buffer += string(char)
			p.state = TRUE2
		} else {
			return fmt.Errorf("unexpected character in TRUE1 state: %c", char)
		}
	case TRUE2:
		if char == 'u' {
			p.buffer += string(char)
			p.state = TRUE3
		} else {
			return fmt.Errorf("unexpected character in TRUE2 state: %c", char)
		}
	case TRUE3:
		if char == 'e' {
			p.addValue(true)
			p.buffer = ""
			p.state = COMMA
		} else {
			return fmt.Errorf("unexpected character in TRUE3 state: %c", char)
		}
	}
	return nil
}

// handleFalseState 处理FALSE状态
func (p *StreamingJsonParser) handleFalseState(char rune) error {
	switch p.state {
	case FALSE1:
		if char == 'a' {
			p.buffer += string(char)
			p.state = FALSE2
		} else {
			return fmt.Errorf("unexpected character in FALSE1 state: %c", char)
		}
	case FALSE2:
		if char == 'l' {
			p.buffer += string(char)
			p.state = FALSE3
		} else {
			return fmt.Errorf("unexpected character in FALSE2 state: %c", char)
		}
	case FALSE3:
		if char == 's' {
			p.buffer += string(char)
			p.state = FALSE4
		} else {
			return fmt.Errorf("unexpected character in FALSE3 state: %c", char)
		}
	case FALSE4:
		if char == 'e' {
			p.addValue(false)
			p.buffer = ""
			p.state = COMMA
		} else {
			return fmt.Errorf("unexpected character in FALSE4 state: %c", char)
		}
	}
	return nil
}

// handleNullState 处理NULL状态
func (p *StreamingJsonParser) handleNullState(char rune) error {
	switch p.state {
	case NULL1:
		if char == 'u' {
			p.buffer += string(char)
			p.state = NULL2
		} else {
			return fmt.Errorf("unexpected character in NULL1 state: %c", char)
		}
	case NULL2:
		if char == 'l' {
			p.buffer += string(char)
			p.state = NULL3
		} else {
			return fmt.Errorf("unexpected character in NULL2 state: %c", char)
		}
	case NULL3:
		if char == 'l' {
			p.addValue(nil)
			p.buffer = ""
			p.state = COMMA
		} else {
			return fmt.Errorf("unexpected character in NULL3 state: %c", char)
		}
	}
	return nil
}

// addValue 添加值到当前容器
func (p *StreamingJsonParser) addValue(value interface{}) {
	if len(p.stack) == 0 {
		// 根值
		p.stack = append(p.stack, value)
		// 在增量模式下，如果是实时模式且已经发送过增量内容，则不再发送完整值
		if !(p.realtime && p.incremental && p.hasIncrementalContent(value)) {
			p.matcher.CheckPatterns(p.path, value)
		}
		return
	}

	parent := p.stack[len(p.stack)-1]

	if arrPtr, isArray := parent.(*[]interface{}); isArray {
		// 添加到数组
		if len(p.arrayIndexes) > 0 {
			index := p.arrayIndexes[len(p.arrayIndexes)-1]
			// 扩展数组以容纳新元素
			for len(*arrPtr) <= index {
				*arrPtr = append(*arrPtr, nil)
			}
			(*arrPtr)[index] = value
		}
	} else if obj, isObject := parent.(map[string]interface{}); isObject {
		// 添加到对象
		if p.currentKey != nil {
			obj[*p.currentKey] = value
		}
	}

	// 在增量模式下，如果是实时模式且已经发送过增量内容，则不再发送完整值
	if !(p.realtime && p.incremental && p.hasIncrementalContent(value)) {
		p.matcher.CheckPatterns(p.path, value)
	}
}

// endObject 结束对象处理
func (p *StreamingJsonParser) endObject() {
	if len(p.stack) > 0 {
		p.stack = p.stack[:len(p.stack)-1]
	}
	// 只有当当前路径的最后一个元素不是数组索引时，才移除路径元素
	// 这样可以保持数组索引在路径中的正确位置
	if len(p.path) > 1 {
		// 检查最后一个路径元素是否为数组索引（整数类型）
		lastElement := p.path[len(p.path)-1]
		if _, isInt := lastElement.(int); !isInt {
			// 如果不是数组索引，则移除路径元素
			p.path = p.path[:len(p.path)-1]
		}
	}
	p.state = COMMA
}

// endArray 结束数组处理
func (p *StreamingJsonParser) endArray() {
	if len(p.stack) > 0 {
		p.stack = p.stack[:len(p.stack)-1]
	}
	if len(p.arrayIndexes) > 0 {
		p.arrayIndexes = p.arrayIndexes[:len(p.arrayIndexes)-1]
	}
	if len(p.path) > 1 {
		p.path = p.path[:len(p.path)-1]
	}
	p.state = COMMA
}

// End 结束解析
func (p *StreamingJsonParser) End() error {
	if len(p.stack) != 1 {
		return errors.New("unexpected end of input: JSON structure is incomplete")
	}
	fmt.Printf("JSON parsing complete: %+v\n", p.stack[0])
	return nil
}

// GetResult 获取解析结果
func (p *StreamingJsonParser) GetResult() interface{} {
	if len(p.stack) > 0 {
		return p.stack[0]
	}
	return nil
}

// getPathKey 生成路径的唯一标识符
func (p *StreamingJsonParser) getPathKey() string {
	var pathStr strings.Builder
	for i, segment := range p.path {
		if i > 0 {
			pathStr.WriteString(".")
		}
		pathStr.WriteString(fmt.Sprintf("%v", segment))
	}
	return pathStr.String()
}

// hasIncrementalContent 检查是否已经发送过增量内容
func (p *StreamingJsonParser) hasIncrementalContent(value interface{}) bool {
	// 只有在实时增量模式下才进行检查
	if !p.realtime || !p.incremental {
		return false
	}
	
	// 只对字符串类型进行增量处理，检查是否已经发送过增量内容
	if _, isString := value.(string); isString {
		pathKey := p.getPathKey()
		_, exists := p.lastSentPos[pathKey]
		return exists
	}
	// 对于其他类型（数字、布尔值、null、对象、数组），不进行增量处理
	return false
}
