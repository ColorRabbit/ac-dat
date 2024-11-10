### AC DAT 说明
#### AC自动机简介
AC自动机（Aho-Corasick automaton）是一种高效的多模式匹配算法，常用于文本搜索、敏感词过滤等场景。它基于Trie树，通过添加fail指针，实现了在文本中快速查找多个模式串的功能。

#### AC自动机的核心思想:

- Trie树: 将所有模式串插入到一棵Trie树中，每个节点代表一个字符。
- fail指针: 每个节点的fail指针指向另一个节点，表示当前节点失配时应该跳转到的节点。fail指针的构建使得AC自动机可以在一次遍历文本的过程中匹配所有模式串。


## 参考文献：
- https://www.hankcs.com/program/algorithm/implementation-and-analysis-of-aho-corasick-algorithm-in-java.html#google_vignette
- https://www.hankcs.com/program/java/triedoublearraytriejava.html#google_vignette
- https://www.hankcs.com/program/algorithm/aho-corasick-double-array-trie.html
- https://www.doc88.com/p-0601666272874.html

## 参考代码：
- https://github.com/hankcs/AhoCorasickDoubleArrayTrie
- https://github.com/hankcs/aho-corasick
- https://github.com/Vonng/ac

## 项目初衷：
- 个人学习
- 参考项目中采用了位运算提高运算时间，[位运算及相关知识参考](https://oi-wiki.org/math/bit/)
- 项目中关于rune和byte相关知识点 https://zhuanlan.zhihu.com/p/248173199

后续将持续优化，并计算其消耗的内存及时间复杂度