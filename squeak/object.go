package squeak

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"github.com/ernilsson/pia/squeak/ast"
	"github.com/ernilsson/pia/squeak/token"
	"net/http"
	"slices"
	"strings"
)

// Object is a broad interface for any data that a Squeak script can process. It does not provide any interface beyond
// the standard library fmt.Stringer, but passing around types of this interface around the interpreter obfuscates the
// meaning behind the value. Hence, this interface is largely just for clearer naming.
type Object interface {
	fmt.Stringer
	Clone() Object
}

type Instance interface {
	Object
	Get(string) Object
	Put(string, Object) Object
}

func NewRequestObject(req *http.Request) *ObjectInstance {
	obj := &ObjectInstance{Properties: make(map[string]Object)}
	obj.Properties["method"] = String{req.Method}
	obj.Properties["url"] = String{req.URL.String()}
	headers := &ObjectInstance{Properties: make(map[string]Object)}
	for k, v := range req.Header {
		headers.Put(k, String{strings.Join(v, ", ")})
	}
	obj.Properties["headers"] = headers
	return obj
}

type BoundBuiltinMethod struct {
	this Object
	impl BuiltinMethod
}

func (b BoundBuiltinMethod) String() string {
	return "builtin:method"
}

func (b BoundBuiltinMethod) Clone() Object {
	return BoundBuiltinMethod{
		this: b.this,
		impl: b.impl,
	}
}

func (b BoundBuiltinMethod) Arity() int {
	return b.impl.arity
}

func (b BoundBuiltinMethod) Call(in *Interpreter, args ...Object) (Object, error) {
	return b.impl.fn(b.this, in, args...)
}

type BuiltinMethod struct {
	arity int
	fn    func(Object, *Interpreter, ...Object) (Object, error)
}

func (b BuiltinMethod) String() string {
	return "builtin:method"
}

func (b BuiltinMethod) Clone() Object {
	return BuiltinMethod{
		arity: b.arity,
		fn:    b.fn,
	}
}

func (b BuiltinMethod) Bind(obj Object) (Callable, error) {
	return BoundBuiltinMethod{
		this: obj,
		impl: b,
	}, nil
}

func NewResponseObject(res *http.Response, body []byte) *ObjectInstance {
	obj := &ObjectInstance{Properties: make(map[string]Object)}
	obj.Properties["status_code"] = Number{float64(res.StatusCode)}
	obj.Properties["status"] = String{res.Status}
	headers := &ObjectInstance{Properties: make(map[string]Object)}
	for k, v := range res.Header {
		headers.Put(k, String{strings.Join(v, ", ")})
	}
	obj.Properties["headers"] = headers

	obj.Properties["json"] = BuiltinMethod{
		arity: 0,
		fn: func(_ Object, _ *Interpreter, _ ...Object) (Object, error) {
			if len(body) == 0 {
				return nil, nil
			}
			var builder Builder
			if err := json.Unmarshal(body, &builder); err != nil {
				return nil, err
			}
			return builder.Object(), nil
		},
	}
	obj.Properties["xml"] = BuiltinMethod{
		arity: 0,
		fn: func(_ Object, _ *Interpreter, _ ...Object) (Object, error) {
			if len(body) == 0 {
				return nil, nil
			}
			var builder Builder
			if err := xml.Unmarshal(body, &builder); err != nil {
				return nil, err
			}
			return builder.Object(), nil
		},
	}
	return obj
}

type Builder struct {
	obj Object
}

func (b *Builder) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	stack := struct {
		slice  []*ObjectInstance
		cursor int
	}{
		slice:  make([]*ObjectInstance, 1),
		cursor: 0,
	}
	push := func(o *ObjectInstance) {
		stack.cursor += 1
		if stack.cursor == len(stack.slice) {
			stack.slice = append(stack.slice, o)
		} else {
			stack.slice[stack.cursor] = o
		}
	}
	pop := func() *ObjectInstance {
		o := stack.slice[stack.cursor]
		stack.cursor -= 1
		return o
	}
	peek := func() *ObjectInstance {
		return stack.slice[stack.cursor]
	}
	element := func(el xml.StartElement) *ObjectInstance {
		attrs := &ObjectInstance{Properties: make(map[string]Object)}
		for _, attr := range el.Attr {
			attrs.Put(attr.Name.Local, String{attr.Value})
		}
		o := &ObjectInstance{Properties: make(map[string]Object)}
		o.Put("_attributes", attrs)
		return o
	}
	stack.slice[0] = element(start)

	for {
		t, _ := d.Token()
		switch t := t.(type) {
		case xml.StartElement:
			el := element(t)
			parent := peek()
			if parent.Get(t.Name.Local) != nil {
				switch container := parent.Get(t.Name.Local).(type) {
				case *List:
					container.slice = append(container.slice, el)
				default:
					parent.Put(t.Name.Local, &List{
						slice: []Object{container, el},
					})
				}
			} else {
				peek().Put(t.Name.Local, el)
			}
			push(el)
		case xml.CharData:
			if strings.TrimSpace(string(t)) == "" {
				continue
			}
			peek().Put("_inner", String{string(t)})
		case xml.EndElement:
			if t.Name == start.Name {
				b.obj = pop()
				return nil
			} else {
				pop()
			}
		}
	}
}

func (b *Builder) UnmarshalJSON(data []byte) error {
	var raw any
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	b.obj, err = b.asObject(raw)
	return err
}

func (b *Builder) Object() Object {
	return b.obj
}

func (b *Builder) asObject(raw any) (Object, error) {
	switch v := raw.(type) {
	case string:
		return String{v}, nil
	case int:
		return Number{float64(v)}, nil
	case float64:
		return Number{v}, nil
	case map[string]any:
		props := make(map[string]Object)
		for k, v := range v {
			prop, err := b.asObject(v)
			if err != nil {
				return nil, err
			}
			props[k] = prop
		}
		return &ObjectInstance{props}, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to object", v)
	}
}

// ObjectInstance is an asObject instance, which consists of a collection of named data as well as behaviours coupled to the
// data.
type ObjectInstance struct {
	Properties map[string]Object
}

func (i *ObjectInstance) String() string {
	sb := strings.Builder{}
	sb.WriteString("Object {")
	for k, v := range i.Properties {
		sb.WriteString(fmt.Sprintf("%s: %s", k, v.String()))
	}
	sb.WriteString("}")
	return sb.String()
}

func (i *ObjectInstance) Clone() Object {
	props := make(map[string]Object)
	for k, v := range i.Properties {
		props[k] = v.Clone()
	}
	return &ObjectInstance{Properties: props}
}

func (i *ObjectInstance) Get(s string) Object {
	return i.Properties[s]
}

func (i *ObjectInstance) Put(s string, object Object) Object {
	i.Properties[s] = object
	return object
}

type Callable interface {
	Object
	Arity() int
	Call(*Interpreter, ...Object) (Object, error)
}

// Function is the callable equivalent of [ast.Function].
type Function struct {
	declaration ast.Function
	closure     *Environment
}

func (f Function) String() string {
	return fmt.Sprintf("function:%s", f.declaration.Name.Lexeme)
}

func (f Function) Clone() Object {
	return Function{
		declaration: f.declaration,
		closure:     f.closure,
	}
}

func (f Function) Arity() int {
	return len(f.declaration.Params)
}

func (f Function) Call(in *Interpreter, args ...Object) (Object, error) {
	scope := NewEnvironment(Parent(f.closure))
	for i, param := range f.declaration.Params {
		scope.Declare(param.Lexeme, args[i])
	}
	uw, err := in.block(scope, f.declaration.Body.Body)
	if err != nil {
		return nil, err
	}
	if uw == nil {
		return nil, nil
	}
	if uw.source.Type != token.Return {
		return nil, fmt.Errorf("%w: unexpected unwinding source %s", ErrRuntimeFault, uw.source.Lexeme)
	}
	return uw.value, nil
}

type Method interface {
	Bind(Object) (Callable, error)
}

type BoundObjectInstanceMethod struct {
	ObjectInstanceMethod
	this *ObjectInstance
}

func (b BoundObjectInstanceMethod) Call(in *Interpreter, args ...Object) (Object, error) {
	closure := NewEnvironment(Parent(in.global), Prefill("this", b.this))
	scope := NewEnvironment(Parent(closure))
	for i, param := range b.declaration.Params {
		scope.Declare(param.Lexeme, args[i])
	}
	uw, err := in.block(scope, b.declaration.Body.Body)
	if err != nil {
		return nil, err
	}
	if uw == nil {
		return nil, nil
	}
	if uw.source.Type != token.Return {
		return nil, fmt.Errorf("%w: unexpected unwinding source %s", ErrRuntimeFault, uw.source.Lexeme)
	}
	return uw.value, nil
}

func (b BoundObjectInstanceMethod) Arity() int {
	return len(b.declaration.Params)
}

type ObjectInstanceMethod struct {
	declaration ast.Method
}

func (m ObjectInstanceMethod) String() string {
	return "method"
}

func (m ObjectInstanceMethod) Clone() Object {
	return &ObjectInstanceMethod{declaration: m.declaration}
}

func (m ObjectInstanceMethod) Bind(obj Object) (Callable, error) {
	i, ok := obj.(*ObjectInstance)
	if !ok {
		return nil, fmt.Errorf("%w: %T cannot be binding target for object method", ErrIllegalArgument, obj)
	}
	return BoundObjectInstanceMethod{
		ObjectInstanceMethod: m,
		this:                 i,
	}, nil
}

// Number is an Object representing a numerical value internally represented as a float64. In Squeak, the notion of
// integers only exists in the lexical and parsing phase. During evaluation, all numerical objects are represented with
// this struct.
type Number struct {
	value float64
}

func (i Number) String() string {
	return strings.TrimRight(fmt.Sprintf("%f", i.value), "0")
}

func (i Number) Clone() Object {
	return Number{
		value: i.value,
	}
}

// String is an Object representing a textual value.
type String struct {
	value string
}

func (s String) String() string {
	return s.value
}

func (s String) Clone() Object {
	return String{value: s.value}
}

// Boolean is an Object representing a boolean value.
type Boolean struct {
	value bool
}

func (b Boolean) String() string {
	if b.value {
		return "true"
	}
	return "false"
}

func (b Boolean) Clone() Object {
	return Boolean{value: b.value}
}

type BoundListMethod struct {
	ListMethod
	this *List
}

func (b BoundListMethod) String() string {
	return "builtin:list:method"
}

func (b BoundListMethod) Clone() Object {
	return BoundListMethod{
		ListMethod: b.ListMethod,
		this:       b.this,
	}
}

func (b BoundListMethod) Arity() int {
	return b.arity
}

func (b BoundListMethod) Call(in *Interpreter, args ...Object) (Object, error) {
	return b.ListMethod.fn(b.this, in, args...)
}

type ListMethod struct {
	arity int
	fn    func(*List, *Interpreter, ...Object) (Object, error)
}

func (l ListMethod) String() string {
	return "builtin:list:method"
}

func (l ListMethod) Clone() Object {
	return ListMethod{
		arity: l.arity,
		fn:    l.fn,
	}
}

func (l ListMethod) Bind(obj Object) (Callable, error) {
	list, ok := obj.(*List)
	if !ok {
		return nil, fmt.Errorf("%w: %T cannot be binding target for list method", ErrIllegalOperation, obj)
	}
	return BoundListMethod{
		ListMethod: l,
		this:       list,
	}, nil
}

// List is a single Object containing a collection of Object values.
type List struct {
	slice []Object
}

func (l *List) String() string {
	items := make([]string, len(l.slice))
	for i := range l.slice {
		items[i] = l.slice[i].String()
	}
	return fmt.Sprintf("[%s]", strings.Join(items, ","))
}

func (l *List) Clone() Object {
	clone := make([]Object, len(l.slice))
	for i, v := range l.slice {
		clone[i] = v.Clone()
	}
	return &List{slice: clone}
}

func (l *List) Get(s string) Object {
	switch s {
	case "add":
		return ListMethod{
			arity: 1,
			fn: func(l *List, _ *Interpreter, args ...Object) (Object, error) {
				l.slice = append(l.slice, args[0])
				return l, nil
			},
		}
	case "length":
		return ListMethod{
			arity: 0,
			fn: func(l *List, _ *Interpreter, _ ...Object) (Object, error) {
				return Number{float64(len(l.slice))}, nil
			},
		}
	case "find":
		return ListMethod{
			arity: 1,
			fn: func(l *List, in *Interpreter, args ...Object) (Object, error) {
				for i, v := range l.slice {
					eq, err := in.isEqual(args[0], v)
					if err != nil {
						return nil, err
					}
					if eq.value {
						return Number{float64(i)}, nil
					}
				}
				return Number{-1}, nil
			},
		}
	case "contains":
		return ListMethod{
			arity: 1,
			fn: func(l *List, in *Interpreter, args ...Object) (Object, error) {
				for _, v := range l.slice {
					eq, err := in.isEqual(args[0], v)
					if err != nil {
						return nil, err
					}
					if eq.value {
						return Boolean{true}, nil
					}
				}
				return Boolean{false}, nil
			},
		}
	case "remove":
		return ListMethod{
			arity: 1,
			fn: func(l *List, _ *Interpreter, args ...Object) (Object, error) {
				idx, ok := args[0].(Number)
				if !ok {
					return nil, fmt.Errorf("%w: index must be a number", ErrIllegalArgument)
				}
				i := int(idx.value)
				if i >= len(l.slice) {
					return nil, fmt.Errorf("%w: index out of range", ErrIllegalArgument)
				}
				l.slice = slices.Delete(l.slice, i, i+1)
				return l, nil
			},
		}
	default:
		return nil
	}
}

func (l *List) Put(string, Object) Object {
	panic(fmt.Errorf("%w: cannot mutate prototype of list data structure", ErrIllegalOperation))
}
