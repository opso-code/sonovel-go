package util

import "testing"

func TestApplyInlineJS_Replace(t *testing.T) {
	got := ApplyInlineJS("r=r.replaceAll('作者：','');", "作者：张三")
	if got != "张三" {
		t.Fatalf("unexpected js replace result: %q", got)
	}
}

func TestApplyInlineJS_DecodeEmbeddedBase64Script(t *testing.T) {
	input := "<script>document.writeln(qsbs.bb('PHA+SGVsbG88L3A+'));</script>"
	js := "var qsbs={_keyStr:\"ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/=\",bb:function(a){var b=\"\",d,c,i,e,f,h,j,g=0;a=a.replace(/[^A-Za-z0-9+\\/=]/g,\"\");while(g<a.length){e=this._keyStr.indexOf(a.charAt(g++));f=this._keyStr.indexOf(a.charAt(g++));h=this._keyStr.indexOf(a.charAt(g++));j=this._keyStr.indexOf(a.charAt(g++));d=(e<<2)|(f>>4);c=((f&15)<<4)|(h>>2);i=((h&3)<<6)|j;b+=String.fromCharCode(d);if(h!=64)b+=String.fromCharCode(c);if(j!=64)b+=String.fromCharCode(i)}return b}};r=r.replace(/<script>\\s*document\\.writeln\\(qsbs\\.bb\\('([^']+)'\\)\\);\\s*<\\/script>/g,function(a,b){return qsbs.bb(b)});"
	got := ApplyInlineJS(js, input)
	if got != "<p>Hello</p>" {
		t.Fatalf("unexpected decoded content: %q", got)
	}
}
