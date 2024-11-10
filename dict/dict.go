package dict

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"sort"
)

const (
	RootState = 0
	FailState = -1

	//MaskBegin  = 0xFF000000
	//MaskEnd    = 0x00FF0000
	//MaskLength = 0x00003FFF
)

func HandleDict(dictPath string) *ACAutomaton {
	dict := make([]string, 0)
	f, err := os.Open(dictPath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	r := bufio.NewReader(f)
	for {
		l, err := r.ReadBytes('\n')
		if err != nil {
			break
		}

		dict = append(dict, string(bytes.TrimSpace(l)))
	}

	return dictToAc(dict)
}

func dictToAc(dict []string) *ACAutomaton {
	rList := make([][]rune, len(dict))
	i := 0
	for _, d := range dict {
		rList[i] = []rune(d)
		i++
	}

	// 生产AC自动机⚙️
	ac := new(ACAutomaton)
	if err := ac.generateAutomaton(rList); err != nil {
		panic(err)
	}

	return ac
}

type ACAutomaton struct {
	DoubleArrayTrie
	Failure map[int]int
}

type DoubleArrayTrie struct {
	Base   []int
	Check  []int
	Output map[int][]rune
}

// generateAutomaton 根据 DAT 生成自动机
func (ac *ACAutomaton) generateAutomaton(rList [][]rune) (err error) {
	// 生成DAT 🌲
	t := new(DAT)
	err = t.BuildTire(rList)
	if err != nil {
		return err
	}

	ac.Base = t.dat.Base
	ac.Check = t.dat.Check
	ac.Output = t.dat.Output
	err = ac.Build(t.trie)

	return
}

func (ac *ACAutomaton) Build(trieNodeList *TrieNodeList) (err error) {
	// 初始化状态
	ac.Failure = make(map[int]int)

	// 使用队列进行 BFS
	queue := make([]*TrieNodeList, 0)
	queue = append(queue, trieNodeList.Children...) // 从根节点开始

	for _, c := range queue {
		ac.Failure[c.Base] = RootState
	}

	for len(queue) > 0 {
		node := queue[0]

		// 遍历 Trie 树:
		// 使用广度优先搜索 (BFS) 遍历 Trie 树的每个节点。
		for _, child := range node.Children {
			if child.Base < RootState {
				continue
			}
			// 找到父节点的失败节点，进行跳转
			failNode := ac.Failure[node.Base]

			var outBase = RootState
			if failNode >= RootState {
				// 设置当前失败节点，指向计算出来的base
				var key = int(child.Code) + failNode
				if key < len(ac.Check) {
					// 根据条件获取下一个节点，同时不能是当前节点的
					if failNode == ac.Check[key] {
						outBase = ac.Base[key]
					}
				}
			}
			ac.Failure[child.Base] = outBase
		}
		// 将子节点入队以便后续处理
		queue = append(queue, node.Children...)
		queue = queue[1:]
	}

	return
}

type TrieNodeList struct {
	Code                            rune
	SubCode                         []rune
	Depth, Left, Right, Index, Base int
	IsEnd                           bool
	Children                        []*TrieNodeList
	Word                            string
}

type DAT struct {
	trie                 *TrieNodeList
	dat                  *DoubleArrayTrie
	originalTrieNodeList *OriginalTrieNodeList
	lastPos              int
	used                 map[int]bool
}

type OriginalTrieNodeList struct {
	OriginalKeys []TrieNodeKeys
}

type TrieNodeKeys []rune

func (t *DAT) BuildTire(rList [][]rune) (err error) {
	if len(rList) == 0 {
		return fmt.Errorf("empty keywords")
	}

	o := new(OriginalTrieNodeList)
	for _, rv := range rList {
		var originalKey TrieNodeKeys = rv
		o.OriginalKeys = append(o.OriginalKeys, originalKey)
	}
	// 按行首字符升序排序
	sort.Slice(o.OriginalKeys, func(i, j int) bool {
		lk := max(len(o.OriginalKeys[i]), len(o.OriginalKeys[j]))
		for m := 0; m < lk; m++ {
			if len(o.OriginalKeys[i]) <= m {
				return true
			}
			if len(o.OriginalKeys[j]) <= m {
				return false
			}
			if string(o.OriginalKeys[i][m]) != string(o.OriginalKeys[j][m]) {
				return string(o.OriginalKeys[i][m]) < string(o.OriginalKeys[j][m])
			}
		}
		return true
	})

	t.trie = new(TrieNodeList)
	t.dat = new(DoubleArrayTrie)
	t.dat.Output = make(map[int][]rune)
	t.used = make(map[int]bool)
	t.originalTrieNodeList = o
	t.trie.Depth = 0
	t.trie.Left = 0
	t.trie.Right = len(rList)
	t.trie.Index = 0
	t.trie.Base = 1

	// 处理根节点
	_, err = t.fetch(t.trie) // Feature: 返回的结构体是否需要处理

	_, err = t.insert(t.trie.Children)

	return
}

func (t *DAT) fetch(trieNodeChild *TrieNodeList) (trieChildrenNodeList []*TrieNodeList, err error) {
	var prev rune = 0

	for k, curTrieNodeChildren := range t.originalTrieNodeList.OriginalKeys[trieNodeChild.Left:trieNodeChild.Right] {
		if len(curTrieNodeChildren) < trieNodeChild.Depth { // 树高度不及预期
			continue
		}

		var cur rune = 0
		if len(curTrieNodeChildren) != trieNodeChild.Depth {
			cur = curTrieNodeChildren[trieNodeChild.Depth]
		}

		if prev > cur {
			continue
		}

		if prev == 0 && cur == 0 { // 设置最终节点
			tmpTrieNodeChild := new(TrieNodeList)
			tmpTrieNodeChild.Word = string(cur)
			tmpTrieNodeChild.Code = cur
			tmpTrieNodeChild.Depth = trieNodeChild.Depth + 1
			tmpTrieNodeChild.Left = trieNodeChild.Left + k
			tmpTrieNodeChild.Right = trieNodeChild.Left + k + 1
			tmpTrieNodeChild.IsEnd = true
			tmpTrieNodeChild.SubCode = trieNodeChild.SubCode
			trieNodeChild.Children = append(trieNodeChild.Children, tmpTrieNodeChild)

			continue
		}

		if prev != cur {
			// 避免切片引用
			var subCode []rune
			subCode = append(subCode, trieNodeChild.SubCode...)
			if cur != 0 {
				subCode = append(subCode, cur)
			}

			// 添加结点
			tmpTrieNodeChild := new(TrieNodeList)
			tmpTrieNodeChild.Word = string(cur)
			tmpTrieNodeChild.Code = cur
			tmpTrieNodeChild.Depth = trieNodeChild.Depth + 1
			tmpTrieNodeChild.Left = trieNodeChild.Left + k
			tmpTrieNodeChild.SubCode = subCode
			if len(trieNodeChild.Children) > 0 { // 设置上一个child节点的right值
				trieNodeChild.Children[len(trieNodeChild.Children)-1].Right = trieNodeChild.Left + k
			}

			trieNodeChild.Children = append(trieNodeChild.Children, tmpTrieNodeChild)
		}

		prev = cur
	}
	// 设置最后一个Node的右节点
	if len(trieNodeChild.Children) > 0 {
		trieNodeChild.Children[len(trieNodeChild.Children)-1].Right = trieNodeChild.Right
	}

	return trieNodeChild.Children, nil
}

func (t *DAT) insert(trieNodeChildrenList []*TrieNodeList) (end int, err error) {
	length := len(trieNodeChildrenList)
	if length <= 0 {
		return -1, nil
	}

	// 因为这里是进行便利 每次的begin都会计算成0，所以pos需要计算出当前便利下可用的节点
	var begin int = RootState
	var pos int = max(int(trieNodeChildrenList[0].Code), t.lastPos+1)
	var codeIdx int = int(trieNodeChildrenList[0].Code)
	t.dat.expandList(pos + codeIdx)

	for {
	findAvailableBegin:
		pos++ // 跟节点计算check的key 从1开始 而不是从0开始，同理HandleFileLine侧state从1开始
		if pos >= len(t.dat.Check)-1 {
			// 扩展数组大小
			t.dat.expandList(pos)
		}
		if 0 != t.dat.Check[pos] {
			goto findAvailableBegin
		}

		// 设置开始的节点，进行遍历计算
		begin = pos - int(trieNodeChildrenList[0].Code)
		if ok := t.used[begin]; ok {
			goto findAvailableBegin
		}

		for i := 0; i < length; i++ {
			trieNodeChild := trieNodeChildrenList[i]
			if (begin + int(trieNodeChild.Code)) >= len(t.dat.Check)-1 {
				// 扩展数组大小
				t.dat.expandList(begin + int(trieNodeChild.Code))
			}
			// 所有的兄弟节点 都可用 才能确定begin值
			if 0 != t.dat.Check[begin+int(trieNodeChild.Code)] {
				i = 0 // 重新循环 获取可以给当前兄弟节点用的所有key
				goto findAvailableBegin
			}
		}
		break
	}

	t.lastPos = pos
	for i := 0; i < length; i++ {
		trieNodeChild := trieNodeChildrenList[i]
		// check值为上一个节点所属路径，算法计算check的key
		t.dat.Check[begin+int(trieNodeChild.Code)] = begin
	}
	t.used[begin] = true

	for i := 0; i < length; i++ {
		trieNodeChild := trieNodeChildrenList[i]
		newTrieNodeChildrenList, _ := t.fetch(trieNodeChild)

		if len(newTrieNodeChildrenList) > 0 {
			var nextBegin int
			nextBegin, err = t.insert(newTrieNodeChildrenList)
			if err != nil {
				continue
			}
			// begin+int(trieNodeChild.Code) 当前base的key 指向 下一个begin值
			// 这个时候begin尚未完全赋值 需要存储下
			t.dat.Base[begin+int(trieNodeChild.Code)] = nextBegin
			trieNodeChildrenList[i].Index = begin + int(trieNodeChild.Code)
			trieNodeChildrenList[i].Base = nextBegin
		} else {
			t.dat.Base[begin+int(trieNodeChild.Code)] = -trieNodeChild.Left + FailState
			trieNodeChildrenList[i].Index = begin + int(trieNodeChild.Code)
			trieNodeChildrenList[i].Base = -trieNodeChild.Left + FailState
			t.dat.Output[trieNodeChildrenList[i].Index] = trieNodeChildrenList[i].SubCode
		}

	}

	end = begin
	return
}

// 扩展 Base 和 Check 数组
func (dat *DoubleArrayTrie) expandList(newSize int) {
	for len(dat.Base) <= newSize {
		dat.Base = append(dat.Base, make([]int, newSize+1-len(dat.Base))...)
	}
	for len(dat.Check) <= newSize {
		dat.Check = append(dat.Check, make([]int, newSize+1-len(dat.Check))...)
	}
}
