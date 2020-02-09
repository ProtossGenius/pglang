package pgl_ana_lex

var SymbolList = map[string]bool{"+":true,"-":true,"*":true,"/":true,"%":true,"!":true,"^":true,"&":true,"~":true,"(":true,")":true,"=":true,"[":true,"]":true,"{":true,"}":true,"==":true,"+=":true,"-=":true,"*=":true,"/=":true,"!=":true,"|=":true,"&=":true,"^=":true,"%=":true,"~=":true,",":true,".":true,";":true,"$":true,"@":true,"#":true,"\\":true,":":true,}

var SymbolCharSet = map[rune]bool{'&':true,']':true,'.':true,'\\':true,'(':true,'[':true,'{':true,'}':true,'@':true,'#':true,'/':true,'%':true,'^':true,')':true,'|':true,',':true,':':true,'$':true,'+':true,'-':true,'*':true,'!':true,'~':true,'=':true,';':true,}

var SymbolCanContinue = map[string]bool{"=":true, "+":true, "-":true, "*":true, "/":true, "!":true, "&":true, "^":true, "%":true, "~":true, }

//some maybe define in another type, but not as symbol. like comment's "//" and "/*"
var SymbolUnuse = map[string]bool{"//":true, "/*":true}
