package ast

import (
	"fmt"
	"strconv"
)

type parameter struct {
	key   string
	value CompositeValue
}

type statementBuilder struct {
	action                string
	entity                string
	declarationIdentifier string
	isValue               bool
	params                []*parameter
	newparams             map[string]interface{}
	currentKey            string
	currentValue          CompositeValue
	listBuilder           *listValueBuilder
	concatenationBuilder  *concatenationValueBuilder
}

func (b *statementBuilder) build() *Statement {
	if b.action == "" && b.entity == "" && b.declarationIdentifier == "" && !b.isValue {
		return nil
	}
	var expr ExpressionNode
	if b.isValue {
		expr = &ValueNode{Value: b.currentValue}
	} else {
		cmdParams := make(map[string]CompositeValue)
		for _, param := range b.params {
			cmdParams[param.key] = param.value
		}
		cmdNode := CmdNode{
			action:     b.action,
			entity:     b.entity,
			ParamNodes: b.newparams,
		}
		fmt.Println(cmdNode)
		expr = &CommandNode{Action: b.action, Entity: b.entity, Params: cmdParams}
	}
	if b.declarationIdentifier != "" {
		decl := &DeclarationNode{Ident: b.declarationIdentifier, Expr: expr}
		return &Statement{Node: decl}
	}
	return &Statement{Node: expr}
}

func (b *statementBuilder) addParamKey(key string) *statementBuilder {
	b.currentKey = key
	return b
}

func (b *statementBuilder) addRef(s string) *statementBuilder {
	if b.newparams == nil {
		b.newparams = make(map[string]interface{})
	}
	b.newparams[b.currentKey] = RefNode{key: s}
	return b
}

func (b *statementBuilder) addAlias(s string) *statementBuilder {
	if b.newparams == nil {
		b.newparams = make(map[string]interface{})
	}
	b.newparams[b.currentKey] = AliasNode{key: s}
	return b
}

func (b *statementBuilder) addHole(s string) *statementBuilder {
	if b.newparams == nil {
		b.newparams = make(map[string]interface{})
	}
	b.newparams[b.currentKey] = HoleNode{key: s}
	return b
}

func (b *statementBuilder) addParamValue(val CompositeValue) *statementBuilder {
	b.currentValue = val
	if b.concatenationBuilder != nil {
		b.concatenationBuilder.add(b.currentValue)
		b.currentValue = nil
	} else if b.listBuilder != nil {
		b.listBuilder.add(b.currentValue)
		b.currentValue = nil
	} else {
		if b.currentKey != "" {
			b.params = append(b.params, &parameter{key: b.currentKey, value: b.currentValue})
			b.currentKey = ""
			b.currentValue = nil
		}
	}

	return b
}

func (b *statementBuilder) newList() *statementBuilder {
	b.listBuilder = &listValueBuilder{}
	return b
}

func (b *statementBuilder) buildList() *statementBuilder {
	if b.listBuilder != nil {
		list := b.listBuilder.build()
		b.listBuilder = nil
		b.addParamValue(list)
	}
	return b
}

func (a *AST) addAction(text string) {
	if IsInvalidAction(text) {
		panic(fmt.Errorf("unknown action '%s'", text))
	}
	a.stmtBuilder.action = text
}

func (a *AST) addEntity(text string) {
	if IsInvalidEntity(text) {
		panic(fmt.Errorf("unknown entity '%s'", text))
	}
	a.stmtBuilder.entity = text
}

func (a *AST) addValue() {
	a.stmtBuilder.isValue = true
}

func (a *AST) addDeclarationIdentifier(text string) {
	a.stmtBuilder.declarationIdentifier = text
}

func (a *AST) NewStatement() {
	a.stmtBuilder = &statementBuilder{}
}

func (a *AST) StatementDone() {

	if stmt := a.stmtBuilder.build(); stmt != nil {
		a.Statements = append(a.Statements, stmt)
	}
	a.stmtBuilder = nil
}

func (a *AST) addParamKey(text string) {
	a.stmtBuilder.addParamKey(text)
}

func (a *AST) addParamValue(text string) {
	var val interface{}
	i, err := strconv.Atoi(text)
	if err == nil {
		val = i
	} else {
		f, err := strconv.ParseFloat(text, 64)
		if err == nil {
			val = f
		} else {
			val = text
		}
	}
	a.stmtBuilder.addParamValue(&interfaceValue{val: val})
}

func (a *AST) addFirstValueInList() {
	a.stmtBuilder.newList()
}
func (a *AST) lastValueInList() {
	a.stmtBuilder.buildList()
}

func (a *AST) addFirstValueInConcatenation() {
	a.stmtBuilder.concatenationBuilder = &concatenationValueBuilder{}
}

func (a *AST) lastValueInConcatenation() {
	if a.stmtBuilder.concatenationBuilder != nil {
		concat := a.stmtBuilder.concatenationBuilder.build()
		a.stmtBuilder.concatenationBuilder = nil
		a.stmtBuilder.addParamValue(concat)
	}
}

func (a *AST) addStringValue(text string) {
	a.stmtBuilder.addParamValue(&interfaceValue{val: text})
}

func (a *AST) addParamRefValue(text string) {
	a.stmtBuilder.addRef(text)
	a.stmtBuilder.addParamValue(&referenceValue{ref: text})
}

func (a *AST) addParamHoleValue(text string) {
	a.stmtBuilder.addHole(text)
	a.stmtBuilder.addParamValue(NewHoleValue(text))
}

func (a *AST) addAliasParam(text string) {
	a.stmtBuilder.addAlias(text)
	a.stmtBuilder.addParamValue(&aliasValue{alias: text})
}

type listValueBuilder struct {
	vals []CompositeValue
}

func (c *listValueBuilder) add(v CompositeValue) *listValueBuilder {
	c.vals = append(c.vals, v)
	return c
}

func (c *listValueBuilder) build() CompositeValue {
	return &listValue{c.vals}
}

type concatenationValueBuilder struct {
	vals []CompositeValue
}

func (c *concatenationValueBuilder) add(v CompositeValue) *concatenationValueBuilder {
	c.vals = append(c.vals, v)
	return c
}

func (c *concatenationValueBuilder) build() CompositeValue {
	return &concatenationValue{c.vals}
}
