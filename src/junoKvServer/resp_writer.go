package junoKvServer

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"log"
)

type respWriter struct {
	buff *bufio.Writer
}

func newResponseWriter(conn net.Conn, size int) *respWriter {
	w := new(respWriter)
	w.buff = bufio.NewWriterSize(conn, size)
	return w
}

func formatInt64ToSlice(v int64) []byte {
	return strconv.AppendInt(nil, int64(v), 10)
}

func (w *respWriter) writeError(err error) {
	if _, berr := w.buff.Write([]byte("-")); berr != nil {
		log.Fatal(berr)
	}

	if err != nil {
		if _, berr := w.buff.Write([]byte(err.Error())); berr != nil {
			log.Fatal()
		}
	}
	w.buff.Write(newLine)
}

func (w *respWriter) writeStatus(status string) {
	w.buff.WriteByte('+')
	w.buff.Write([]byte(status))
	w.buff.Write(newLine)
}

func (w *respWriter) writeInteger(n int64) {
	w.buff.WriteByte(':')
	w.buff.Write(formatInt64ToSlice(n))
	w.buff.Write(newLine)
}

func (w *respWriter) writeBulk(b []byte) {
	w.buff.WriteByte('$')
	if b == nil {
		w.buff.Write(nullBulk)
	} else {
		w.buff.Write([]byte(strconv.Itoa(len(b))))
		w.buff.Write(newLine)
		w.buff.Write(b)
	}

	w.buff.Write(newLine)
}

func (w *respWriter) writeArray(lst []interface{}) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(nullArray)
		w.buff.Write(newLine)
	} else {
		w.buff.Write([]byte(strconv.Itoa(len(lst))))
		w.buff.Write(newLine)

		for i := 0; i < len(lst); i++ {
			switch v := lst[i].(type) {
			case []interface{}:
				w.writeArray(v)
			case [][]byte:
				w.writeSliceArray(v)
			case []byte:
				w.writeBulk(v)
			case nil:
				w.writeBulk(nil)
			case int64:
				w.writeInteger(v)
			case string:
				w.writeStatus(v)
			case error:
				w.writeError(v)
			default:
				panic(fmt.Sprintf("invalid array type %T %v", lst[i], v))
			}
		}
	}
}

func (w *respWriter) writeSliceArray(lst [][]byte) {
	w.buff.WriteByte('*')
	if lst == nil {
		w.buff.Write(nullArray)
		w.buff.Write(newLine)
	} else {
		w.buff.Write([]byte(strconv.Itoa(len(lst))))
		w.buff.Write(newLine)

		for i := 0; i < len(lst); i++ {
			w.writeBulk(lst[i])
		}
	}
}

func (w *respWriter) writeMap(m map[string][]byte) {
	w.buff.WriteByte('*')
	if m == nil {
		w.buff.Write(nullArray)
		w.buff.Write(newLine)
	} else {
		w.buff.Write([]byte(strconv.Itoa(len(m) * 2)))
		w.buff.Write(newLine)

		for k, v := range m {
			w.writeBulk([]byte(k))
			w.writeBulk(v)
		}
	}
}

func (w *respWriter) writeBulkFrom(n int64, rb io.Reader) {
	w.buff.WriteByte('$')
	w.buff.Write([]byte(strconv.FormatInt(n, 10)))
	w.buff.Write(newLine)

	io.Copy(w.buff, rb)
	w.buff.Write(newLine)
}

func (w *respWriter) flush() {
	w.buff.Flush()
}
