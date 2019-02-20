### 从Go标准库看字符串匹配算法

Go的标准库本身质量非常高，本文主要深入strings库，从源代码中探查字符串匹配常用算法的具体实现

我们先看一个简单的例子开始。

在目标字符串中检查是否有子串等于匹配文本,这是非常常见的操作. 最容易让人想到的算法就是从目标字符串第一位开始，逐个字符与待匹配文本比较.匹配成功则指针右移继续比较，要不然从目标文本的第二位开始和待匹配文本继续逐个比较。如果指针先到达待匹配文本末尾，则匹配成功，要不然匹配失败。该算法称之为朴素算法，非常容易理解,但效率也比较慢.具体实现如下：
```
#include<stdio.h>
#include<string.h>
void search(char *pat, char *txt)
{
    int M = strlen(pat);
    int N = strlen(txt);
    int i;
    // 只需要遍历目标文本 N-M 次， 因为从目标文本的 N-M 位开始的子串，长度永远小于 M， 所以不会匹配成功
    for (i = 0; i <= N - M; i++)  
    {
        int j;
        for (j = 0; j < M; j++)
        {
            if (txt[i+j] != pat[j])
                break;
        }
        if (j == M)
        {
           printf("Pattern found at index %d \n", i);
        }
    }
}

int main()
{
   char *txt = "AABAACAADAABAAABAA";
   char *pat = "AABA";
   search(pat, txt);
   return 0;
}
```
Go 标准库中的strings.Contains函数使用了Rabin-Karp算法, 主要思想如下:

假设匹配文本的长度为M,目标文本的长度为N
1. 计算匹配文本的hash值
2. 计算目标字符串中每个长度为M的子串的hash值（需要计算N-M+1次）
3. 比较hash值, 如果hash值不同，字符串必然不匹配，如果hash值相同，还需要使用朴素算法再次判断

步骤2中每次都要重新计算hash, Rabin-Karp算法的优点在于设计了一个特别的hash算法,使其在计算下一个子串的hash时可以利用之前的hash结果, 以达到加速计算的效果。将每一个字节看作数字, 选择一个比较大的质数作为base. 字节的值是包含在基数之内的

举例说明：
文本为”abracadabra”,base为101,那么 hash(“abr”) = 97 * 101的2次方 + 98 * 101的1次方 + 114 * 101的0次方= 999509
下一个子串 “bra”的hash值为 98 * 101的2次方 + 114 * 101的1次方 + 97 * 101的0次方. 我们可以利用之前”abr”的hash值, 写成:

//　　　　　　　base　　old hash　　　　　　new ‘a’　　　　old ‘a’ * base
hash(“bra”) = 1011 * hash(“abr”) + (97 × 101的0次方) - (97 × 101的3次方)

可以看出hash算法里要点是确立一个非常大的数字作为base,同时根据子串长度得到乘数因子(上述的 101的3次方，其实就是base的len(待匹配文本)次方).
```
src/strings/strings_amd64.go相关代码注释

// 选择非常大的一个质数16777619 作为 base 
const primeRK = 16777619   

// hashStr 返回子串的hash值和乘数因子
func hashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*primeRK + uint32(sep[i])  //计算hash值
	}
    // 计算(最高位 + 1)位的乘数因子, 使用位移, 没有使用 i--, 可以有效减少循环次数. i >>=1 相当于遍历二进制的每一位
	var pow, sq uint32 = 1, primeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// Index 返回sep在s里第一次匹配时的index, 无法匹配则返回-1.
func Index(s, sep string) int {
	n := len(sep)
    // 先分析一些常见情况, 起到进一步加速的效果
	switch {   
	case n == 0:
		return 0
	case n == 1:  //如果为一个字节,则调用IndixByte(汇编语言)
		return IndexByte(s, sep[0])
	case n <= shortStringLen:  //如果sep的长度小于31且大于1, 则使用汇编代码(也是一种优化). 
		return indexShortStr(s, sep)  
	case n == len(s):  
		if sep == s {
			return 0
		}
		return -1
	case n > len(s):
		return -1
	}
	// 使用Rabin-Karp算法匹配
    // 步骤1 初始计算待匹配的文本的hash值和乘数因子, 
	hashsep, pow := hashStr(sep)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*primeRK + uint32(s[i])  // 步骤2 计算长度跟sep一样的s子串的hash值
	}
	if h == hashsep && s[:n] == sep {
		return 0
	}
	for i := n; i < len(s); {
        // 利用先前的hash值, 计算新的hash值 
		h *= primeRK  // 乘以base
		h += uint32(s[i]) // 加上下一个字符的 hash 值
		h -= pow * uint32(s[i-n]) // 减去先前子串的第一个字符的hash值
		i++
        // 如果hash相等则继续使用朴素算法比较, 如果hash不一致,则直接用下一个匹配
		if h == hashsep && s[i-n:i] == sep {   
			return i - n
		}
	}
	return -1
}
strings库里还实现了BM算法， 在这之前，我们先来看另一个非常经典的KMP算法
```
假设检查bacbababaabcbab是否包含abababca, 此时发现第6位不一样

bacbababaabcbab   
    abababca
         |
       第六位

朴素算法：
bacbababaabcbab   
     abababca
     |
   移动一位后开始重新比较

KMP算法：
bacbababaabcbab   
      abababca
      |
直接移动两位后开始重新比较
如果按朴素算法则按上面所示需要搜索词移一位后重新从第一位开始匹配。仔细想想, 前5个字符ababa已经匹配成功,也就是我们已经知道双方的文本, 通过提前的计算，可以多移几位, 而不仅仅移一位. 这样可以加快搜索

KMP算法的主要原理如下:
s为目标文本, 长度为m
p为搜索词,长度为n
假设p[i]与s[x]匹配失败,那么p[i-1]与s[x-1]是匹配成功的, 则试图找到一个索引 j, 使得p[0:j] = p[i-j-1:i-1] (p[0:j] 包含p[j])
如果有则s[x]继续与p[j+1]进行比较, 相当于搜索词移动i-j-1位
无则s[x]与p[0]比较. (具体代码实现时无可以表示为-1, 这样+1 后正好为0) 相当于搜索词移动i位
```
void cal(char *p, int *next)
{
    int i;
    int k;
    /*第一次字符前面没有索引了, 算corner case, 直接赋值为-1*/
    next[0] = -1;
    /* 循环每一个索引, 并计算next值 */
    for (i = 1; p[i] != '\0'; i++) {
        /* 获取前一个索引的next值 */
        k = next[i - 1];
        /* 当p[i] != p[k + 1]时, 则令 k = next[k], 直到相等或者k == -1 退出*/
        while (p[i] != p[k + 1]) {
            if (k == -1) {
                k = -2;
                break;
            }
            k = next[k];
        }
        /*  1. p[i] == p[k + 1] 则 i对应的next值为 ++k
            2. 无索引时, k= -2, 则++k正好为-1
        */
        next[i] = ++k;
    }
}

int kmp(char *p, char*t)
{
    /*next为数组, 存储搜索词里每一个索引对应的next值, 使得 p[0:next[i]] == p[i-j-1:i-1]*/
    int next[strlen(p)];
    cal(p, next);
    int i, j;
    i = 0;
    j = 0;
    while (p[i] != '\0' && t[j] != '\0') {
        if (p[i] == t[j]) {
            /* 值相等, 则指针 i, j 都递增 */
            i++;
            j++;
        } else {
            if (i == 0) {
                j++;
                continue;
            }
            i = next[i - 1] + 1;
        }
    }
    if (p[i] == '\0') {
        return 0;
    } else {
        return 1;
    }
}
```
Go语言里在 strings/search.go 里使用了Boyer-Moore字符串搜索算法, 这个思想和KMP类似,都是根据Pattern自身计算出移动的步数. 有两个优化点:
1. BM算法是从后向前逐渐匹配.
2. kmp里的通过已匹配的文本增加移动步数的叫做好规则，那么BM里同时还增加了坏规则

假定Text为”HERE IS A SIMPLE EXAMPLE”，Pattern为”EXAMPLE”。
当T[i] != P[j], P[j]右边都匹配时时, 具体的移动规则如下:
坏字符规则: 此时T[i]定义为坏字符, 如果P[0..j-1]中包含T[i]这个字符, 则移动T使坏字符与相等的字符对齐, 如果不包含,则直接移动len(P)

HERE IS A SIMPLE EXAMPLE
             |
       EXAMPLE

此时P为坏字符, 因EXAMPLE包含P, 则T的i指针右移二位使之对齐，然后重新开始从P的末端继续匹配（下面打X处）.

HERE IS A SIMPLE EXAMPLE
             | X
         EXAMPLE

如下场景，T中的M与P中的E不匹配, 按Go的代码实现，是移动两位（取该字符到P末尾的最短距离），没完全按上面的规则实现
大家是不是发现没有跳跃前进，反而匹配又倒回到之前已完成的匹配过程。 Go代码这么做是为了实现简单。 
因为还有好规则可以保证最终的移动步数是正确的
ABCADADEEFXYZ
   | 
 AYEDADE
移动为
  ABCADADEEFXYZ
     | X   
 AYEDADE

好后缀规则: 当发生不匹配时,之前已经匹配成功的,称之为好字符. 如下I和A不匹配, 后面的MPLE就是好后缀. 首先检查P里是否好后缀只出现过一次: 比如此时的MPLE作为好后缀在整个字符串EXAMPLE中只出现过一次

不是, 则移动P使T中的好后缀与P中长度相等的字符串对齐
是, 则继续检查好后缀的所有后缀（比如PLE,PL,E)是否和同等长度的P前缀相等, 如果相等则移动P使之对齐, 不相等则移动 len(P).
这里相当于要求后缀必须出现在P的首部, 如果非首部, 因前缀的前一个字符必然不相等,则整个字符串肯定无法匹配
HERE IS A SIMPLE EXAMPLE
            ||||
         EXAMPLE
MPLE， PLE，LE没法和首部匹配，但后缀E和P前缀相等, 则移动T使其对齐，从打X出继续从后向前比较
HERE IS A SIMPLE EXAMPLE
               |     X
               EXAMPLE
            
具体的代码注释如下：
```
func makeStringFinder(pattern string) *stringFinder {
	f := &stringFinder{
		pattern:        pattern,
		goodSuffixSkip: make([]int, len(pattern)),
	}
	// last 是pattern最后一个字符的索引
	last := len(pattern) - 1

	// 创建坏字符表，记录不匹配时T的i指针移动步数
	// 第一阶段，初始化256个字符全部移动 len(pattern) 步
	for i := range f.badCharSkip {
		f.badCharSkip[i] = len(pattern)
	}

	// 第二阶段：从左到右遍历pattern，更新其索引与P末尾的距离，结果就是该字符到末尾的最小距离
	// 没有计算last byte的距离, 因为移动至少要一步。 没有0步。 
	for i := 0; i < last; i++ {
		f.badCharSkip[pattern[i]] = last - i
	}

	// 创建好后缀表
	// 第一阶段: 此时pattern[i+1:]都是已经匹配的，且好后缀只出现了一次
	// 计算T中的指针要移动的步数
	lastPrefix := last
	for i := last; i >= 0; i-- {
		if HasPrefix(pattern, pattern[i+1:]) {
			lastPrefix = i + 1
		}
		// 好后缀时T的指针移动分两步，首先移动到与 pattern的末尾对齐，即 last - i
		// lastPrefix 用来记录 pattern[i+1:]中所有后缀与同等长度的前缀相等时的最大索引
		// 然后移动 lastPrefix步
		f.goodSuffixSkip[i] = lastPrefix + last - i
	}
	// 第二阶段: 好后缀在pattern前面部分还出现过, 如下计算相应的移动步数
	// 会覆盖之前第一阶段的部分值。但好后缀出现过移动步数比没出现的小。所以最终值是正确的
	// 举例： "mississi" 中好后缀是issi, 在pattern[1]处出现过，所以移动步数为 last-i  +  lenSuffix
	for i := 0; i < last; i++ {
		lenSuffix := longestCommonSuffix(pattern, pattern[1:i+1])
		if pattern[i-lenSuffix] != pattern[last-lenSuffix] {
			// (last-i) is the shift, and lenSuffix is len(suffix).
			f.goodSuffixSkip[last-lenSuffix] = lenSuffix + last - i
		}
	}
	return f
}
// longestCommonSuffix 仅仅比较两个字符串的共同后缀的长度, 没有则为0
func longestCommonSuffix(a, b string) (i int) {
	for ; i < len(a) && i < len(b); i++ {
		if a[len(a)-1-i] != b[len(b)-1-i] {
			break
		}
	}
	return
}

// next 主要返回p在text里第一次匹配时的索引, 不匹配则返回-1
func (f *stringFinder) next(text string) int {
    // i 是T(即变量text)中要检查的字符索引, j为P中要检查的字符索引

    // 因从后向前比较, 所以i初始化为P的最后一位索引
	i := len(f.pattern) - 1
	for i < len(text) {
        // 每次比较时都从p的最后一位开始比较
		j := len(f.pattern) - 1
		for j >= 0 && text[i] == f.pattern[j] {
			i--
			j--
		}
        // j为负数,说明匹配成功, 则直接返回 i+ 1 
		if j < 0 {
			return i + 1
		}
        // j为非负, 表明text[i] != f.pattern[j], 则从坏字符表和好后缀表中获取分别获取i需要移动的步数, 取最大值并使移动到新位置
		i += max(f.badCharSkip[text[i]], f.goodSuffixSkip[j])
	}
	return -1
}
```

from : https://xixiliguo.github.io/post/strings-algorithm-golang/
