package ffmt

import (
	"strings"
)

// 这就是一个二叉树
type node struct {
	child *node   // 子节点 左
	next  *node   // 下节点 右
	value builder // 数据
	colon int     // 这是用冒号分割的数据 冒号的位置 用来对齐冒号后面的数据
}

// 缩进空的括号
func (n *node) lrPos() {
	b := n
	for x := b; x != nil && x.next != nil; x = x.next {
		if x.child != nil {
			continue
		}
		ss := x.next.value.String()
		if len(ss) == 2 && (ss[1] == ')' || ss[1] == ']' || ss[1] == '}') {
			x.mergeNext(1)
		}
	}
}

// 对齐数组类型的数据
func (n *node) tablePos() {
	ms := []int{}
	b := n
	max := 0
	for x := b; x != nil; x = x.next {
		if x.colon > 0 || x.child != nil {
			return
		}
		ll := strLen(x.value.String())
		ms = append(ms, ll)
		if ll > max {
			max = ll
		}
	}

	if max < 10 {
		n.merge(9, ms)
	} else if max < 18 {
		n.merge(5, ms)
	} else if max < 30 {
		n.merge(3, ms)
	}
}

// 合并到下一节点
func (n *node) merge(m int, ms []int) {
	l := len(ms)
	col := 0
	for i := 0; i != m; i++ {
		z := m - i
		if l > z && l%z == 0 {
			col = z
			break
		}
	}
	if col > 1 {
		n.mergeNextSize(col, ms)
	}
}

// 合并到下一节点指定长度
func (n *node) mergeNextSize(s int, ms []int) {
	lmax := make([]int, s)
	for j := 0; j != s; j++ {
		for i := 0; i*s < len(ms); i++ {
			b := i*s + j
			if ms[b] > lmax[j] {
				lmax[j] = ms[b]
			}
		}
	}
	for i := 1; i < len(lmax); i++ {
		lmax[i] += lmax[i-1]
	}
	for x := n; x != nil; x = x.next {
		for i := 0; i < s-1 && x.next != nil; i++ {
			x.mergeNext(lmax[i])
		}
	}
}

// 空格 写入缓冲
func (n *node) spac(i int) {
	for k := 0; k < i; k++ {
		n.value.WriteByte(Space)
	}
	return
}

// 合并下一个节点到当前节点
func (n *node) mergeNext(max int) {
	n.spac(max - strLen(n.value.String()))
	n.value.WriteString(n.next.value.String())
	putBuilder(n.next.value)
	n.next = n.next.next
}

// 对齐冒号后面的数据
func (n *node) colonPos() {
	b := n
	for b != nil {
		m := 0
		for x := b; x != nil; x = x.next {
			if x.colon <= 0 {
				continue
			}
			bl := strLen(x.value.String()[:x.colon])
			if bl > m {
				m = bl
			}
			if x.child != nil {
				break
			}
		}
		for x := b; x != nil; x = x.next {

			if x.colon > 0 {
				bl := strLen(x.value.String()[:x.colon])
				if m-bl > 0 {
					t := strings.Replace(x.value.String(), colSym, colSym+spac(m-bl), 1)
					x.value.Reset()
					x.value.WriteString(t)
				}
			}
			b = x.next
			if x.child != nil {
				break
			}
		}
	}
	return
}

func (n *node) put() {
	if n.value != nil {
		putBuilder(n.value)
		n.value = nil
	}
	if n.child != nil {
		n.child.put()
	}
	if n.next != nil {
		n.next.put()
	}
	return
}

func (n *node) String() string {
	buf := getBuilder()
	defer putBuilder(buf)
	n.strings(0, buf)
	s := buf.String()
	if len(s) >= 2 {
		return s[2:]
	}
	return ""
}

func (n *node) strings(d int, buf builder) {
	buf.WriteString(spacing(d))
	buf.WriteString(n.value.String())
	if n.child != nil {
		n.child.strings(d+1, buf)
	}
	if x := n.next; x != nil {
		x.strings(d, buf)
	}
	return
}

func (n *node) toChild() (e *node) {
	if n.child == nil {
		buf := getBuilder()
		n.child = &node{
			value: buf,
		}
	}
	return n.child
}

func (n *node) toNext() (e *node) {
	if n.next == nil {
		buf := getBuilder()
		n.next = &node{
			value: buf,
		}
	}
	return n.next
}

func getDepth(a string) int {
	for i := 0; i != len(a); i++ {
		switch a[i] {
		case Space:
		case ',':
			return i + 1
		default:
			return i
		}
	}
	return 0
}

func stringToNode(a string) *node {
	ss := strings.Split(a, "\n")
	depth := 0
	o := &node{}
	x := o
	buf := getBuilder()
	x.value = buf
	st := []*node{}
	for i := 0; i != len(ss); i++ {
		b := ss[i]
		d := getDepth(b)
		switch {
		case d == depth:
			x = x.toNext()
		case d > depth:
			st = append(st, x)
			x = x.toChild()
		case d < depth:
			if len(st) == 0 {
				x = x.toNext()
			} else {
				x = st[len(st)-1]
				if x != nil {
					st = st[:len(st)-1]
					x.child.colonPos() // 冒号后对其
					x.child.tablePos() // 数组对其
					x.child.lrPos()    // 空括号合并
					x = x.toNext()
				}
			}
		}

		depth = d
		if d > 0 {
			d--
		}
		x.value.WriteString(b[d:])
		x.colon = strings.Index(x.value.String(), colSym)
	}
	return o
}

// Align returns align structured strings
func Align(a string) string {
	s := stringToNode(a)
	defer s.put()
	return s.String()
}
