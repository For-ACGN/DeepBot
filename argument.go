package deepbot

// textToArgN splits a text into individual argument strings,
// following the Windows conventions documented at
// http://daviddeley.com/autohotkey/parameters/parameters.htm#WINARGV
func textToArgN(text string, n int) []string {
	if n < 1 {
		panic("invalid n in textToArgN")
	}
	var args []string
	for len(text) > 0 {
		if text[0] == ' ' || text[0] == '\t' {
			text = text[1:]
			continue
		}
		if len(args) == n-1 {
			args = append(args, text)
			break
		}
		var arg []byte
		arg, text = readNextArg(text)
		args = append(args, string(arg))
	}
	return args
}

// readNextArg splits command line string cmd into next
// argument and command line remainder.
func readNextArg(cmd string) (arg []byte, rest string) {
	var b []byte
	var inQuote bool
	var nSlash int
	for ; len(cmd) > 0; cmd = cmd[1:] {
		c := cmd[0]
		switch c {
		case ' ', '\t':
			if !inQuote {
				return appendBSBytes(b, nSlash), cmd[1:]
			}
		case '"':
			b = appendBSBytes(b, nSlash/2)
			if nSlash%2 == 0 {
				// use "Prior to 2008" rule from
				// http://daviddeley.com/autohotkey/parameters/parameters.htm
				// section 5.2 to deal with double quotes
				if inQuote && len(cmd) > 1 && cmd[1] == '"' {
					b = append(b, c)
					cmd = cmd[1:]
				}
				inQuote = !inQuote
			} else {
				b = append(b, c)
			}
			nSlash = 0
			continue
		case '\\':
			nSlash++
			continue
		}
		b = appendBSBytes(b, nSlash)
		nSlash = 0
		b = append(b, c)
	}
	return appendBSBytes(b, nSlash), ""
}

// appendBSBytes appends n '\\' bytes to b and returns the resulting slice.
func appendBSBytes(b []byte, n int) []byte {
	for ; n > 0; n-- {
		b = append(b, '\\')
	}
	return b
}
