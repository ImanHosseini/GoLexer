package main

import (
	"fmt"
	"text/scanner"
	"unicode/utf8"
	"strings"
	"unicode"
	"sort"
	"bufio"
	"os"
	"io"
)

type itemType int
const(
	itemError itemType=iota
	itemAuto
	itemBool
	itemBreak
	itemCase
	itemChar
	itemClass
	itemConst
	itemContinue
	itemDefault
	itemDestruct
	itemDouble
	itemElse
	itemFalse
	itemFloat
	itemFor
	itemForeach
	itemGoto
	itemIf
	itemIn
	itemInherit
	itemInt
	itemLong
	itemNew
	itemPrivate
	itemProcedure
	itemPublic
	itemRepeat
	itemReturn
	itemSizeof
	itemStatic
	itemString
	itemSwitch
	itemTrue
	itemUntil
	itemVoid
	itemEqual
	itemNotequal
	itemLessorequal
	itemLessthan
	itemBiggerthan
	itemAssignment
	itemNot
	itemArithmeticand
	itemLogicaland
	itemArithmeticor
	itemLogicalor
	itemLogicalarithmeticxor
	itemProduction
	itemAdd
	itemIncrement
	itemDecrement
	itemSubandunaryminus
	itemDiv
	itemMod
	itemOpeningcurlybrace
	itemClosingcurlybrace
	itemOpeningparanthesis
	itemClosingparanthesis
	itemOpeningbrace
	itemClosingbrace
	itemDot
	itemComma
	itemColon
	itemSemicolon
	itemStr
	itemCtr
	itemMComment
	itemOp
	itemKeyword
	itemIdentifier
	itemNumber
	itemText
	itemComment
	itemEOF
)

var keyw = []string {"auto", "bool","break","case","char","class","const","continue","default","destruct","double","else","false","float","for","foreach","goto","if","in","inherit","int","long","new","private","procedure","public","repeat","return","sizeof","static","string","switch","true","until","void"}
var letterz = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
var opz = map[string]string{
	"==": "equal",
	"!=":  "not equal",
	"<=": "less or equal",
	"<": "less than",
	">": "bigger than",
	">=" : "bigger or equal",
	"=" : "assignment",
	"!" : "not",
	"&" : "arithmetic and",
	"&&" : "logical and",
	"|" : "arithmetic or",
	"||" : "logical or",
	"^" : "logical/arithmetic xor",
	"*" : "production",
	"+" : "add",
	"++" : "increment",
	"--" : "decrement",
	"-" : "sub and unary minus",
	"/" : "div",
	"%" : "mod",
	"{" : "opening curly brace",
	"}" : "closing curly brace",
	"(" : "opening parenthesis",
	")" : "closing parenthesis",
	"[" : "opening brace",
	"]" : "closing brace",
	"." : "dot",
	"," : "comma",
	":" : "colon",
	";" : "semi_colon",
}

var testyOp = map[rune]string{
	'=' : "assignment",
	'+' : "add",
	'<': "less than",
	'>': "bigger than",
	'-' : "sub and unary minus",
	'!' : "not",
	'&' : "arithmetic and",
	'|' : "arithmetic or",
}

var singleOp = map[rune]string{
	'^' : "logical/arithmetic xor",
	'*' : "production",
	'/' : "div",
	'%' : "mod",
	'{' : "opening curly brace",
	'}' : "closing curly brace",
	'(' : "opening parenthesis",
	')' : "closing parenthesis",
	'[' : "opening brace",
	']' : "closing brace",
	'.' : "dot",
	',' : "comma",
	':' : "colon",
	';' : "semi_colon",
}


type stateFn func(*lexer) stateFn

type item struct{
	typ itemType
	val string
}

func (i item) String() string{
	// Details of printing
	switch i.typ{
	case itemEOF:
		return "EOF"
	case itemError:
		return i.val
	}
	//if len(i.val)>10{
	//	return fmt.Sprintf("%.10q ...",i.val)
	//}
	return fmt.Sprintf("%q",i.val)
}

type lexer struct{
	name string
	input string
	start int
	pos int
	lcount int
	nlc int
	width int
	state stateFn
	items chan item
}

func (l *lexer) errorf(format string,args ...interface{}) stateFn {
	itm := item{
		itemError,
		fmt.Sprintf(format, args...),
	}
	fmt.Println(itm)
	l.items <- itm
	return nil
}

func (l *lexer) next() (rune rune){
	if l.pos>=len(l.input){
		l.width=0
		return scanner.EOF
	}

	rune,l.width = utf8.DecodeRuneInString(l.input[l.pos:])

	l.pos+=l.width
	return rune
}

func (l *lexer) ignore(){
	l.start=l.pos
}

func (l *lexer) backup(){
	l.pos-=l.width
}

func (l *lexer) peek() rune{
	runne:=l.next()
	l.backup()
	return runne
}

func (l *lexer) emit (t itemType){

	itm := item{t,l.input[l.start:l.pos]}
	switch t{
	case itemStr:
		fmt.Printf("line %d: string\n",l.lcount)
		break
	case itemCtr:
		fmt.Printf("line %d: char\n",l.lcount)
		break
	case itemOp:
		//fmt.Println(itm.val)
		fmt.Printf("line %d: operator %q\n",l.lcount,opz[itm.val])
		break
	case itemComment:
		fmt.Printf("line %d: comment \"%s\"\n",l.lcount,itm.val)
		break
	case itemKeyword:
		fmt.Printf("line %d: keyword %q\n",l.lcount,itm.val)
		break
	case itemIdentifier:
		fmt.Printf("line %d: id %q\n",l.lcount,itm.val)
		break
	case itemNumber:
		//fmt.Println(itm.val)
		if strings.Contains(itm.val,"0x"){
			fmt.Printf("line %d: integer\n",l.lcount)
			break
		}
		if strings.Contains(itm.val,".") || strings.Contains(itm.val,"e") || strings.Contains(itm.val,"E"){
			fmt.Printf("line %d: real\n",l.lcount)
			break
		}else{
			fmt.Printf("line %d: integer\n",l.lcount)
			break
		}
	case itemMComment:
		if l.nlc!=l.lcount{
			fmt.Printf("line %d-%d: comment \"%s\"\n",l.lcount,l.nlc,itm.val[2:len(itm.val)-2])
		}else{
			//fmt.Printf("line %d: comment \"%s\"\n",l.lcount,itm.val[:len(itm.val)-2])
			fmt.Printf("line %d-%d: comment \"%s\"\n",l.lcount,l.nlc,itm.val[2:len(itm.val)-2])
		}
		break
	}


	l.items <- itm

	l.start=l.pos
}
func (l *lexer) accept (valid string) bool {
	if strings.IndexRune(valid,l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string){
	for strings.IndexRune(valid,l.next())>=0{

	}
	l.backup()
}

func (l * lexer) nextItem() item{
	for{
		select{
		case item := <-l.items:
			return item
		default:
			l.state=l.state(l)
		}
	}
	panic("reached here !")
}


func lex(name, input string) *lexer{
	l:=&lexer{
		name:name,
		input:input,
		lcount:1,
		nlc:1,
		state : lexNumber,
		items:make(chan item,10000),
	}
	return l
}

func lexText (l *lexer) stateFn{
	for{
		if strings.HasPrefix(l.input[l.pos:],"@@"){
			if l.pos>l.start{
				l.emit(itemText)
			}
			return lexSComment
		}
		if strings.HasPrefix(l.input[l.pos:],"/@"){
			if l.pos>l.start{
				l.emit(itemText)
			}
			return lexMComment
		}
		if strings.HasPrefix(l.input[l.pos:],"\""){
			//fmt.Println("aaa")
			if l.pos>l.start{
				l.emit(itemText)
			}
			return lexStr
		}
		if strings.HasPrefix(l.input[l.pos:],"'"){
			if l.pos>l.start{
				l.emit(itemText)
			}
			return lexCtr
		}

		r:=l.next()
		switch {

		case r==scanner.EOF:
			return nil
		case r=='\n':
			l.ignore()
			l.nlc++
			l.lcount++
			// fmt.Println("HI")
			return lexText
		case unicode.IsSpace(r):
			l.ignore()
			return lexText

		case (r=='.' && ('0' <= l.peek() && l.peek()<='9'))  || (r=='-' && (('0' <= l.peek() && l.peek()<='9') || l.peek()=='.')) || '0' <= r && r<='9':
			l.backup()
			return lexNumber
		case isLetter(r) || r=='_':
			l.backup()
			return lexIdentifier
		case func()bool{ _, ok := singleOp[r]; return ok}():
			l.emit(itemOp)
			return lexText
		case func()bool{ _, ok := testyOp[r]; return ok}():
			l.next()
			str := l.input[l.start:l.pos]
			//fmt.Println(str)
			if func()bool{ _, ok := opz[str]; return ok}(){
				l.emit(itemOp)
				return lexText
			}else{
				l.backup()
				l.emit(itemOp)
				return lexText
			}
			return nil

		}
		if l.pos>l.start{
			l.emit(itemText)
			return lexText
		}
		l.emit(itemEOF)
		return nil
	}

}
func lexIdentifier (l *lexer) stateFn{
	// If we're here, it doesn't begin with numbers
	//l.acceptRun(letterz+"0123456789")
	for{
		l.acceptRun(letterz+"0123456789")
		if l.accept("_"){
			if l.accept("_"){
				if l.accept(letterz+"0123456789"){
					continue
				}

				l.pos=l.pos-2
				l.emit(itemIdentifier)
				return nil

			}else
			if l.accept(letterz+"0123456789"){
				continue
			}else{
				l.pos--
				tok := l.input[l.start:l.pos]
				i := sort.Search(len(keyw),
					func(i int) bool { return keyw[i] >= tok })
				if i < len(keyw) && keyw[i] == tok {
					// Keyword
					// fmt.Printf("found \"%s\" at files[%d]\n", keyw[i], i)
					l.emit(itemKeyword)
				} else{
					// Id
					l.emit(itemIdentifier)
				}
				return nil
			}

		}
		tok := l.input[l.start:l.pos]
		i := sort.Search(len(keyw),
			func(i int) bool { return keyw[i] >= tok })
		if i < len(keyw) && keyw[i] == tok {
			// Keyword
			// fmt.Printf("found \"%s\" at files[%d]\n", keyw[i], i)
			l.emit(itemKeyword)
		} else{
			// Id
			l.emit(itemIdentifier)

		}
		return lexText
	}


	return lexText
}

func lexSComment(l *lexer) stateFn {
	l.start += len("@@")
	l.pos += len("@@")
	for nextr := l.peek(); (nextr != '\n' && nextr != scanner.EOF); {
		l.next()
		nextr = l.peek()
	}
	// l.pos+=len("\n")
	//fmt.Println(l.input[l.start:l.pos])
	l.emit(itemComment)
	return lexText
}

func lexStr (l *lexer) stateFn{
	//l.start+=len("\"")
	//l.pos+=len("\"")
	l.next()
	for nextr:=l.peek();nextr!='"';{
		if nextr==scanner.EOF{
			return nil
		}
		if l.next()=='\\'{
			if l.peek()=='"' || l.peek()=='\\'{
				l.next()
			}
		}
		nextr=l.peek()
	}
	l.next()
	//fmt.Println(l.input[l.start:l.pos])
	l.emit(itemStr)
	return lexText
}

func lexCtr (l *lexer) stateFn{
	//l.start+=len("'")
	//l.pos+=len("'")
	l.next()
	//fmt.Println("here")
	if l.accept("\\"){
		if l.accept("rntbv'\\"){
			if l.accept("'"){
				//l.pos+=len("'")
				l.emit(itemCtr)
				return lexText
			}else{
				return nil
			}

		}else{
			return nil
		}
	}else{
		if !l.accept("'"){
			l.next()
			if l.accept("'"){
				//l.pos+=len("'")
				l.emit(itemCtr)
				return lexText
			}
		}
		return nil
	}

}

func lexMComment(l *lexer) stateFn{
	l.next()
	l.next()
	for{
		if l.peek()=='\n'{
			l.nlc++
		}else{
			if l.peek()==scanner.EOF{
				return nil
			}

		}
		if l.accept("@") {
			if l.peek()=='/'{
				l.next()
				break
			}
			continue
		}

		l.next()
	}
	// l.pos+=len("\n")
	//fmt.Println(l.input[l.start:l.pos])
	l.emit(itemMComment)
	l.lcount=l.nlc
	return lexText
}

func isLetter(s rune) bool {

	if (s < 'a' || s > 'z') && (s < 'A' || s > 'Z') {
		return false
	}

	return true
}

func lexNumber(l *lexer) stateFn{
	l.accept("+-")
	digits:="0123456789"
	digitshex:="0123456789abcdefABCDEF"
	if l.accept("0") {
		if l.accept("x") {
			l.acceptRun(digitshex)
			l.emit(itemNumber)
			return lexText
		}
	}

	l.acceptRun(digits)
	if l.accept("."){
		l.acceptRun(digits)
	}else if l.accept("eE"){
		if l.accept("+-"){
			if ('0' <= l.peek() && l.peek()<='9'){
				l.acceptRun(digits)
			}else{
				l.backup()
				l.backup()
			}
		}else{
			if ('0' <= l.peek() && l.peek()<='9'){
				l.acceptRun(digits)
			}else{
				l.backup()
			}
		}

	}
	//if isLetter(l.peek()){
	//	l.emit(itemNumber)
	//	return lexText
	//	//l.next()
	//	//// fmt.Println("bad number syntax")
	//	//// l.emit(itemError)
	//	//return l.errorf("bad number syntax: %q",l.input[l.start:l.pos])
	//}
	// fmt.Println(l.input[l.start:l.pos])
	l.emit(itemNumber)
	return lexText
}

func (l *lexer) run(){
	// fmt.Println("hi")
	state := lexText
	for state!=nil {
		state=state(l)
	}
	close(l.items)
}

func main(){
	//inp:="aa\n"
	//e:=string(scanner.EOF)
	//inp=inp+e
	var str string
	//str=""
	in := bufio.NewReader(os.Stdin)

	str=""
	strp, err := in.ReadString('\n')
	for err!=io.EOF{
		str=str+strp
		strp, err = in.ReadString('\n')
		//if strings.Contains(strp,"exit"){
		//	break
		//}
		////fmt.Print(str)
	}
	str=str+strp

	//str = "int b = 0xx34;\n0x4e3\n.23e2\n36.2e5\n3e+2\n-4e-23e3\n\"com 'h' fg\"\n/@ erjk @/ /@ lkfjlek krejng @/ elr /@elrk@@rkg@/ 0234\n@@ rkg /@lek@/ elrk @@ erj~~~\n23er\n234Ee2\n23e23e23e\n__e__e23__23e3\n-0x34\n34e34-e3-23-.24\n++34++34+-34--34\na_b__d\nc/@@ string @/ int;\n/@@/\n-20x2\n-20e-2\n-20e--2"
	//str="\n\n\n /@ multi \n a@/"

	l:=lex("test",str)
	l.run()
}

